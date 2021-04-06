package consumer

import (
	"fmt"
	"sync"

	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

type Registry interface {
	// Register creates a new entry for the given Consumer identified by id.
	Register(id string) (Consumer, error)

	// Deregister deregisters the consumer identified by id.
	Deregister(id string) error
}

// registry is a concrete implementation of the registry interface.
type registry struct {
	Registry

	// consumers holds a set of registered consumers.
	consumers map[string]Consumer

	//
	resources resources.RemoteResources

	// mutex synchronizes access to the consumer listing.
	mutex *sync.Mutex
}

// NewRegistry returns a new instance of the registry.
func NewRegistry(rr resources.RemoteResources) Registry {
	return &registry{
		consumers: make(map[string]Consumer),
		resources: rr,
		mutex:     &sync.Mutex{},
	}
}

func (r *registry) Register(id string) (Consumer, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, found := r.consumers[id]; found {
		return nil, fmt.Errorf("Consumer with id %s already exists", id)
	}
	r.consumers[id] = newConsumer(id, r.resources)
	return r.consumers[id], nil
}

func (r *registry) Deregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.consumers, id)
	r.resources.DeleteProvider(id)
	return nil
}
