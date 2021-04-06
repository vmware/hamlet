// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publisher

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
)

// Registry maintains an active set of publishers and provides a mechanism to
// interact with them.
type Registry interface {
	// Register creates a new entry for the given publisher identified by id.
	Register(id string) (Publisher, error)

	// Deregister deregisters the publisher identified by id.
	Deregister(id string) error

	// notify all the publishers of a change
	Notify(id string, obj *any.Any, op rd.StreamResponse_Operation) error
}

// registry is a concrete implementation of the registry interface.
type registry struct {
	Registry

	// stateProvider provides the mechanism to query the federated service
	// mesh owner implementation for the current state of a particular type
	// of resources.
	// stateProvider state.StateProvider

	// publishers holds a set of registered publisher.
	publishers map[string]Publisher

	// mutex synchronizes access to the publisher listing.
	mutex *sync.Mutex
}

// NewRegistry returns a new instance of the registry.
func NewRegistry() Registry {
	return &registry{
		publishers: make(map[string]Publisher),
		mutex:      &sync.Mutex{},
	}
}

func (r *registry) Register(id string) (Publisher, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, found := r.publishers[id]; found {
		return nil, fmt.Errorf("Publisher with id %s already exists", id)
	}

	r.publishers[id] = newPublisher(id)

	return r.publishers[id], nil
}

func (r *registry) Deregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	log.Infof("Publisher DERegister SIZE %d %p\n", len(r.publishers), r)

	delete(r.publishers, id)
	return nil
}

func (r *registry) Notify(id string, obj *any.Any, op rd.StreamResponse_Operation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, v := range r.publishers {
		obj := &rd.StreamResponse{
			ResourceUrl: obj.TypeUrl,
			Resource:    obj,
			Operation:   op,
		}
		e := v.NotifyStream(obj)
		if e != nil {
			return e
		}
	}
	return nil
}
