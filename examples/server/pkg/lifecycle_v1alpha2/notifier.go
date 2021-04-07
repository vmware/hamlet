// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	"time"

	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/server"
)

// notifyResourceChanges notifies consumers about the changes in resources.
func notifyResourceChanges(srv server.Server) {
	// Create a new service.
	svc := &types2.FederatedService{
		Name: "svc",
		Fqdn: "svc.srv.foo.com",
	}
	if err := srv.Upsert(svc.Fqdn, svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while creating service")
		return
	}
	log.WithField("svc", svc).Infof("server:LocalResources: Created a service %s", svc.Fqdn)

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Update an existing service.
	svc.Name = "svc_blue"
	if err := srv.Upsert(svc.Fqdn, svc); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while updating service")
		return
	}
	log.WithField("svc", svc).Infof("server:LocalResources: Updated a service %s", svc.Fqdn)

	// Wait for some time.
	time.Sleep(1 * time.Second)

	// Delete an existing service.
	if err := srv.Delete(svc.Fqdn); err != nil {
		log.WithField("svc", svc).Errorln("Error occurred while deleting service")
		return
	}
	log.WithField("svc", svc).Infof("server:LocalResources: Deleted a service %s", svc.Fqdn)

	// Wait for some time.
	time.Sleep(1 * time.Second)
}

// create notification routines that will periodically
// update the service manifest
func createNotificationTask(srv server.Server) {
	// Notify the consumers about changes to resources.
	go notifyResourceChanges(srv)
}
