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
)

type ResourceObserver interface {
	// OnCreate is called when a new resource is created.
	OnCreate(resourceUrl string, dt *any.Any) error

	// OnUpdate is called when an existing resource is updated.
	OnUpdate(resourceUrl string, dt *any.Any) error

	// OnDelete is called when an existing resource is deleted.
	OnDelete(resourceUrl string, dt *any.Any) error
}

// Resources are stored in the resource registry where
// items can be added that needs to be passed on to the hamlet federated mesh.

type RemoteResources interface {

	// create/update a resource in registery,
	// called by publisher
	Upsert(providerId string, resourceId string, message proto.Message) error

	// delete a resource from register, called by publisher
	Delete(providerId string, resourceId string) error

	// delete a provider
	DeleteProvider(providerId string) error

	// WatchFederatedServices watches for notifications related to federated
	// services on the federated service mesh owner.
	WatchRemoteResources(id string, observer ResourceObserver) error
}

// resources is a concrete implementation of the Resources API that publishes
// messages to all registered federated service mesh consumers.
type remoteResources struct {
	// consumerRegistry holds an active set of registered federated service
	// mesh consumers.
	resources map[string]map[string]*any.Any
	observers map[string]ResourceObserver
	mutex     *sync.Mutex
}

// NewResources returns a new instances of the Resources API implementation.
func NewRemoteResources() RemoteResources {
	return &remoteResources{
		resources: make(map[string]map[string]*any.Any),
		observers: make(map[string]ResourceObserver),
		mutex:     &sync.Mutex{}}
}

// notifyObserver notifies the observer about a particular event on a particular
// resource as received from the federated service mesh owner.
func (r *remoteResources) notifyObserver(observer ResourceObserver, res *rd.StreamResponse) error {
	switch res.Operation {
	case rd.StreamResponse_CREATE:
		return observer.OnCreate(res.ResourceUrl, res.Resource)
	case rd.StreamResponse_UPDATE:
		return observer.OnUpdate(res.ResourceUrl, res.Resource)
	case rd.StreamResponse_DELETE:
		return observer.OnDelete(res.ResourceUrl, res.Resource)
	default:
		log.WithField("operation", res.Operation).Errorln("Unable to handle operation")
		return nil
	}
}

func (r *remoteResources) WatchRemoteResources(id string, observer ResourceObserver) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.observers[id]; ok {
		return fmt.Errorf("Observer with id %s already exists.", id)
	}
	r.observers[id] = observer
	for _, p := range r.resources {
		for _, v := range p {
			err := observer.OnCreate(v.TypeUrl, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *remoteResources) Upsert(providerId string, resourceId string, message proto.Message) error {
	obj, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	p, ok := r.resources[providerId]
	if !ok {
		p = make(map[string]*any.Any)
		r.resources[providerId] = p
	}
	if _, ok := p[resourceId]; ok {
		p[resourceId] = obj
		for _, o := range r.observers {
			err := o.OnUpdate(obj.TypeUrl, obj)
			if err != nil {
				return err
			}
		}
	} else {
		p[resourceId] = obj
		for _, o := range r.observers {
			err := o.OnCreate(obj.TypeUrl, obj)
			if err != nil {
				return err
			}

		}
	}
	return nil
}

func (r *remoteResources) Delete(providerId, resourceId string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	p, ok := r.resources[providerId]
	if !ok {
		return fmt.Errorf("Unable to find provider id = %s when trying to delete %s", providerId, resourceId)
	}
	if obj, ok := p[resourceId]; ok {
		delete(p, resourceId)
		for _, o := range r.observers {
			o.OnDelete(obj.TypeUrl, obj)
		}
	} else {
		return fmt.Errorf("Unable to find resource with id = %s", resourceId)
	}
	return nil
}

func (r *remoteResources) DeleteProvider(providerId string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	p, ok := r.resources[providerId]
	if !ok {
		return fmt.Errorf("Unable to find provider id = %s when trying to delete provider", providerId)
	}
	for rid, obj := range p {
		delete(p, rid)
		for _, o := range r.observers {
			o.OnDelete(obj.TypeUrl, obj)
		}
	}
	delete(r.resources, providerId)
	return nil
}
