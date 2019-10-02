// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"github.com/vmware/hamlet/pkg/server/consumer"
)

// Resources provides a mechanism for notifying federated service mesh consumers
// about changes to resources in a federated service mesh owner.
type Resources interface {
	// Create notifies the creation of a resource.
	Create(message proto.Message) error

	// Update notifies the updation of a resource.
	Update(message proto.Message) error

	// Delete notifies the deletion of a resource.
	Delete(message proto.Message) error
}

// resources is a concrete implementation of the Resources API that publishes
// messages to all registered federated service mesh consumers.
type resources struct {
	// consumerRegistry holds an active set of registered federated service
	// mesh consumers.
	consumerRegistry consumer.Registry
}

// NewResources returns a new instances of the Resources API implementation.
func NewResources(consumerRegistry consumer.Registry) Resources {
	return &resources{consumerRegistry: consumerRegistry}
}

// notifyConsumers notifies all the registered federated service mesh consumers
// about the given resource change.
func (r *resources) notifyConsumers(obj *any.Any, op rd.StreamResponse_Operation) error {
	for _, consumer := range r.consumerRegistry.GetAll() {
		obj := &rd.StreamResponse{
			ResourceUrl: obj.TypeUrl,
			Resource:    obj,
			Operation:   op,
		}
		if err := consumer.NotifyStream(obj); err != nil {
			log.WithField("err", err).Errorln("Error occurred while notifying consumer")
			return err
		}
	}
	return nil
}

func (r *resources) Create(message proto.Message) error {
	obj, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	return r.notifyConsumers(obj, rd.StreamResponse_CREATE)
}

func (r *resources) Update(message proto.Message) error {
	obj, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	return r.notifyConsumers(obj, rd.StreamResponse_UPDATE)
}

func (r *resources) Delete(message proto.Message) error {
	obj, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	return r.notifyConsumers(obj, rd.StreamResponse_DELETE)
}
