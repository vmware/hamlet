// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publisher

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
)

// MaxStreamBufferSize represents the maximum number of elements a stream can
// have buffered before a stream consumes it.
var MaxStreamBufferSize uint32 = 4096

type ResourceRegistry interface {
	GetFull(resourceUrl string) (map[string](*any.Any), error)

	// create/update a resource in registry,
	// called by publisher
	Upsert(resourceId string, message proto.Message) error

	// delete a resource from register, called by publisher
	Delete(resourceId string) error
}

// WatchResponse holds the information about a resource change event to be
// notified to the watcher.
type WatchResponse struct {
	// Object is the information about the change in a resource that's being
	// watched.
	Object *rd.StreamResponse

	// Closed tells if the stream being watched was closed.
	Closed bool

	// Error tells of any errors while processing the stream.
	Error error
}

// Publisher represents an instance of a federated service mesh publisher.
type Publisher interface {
	// InitStream initializes a resource stream for the publisher.
	InitStream(resourceUrl string, resourceRegistry ResourceRegistry) error

	// NotifyStream lazily notifies the relevant stream, if it exists, about
	// a change in a particular resource. This comes from publisher registry.
	NotifyStream(obj *rd.StreamResponse) error

	// for the stream side.
	// WatchStream publishes changes to resources that are being watched.
	WatchStream(resourceUrl string) (<-chan WatchResponse, error)
	// process ack/nack call response
	ProcessAckNack(obj *rd.StreamRequest)

	// CloseStream closes a resource stream.
	CloseStream(resourceUrl string) error
}

// publisher is a concrete implementation of the publisher API.
type publisher struct {
	Publisher

	// id represents the unique identifier for the publisher.
	id string

	// streams represent the set of streams that are currently active and
	// consuming the published info .
	streams map[string]chan WatchResponse

	// resource registry will be local resources that need
	// to be communicated to remote.
	resourceRegistry ResourceRegistry

	// mutex synchronizes the access to streams.
	mutex *sync.Mutex
}

// newPublisher returns a new instance of a publisher for the given id.
func newPublisher(id string) Publisher {
	return &publisher{
		id:      id,
		streams: make(map[string]chan WatchResponse),
		mutex:   &sync.Mutex{},
	}
}

func (c *publisher) InitStream(resourceUrl string, resourceRegistry ResourceRegistry) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, found := c.streams[resourceUrl]; found {
		return fmt.Errorf("Publisher already publishing to stream %s", resourceUrl)
	}
	c.resourceRegistry = resourceRegistry
	messages, err := resourceRegistry.GetFull(resourceUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"resourceUrl": resourceUrl,
			"err":         err,
		}).Errorln("Error occurred while retrieving state")
		return err
	}

	c.streams[resourceUrl] = make(chan WatchResponse, MaxStreamBufferSize)

	for id, message := range messages {

		if err := c.createStreamObjectFromAny(id, message); err != nil {
			log.WithFields(log.Fields{
				"resourceUrl": resourceUrl,
				"message":     message,
				"err":         err,
			}).Errorln("Error occurred while creating stream object")
			return err
		}
	}
	return nil
}

// createStreamObject publishes the given message to a relevant stream with the
// create operation.
func (c *publisher) createStreamObject(id string, message proto.Message) error {
	res, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}
	return c.createStreamObjectFromAny(id, res)
}

func (c *publisher) getNewNonce() string {
	return shortuuid.New()
}
func (c *publisher) createStreamObjectFromAny(id string, res *any.Any) error {
	obj := &rd.StreamResponse{
		ResourceUrl: res.TypeUrl,
		ResourceId:  id,
		Resource:    res,
		Operation:   rd.StreamResponse_UPSERT,
		Nonce:       c.getNewNonce(),
	}
	return c.notifyStream(obj.ResourceUrl, WatchResponse{Object: obj})
}

// notifyStream publishes the watch response to the given stream without
// blocking. If the buffer is full, this method returns an error.
func (c *publisher) notifyStream(resourceUrl string, wr WatchResponse) error {
	select {
	case c.streams[resourceUrl] <- wr:
		log.WithFields(log.Fields{
			"publisher":   c.id,
			"resourceUrl": resourceUrl,
			"wr":          wr,
		}).Debugf("Added object to stream")
	default:
		// TODO: Provide a better mechanism for handling buffer overflows.
		log.WithFields(log.Fields{
			"publisher":   c.id,
			"resourceUrl": resourceUrl,
			"wr":          wr,
		}).Errorln("Stream full, discarding object")
		return fmt.Errorf("Discarding object due to overflow in stream %s", resourceUrl)
	}
	return nil
}

func (c *publisher) NotifyStream(obj *rd.StreamResponse) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, found := c.streams[obj.ResourceUrl]; !found {
		return nil
	}
	obj.Nonce = c.getNewNonce()
	return c.notifyStream(obj.ResourceUrl, WatchResponse{Object: obj})
}

func (c *publisher) WatchStream(resourceUrl string) (<-chan WatchResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 	log.Infof("Publisher id=%s watching for %s\n", c.id, resourceUrl)
	stream, found := c.streams[resourceUrl]
	if !found {
		return nil, fmt.Errorf("Publisher hasn't subscribed to stream %s", resourceUrl)
	}
	return stream, nil
}

func (c *publisher) ProcessAckNack(obj *rd.StreamRequest) {
	if obj.Status.Code != 0 {
		log.Errorf("Publisher id=%s Got ACK/NACK %v\n", c.id, obj)
	} else {
		log.Debugf("Publisher id=%s Got ACK/NACK %v\n", c.id, obj)
	}
}
func (c *publisher) CloseStream(resourceUrl string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.notifyStream(resourceUrl, WatchResponse{Closed: true}); err != nil {
		log.WithFields(log.Fields{
			"resourceUrl": resourceUrl,
			"err":         err,
		}).Errorln("Error occurred while publishing stream closure")
		return err
	}

	close(c.streams[resourceUrl])
	delete(c.streams, resourceUrl)
	return nil
}
