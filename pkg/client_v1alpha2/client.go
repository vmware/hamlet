// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_v1alpha2

import (
	"context"
	"crypto/tls"

	log "github.com/sirupsen/logrus"
	"github.com/vmware/hamlet/pkg/v1alpha2/client"
)

// Client is an abstraction over the federated resource discovery protocol to be
// implemented by the client.
type ClientV1Alpha2 interface {
}

// client implements the federated resource discovery interface.
type clientV1Alpha2 struct {
	context string
	client  client.Client
}

// NewClient creates a new client instance.
func NewClient(serverAddr string, tlsConfig *tls.Config, context string) (ClientV1Alpha2, error) {
	c := &clientV1Alpha2{}
	c.context = context
	var err error
	c.client, err = client.NewClient(serverAddr, tlsConfig)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating client.")
		return nil, err
	}
	return c, nil
}

func (c *clientV1Alpha2) Start() {
	err := c.client.Start(context.Background(), "", c.context)
	if err != nil {
		log.WithField("err", err).Errorln("Error occurred while creating client.")
	}
}

// add/remove service operations
// discovered operations
