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

//go:generate mockgen --source=../../api/resourcediscovery/v1alpha1/resource_discovery.pb.go --destination=../../mocks/api/resourcediscovery/v1alpha1/resource_discovery.pb.go

package client

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
