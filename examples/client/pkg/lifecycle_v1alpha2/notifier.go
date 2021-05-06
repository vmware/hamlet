// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"time"

	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/client"
)

// notifyResourceChanges notifies consumers about the changes in resources.
func notifyResourceChanges(cl client.Client) {
	id := shortuuid.New()
	for {
		var ep []*types2.FederatedService_Endpoint
		ep = append(ep, &types2.FederatedService_Endpoint{
			Address: "127.1.2.3",
			Port:    443,
		})
		var ins []*types2.FederatedService_Instance
		ins = append(ins, &types2.FederatedService_Instance{
			Id:       "inst-pool-1",
			Protocol: "HTTP",
			Metadata: map[string]string{"port": "1234"},
		})
		// Create a new service.
		svc := &types2.FederatedService{
			Name:      "svc",
			Fqdn:      "svc." + id + ".bar.com",
			Endpoints: ep,
			Instances: ins,
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

// create notification routines that will periodically
// update the service manifest
func createNotificationTask(cl client.Client) {
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
}
