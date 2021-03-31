package provider

import (
	"fmt"
	"sync"

	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

type Registry interface {
	// Register creates a new entry for the given Provider identified by id.
	Register(id string) (Provider, error)

	// Deregister deregisters the provider identified by id.
	Deregister(id string) error
}

// registry is a concrete implementation of the registry interface.
type registry struct {
	Registry

	// providers holds a set of registered consumers.
	providers map[string]Provider

	//
	resources resources.RemoteResources

	// mutex synchronizes access to the consumer listing.
	mutex *sync.Mutex
}

// NewRegistry returns a new instance of the registry.
func NewRegistry(rr resources.RemoteResources) Registry {
	return &registry{
		providers: make(map[string]Provider),
		resources: rr,
		mutex:     &sync.Mutex{},
	}
}

func (r *registry) Register(id string) (Provider, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, found := r.providers[id]; found {
		return nil, fmt.Errorf("Provider with id %s already exists", id)
	}
	r.providers[id] = newProvider(id, r.resources)
	return r.providers[id], nil
}

func (r *registry) Deregister(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.providers, id)
	r.resources.DeleteProvider(id)
	return nil
}
