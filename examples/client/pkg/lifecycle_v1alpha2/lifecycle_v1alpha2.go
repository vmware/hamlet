// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	types "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/client"
	"github.com/vmware/hamlet/pkg/tls"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/provider"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

// federatedServiceObserver observes for updates related to federated services.
type federatedServiceObserver struct {
	client.FederatedServiceObserver
}

func (o *federatedServiceObserver) OnCreate(fs *types.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was created")
	return nil
}

func (o *federatedServiceObserver) OnUpdate(fs *types.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was updated")
	return nil
}

func (o *federatedServiceObserver) OnDelete(fs *types.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was deleted")
	return nil
}

// Start starts the client lifecycle.
func Start(rootCACert string, peerCert string, peerKey string, serverAddr string, insecureSkipVerify bool, context string) {
	// Prepare the client instance. Alternative functions for tls.Config exist in the ./pkg/tls/tls.go
	tlsConfig := tls.PrepareClientConfig(rootCACert, peerCert, peerKey, insecureSkipVerify)
	cl, err := client.NewClient(serverAddr, tlsConfig)
	if err != nil {
		log.WithFields(log.Fields{
			"address": serverAddr,
			"err":     err,
		}).Errorln("Error connecting to server")
	}

	// Setup the shutdown goroutine.
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChannel
		os.Exit(0)
	}()
	// setup the registery
	localResources := resources.NewLocalResources()
	consumerRegistry := consumer.NewRegistry()
	co, err33 := consumerRegistry.Register("Client Consumer")
	co.InitStream("", localResources)
	// now localResource can have api call's to ti
	// using upsert/delete and it will be kept in sync with remote connections.
	// on the stream side
	// co has the WatchStream channel to receive data to put on the stream
	// co will accept ack via ProcessResponse api

	remoteResources := resources.NewRemoteResources(consumerRegistry)
	providerRegistry := provider.NewRegistry(remoteResources)
	// connect the client
	pr, err23 := providerRegistry.Register("client-provider-1")
	// now provider can accept data and
	var observer resources.ResourceObserver
	remoteResources.WatchRemoteResources("remote-watcher-1", observer)
	// stream data is accepted by pr.AcceptStreamData
	// ack/nack are generatedy by pr.WatchStream

	// Watch for federated service notifications.
	err = cl.WatchFederatedServices(context.Background(), &federatedServiceObserver{})
	if err != nil && err != io.EOF {
		log.WithField("err", err).Fatalln("Error occurred while watching federated services")
	}
}
