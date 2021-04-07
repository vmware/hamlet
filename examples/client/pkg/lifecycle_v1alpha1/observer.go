// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha1

import (
	log "github.com/sirupsen/logrus"
	types1 "github.com/vmware/hamlet/api/types/v1alpha1"
	"github.com/vmware/hamlet/pkg/client"
)

// federatedServiceObserver observes for updates related to federated services.
type federatedServiceObserver struct {
	client.FederatedServiceObserver
}

func (o *federatedServiceObserver) OnCreate(fs *types1.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was created")
	return nil
}

func (o *federatedServiceObserver) OnUpdate(fs *types1.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was updated")
	return nil
}

func (o *federatedServiceObserver) OnDelete(fs *types1.FederatedService) error {
	log.WithField("fs", fs).Infoln("Federated service was deleted")
	return nil
}
