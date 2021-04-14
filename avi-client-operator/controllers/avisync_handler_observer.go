// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"github.com/go-logr/logr"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/access"
)

// federatedServiceObserver observes for updates related to federated services.
type federatedServiceObserver struct {
	access.FederatedServiceObserverV1Alpha2
	log logr.Logger
}

func (o *federatedServiceObserver) OnCreate(providerId string, fs *types2.FederatedService) error {
	o.log.Info("client:RemoteResources:Federated Created", "service", fs.GetFqdn(), "provider", providerId, "fs", fs)
	return nil
}

func (o *federatedServiceObserver) OnUpdate(providerId string, fs *types2.FederatedService) error {
	o.log.Info("client:RemoteResources:Federated Updated", "service", fs.GetFqdn(), "provider", providerId, "fs", fs)
	return nil
}

func (o *federatedServiceObserver) OnDelete(providerId string, fs *types2.FederatedService) error {
	o.log.Info("client:RemoteResources:Federated Deleted", "service", fs.GetFqdn(), "provider", providerId, "fs", fs)
	return nil
}
