// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	Log         logr.Logger
	Scheme      *runtime.Scheme
	syncMap     map[string]AVISyncHandler
	syncMapLock sync.Mutex
}

func NewAVISyncReconciler(mgr manager.Manager) *AVISyncReconciler {
	r := &AVISyncReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("AVISync"),
		Scheme:  mgr.GetScheme(),
		syncMap: make(map[string]AVISyncHandler),
	}
	return r
}

// +kubebuilder:rbac:groups=hamlet.tanzu.vmware.com,resources=avisyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hamlet.tanzu.vmware.com,resources=avisyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;watch;list

func (r *AVISyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("avisync", req.NamespacedName)

	aviSyncResource := &hamletv1alpha1.AVISync{}
	err := r.Get(ctx, req.NamespacedName, aviSyncResource)
	if err != nil {
		if errors.IsNotFound(err) {
			// log.Info("AviSync Resource not found. Ignoring.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get AviSync Resource")
		return ctrl.Result{}, err
	}

	log.Info("AVISync", "name", aviSyncResource.Name, "spec", fmt.Sprintf("%v", aviSyncResource.Spec))

	// check if the object is being removed and, in this case, delete all related objects
	finalizing, err := r.finalizerCheck(log, aviSyncResource)
	if err != nil {
		return reconcile.Result{}, err
	}

	if finalizing {
		// err = r.reconcileRemoval(instance, deployment, configMap, staticClientsPasswords)
		log.Info("Removing the AviSync Client for", "name", aviSyncResource.Name)
		r.removeSync(log, aviSyncResource.Name)
		return reconcile.Result{}, r.finalizerDone(log, aviSyncResource)
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
		aviSyncResource.Status.AviSyncSecretOk = err.Error()
		if err2 := r.updateAviSync(log, aviSyncResource); err2 != nil {
			err = fmt.Errorf("Error when Updating CRD %s\n%s", err2.Error(), err.Error())
		}
		return ctrl.Result{}, err
	} else if err != nil {
		log.Error(err, "Failed to get secrets")
		r.removeSync(log, aviSyncResource.Name)
		aviSyncResource.Status.AviSyncSecretOk = err.Error()
		if err2 := r.updateAviSync(log, aviSyncResource); err2 != nil {
			err = fmt.Errorf("Error when Updating CRD %s\n%s", err2.Error(), err.Error())
		}
		return ctrl.Result{}, err
	}
	log.Info(fmt.Sprintf("Secret found kind=%s type=%s", aviSecret.Kind, aviSecret.Type))

	if aviSecret.Type != AVISYNC_ExpSecretType {
		err = fmt.Errorf("Secret %s is of type %s, was expecting it to be of type %s",
			aviSyncResource.Name, aviSecret.Type, AVISYNC_ExpSecretType)
		log.Info(err.Error())
		aviSyncResource.Status.AviSyncSecretOk = err.Error()
		if err2 := r.updateAviSync(log, aviSyncResource); err2 != nil {
			err = fmt.Errorf("Error when Updating CRD %s\n%s", err2.Error(), err.Error())
		}
		return ctrl.Result{}, err
	}
	for _, key := range AVISYNC_SECRET_KEYS {
		if _, ok := aviSecret.Data[key]; !ok {
			err = fmt.Errorf("Secret %s is missing a required key %s or the key is empty",
				aviSyncResource.Name, key)
			log.Info(err.Error())
			r.removeSync(log, aviSyncResource.Name)
			aviSyncResource.Status.AviSyncSecretOk = err.Error()
			if err2 := r.updateAviSync(log, aviSyncResource); err2 != nil {
				err = fmt.Errorf("Error when Updating CRD %s\n%s", err2.Error(), err.Error())
			}
			return ctrl.Result{}, err
		}
	}
	secret := &AviSyncSecret{
		AviPassword:      string(aviSecret.Data[AVISYNC_SECRET_AVIPassword]),
		AviUsername:      string(aviSecret.Data[AVISYNC_SECRET_AVIUsername]),
		AviTenant:        string(aviSecret.Data[AVISYNC_SECRET_AVITenant]),
		HamletServerCert: aviSecret.Data[AVISYNC_SECRET_HamletServerCert],
		HamletToken:      string(aviSecret.Data[AVISYNC_SECRET_HamletToken])}

	if aviSyncResource.Status.AviSyncSecretOk != "ok" {
		aviSyncResource.Status.AviSyncSecretOk = "ok"
		if err := r.updateAviSync(log, aviSyncResource); err != nil {
			return ctrl.Result{}, err
		}
	}
	r.upsertSync(log, aviSyncResource, secret)
	return ctrl.Result{}, nil
}

func (r *AVISyncReconciler) updateAviSync(log logr.Logger, instance *hamletv1alpha1.AVISync) error {
	if err := r.Update(context.Background(), instance); err != nil {
		log.Info("ERROR: when updating aviSyncResource instance", "name", instance.GetName(), "err", err)
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
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

func (r *AVISyncReconciler) removeSync(log logr.Logger, name string) {
	r.syncMapLock.Lock()
	defer r.syncMapLock.Unlock()
	if sm, ok := r.syncMap[name]; ok {
		sm.StopSync()
		delete(r.syncMap, name)
	}
}

func (r *AVISyncReconciler) upsertSync(log logr.Logger, instance *hamletv1alpha1.AVISync, secret *AviSyncSecret) {
	r.syncMapLock.Lock()
	defer r.syncMapLock.Unlock()
	name := instance.Name
	if sm, ok := r.syncMap[name]; ok {
		log.Info("Updating an existing instance of sync", "name", instance.Name)
		err := sm.Update(instance, secret)
		if err != nil {
			log.Error(err, "While updating sync", "name", instance.Name)
		}
	} else {
		log.Info("Creating a new instance of sync", "name", instance.Name)
		aSync := newAviSyncHandler(instance, secret)
		r.syncMap[name] = aSync
		err := aSync.StartSync()
		if err != nil {
			log.Error(err, "While starting sync", "name", instance.Name)
		}
	}

}
