// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	types "github.com/vmware/hamlet/api/types/v1alpha1"
)

// FederatedServiceObserver is an observer that receives notifications when
// federated services are created, updated, or deleted.
type FederatedServiceObserver interface {
	// OnCreate is called when a new federated service is created.
	OnCreate(*types.FederatedService) error

	// OnUpdate is called when an existing federated service is updated.
	OnUpdate(*types.FederatedService) error

	// OnDelete is called when an existing federated service is deleted.
	OnDelete(*types.FederatedService) error
}

// federatedServiceObserverDelegate is a delegate type that unmarshalls generic
// resources to federated services before notifying a FederatedServiceObserver.
type federatedServiceObserverDelegate struct {
	ResourceObserver

	// observer represents an instance of FederatedServiceObserver.
	observer FederatedServiceObserver
}

func (d *federatedServiceObserverDelegate) OnCreate(r *any.Any) error {
	fs := &types.FederatedService{}
	if err := ptypes.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnCreate(fs)
}

func (d *federatedServiceObserverDelegate) OnUpdate(r *any.Any) error {
	fs := &types.FederatedService{}
	if err := ptypes.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnUpdate(fs)
}

func (d *federatedServiceObserverDelegate) OnDelete(r *any.Any) error {
	fs := &types.FederatedService{}
	if err := ptypes.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnDelete(fs)
}

func (c *client) WatchFederatedServices(ctx context.Context, observer FederatedServiceObserver) error {
	d := &federatedServiceObserverDelegate{observer: observer}
	return c.WatchResources(ctx, "type.googleapis.com/federation.types.v1alpha1.FederatedService", d)
}
