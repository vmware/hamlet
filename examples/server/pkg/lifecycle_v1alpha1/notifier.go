// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha1

import (
	"time"

	log "github.com/sirupsen/logrus"
	types "github.com/vmware/hamlet/api/types/v1alpha1"
	"github.com/vmware/hamlet/pkg/server"
)

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
