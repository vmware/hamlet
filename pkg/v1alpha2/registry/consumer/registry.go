// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/any"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
)

// Registry maintains an active set of consumers and provides a mechanism to
// interact with them.
type Registry interface {
	// Register creates a new entry for the given consumer identified by id.
	Register(id string) (Consumer, error)

	// Deregister deregisters the consumer identified by id.
	Deregister(id string) error

	// notify all the consumers of a change
	Notify(id string, obj *any.Any, op rd.StreamResponse_Operation) error
}

// registry is a concrete implementation of the registry interface.
type registry struct {
	Registry

	// stateProvider provides the mechanism to query the federated service
	// mesh owner implementation for the current state of a particular type
	// of resources.
	// stateProvider state.StateProvider

	// consumers holds a set of registered consumers.
	consumers map[string]Consumer

	// mutex synchronizes access to the consumer listing.
	mutex *sync.Mutex
}

// NewRegistry returns a new instance of the registry.
func NewRegistry() Registry {
	return &registry{
		consumers: make(map[string]Consumer),
		mutex:     &sync.Mutex{},
	}
}

func (r *registry) Register(id string) (Consumer, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, found := r.consumers[id]; found {
		return nil, fmt.Errorf("Consumer with id %s already exists", id)
	}

	r.consumers[id] = newConsumer(id)
	return r.consumers[id], nil
}

func (r *registry) Deregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.consumers, id)
	return nil
}

func (r *registry) Notify(id string, obj *any.Any, op rd.StreamResponse_Operation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, v := range r.consumers {
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
