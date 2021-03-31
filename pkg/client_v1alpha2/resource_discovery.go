// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_v1alpha2

import (
	"context"

	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
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

func (c *client) WatchResources(ctx context.Context, resourceUrl, connectionContext string, observer ResourceObserver) error {
	// Create a stream instance.

	stream, err := c.dsClient.EstablishStream(ctx)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating stream instance")
		return err
	}

	// Create the initial stream request.
	bReq := &BidirectionalStream{
		Request: &rd.StreamRequest{ResourceUrl: resourceUrl, Context: connectionContext}}
	err = stream.Send(bReq)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while sending first stream request")
		return err
	}
	// if publishing add the registry
	// publishing := true
	// // Identify consumer.
	// if publishing {
	// 	consumer, err := c.consumerRegistry.Register(streamId)
	// 	if err != nil {
	// 		log.WithFields(log.Fields{
	// 			"initReq": initReq,
	// 			"err":     err,
	// 		}).Errorln("Error occurred while retrieving consumer")
	// 		return err
	// 	}
	// }
	// Loop until the stream has ended.
	for {
		// Send the stream request.
		// err := stream.Send(req)
		// if err != nil {
		// 	log.WithField("err", err).Errorln("Error occurred while sending stream request")
		// 	return err
		// }

		// // Receive a message.
		res, err := stream.Recv()
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while consuming stream response")
			return err
		}
		streamRequest := res.GetRequest()
		steramResponse := res.GetResponse()

		if steramResponse != nil {
			err := c.notifyObserver(observer, steramResponse)
		}

		if streamRequest != nil {
			c.consumerRegistry.notify(streamRequest)
		}
		// // Notify the observer.
		// err = c.notifyObserver(observer, res)

		// // Prepare the acknowledgement.
		// c.prepareAcknowledgement(req, res.Nonce, err)
	}
}
func (c *client) sendDataToStream(stream *rd.DiscoveryService_EstablishStreamClient, data *rd.BidirectionalStream) error {
	err := (*stream).Send(data)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while sending stream request")
		return err
	}
	return nil
}
