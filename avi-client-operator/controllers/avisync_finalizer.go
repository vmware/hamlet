/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	hamletv1alpha1 "github.com/vmware/hamlet/avi-client-operator/api/v1alpha1"
)

// finalizerDone marks the instance as "we are done with it, you can remove it now"
// Removal of the `instance` is blocked until we run this function, so make sure you don't
// forget about calling it...
func (r *AVISyncReconciler) finalizerDone(log logr.Logger, instance *hamletv1alpha1.AVISync) error {
	// Helper functions to check and remove string from a slice of strings.
	removeString := func(slice []string, s string) (result []string) {
		for _, item := range slice {
			if item == s {
				continue
			}
			result = append(result, item)
		}
		return
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		panic(fmt.Sprintf("finalizerDone() called on %s when it was not being deleted", instance.GetName()))
	}

	log.Info("AviSync is being cleaned up.", "name", instance.GetName())
	instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, AVISYNC_Finalizer)
	if err := r.Update(context.Background(), instance); err != nil {
		log.Info("ERROR: when updating aviSyncResource instance", "name", instance.GetName(), "err", err)
		if !apierrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

// finalizerCheck checks if the object is being finalized and, in that case,
// remove all the related objects
func (r *AVISyncReconciler) finalizerCheck(log logr.Logger, instance *hamletv1alpha1.AVISync) (bool, error) {
	// Helper functions to check and remove string from a slice of strings.
	containsString := func(slice []string, s string) bool {
		for _, item := range slice {
			if item == s {
				return true
			}
		}
		return false
	}

	finalizing := false
	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, AVISYNC_Finalizer) {
			log.Info(fmt.Sprintf("AviSync %s does not have finalizer %s registered: adding it", instance.GetName(), AVISYNC_Finalizer))
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, AVISYNC_Finalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return false, err
			}
		}
	} else {
		log.Info("AviSync is being deleted", "name", instance.GetName())
		finalizing = true
	}

	return finalizing, nil
}
