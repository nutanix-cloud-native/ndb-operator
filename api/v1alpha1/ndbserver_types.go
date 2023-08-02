/*
Copyright 2022-2023 Nutanix, Inc.

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

// NDBServerSpec defines the desired state of NDBServer
type NDBServerSpec struct {
	// +kubebuilder:validation:Required
	Server string `json:"server"`
	// +kubebuilder:validation:Required
	CredentialSecret string `json:"credentialSecret"`
	// +kubebuilder:default:=false
	// +optional
	// Skip server's certificate and hostname verification
	SkipCertificateVerification bool `json:"skipCertificateVerification"`
}

// NDBServerStatus defines the observed state of NDBServer
type NDBServerStatus struct {
	Status           string                           `json:"status"`
	LastUpdated      string                           `json:"lastUpdated"`
	Databases        map[string]NDBServerDatabaseInfo `json:"databases"`
	ReconcileCounter ReconcileCounter                 `json:"reconcileCounter"`
}

type ReconcileCounter struct {
	Database int `json:"database"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:shortName={"ndb","ndbs"}
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Updated At",type=string,JSONPath=`.status.lastUpdated`

// NDBServer is the Schema for the ndbservers API
type NDBServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NDBServerSpec   `json:"spec,omitempty"`
	Status NDBServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NDBServerList contains a list of NDBServer
type NDBServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NDBServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NDBServer{}, &NDBServerList{})
}

// Database related info to be stored in the status field of the NDB CR
type NDBServerDatabaseInfo struct {
	Name          string `json:"name"`
	Id            string `json:"id"`
	Status        string `json:"status"`
	DBServerId    string `json:"dbServerId"`
	TimeMachineId string `json:"timeMachineId"`
	IPAddress     string `json:"ipAddress"`
}
