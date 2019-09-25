// Copyright 2019 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"

	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// ResourceObserver is a generic resource observer that receives notifications
// when resources are created, updated, or deleted.
type ResourceObserver interface {
	// OnCreate is called when a new resource is created.
	OnCreate(*any.Any) error

	// OnUpdate is called when an existing resource is updated.
	OnUpdate(*any.Any) error

	// OnDelete is called when an existing resource is deleted.
	OnDelete(*any.Any) error
}

// notifyObserver notifies the observer about a particular event on a particular
// resource as received from the federated service mesh owner.
func (c *client) notifyObserver(observer ResourceObserver, res *rd.StreamResponse) error {
	switch res.Operation {
	case rd.StreamResponse_CREATE:
		return observer.OnCreate(res.Resource)
	case rd.StreamResponse_UPDATE:
		return observer.OnUpdate(res.Resource)
	case rd.StreamResponse_DELETE:
		return observer.OnDelete(res.Resource)
	default:
		log.WithField("operation", res.Operation).Errorln("Unable to handle operation")
		return nil
	}
}

// prepareAcknowledgement prepares the acknowledgement for a previously consumed
// notification from the federated service mesh owner.
func (c *client) prepareAcknowledgement(req *rd.StreamRequest, nonce string, err error) {
	req.ResponseNonce = nonce
	if err == nil {
		req.Status = &status.Status{Code: int32(code.Code_OK)}
	} else {
		log.WithField("err", err).Errorln("Error occurred while processing stream response")
		req.Status = &status.Status{Code: int32(code.Code_UNAVAILABLE), Message: err.Error()}
	}
}

func (c *client) WatchResources(ctx context.Context, resourceUrl string, observer ResourceObserver) error {
	// Create a stream instance.
	stream, err := c.dsClient.EstablishStream(ctx)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating stream instance")
		return err
	}

	// Create the initial stream request.
	req := &rd.StreamRequest{ResourceUrl: resourceUrl}

	// Loop until the stream has ended.
	for {
		// Send the stream request.
		err := stream.Send(req)
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while sending stream request")
			return err
		}

		// Receive a message.
		res, err := stream.Recv()
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while consuming stream response")
			return err
		}

		// Notify the observer.
		err = c.notifyObserver(observer, res)

		// Prepare the acknowledgement.
		c.prepareAcknowledgement(req, res.Nonce, err)
	}
}
