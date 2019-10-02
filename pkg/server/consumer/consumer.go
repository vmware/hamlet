// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package consumer

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"github.com/vmware/hamlet/pkg/server/state"
)

// MaxStreamBufferSize represents the maximum number of elements a stream can
// have buffered before a consumer consumes it.
var MaxStreamBufferSize uint32 = 4096

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

// Consumer represents an instance of a federated service mesh consumer.
type Consumer interface {
	// InitStream initializes a resource stream for the consumer.
	InitStream(resourceUrl string) error

	// NotifyStream lazily notifies the relevant stream, if it exists, about
	// a change in a particular resource.
	NotifyStream(obj *rd.StreamResponse) error

	// WatchStream publishes changes to resources that are being watched.
	WatchStream(resourceUrl string) (<-chan WatchResponse, error)

	// CloseStream closes a resource stream for the consumer.
	CloseStream(resourceUrl string) error
}

// consumer is a concrete implementation of the consumer API.
type consumer struct {
	Consumer

	// id represents the unique identifier for the consumer.
	id string

	// stateProvider provides the mechanism to query the federated service
	// mesh owner implementation for the current state of a particular type
	// of resources.
	stateProvider state.StateProvider

	// streams represent the set of streams that are currently subscribed to
	// by the federated service mesh consumer.
	streams map[string]chan WatchResponse

	// mutex synchronizes the access to streams.
	mutex *sync.Mutex
}

// newConsumer returns a new instance of a consumer for the given id.
func newConsumer(id string, stateProvider state.StateProvider) Consumer {
	return &consumer{
		id:            id,
		stateProvider: stateProvider,
		streams:       make(map[string]chan WatchResponse),
		mutex:         &sync.Mutex{},
	}
}

func (c *consumer) InitStream(resourceUrl string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.streams[resourceUrl]; found {
		return fmt.Errorf("Consumer already subscribed to stream %s", resourceUrl)
	}

	messages, err := c.stateProvider.GetState(resourceUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"resourceUrl": resourceUrl,
			"err":         err,
		}).Errorln("Error occurred while retrieving state")
		return err
	}

	c.streams[resourceUrl] = make(chan WatchResponse, MaxStreamBufferSize)

	for _, message := range messages {
		if err := c.createStreamObject(message); err != nil {
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
func (c *consumer) createStreamObject(message proto.Message) error {
	res, err := ptypes.MarshalAny(message)
	if err != nil {
		log.WithField("err", err).Errorln("Failed to marshal proto message")
		return err
	}

	obj := &rd.StreamResponse{
		ResourceUrl: res.TypeUrl,
		Resource:    res,
		Operation:   rd.StreamResponse_CREATE,
	}
	return c.notifyStream(obj.ResourceUrl, WatchResponse{Object: obj})
}

// notifyStream publishes the watch response to the given stream without
// blocking. If the buffer is full, this method returns an error.
func (c *consumer) notifyStream(resourceUrl string, wr WatchResponse) error {
	select {
	case c.streams[resourceUrl] <- wr:
		log.WithFields(log.Fields{
			"consumer":    c.id,
			"resourceUrl": resourceUrl,
			"wr":          wr,
		}).Infoln("Added object to stream")
	default:
		// TODO: Provide a better mechanism for handling buffer overflows.
		log.WithFields(log.Fields{
			"consumer":    c.id,
			"resourceUrl": resourceUrl,
			"wr":          wr,
		}).Errorln("Stream full, discarding object")
		return fmt.Errorf("Discarding object due to overflow in stream %s", resourceUrl)
	}
	return nil
}

func (c *consumer) NotifyStream(obj *rd.StreamResponse) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.streams[obj.ResourceUrl]; !found {
		return nil
	}
	return c.notifyStream(obj.ResourceUrl, WatchResponse{Object: obj})
}

func (c *consumer) WatchStream(resourceUrl string) (<-chan WatchResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	stream, found := c.streams[resourceUrl]
	if !found {
		return nil, fmt.Errorf("Consumer hasn't subscribed to stream %s", resourceUrl)
	}
	return stream, nil
}

func (c *consumer) CloseStream(resourceUrl string) error {
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
