// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/hamlet/pkg/tls"
	"github.com/vmware/hamlet/pkg/v1alpha2/server"
)

// Start starts the server lifecycle.
func Start(rootCACerts []string, peerCert string, peerKey string, port uint32, connectionContext string) {
	// Initialize the server. Alternative functions for tls.Config exist in the ./pkg/tls/tls.go
	tlsConfig := tls.PrepareServerConfig(rootCACerts, peerCert, peerKey)
	s, err := server.NewServer(port, tlsConfig, connectionContext)
	if err != nil {
		log.WithField("err", err).Fatalln("Error occurred while creating the server instance")
	}

	// Setup the shutdown goroutine.
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChannel
		if err := s.Stop(); err != nil {
			log.WithField("err", err).Errorln("Error occurred while starting the server")
		}
		os.Exit(0)
	}()

	// Run the background resource change notifier.
	createNotificationTask(s)
	// Watch for federated service notifications.
	err = s.WatchRemoteResources("w1", &federatedServiceObserver{})
	if err != nil && err != io.EOF {
		log.WithField("err", err).Fatalln("Error occurred while watching federated services")
	}

	// Start the server.
	if err := s.Start(); err != nil {
		log.WithField("err", err).Errorln("Error occurred while starting the server")
	}
}
