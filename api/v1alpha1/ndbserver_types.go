/*
Copyright 2022-2026 Nutanix, Inc.

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

// SecretReference references a Secret by name and namespace (e.g. for NDB API credentials in a restricted namespace).
type SecretReference struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

// NDBServerSpec defines the desired state of NDBServer
type NDBServerSpec struct {
	// +kubebuilder:validation:Required
	Server string `json:"server"`
	// Reference to the secret holding NDB API credentials. The secret can live in a restricted namespace
	// so that developers with access to NDBServer do not need access to the secret.
	// +kubebuilder:validation:Required
	CredentialSecretRef SecretReference `json:"credentialSecretRef"`
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName={"ndb","ndbs"}
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Updated At",type=string,JSONPath=`.status.lastUpdated`

// NDBServer is the Schema for the ndbservers API (cluster-scoped).
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

// MongoNodeInfo holds the VM hostname and IP for a single MongoDB HA replica set member.
// IsArbiter is true for arbiter nodes, which are excluded from headless Service and URI creation.
type MongoNodeInfo struct {
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	IsArbiter bool   `json:"isArbiter"`
}

// Database related info to be stored in the status field of the NDB CR
type NDBServerDatabaseInfo struct {
	Name          string `json:"name"`
	Id            string `json:"id"`
	Status        string `json:"status"`
	DBServerId    string `json:"dbServerId"`
	TimeMachineId string `json:"timeMachineId"`
	IPAddress     string `json:"ipAddress"`
	Type          string `json:"type"`
	// MongoNodes holds per-node hostname and IP for MongoDB HA replica set members.
	// Populated only for MongoDB HA databases; nil for all other database types.
	// +optional
	MongoNodes []MongoNodeInfo `json:"mongoNodes,omitempty"`
}
