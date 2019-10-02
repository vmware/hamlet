// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	types "github.com/vmware/hamlet/api/types/v1alpha1"
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

// notifyResourceChanges notifies consumers about the changes in resources.
func notifyResourceChanges(s server.Server) {
	// Create a new service.
	svc := &types.FederatedService{
		Name: "svc",
		Id:   "svc.foo.com",
	}
	if err := s.Resources().Create(svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while creating service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully created a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Update an existing service.
	svc.Id = "svc.acme.com"
	if err := s.Resources().Update(svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while updating service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully updated a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Delete an existing service.
	if err := s.Resources().Delete(svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while deleting service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully deleted a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)
}

// Start starts the server lifecycle.
func Start(rootCACerts []string, peerCert string, peerKey string, port uint32) {
	// Initialize the server.
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
