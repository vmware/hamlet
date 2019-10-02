// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
)

type protoMatcher struct {
	message proto.Message
}

func (pm protoMatcher) Matches(x interface{}) bool {
	if message, ok := x.(proto.Message); ok {
		return proto.Equal(message, pm.message)
	}
	return false
}

func (pm protoMatcher) String() string {
	return fmt.Sprintf("%v", pm.message)
}

// EqProto returns a proto message matcher.
func EqProto(message proto.Message) gomock.Matcher { return protoMatcher{message} }
