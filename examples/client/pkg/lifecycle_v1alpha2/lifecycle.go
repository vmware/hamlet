// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/hamlet/pkg/tls"
	"github.com/vmware/hamlet/pkg/v1alpha2/client"
)

// Start starts the client lifecycle.
func Start(rootCACert string, peerCert string, peerKey string, serverAddr string, insecureSkipVerify bool, connectionContext string) {
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

	ctx := context.Background()
	// create task to generate the service change notifications
	createNotificationTask(cl)
	// Watch for federated service notifications.
	err = cl.WatchRemoteResources("w1", &federatedServiceObserver{})
	if err != nil && err != io.EOF {
		log.WithField("err", err).Fatalln("Error occurred while watching federated services")
	}

	err = cl.Start(ctx, "type.googleapis.com/federation.types.v1alpha2.FederatedService", connectionContext)
	if err != nil {
		log.WithField("err", err).Fatalln("Error occurred while starting client")
	}

}
