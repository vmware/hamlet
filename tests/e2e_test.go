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

//go:generate mockgen --source=../pkg/client/federated_service.go --destination=../mocks/pkg/client/federated_service.go

package tests

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	types "github.com/vmware/hamlet/api/types/v1alpha1"
	mock_client "github.com/vmware/hamlet/mocks/pkg/client"
	"github.com/vmware/hamlet/pkg/client"
	"github.com/vmware/hamlet/pkg/server"
	"github.com/vmware/hamlet/pkg/server/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// randomPort returns a free TCP port.
func randomPort() uint32 {
	listener, err := net.Listen("tcp", ":0")
	Expect(err).NotTo(HaveOccurred())

	port := listener.Addr().(*net.TCPAddr).Port

	err = listener.Close()
	Expect(err).NotTo(HaveOccurred())
	return uint32(port)
}

var cachedPortNumber uint32

func cachedPort() uint32 {
	if cachedPortNumber == 0 {
		cachedPortNumber = randomPort()
	}
	return cachedPortNumber
}

type emptyProvider struct {
	state.StateProvider
}

func (p *emptyProvider) GetState(string) ([]proto.Message, error) {
	return []proto.Message{}, nil
}

var _ = Describe("e2e tests", func() {

	Context("with a server/client instance", func() {
		var (
			cl       client.Client
			sr       server.Server
			mockCtrl *gomock.Controller
			err      error
		)

		BeforeEach(func() {
			port := cachedPort()
			sr, err = server.NewServer(port, nil, &emptyProvider{})
			Expect(err).NotTo(HaveOccurred())

			go func() {
				err = sr.Start()
				Expect(err).NotTo(HaveOccurred())
			}()

			cl, err = client.NewClient(fmt.Sprintf("[::]:%d", port), nil)
			Expect(err).NotTo(HaveOccurred())

			mockCtrl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			mockCtrl.Finish()

			err = sr.Stop()
			Expect(err).NotTo(HaveOccurred())

			mockCtrl = nil
			cl = nil
			sr = nil
			err = nil
		})

		Context("when there's an update on the server", func() {
			It("should notify a consumer", func() {
				mockObs := mock_client.NewMockFederatedServiceObserver(mockCtrl)

				go func() {
					defer GinkgoRecover()

					err = cl.WatchFederatedServices(context.Background(), mockObs)
					Expect(err).NotTo(HaveOccurred())
				}()

				time.Sleep(2 * time.Second)

				svc1 := types.FederatedService{
					Name: "svc1",
					Id:   "svc1.foo.com",
				}
				mockObs.EXPECT().OnCreate(EqProto(&svc1)).Return(nil)
				err = sr.Resources().Create(&svc1)
				Expect(err).NotTo(HaveOccurred())

				svc2 := svc1
				svc2.Name = "foo svc"
				mockObs.EXPECT().OnUpdate(EqProto(&svc2)).Return(nil)
				err = sr.Resources().Update(&svc2)
				Expect(err).NotTo(HaveOccurred())

				svc3 := svc2
				mockObs.EXPECT().OnDelete(EqProto(&svc3)).Return(nil)
				err = sr.Resources().Delete(&svc3)
				Expect(err).NotTo(HaveOccurred())

				time.Sleep(2 * time.Second)
			})
		})
	})

})
