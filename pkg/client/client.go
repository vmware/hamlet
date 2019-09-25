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

package client

import (
	"context"
	"crypto/tls"
	"time"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Client is an abstraction over the federated resource discovery protocol to be
// implemented by the client.
type Client interface {
	// WatchResources watches for the specified resource on the federated
	// service mesh owner and notifies via the supplied observer.
	WatchResources(ctx context.Context, resourceUrl string, observer ResourceObserver) error

	// WatchFederatedServices watches for notifications related to federated
	// services on the federated service mesh owner.
	WatchFederatedServices(ctx context.Context, observer FederatedServiceObserver) error
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

	// dsClient is the gRPC client for the federated resource discovery
	// protocol.
	dsClient rd.DiscoveryServiceClient
}

// NewClient creates a new client instance.
func NewClient(serverAddr string, tlsConfig *tls.Config) (Client, error) {
	client := &client{serverAddr: serverAddr, tlsConfig: tlsConfig}

	// Prepare the dial options.
	dialOptions := []grpc.DialOption{grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             20 * time.Second,
		PermitWithoutStream: true,
	})}
	if client.tlsConfig != nil {
		creds := credentials.NewTLS(client.tlsConfig)
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(creds))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	// Connect with the server.
	var err error
	client.conn, err = grpc.Dial(client.serverAddr, dialOptions...)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while connecting to the server")
		return nil, err
	}

	// Create the federated resource discovery client.
	client.dsClient = rd.NewDiscoveryServiceClient(client.conn)
	return client, nil
}
