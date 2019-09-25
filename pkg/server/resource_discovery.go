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

package server

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha1"
	"github.com/vmware/hamlet/pkg/server/consumer"
)

// discoveryServiceServer is a concrete implementation of the resource discovery
// protocol server.
type discoveryServiceServer struct {
	rd.DiscoveryServiceServer

	// consumerRegistry holds an active set of registered federated service
	// mesh consumers.
	consumerRegistry consumer.Registry
}

// newDiscoveryServiceServer returns a new instance of the
// DiscoveryServiceServer implementation.
func newDiscoveryServiceServer(consumerRegistry consumer.Registry) rd.DiscoveryServiceServer {
	return &discoveryServiceServer{consumerRegistry: consumerRegistry}
}

func (s *discoveryServiceServer) EstablishStream(stream rd.DiscoveryService_EstablishStreamServer) error {
	// Retrieve initial consumer message.
	initReq, err := stream.Recv()
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while establishing stream")
		return err
	}

	// Generate a unique consumer ID.
	id, err := uuid.NewUUID()
	if err != nil {
		log.WithField("err", err).Errorln("Couldn't generate UUID")
		return err
	}
	streamId := id.String()

	// TODO: Simplify stream registration.

	// Identify consumer.
	consumer, err := s.consumerRegistry.Register(streamId)
	if err != nil {
		log.WithFields(log.Fields{
			"initReq": initReq,
			"err":     err,
		}).Errorln("Error occurred while retrieving consumer")
		return err
	}

	// Initialize stream.
	if err := consumer.InitStream(initReq.ResourceUrl); err != nil {
		log.WithFields(log.Fields{
			"initReq": initReq,
			"err":     err,
		}).Errorln("Error occurred while initializing stream")
		return err
	}
	log.WithField("initReq", initReq).Infoln("Consumer successfully subscribed to stream")

	// Close stream when exiting.
	defer func() {
		if err := consumer.CloseStream(initReq.ResourceUrl); err != nil {
			log.WithFields(log.Fields{
				"initReq": initReq,
				"err":     err,
			}).Errorln("Error occurred while closing stream")
		}

		if err := s.consumerRegistry.Deregister(streamId); err != nil {
			log.WithFields(log.Fields{
				"initReq": initReq,
				"err":     err,
			}).Errorln("Error occurred during consumer deregistration")
		}
	}()

	// Initialize watch.
	watchChan, err := consumer.WatchStream(initReq.ResourceUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"initReq": initReq,
			"err":     err,
		}).Errorln("Error occurred while initializing watch")
		return err
	}

	// Watch for events and notify consumer.
	var req *rd.StreamRequest = initReq
	for {
		select {
		case watchResp := <-watchChan:
			if watchResp.Error != nil {
				log.WithFields(log.Fields{
					"initReq": initReq,
					"err":     watchResp.Error,
				}).Errorln("Error occurred while watching")
				return err
			}

			if watchResp.Closed {
				log.WithField("req", req).Infoln("Watch stream was closed")
				return nil
			}

			if err = stream.Send(watchResp.Object); err != nil {
				log.WithFields(log.Fields{
					"req": req,
					"res": watchResp.Object,
					"err": watchResp.Error,
				}).Errorln("Error occurred while sending stream message")
				return err
			}

			if req, err = stream.Recv(); err != nil {
				log.WithField("err", err).Errorln("Error occurred while consuming acknowledgement")
				return err
			}

			// TODO: Handle ACKs/NACKs
			log.WithField("req", req).Infoln("Received acknowledgement")

		// Handle client disconnect.
		case <-stream.Context().Done():
			log.WithField("err", stream.Context().Err()).Errorln("Stream context was done")
			return nil
		}
	}

	return nil
}
