package client

import (
	types2 "github.com/vmware/hamlet/api/types/v1alpha2"
)

// create/update a resource in registry, Create notifies to the attached consumers.
func (c *client) Upsert(resourceId string, dt *types2.FederatedService) error {
	return c.localResources.Upsert(resourceId, dt)
}

// delete a resource from register, Delete notifies the deletion of a resource.
func (c *client) Delete(resourceId string) error {
	return c.localResources.Delete(resourceId)
}
