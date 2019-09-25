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
