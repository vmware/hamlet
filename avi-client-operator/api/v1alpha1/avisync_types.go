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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AVISyncSpec defines the desired state of AVISync
type AVISyncSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	AviControllerLocation string `json:"avi-controller-location"`
	AviControllerVersion  string `json:"avi-controller-version,omitempty"`
	// secrets will be used for getting
	// avi tenant name , username, password

	// grpc connection
	HamletServerLocation string `json:"hamlet-server-location"`
	// hamlet token is in the secret
	// A secret by the same name as the AVISync is needed
	/* Secret  Schema
	{
		"avi-username": "",
		"avi-password": "",
		"avi-tenant": "",
		"hamlet-token": "",
		"hamlet-server-cert": ""
	}
	*/
}

// AVISyncStatus defines the observed state of AVISync
type AVISyncStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	AviSyncSecretOk                     string `json:"avi-sync-secret-ok"`
	AviConnectionHealth                 string `json:"avi-connection-health"`
	SuccessfulAviConnectionTimestamp    string `json:"successful-avi-connection-timestamp"`
	SuccessfulHamletConnectionTimestamp string `json:"successful-hamlet-connection-timestamp"`
	FailedAviConnectionTimestamp        string `json:"failed-avi-connection-timestamp"`
	FailedHamletConnectionTimestamp     string `json:"failed-hamlet-connection-timestamp"`
	AviConnectionRetryCount             int32  `json:"avi-connection-retry-count"`
	HamletConnectionRetryCount          int32  `json:"hamlet-connection-retry-count"`
}

// +kubebuilder:object:root=true

// AVISync is the Schema for the avisyncs API
type AVISync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AVISyncSpec   `json:"spec,omitempty"`
	Status AVISyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AVISyncList contains a list of AVISync
type AVISyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AVISync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AVISync{}, &AVISyncList{})
}
