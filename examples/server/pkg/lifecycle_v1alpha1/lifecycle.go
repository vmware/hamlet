// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha1

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/hamlet/pkg/server"
	"github.com/vmware/hamlet/pkg/server/state"
	"github.com/vmware/hamlet/pkg/tls"
)

// emptyProvider is a sample state provider implementation that always returns a
// default empty set of resources.
type emptyProvider struct {
	state.StateProvider
}

func (p *emptyProvider) GetState(string) ([]proto.Message, error) {
	return []proto.Message{}, nil
}

// Start starts the server lifecycle.
func Start(rootCACerts []string, peerCert string, peerKey string, port uint32) {
	// Initialize the server. Alternative functions for tls.Config exist in the ./pkg/tls/tls.go
	tlsConfig := tls.PrepareServerConfig(rootCACerts, peerCert, peerKey)
	s, err := server.NewServer(port, tlsConfig, &emptyProvider{})
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
	go func() {
		for {
			// Notify the consumers about changes to resources.
			notifyResourceChanges(s)
		}
	}()

	// Start the server.
	if err := s.Start(); err != nil {
		log.WithField("err", err).Errorln("Error occurred while starting the server")
	}
}
