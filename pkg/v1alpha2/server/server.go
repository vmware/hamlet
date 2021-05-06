// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/server/state"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/publisher"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Server represents the API for managing the resource discovery server.
type Server interface {
	// Start starts the server.
	Start() error

	// Stop stops the server.
	Stop() error

	WatchRemoteResources(id string, observer access.FederatedServiceObserverV1Alpha2) error
	UnwatchRemoteResources(id string) error
	// create/update a resource in registry, Create notifies to the attached publishers.
	Upsert(resourceId string, dt *types2.FederatedService) error
	// delete a resource from register, Delete notifies the deletion of a resource.
	Delete(resourceId string) error
}

// server is a concrete implementation of the Server API.
type server struct {
	rd.DiscoveryServiceServer
	access.RemoteResourceAccessV1Alpha2
	access.LocalResourceAccessV1Alpha2
	// port represents the port on which the federated resource discovery
	// protocol is served.
	port uint32

	// tlsConfig is a TLS configuration that support mTLS (mutual TLS) with
	// the federated service mesh publisher.
	tlsConfig *tls.Config

	// listener represents the TCP connection listener for the server.
	listener net.Listener

	// grpcServer represents the gRPC server that services the federated
	// service mesh publisher's requests.
	grpcServer *grpc.Server

	// stateProvider provides the mechanism to query the federated service
	// mesh owner implementation for the current state of a particular type
	// of resources.
	stateProvider state.StateProvider

	publisherRegistry publisher.Registry
	consumerRegistry  consumer.Registry
	// dsServer is an implementation of the server APIs for the federated
	// resource discovery protocol.
	dsServer rd.DiscoveryServiceServer

	// mutex synchronizes the access to streams.
	streamSendMutex *sync.Mutex

	// connection context/token
	connectionContext string
}

// NewServer returns a new instance of the server given a port to run on, the
// TLS configuration, and a state consumer.
func NewServer(port uint32, tlsConfig *tls.Config, connectionContext string) (Server, error) {
	// Create the listener.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(log.Fields{
			"port": port,
			"err":  err,
		}).Errorln("Error occurred while creating the listener")
		return nil, err
	}
	// Identify the server option.
	serverOptions := []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             40 * time.Second,
		PermitWithoutStream: true,
	})}
	if tlsConfig != nil {
		serverOptions = append(serverOptions, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	// Return a new instance of the server.
	remoteResource := resources.NewRemoteResources()
	consumerRegistry := consumer.NewRegistry(remoteResource)

	publisherRegistry := publisher.NewRegistry()
	localResources := resources.NewLocalResources(publisherRegistry)
	s := &server{
		port:              port,
		tlsConfig:         tlsConfig,
		listener:          listener,
		grpcServer:        grpc.NewServer(serverOptions...),
		publisherRegistry: publisherRegistry,
		consumerRegistry:  consumerRegistry,
		streamSendMutex:   &sync.Mutex{},
		connectionContext: connectionContext,
	}
	s.LocalResources = localResources
	s.RemoteResources = remoteResource
	rd.RegisterDiscoveryServiceServer(s.grpcServer, s)
	return s, nil
}

func (s *server) Start() error {
	log.WithField("address", s.listener.Addr()).Infoln("Starting to listen to requests")
	serv := s.grpcServer.Serve(s.listener)

	return serv
}

func (s *server) Stop() error {
	log.Infoln("Stopping the gRPC server")
	s.grpcServer.Stop()
	log.Infoln("Stopped the gRPC server")
	return nil
}
