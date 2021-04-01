// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/tls"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
	"github.com/vmware/hamlet/pkg/v1alpha2/server"
)

type federatedServiceObserver struct {
	access.FederatedServiceObserverV1Alpha2
}

func (o *federatedServiceObserver) OnCreate(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was created")
	return nil
}

func (o *federatedServiceObserver) OnUpdate(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was updated")
	return nil
}

func (o *federatedServiceObserver) OnDelete(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was deleted")
	return nil
}

// notifyResourceChanges notifies consumers about the changes in resources.
func notifyResourceChanges(srv server.Server) {
	// Create a new service.
	svc := &types2.FederatedService{
		Name: "svc",
		Fqdn: "svc.foo.com",
	}
	if err := srv.Upsert(svc.Fqdn, svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while creating service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully created a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Update an existing service.
	svc.Name = "svc_blue"
	if err := srv.Upsert(svc.Fqdn, svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while updating service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully updated a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Delete an existing service.
	if err := srv.Delete(svc.Fqdn); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while deleting service")
		return
	}
	log.WithField("svc", svc).Infoln("Successfully deleted a service")

	// Wait for some time.
	time.Sleep(1 * time.Second)
}

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
