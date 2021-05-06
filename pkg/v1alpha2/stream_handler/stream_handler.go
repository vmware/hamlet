// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stream_handler

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/publisher"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

type StreamInterface interface {
	Send(*rd.BidirectionalStream) error
	Context() context.Context
	Recv() (*rd.BidirectionalStream, error)
}

func Handler(
	streamId, resourceUrl, connectionContext string,
	stream StreamInterface,
	localResources resources.LocalResources,
	publisherRegistry publisher.Registry,
	consumerRegistry consumer.Registry,
	sendInitialReqPacket bool) error {

	sendLock := &sync.Mutex{}

	// add to publisher registry
	pub, err := publisherRegistry.Register(streamId)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating publisher registry")
		return err
	}
	defer func(streamId string) {
		err := publisherRegistry.Deregister(streamId)
		if err != nil {
			log.WithField("err", err).Errorf("Error occurred while deRegistering publisher %s", streamId)
		}
	}(streamId)

	err = pub.InitStream(resourceUrl, localResources)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while init call on publisher registry")
		return err
	}

	pubChan, err := pub.WatchStream(resourceUrl)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating watch stream")
		return err
	}
	// Register the consumer
	pr, err := consumerRegistry.Register(streamId)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating consumer registry")
		return err
	}
	defer func(streamId string) {
		err := consumerRegistry.Deregister(streamId)
		if err != nil {
			log.WithField("err", err).Errorf("Error occurred while deRegistering consumer %s", streamId)
		}
	}(streamId)

	if sendInitialReqPacket {
		// Create the initial stream request.
		req := &rd.StreamRequest{ResourceUrl: resourceUrl, Context: connectionContext}
		// Send the stream request.
		sendLock.Lock()
		err = stream.Send(&rd.BidirectionalStream{Request: req})
		sendLock.Unlock()
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while sending initial stream request")
			return err
		}
	}

	// create a channel for processing received data
	recvDataChan := make(chan *rd.BidirectionalStream, 100)
	go func() {
		for {
			r, err := stream.Recv()
			if err != nil {
				log.WithField("err", err).Errorln("Error occurred on stream recv call")
				close(recvDataChan)
				return
			}
			recvDataChan <- r
		}
	}()

	for {
		done := false
		select {
		case recvData, recvDataOk := <-recvDataChan:
			if !recvDataOk {
				done = true
			}
			resp := recvData.GetResponse()
			req := recvData.GetRequest()
			// log.Infof("GOT DATA FROM Stream resp=%t req=%t\n", resp != nil, req != nil)
			if resp != nil {
				respAckNack, err := pr.AcceptStreamData(resp)
				if err != nil {
					log.WithField("err", err).Errorln("Error occurred while consuming response in producer")
				}
				if respAckNack != nil {
					r := &rd.BidirectionalStream{}
					respAckNack.Context = connectionContext
					r.Request = respAckNack
					sendLock.Lock()
					err = stream.Send(r)
					sendLock.Unlock()
					if err != nil {
						log.WithField("err", err).Errorln("Error occurred while sending stream response data.")
					}
				}
			}
			if req != nil {
				if req.ResponseNonce != "" {
					pub.ProcessAckNack(req)
				} else {
					log.Infof("got Request without Nonce(Initial setup Request). %v\n", req)
				}
			}

		case sendData, sendDataOk := <-pubChan:
			if !sendDataOk {
				done = true
			}
			if sendData.Error != nil {
				log.Errorf("Error while waiting on pubChan %v\n", sendData.Error)
			}
			r := &rd.BidirectionalStream{}
			r.Response = sendData.Object
			r.Response.Context = connectionContext
			sendLock.Lock()
			err = stream.Send(r)
			sendLock.Unlock()
			if err != nil {
				log.WithField("err", err).Errorln("Error occurred while sending stream data.")
			}
			// Handle client disconnect.
		case <-stream.Context().Done():
			log.WithFields(log.Fields{
				"connectionContext": connectionContext,
				"err":               stream.Context().Err(),
			}).Warnln("Stream context was done")
			done = true
		}
		if done {
			break
		}
	}
	log.Infof("Stream Handler Done\n")
	return nil
}
