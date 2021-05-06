// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
	rd "github.com/vmware/hamlet/api/resourcediscovery/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/stream_handler"
)

func (s *server) EstablishStream(stream rd.DiscoveryService_EstablishStreamServer) error {
	resourceUrl := "type.googleapis.com/federation.types.v1alpha2.FederatedService"
	// Retrieve initial publisher message.
	initData, err := stream.Recv()
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while establishing stream")
		return err
	}

	// check if request field is there
	initReq := initData.Request
	if initReq == nil {
		log.WithField("err", err).Errorf("Did not receive request as the first message after stream creation. %v", initData)
		return err
	}

	// check context
	if initReq.Context != s.connectionContext {
		log.WithField("err", err).Errorf("Received a unknown connection context:%s", initReq.Context)
		return err
	}

	// Generate a unique publisher ID.
	streamId := shortuuid.New()
	err = stream_handler.Handler(streamId, resourceUrl, s.connectionContext,
		stream, s.LocalResources, s.publisherRegistry, s.consumerRegistry, true)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while working with stream")
		return err
	}
	return nil
}
