// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

type ResourcesHandler interface {
	// create/update a resource in registery,
	// called by publisher
	Upsert(consumerId string, resourceId string, message *any.Any) error

	// delete a resource from register, called by publisher
	Delete(consumerId string, resourceId string) error

	// delete a consumer
	DeleteProvider(providerId string) error
}
type Consumer interface {
	// this is a single stream that is trying to publish
	// services into the resource registry
	// we will consume messages and provide acknowledgement
	// here.

	// AcceptStream accepts messages from the consumer stream
	AcceptStreamData(obj *rd.StreamResponse) (*rd.StreamRequest, error)
}

type consumer struct {
	Consumer

	// id represents the unique identifier for the consumer.
	id string

	// Remote resources
	resourceHandler ResourcesHandler
	// mutex synchronizes the access to streams.
	mutex *sync.Mutex
}

func newConsumer(id string, resourceHandler ResourcesHandler) Consumer {
	return &consumer{
		id:              id,
		resourceHandler: resourceHandler,
		mutex:           &sync.Mutex{},
	}
}

func (p *consumer) AcceptStreamData(dt *rd.StreamResponse) (*rd.StreamRequest, error) {
	res := dt.GetResource()
	id := dt.GetResourceId()
	nonce := dt.GetNonce()
	var err error = nil
	if dt.Operation == rd.StreamResponse_UPSERT {
		log.Debugf("Consumer : Receive Upsert with id %s nonce %s\n", id, dt.Nonce)
		p.resourceHandler.Upsert(p.id, id, res)
	} else if dt.Operation == rd.StreamResponse_DELETE {
		log.Debugf("Consumer : Receive Delete with id %s nonce %s\n", id, dt.Nonce)
		p.resourceHandler.Delete(p.id, id)
	} else {
		err = fmt.Errorf("Error occurred while parsing the operation type %v", dt.GetOperation())
	}
	return p.prepareAcknowledgement(nonce, err), err
}

// prepareAcknowledgement prepares the acknowledgement for a previously consumed
// notification from the federated service mesh owner.
func (p *consumer) prepareAcknowledgement(nonce string, err error) *rd.StreamRequest {
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
