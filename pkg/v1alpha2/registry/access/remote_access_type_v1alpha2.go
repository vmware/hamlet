// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

// FederatedServiceObserver is an observer that receives notifications when
// federated services are created, updated, or deleted.
type FederatedServiceObserverV1Alpha2 interface {
	// OnUpsert is called when a new federated service is created or update.
	OnUpsert(providerId string, svc *types2.FederatedService) error

	// OnDelete is called when an existing federated service is deleted.
	OnDelete(providerId string, svc *types2.FederatedService) error
}

// federatedServiceObserverDelegate is a delegate type that unmarshalls generic
// resources to federated services before notifying a FederatedServiceObserver.
type federatedServiceObserverDelegate struct {
	FederatedServiceObserverV1Alpha2
	// observer represents an instance of FederatedServiceObserver.
	observer FederatedServiceObserverV1Alpha2
}

func (d *federatedServiceObserverDelegate) OnUpsert(resourceUrl, providerId string, r *any.Any) error {
	fs := &types2.FederatedService{}
	if err := ptypes.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnUpsert(providerId, fs)
}

func (d *federatedServiceObserverDelegate) OnDelete(resourceUrl, providerId string, r *any.Any) error {
	fs := &types2.FederatedService{}
	if err := ptypes.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnDelete(providerId, fs)
}

type RemoteResourceAccessV1Alpha2 struct {
	RemoteResources resources.RemoteResources
}

func (c *RemoteResourceAccessV1Alpha2) WatchRemoteResources(id string, observer FederatedServiceObserverV1Alpha2) error {
	d := &federatedServiceObserverDelegate{observer: observer}
	return c.RemoteResources.WatchRemoteResources(id, d)
}

func (c *RemoteResourceAccessV1Alpha2) UnwatchRemoteResources(id string) error {
	return c.RemoteResources.UnwatchRemoteResources(id)
}
