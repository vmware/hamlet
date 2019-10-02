// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	types "github.com/vmware/hamlet/api/types/v1alpha1"
	"github.com/vmware/hamlet/pkg/client"
	"github.com/vmware/hamlet/pkg/tls"
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
func Start(rootCACert string, peerCert string, peerKey string, serverAddr string, insecureSkipVerify bool) {
	// Prepare the client instance.
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

	// Watch for federated service notifications.
	err = cl.WatchFederatedServices(context.Background(), &federatedServiceObserver{})
	if err != nil && err != io.EOF {
		log.WithField("err", err).Fatalln("Error occurred while watching federated services")
	}
}
