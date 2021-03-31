// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/provider"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Client is an abstraction over the federated resource discovery protocol to be
// implemented by the client.
type Client interface {
	Start(ctx context.Context, resourceUrl, connectionContext string) error

	WatchRemoteResources(observer FederatedServiceObserver) error
	// create/update a resource in registry, Create notifies to the attached consumers.
	Upsert(resourceId string, dt *types2.FederatedService) error
	// delete a resource from register, Delete notifies the deletion of a resource.
	Delete(resourceId string) error
}

// client implements the federated resource discovery interface.
type client struct {
	// serverAddr is the address of the federated service mesh owner.
	serverAddr string

	// tlsConfig is a TLS configuration that support mTLS (mutual TLS) with
	// the federated service mesh owner.
	tlsConfig *tls.Config

	// conn is the client connection with the federated service mesh owner.
	conn *grpc.ClientConn

	//dial option s
	dialOptions []grpc.DialOption

	localResources  resources.LocalResources
	remoteResources resources.RemoteResources
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
		localResources:  resources.NewLocalResources(),
		remoteResources: resources.NewRemoteResources(),
		streamSendMutex: &sync.Mutex{},
	}

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

func (c *client) Start(ctx context.Context, resourceUrl, connectionContext string) error {
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

	// Create the initial stream request.
	req := &rd.StreamRequest{ResourceUrl: resourceUrl, Context: connectionContext}
	// Send the stream request.
	err = c.sendStreamData(stream, &rd.BidirectionalStream{Request: req})
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while sending stream request")
		return err
	}

	// Publisher Setup
	// setup the registery
	// localResources := resources.NewLocalResources()
	consumerRegistry := consumer.NewRegistry()
	co, err := consumerRegistry.Register("Client Consumer")
	co.InitStream(resourceUrl, c.localResources)
	// now localResource can have api call's to ti
	// using upsert/delete and it will be kept in sync with remote connections.
	// on the stream side
	// co has the WatchStream channel to receive data to put on the stream
	// co will accept ack via ProcessResponse api

	// remoteResources := resources.NewRemoteResources(consumerRegistry)
	providerRegistry := provider.NewRegistry(c.remoteResources)
	// connect the client
	pr, err := providerRegistry.Register("client-provider-1")
	// now provider can accept data and
	// var observer resources.ResourceObserver
	// remoteResources.WatchRemoteResources("remote-watcher-1", observer)

	// stream data is accepted by pr.AcceptStreamData
	// ack/nack are generatedy by pr.WatchStream
	// list of remote service are accessed via obeserver
	// or by iterating on remoteResoruces DS.

	// application needs to see
	// 1. create client or server
	// 2. create stream and attach it to providers/consumers.

	// 3. create Registry for provider/consumer
	// 4. attach provider consumer to a stream created on a client.
	// 5. access registry via observer ... or direct access via add/remove
	coRespChan, err := co.WatchStream("")
	recvDataChan := make(chan *rd.BidirectionalStream)
	func() {
		r, err := stream.Recv()
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while consuming stream response")
			close(recvDataChan)
			return
		}
		recvDataChan <- r
	}()
	// Loop until the stream has ended.
	for {
		done := false
		select {
		case recvData, recvDataOk := <-recvDataChan:
			if !recvDataOk {
				done = true
			}
			resp := recvData.GetResponse()
			req := recvData.GetRequest()
			if resp != nil {
				respAckNack, err := pr.AcceptStreamData(resp)
				if err != nil {
					log.WithField("err", err).Errorln("Error occurred while consuming response in producer")
				}
				if respAckNack != nil {
					r := &rd.BidirectionalStream{}
					r.Request = respAckNack
					c.sendStreamData(stream, r)
				}
			}
			if req != nil {
				co.ProcessAckNack(req)
			}

		case sendData, sendDataOk := <-coRespChan:
			if !sendDataOk {
				done = true
			}
			fmt.Printf("got data \n")
			r := &rd.BidirectionalStream{}
			r.Response = sendData.Object
			c.sendStreamData(stream, r)
		}
		if done {
			break
		}
	}
}
