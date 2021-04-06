// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/publisher"
)

// Resources are stored in the resource registry where
// items can be added that needs to be passed on to the hamlet federated mesh.

type LocalResources interface {
	// create/update a resource in registry, Create notifies to the attached publishers.
	Upsert(resourceId string, message proto.Message) error

	// delete a resource from register, Delete notifies the deletion of a resource.
	Delete(resourceId string) error

	// get a snapshot of all the items stored in the registery
	// used when a new stream is added as watcher for getting intial snapshot.
	GetFull(resourceUrl string) ([](*any.Any), error)
}

// resources is a concrete implementation of the Resources API that publishes
// messages to all registered federated service mesh publishers.
type localResources struct {
	resources          map[string]*any.Any
	publisherRegistery publisher.Registry
	// mutex synchronizes the access to streams.
	mutex *sync.Mutex
}

// NewResources returns a new instances of the Resources API implementation.
func NewLocalResources(publisherReg publisher.Registry) LocalResources {
	return &localResources{
		resources:          make(map[string]*any.Any),
		publisherRegistery: publisherReg,
		mutex:              &sync.Mutex{}}
}

// notifyPublishers notifies all the registered federated service mesh publishers
// about the given resource change.
func (r *localResources) notifyPublishers(id string, obj *any.Any, op rd.StreamResponse_Operation) error {
	if err := r.publisherRegistery.Notify(id, obj, op); err != nil {
		log.WithField("err", err).Errorln("Error occurred while notifying publisher")
		return err
	}
	return nil
}

func (r *localResources) Upsert(id string, message proto.Message) error {
	obj, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.resources[id]; ok {
		r.resources[id] = obj
		return r.notifyPublishers(id, obj, rd.StreamResponse_UPDATE)
	} else {
		r.resources[id] = obj
		return r.notifyPublishers(id, obj, rd.StreamResponse_CREATE)
	}
}

func (r *localResources) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if obj, ok := r.resources[id]; ok {
		delete(r.resources, id)
		return r.notifyPublishers(id, obj, rd.StreamResponse_DELETE)
	}
	return fmt.Errorf("Object not found with id %s", id)
}

func (r *localResources) GetFull(resourceUrl string) ([](*any.Any), error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	dt := [](*any.Any){}
	for _, obj := range r.resources {
		if obj.TypeUrl == resourceUrl || resourceUrl == "" {
			dt = append(dt, obj)
		}
	}
	return dt, nil
}
