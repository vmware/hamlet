// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:generate mockgen --source=../../api/resourcediscovery/v1alpha1/resource_discovery.pb.go --destination=../../mocks/api/resourcediscovery/v1alpha1/resource_discovery.pb.go

package client_v1alpha2

import (
	"crypto/tls"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("using the client API", func() {

	Context("when creating a new client instance", func() {
		It("should return a client instance with the parameters set", func() {
			tlsConfig := &tls.Config{}
			c, err := NewClient("address:port", tlsConfig)
			cl := c.(*client)
			Expect(err).NotTo(HaveOccurred())
			Expect(cl).NotTo(BeNil())
			Expect(cl.serverAddr).To(Equal("address:port"))
			Expect(cl.tlsConfig).To(Equal(tlsConfig))
		})
	})

})
