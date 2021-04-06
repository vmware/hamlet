// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/tls"
	"github.com/vmware/hamlet/pkg/v1alpha2/client"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
)

// federatedServiceObserver observes for updates related to federated services.
type federatedServiceObserver struct {
	access.FederatedServiceObserverV1Alpha2
}

func (o *federatedServiceObserver) OnCreate(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infof("client:RemoteResources:Federated service %s was created from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}

func (o *federatedServiceObserver) OnUpdate(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infof("client:RemoteResources:Federated service %s was updated from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}

func (o *federatedServiceObserver) OnDelete(providerId string, fs *types2.FederatedService) error {
	log.WithField("fs", fs).Infof("client:RemoteResources:Federated service %s was deleted from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}

// notifyResourceChanges notifies consumers about the changes in resources.
func notifyResourceChanges(cl client.Client) {
	id := shortuuid.New()
	for {
		// Create a new service.
		svc := &types2.FederatedService{
			Name: "svc",
			Fqdn: "svc." + id + ".bar.com",
		}
		if err := cl.Upsert(svc.Fqdn, svc); err != nil {
			log.WithField("svc", svc).Errorln("Error occurred while creating service")
			return
		}
		log.WithField("svc", svc).Infof("client:LocalResources: Created a service %s", svc.Fqdn)

		// Wait for some time.
		time.Sleep(1 * time.Second)

		// Update an existing service.
		svc.Name = "svc_blue"
		if err := cl.Upsert(svc.Fqdn, svc); err != nil {
			log.WithField("svc", svc).Errorln("Error occurred while updating service")
			return
		}
		log.WithField("svc", svc).Infof("client:LocalResources: Updated a service %s", svc.Fqdn)

		// Wait for some time.
		time.Sleep(1 * time.Second)

		// Delete an existing service.
		if err := cl.Delete(svc.Fqdn); err != nil {
			log.WithField("svc", svc).Errorln("Error occurred while deleting service")
			return
		}
		log.WithField("svc", svc).Infof("client:LocalResources: Deleted a service %s", svc.Fqdn)

		// Wait for some time.
		time.Sleep(1 * time.Second)
	}
}

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

	go func() {
		// Run the background resource change notifier.
		// stagger multiple notifiers
		go func() {
			// Notify the consumers about changes to resources.
			notifyResourceChanges(cl)
		}()
		// Wait for some time.
		time.Sleep(1 * time.Second)
		go func() {
			notifyResourceChanges(cl)
		}()
		// Wait for some time.
		time.Sleep(1 * time.Second)
		go func() {
			notifyResourceChanges(cl)
		}()
	}()
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
