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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	aviClients "github.com/avinetworks/sdk/go/clients"
	aviSession "github.com/avinetworks/sdk/go/session"
	"github.com/go-logr/logr"
	hamletv1alpha1 "github.com/vmware/hamlet/avi-client-operator/api/v1alpha1"
	hamletClient "github.com/vmware/hamlet/pkg/v1alpha2/client"
	"istio.io/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

type AVISyncHandler interface {
	Update(instance *hamletv1alpha1.AVISync, secret *AviSyncSecret) error
	StopSync() error
	StartSync() error
}

type aviSyncHandler struct {
	crd               *hamletv1alpha1.AVISync
	secret            *AviSyncSecret
	log               logr.Logger
	ctx               context.Context
	ctxCancel         context.CancelFunc
	retryTimeInterval time.Duration
	syncStopDone      chan bool
	startStopMutex    sync.Mutex
}

// create a new instance of a single sync client
func newAviSyncHandler(instance *hamletv1alpha1.AVISync, secret *AviSyncSecret) AVISyncHandler {
	r := &aviSyncHandler{
		log:               ctrl.Log.WithName("controllers").WithName("AVISync").WithName(instance.Name),
		crd:               instance,
		secret:            secret,
		syncStopDone:      make(chan bool),
		retryTimeInterval: 30 * time.Second, // client connection retry time interval
	}
	return r
}

// compare if the config of the sync has changed
func (sh *aviSyncHandler) CompareEqual(instance *hamletv1alpha1.AVISync, secret *AviSyncSecret) bool {
	return sh.secret.CompareEqual(secret) &&
		sh.crd.Spec.HamletServerLocation == instance.Spec.HamletServerLocation &&
		sh.crd.Spec.AviControllerLocation == instance.Spec.AviControllerLocation &&
		sh.crd.Spec.AviControllerVersion == instance.Spec.AviControllerVersion
}

// update the sync handler with config, nop if config has not changed.
func (sh *aviSyncHandler) Update(instance *hamletv1alpha1.AVISync, secret *AviSyncSecret) error {
	if !sh.CompareEqual(instance, secret) {
		// delete the current sync
		// start a new sync
		err := sh.StopSync()
		if err != nil {
			return err
		}
		sh.crd = instance
		sh.secret = secret
		return sh.StartSync()

	}
	return nil
}

// start the sync process after the config has been updated
func (sh *aviSyncHandler) StartSync() error {
	log.Infof("Starting Sync...")
	sh.startStopMutex.Lock()
	if sh.ctxCancel != nil {
		sh.startStopMutex.Unlock()
		return errors.New("Start Sync is already active.")
	}
	sh.ctx = context.Background()
	ctx, cancel := context.WithCancel(sh.ctx)
	sh.ctxCancel = cancel
	sh.startStopMutex.Unlock()
	go func() {
		loopCount := 0
		done := false
		for {
			// client connection retry loop
			connCh := sh.tryConnect(ctx)
			select {
			case err := <-connCh:
				// a single iteration of the loop has concluded
				if err != nil {
					sh.log.Info("Client connection", "loop iteration", loopCount, " ends with error", err.Error())
				} else {
					sh.log.Info("Client connection ended with no errors", "loop iteration", loopCount)
				}
			case <-ctx.Done():
				done = true
			}
			if !done { // wait for retry time.
				select {
				case <-time.After(sh.retryTimeInterval):
					break
				case <-ctx.Done():
					done = true
				}
			}
			if done {
				break
			}
			// Wait for some time.
			time.Sleep(sh.retryTimeInterval)
			loopCount++
		}
		sh.startStopMutex.Lock()
		sh.syncStopDone <- true
		sh.ctxCancel = nil
		sh.startStopMutex.Unlock()
	}()
	log.Infof("Done Starting Sync...")
	return nil
}

// stop the current sync process
func (sh *aviSyncHandler) StopSync() error {
	sh.startStopMutex.Lock()
	if sh.ctxCancel == nil {
		sh.startStopMutex.Unlock()
		return errors.New("Start Sync is not active can't run stop sync.")
	}
	// trigger the cancel
	sh.log.Info("Stopping current client")
	sh.ctxCancel()
	sh.startStopMutex.Unlock()
	<-sh.syncStopDone
	sh.log.Info("Done Stopping current client")
	return nil
}

func (sh *aviSyncHandler) tryConnect(ctx context.Context) <-chan error {
	retCh := make(chan error)
	go func() {
		// Prepare the client instance. Alternative functions for tls.Config exist in the ./pkg/tls/tls.go
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(sh.secret.HamletServerCert) {
			retCh <- errors.New("credentials: failed to append certificate " + string(sh.secret.HamletServerCert))
			return
		}
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            cp,
		}
		hClient, err := hamletClient.NewClient(sh.crd.Spec.HamletServerLocation, tlsConfig)
		if err != nil {
			retCh <- errors.New(fmt.Sprintf("Error creating a new Client and connecting to server at address %s, error %s",
				sh.crd.Spec.HamletServerLocation, err.Error()))
			return
		}

		aClient, err := aviClients.NewAviClient(sh.crd.Spec.AviControllerLocation,
			sh.secret.AviUsername,
			aviSession.SetPassword(sh.secret.AviPassword),
			aviSession.SetTenant(sh.secret.AviTenant),
			aviSession.SetVersion(sh.crd.Spec.AviControllerVersion),
			aviSession.SetInsecure) // todo: Move to secure connection for avi controller.
		if err != nil {
			retCh <- errors.New(fmt.Sprintf("Error creating avi client connection to %s with username %s tenant %s and api version %s error: %s",
				sh.crd.Spec.AviControllerLocation, sh.secret.AviUsername, sh.secret.AviTenant, sh.crd.Spec.AviControllerVersion,
				err.Error()))
			hClient.Close()
			return
		}
		// run the main loop of client with ability to terminate
		// on termination, retry if it's needed.
		// create task to generate the service change notifications
		// createNotificationTask(cl)
		internalCtx, internalCtxCancel := context.WithCancel(ctx)
		sh.createAviSync(internalCtx, internalCtxCancel, hClient, aClient)
		// Watch for federated service notifications.
		err = hClient.WatchRemoteResources("w1", &federatedServiceObserver{log: sh.log})
		if err != nil && err != io.EOF {
			sh.log.Error(err, "Error occurred while watching federated services")
		}

		err = hClient.Start(internalCtx, "type.googleapis.com/federation.types.v1alpha2.FederatedService", sh.secret.HamletToken)
		if err != nil {
			internalCtxCancel()
			hClient.Close()
			retCh <- errors.New("Error occurred while starting client " + err.Error())
			return
		}
		retCh <- nil
	}()
	return retCh
}
