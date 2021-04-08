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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	hamletv1alpha1 "github.com/vmware/hamlet/avi-client-operator/api/v1alpha1"
)

const (
	AVISYNC_ExpSecretType           = "hamlet.tanzu.vmware.com/avi-hamlet-client"
	AVISYNC_Finalizer               = "avisync.finalizer.hamlet.tanzu.vmware.com"
	AVISYNC_SECRET_AVIUsername      = "avi-username"
	AVISYNC_SECRET_AVIPassword      = "avi-password"
	AVISYNC_SECRET_AVITenant        = "avi-tenant"
	AVISYNC_SECRET_HamletToken      = "hamlet-token"
	AVISYNC_SECRET_HamletServerCert = "hamlet-server-cert"
)

var AVISYNC_SECRET_KEYS = []string{
	AVISYNC_SECRET_AVIUsername,
	AVISYNC_SECRET_AVIPassword,
	AVISYNC_SECRET_AVITenant,
	AVISYNC_SECRET_HamletToken,
	AVISYNC_SECRET_HamletServerCert,
}

// AVISyncReconciler reconciles a AVISync object
type AVISyncReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hamlet.tanzu.vmware.com,resources=avisyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hamlet.tanzu.vmware.com,resources=avisyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secret,verbs=get;watch;list

func (r *AVISyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("avisync", req.NamespacedName)

	aviSyncResource := &hamletv1alpha1.AVISync{}
	err := r.Get(ctx, req.NamespacedName, aviSyncResource)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("AviSync Resource not found. Ignoring.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get AviSync Resource")
		return ctrl.Result{}, err
	}

	// foundSecret := &appsv1.Secret{}
	// err = r.Get(ctx, types.NamespacedName{Name: hw.Name, Namespace: hw.Namespace}, found)
	log.Info(fmt.Sprintf("AVISync[%s] = %+v", aviSyncResource.Name, aviSyncResource.Spec))

	// check if the object is being removed and, in this case, delete all related objects
	finalizing, err := r.finalizerCheck(log, aviSyncResource)
	if err != nil {
		return reconcile.Result{}, err
	}

	if finalizing {
		// err = r.reconcileRemoval(instance, deployment, configMap, staticClientsPasswords)
		log.Info("Removing the AviSync Client for", "name", aviSyncResource.Name)
		r.removeSync(log, aviSyncResource.Name)
		r.finalizerDone(log, aviSyncResource)
		if err := r.Update(ctx, aviSyncResource); err != nil {
			log.Info("ERROR: when updating aviSyncResource instance", "name", aviSyncResource.GetName(), "err", err)
			if !apierrors.IsNotFound(err) {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	aviSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      aviSyncResource.Name,
		Namespace: aviSyncResource.Namespace}, aviSecret)
	if err != nil && errors.IsNotFound(err) {
		// of course we expect no deployment to be found
		log.Info("Secret not found.",
			"namespace", aviSyncResource.Namespace,
			"secret name", aviSyncResource.Name)
		r.removeSync(log, aviSyncResource.Name)
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get secrets")
		r.removeSync(log, aviSyncResource.Name)
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Secret found kind=%s type=%s data=%+v", aviSecret.Kind, aviSecret.Type, aviSecret.Data))
	// for k, v := range aviSecret.StringData {
	// 	log.Info("Secret string data ", "key", k, "value", v)
	// }
	if aviSecret.Type != AVISYNC_ExpSecretType {
		err = fmt.Errorf("Secret %s is of type %s, was expecting it to be of type %s",
			aviSyncResource.Name, aviSecret.Type, AVISYNC_ExpSecretType)
		log.Info(err.Error())
		return ctrl.Result{}, err
	}
	for _, key := range AVISYNC_SECRET_KEYS {
		if _, ok := aviSecret.Data[key]; !ok {
			err = fmt.Errorf("Secret %s is missing a required key %s or the key is empty",
				aviSyncResource.Name, key)
			log.Info(err.Error())
			r.removeSync(log, aviSyncResource.Name)
			return ctrl.Result{}, err
		}
	}
	// for k, v := range aviSecret.Data {
	// 	log.Info("Secret data ", "key", k, "value", v)
	// }
	r.upsertSync(log, aviSyncResource, aviSecret.Data)
	return ctrl.Result{}, nil
}

func (r *AVISyncReconciler) removeSync(log logr.Logger, name string) {

}

func (r *AVISyncReconciler) upsertSync(log logr.Logger, instance *hamletv1alpha1.AVISync, secret map[string][]byte) {

}

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

func (r *AVISyncReconciler) SetupWithManager(mgr ctrl.Manager, ns string) error {
	mapFn := handler.ToRequestsFunc(
		func(a handler.MapObject) []reconcile.Request {
			if a.Meta.GetNamespace() == ns || ns == "" {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      a.Meta.GetName(),
						Namespace: a.Meta.GetNamespace(),
					}},
				}
			}
			return []reconcile.Request{}
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&hamletv1alpha1.AVISync{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: mapFn}).
		Complete(r)
}
