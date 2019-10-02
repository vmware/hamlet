// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package state

import "github.com/golang/protobuf/proto"

// StateProvider provides a mechanism to detect the state of resources in a
// federated service mesh owner.
type StateProvider interface {
	// GetState returns the set of currently available resources of the
	// given type.
	GetState(resourceUrl string) ([]proto.Message, error)
}
