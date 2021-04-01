// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/provider"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
	"github.com/vmware/hamlet/pkg/v1alpha2/stream_handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Client is an abstraction over the federated resource discovery protocol to be
// implemented by the client.
type Client interface {
	Start(ctx context.Context, reosourceId, connectionContext string) error

	WatchRemoteResources(id string, observer access.FederatedServiceObserverV1Alpha2) error
	UnwatchRemoteResources(id string) error
	// create/update a resource in registry, Create notifies to the attached consumers.
	Upsert(resourceId string, dt *types2.FederatedService) error
	// delete a resource from register, Delete notifies the deletion of a resource.
	Delete(resourceId string) error
}

// client implements the federated resource discovery interface.
type client struct {
	access.RemoteResourceAccessV1Alpha2
	access.LocalResourceAccessV1Alpha2
	// serverAddr is the address of the federated service mesh owner.
	serverAddr string

	// tlsConfig is a TLS configuration that support mTLS (mutual TLS) with
	// the federated service mesh owner.
	tlsConfig *tls.Config

	// conn is the client connection with the federated service mesh owner.
	conn *grpc.ClientConn

	//dial option s
	dialOptions []grpc.DialOption

	// dsClient is the gRPC client for the federated resource discovery
	// protocol.
	dsClient rd.DiscoveryServiceClient

	// mutex synchronizes the access to streams.
	streamSendMutex *sync.Mutex
}

// NewClient creates a new client instance.
func NewClient(serverAddr string, tlsConfig *tls.Config) (Client, error) {
	client := &client{
		serverAddr:      serverAddr,
		tlsConfig:       tlsConfig,
		streamSendMutex: &sync.Mutex{}}
	client.LocalResources = resources.NewLocalResources()
	client.RemoteResources = resources.NewRemoteResources()

	// Prepare the dial options.
	client.dialOptions = []grpc.DialOption{grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	})}
	if client.tlsConfig != nil {
		creds := credentials.NewTLS(client.tlsConfig)
		client.dialOptions = append(client.dialOptions, grpc.WithTransportCredentials(creds))
	} else {
		client.dialOptions = append(client.dialOptions, grpc.WithInsecure())
	}
	return client, nil
}

func (c *client) sendStreamData(stream rd.DiscoveryService_EstablishStreamClient, data *rd.BidirectionalStream) error {
	c.streamSendMutex.Lock()
	defer c.streamSendMutex.Unlock()
	err := stream.Send(data)

	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while send data to stream")
	}
	return err
}

func (c *client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// start the processing on client. Blocking call.
func (c *client) Start(ctx context.Context, resourceUrl, connectionContext string) error {
	// resourceUrl := "type.googleapis.com/federation.types.v1alpha2.FederatedService"
	// Connect with the server.
	var err error
	c.conn, err = grpc.Dial(c.serverAddr, c.dialOptions...)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while connecting to the server")
		return err
	}

	// Create the federated resource discovery client.
	c.dsClient = rd.NewDiscoveryServiceClient(c.conn)

	// Create a stream instance.
	stream, err := c.dsClient.EstablishStream(ctx)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating stream instance")
		return err
	}

	// only one stream
	streamId := "client-stream-1"
	// setup the registry
	consumerRegistry := consumer.NewRegistry()
	providerRegistry := provider.NewRegistry(c.RemoteResources)

	err = stream_handler.Handler(streamId, resourceUrl, connectionContext,
		stream, c.LocalResources, consumerRegistry, providerRegistry, true)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while working with stream")
		return err
	}
	return nil
}
