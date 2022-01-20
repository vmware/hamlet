// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package access

import (
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
	"github.com/vmware/hamlet/pkg/v1alpha2/registry/resources"
)

type LocalResourceAccessV1Alpha2 struct {
	LocalResources resources.LocalResources
}

// create/update a resource in registry, Create notifies to the attached publisher.
func (c *LocalResourceAccessV1Alpha2) Upsert(resourceId string, dt *types2.FederatedService) error {
	return c.LocalResources.Upsert(resourceId, dt)
}

// delete a resource from register, Delete notifies the deletion of a resource.
func (c *LocalResourceAccessV1Alpha2) Delete(resourceId string) error {
	return c.LocalResources.Delete(resourceId)
}

// delete a resource from register, Delete notifies the deletion of a resource.
func (c *LocalResourceAccessV1Alpha2) GetAllResourceID(resourceType string) []string {
	return c.LocalResources.GetAllResourceID(resourceType)
}