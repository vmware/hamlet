// Copyright 2019 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"github.com/vmware/hamlet/pkg/server/consumer"
	"github.com/vmware/hamlet/pkg/server/resources"
	"github.com/vmware/hamlet/pkg/server/state"
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

	// Resources returns an instance of the Resources API to publish events
	// to federated service mesh consumers.
	Resources() resources.Resources
}

// server is a concrete implementation of the Server API.
type server struct {
	// port represents the port on which the federated resource discovery
	// protocol is served.
	port uint32

	// tlsConfig is a TLS configuration that support mTLS (mutual TLS) with
	// the federated service mesh consumer.
	tlsConfig *tls.Config

	// listener represents the TCP connection listener for the server.
	listener net.Listener

	// grpcServer represents the gRPC server that services the federated
	// service mesh consumer's requests.
	grpcServer *grpc.Server

	// stateProvider provides the mechanism to query the federated service
	// mesh owner implementation for the current state of a particular type
	// of resources.
	stateProvider state.StateProvider

	// consumerRegistry holds an active set of registered federated service
	// mesh consumers.
	consumerRegistry consumer.Registry

	// dsServer is an implementation of the server APIs for the federated
	// resource discovery protocol.
	dsServer rd.DiscoveryServiceServer

	// resources represents the API that allows the federated service mesh
	// owner implementation to notify federated service mesh consumers about
	// changes to resources.
	resources resources.Resources
}

// NewServer returns a new instance of the server given a port to run on, the
// TLS configuration, and a state provider.
func NewServer(port uint32, tlsConfig *tls.Config, stateProvider state.StateProvider) (Server, error) {
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
	consumerRegistry := consumer.NewRegistry(stateProvider)
	s := &server{
		port:             port,
		tlsConfig:        tlsConfig,
		listener:         listener,
		grpcServer:       grpc.NewServer(serverOptions...),
		stateProvider:    stateProvider,
		consumerRegistry: consumerRegistry,
		dsServer:         newDiscoveryServiceServer(consumerRegistry),
		resources:        resources.NewResources(consumerRegistry),
	}
	rd.RegisterDiscoveryServiceServer(s.grpcServer, s.dsServer)
	return s, nil
}

func (s *server) Start() error {
	log.WithField("address", s.listener.Addr()).Infoln("Starting to listen to requests")
	return s.grpcServer.Serve(s.listener)
}

func (s *server) Stop() error {
	log.Infoln("Stopping the gRPC server")
	s.grpcServer.Stop()
	log.Infoln("Stopped the gRPC server")
	return nil
}

func (s *server) Resources() resources.Resources {
	return s.resources
}
