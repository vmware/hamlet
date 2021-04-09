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
	"time"

	aviClients "github.com/avinetworks/sdk/go/clients"
	hamletClient "github.com/vmware/hamlet/pkg/v1alpha2/client"
)

func (sh *aviSyncHandler) createAviSync(ctx context.Context, ctxCancel context.CancelFunc, hClient hamletClient.Client, aClient *aviClients.AviClient) <-chan error {
	// connect to avi services
	// discover all virtual services in avi
	// pass the virtual service as entry to hamlet
	retCh := make(chan error)
	go func() {
		loopCount := 0
		done := false
		for {
			cv, err := aClient.AviSession.GetControllerVersion()
			sh.log.Info("AVI Controller ", "iteration", loopCount, "Version", cv, "Error", err.Error())
			// for any error we can call ctxCancel()
			select {
			case <-ctx.Done():
				done = true
			case <-time.After(sh.retryTimeInterval):
				break
			}
			if done {
				break
			}
			loopCount++
		}
		retCh <- nil
	}()
	return retCh
}
