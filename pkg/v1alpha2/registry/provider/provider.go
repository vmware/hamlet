// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	types1 "github.com/vmware/hamlet/api/types/v1alpha1"
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

type ResourcesHandler interface {
	// create/update a resource in registery,
	// called by publisher
	Upsert(providerId string, resourceId string, message proto.Message) error

	// delete a resource from register, called by publisher
	Delete(providerId string, resourceId string) error

	// delete a provider
	DeleteProvider(providerId string) error
}
type Provider interface {
	// this is a single stream that is trying to publish
	// services into the resource registry
	// we will consume messages and provide acknowledgement
	// here.

	// AcceptStream accepts messages from the provider stream
	AcceptStreamData(obj *rd.StreamResponse) (*rd.StreamRequest, error)
}

type provider struct {
	Provider

	// id represents the unique identifier for the consumer.
	id string

	resourceHandler ResourcesHandler
	// mutex synchronizes the access to streams.
	mutex *sync.Mutex
}

func newProvider(id string, resourceHandler ResourcesHandler) Provider {
	return &provider{
		id:              id,
		resourceHandler: resourceHandler,
		mutex:           &sync.Mutex{},
	}
}

func (p *provider) AcceptStreamData(dt *rd.StreamResponse) (*rd.StreamRequest, error) {
	res := dt.GetResource()
	id := ""
	nonce := dt.GetNonce()
	var err error = nil
	if res.GetTypeUrl() == "federation.types.v1alpha1.FederatedService" {
		fs := &types1.FederatedService{}
		res := dt.GetResource()
		if err := ptypes.UnmarshalAny(res, fs); err != nil {
			log.WithFields(log.Fields{
				"resource": res,
				"err":      err,
			}).Errorln("Error occurred while unmarshalling a federated service v1alpha1")
			return p.prepareAcknowledgement(nonce, err), err
		}
		id = fs.GetFqdn()
	} else if res.GetTypeUrl() == "federation.types.v1alpha2.FederatedService" {
		fs := &types2.FederatedService{}
		res := dt.GetResource()
		if err := ptypes.UnmarshalAny(res, fs); err != nil {
			log.WithFields(log.Fields{
				"resource": res,
				"err":      err,
			}).Errorln("Error occurred while unmarshalling a federated service v1alpha2")
			return p.prepareAcknowledgement(nonce, err), err
		}
		id = fs.GetFqdn()
	} else {
		err = fmt.Errorf("Error occurred while parsing the restream response data format with type %s", res.GetTypeUrl())
		return p.prepareAcknowledgement(nonce, err), err
	}

	if dt.Operation == rd.StreamResponse_CREATE || dt.Operation == rd.StreamResponse_UPDATE {
		p.resourceHandler.Upsert(p.id, id, res)
	} else if dt.Operation == rd.StreamResponse_DELETE {
		p.resourceHandler.Delete(p.id, id)
	} else {
		err = fmt.Errorf("Error occurred while parsing the operation type %v", dt.GetOperation())
	}
	return p.prepareAcknowledgement(nonce, err), err
}

// prepareAcknowledgement prepares the acknowledgement for a previously consumed
// notification from the federated service mesh owner.
func (p *provider) prepareAcknowledgement(nonce string, err error) *rd.StreamRequest {
	req := &rd.StreamRequest{}
	req.ResponseNonce = nonce
	if err == nil {
		req.Status = &status.Status{Code: int32(code.Code_OK)}
	} else {
		log.WithField("err", err).Errorln("Error occurred while processing stream response")
		req.Status = &status.Status{Code: int32(code.Code_UNAVAILABLE), Message: err.Error()}
	}
	return req
}
