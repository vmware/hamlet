// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stream_handler

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/consumer"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/provider"
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
	consumerRegistry consumer.Registry,
	providerRegistry provider.Registry,
	sendInitialReqPacket bool) error {

	sendLock := &sync.Mutex{}

	// add to consumer registry
	co, err := consumerRegistry.Register(streamId)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating consumer registry")
		return err
	}
	err = co.InitStream(resourceUrl, localResources)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while init call on consumer registry")
		return err
	}

	coRespChan, err := co.WatchStream(resourceUrl)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating watch stream")
		return err
	}
	// Register the Provider
	pr, err := providerRegistry.Register(streamId)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating provider registry")
		return err
	}

	if sendInitialReqPacket {
		// Create the initial stream request.
		req := &rd.StreamRequest{ResourceUrl: resourceUrl, Context: connectionContext}
		// Send the stream request.
		err = stream.Send(&rd.BidirectionalStream{Request: req})
		if err != nil {
			log.WithField("err", err).Errorln("Error occurred while sending initial stream request")
			return err
		}
	}

	// create a channel for processing received data
	recvDataChan := make(chan *rd.BidirectionalStream)
	go func() {
		r, err := stream.Recv()
		log.Infof("GOT DATA FROM RECV\n")

		if err != nil {
			log.WithField("err", err).Errorln("Error occurred on stream recv call")
			close(recvDataChan)
			return
		}
		recvDataChan <- r
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
			log.Infof("GOT DATA FROM Stream resp=%t req=%t\n", resp != nil, req != nil)
			if resp != nil {
				respAckNack, err := pr.AcceptStreamData(resp)
				if err != nil {
					log.WithField("err", err).Errorln("Error occurred while consuming response in producer")
				}
				if respAckNack != nil {
					r := &rd.BidirectionalStream{}
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
				co.ProcessAckNack(req)
			}

		case sendData, sendDataOk := <-coRespChan:
			log.Infof("GOT ACK/NACK DATA FROM CONSUMER to send back %v\n", sendData)
			if !sendDataOk {
				done = true
			}
			r := &rd.BidirectionalStream{}
			r.Response = sendData.Object
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
	return nil
}
