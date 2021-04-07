// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lifecycle_v1alpha2

import (
	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
)

type federatedServiceObserver struct {
	access.FederatedServiceObserverV1Alpha2
}

func (o *federatedServiceObserver) OnCreate(providerId string, fs *types2.FederatedService) error {
	log.Infof("server:RemoteResources:Federated service %s was created from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}

func (o *federatedServiceObserver) OnUpdate(providerId string, fs *types2.FederatedService) error {
	log.Infof("server:RemoteResources:Federated service %s was updated from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}

func (o *federatedServiceObserver) OnDelete(providerId string, fs *types2.FederatedService) error {
	log.Infof("server:RemoteResources:Federated service %s was deleted from provider %s\n", fs.GetFqdn(), providerId)
	return nil
}
