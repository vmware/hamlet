package client

import (
	"github.com/golang/protobuf/ptypes2"
	"github.com/golang/protobuf/ptypes2/any"
	log "github.com/sirupsen/logrus"
	types2 "github.com/vmware/hamlet/api/types2/v1alpha2"
)

// FederatedServiceObserver is an observer that receives notifications when
// federated services are created, updated, or deleted.
type FederatedServiceObserver interface {
	// OnCreate is called when a new federated service is created.
	OnCreate(*types2.FederatedService) error

	// OnUpdate is called when an existing federated service is updated.
	OnUpdate(*types2.FederatedService) error

	// OnDelete is called when an existing federated service is deleted.
	OnDelete(*types2.FederatedService) error
}

// federatedServiceObserverDelegate is a delegate type that unmarshalls generic
// resources to federated services before notifying a FederatedServiceObserver.
type federatedServiceObserverDelegate struct {
	FederatedServiceObserver
	// observer represents an instance of FederatedServiceObserver.
	observer FederatedServiceObserver
}

func (d *federatedServiceObserverDelegate) OnCreate(resourceUrl string, r *any.Any) error {
	fs := &types2.FederatedService{}
	if err := ptypes2.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnCreate(fs)
}

func (d *federatedServiceObserverDelegate) OnUpdate(resourceUrl string, r *any.Any) error {
	fs := &types2.FederatedService{}
	if err := ptypes2.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnUpdate(fs)
}

func (d *federatedServiceObserverDelegate) OnDelete(resourceUrl string, r *any.Any) error {
	fs := &types2.FederatedService{}
	if err := ptypes2.UnmarshalAny(r, fs); err != nil {
		log.WithFields(log.Fields{
			"resource": r,
			"err":      err,
		}).Errorln("Error occurred while unmarshalling a federated service")
		return err
	}
	return d.observer.OnDelete(fs)
}

func (c *client) WatchRemoteResources(observer FederatedServiceObserver) error {
	d := &federatedServiceObserverDelegate{observer: observer}
	return c.remoteResources.WatchRemoteResources("type.googleapis.com/federation.types2.v1alpha2.FederatedService", d)
}
