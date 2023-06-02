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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DataProtectionSpec defines the desired state of DataProtection
type DataProtectionSpec struct {
	// +kubebuilder:validation:Enum=restore
	Type     string  `json:"type,omitempty"`
	Database string  `json:"database,omitempty"` // represents Database CR Name
	Name     string  `json:"name,omitempty"`
	Restore  Restore `json:"restore,omitempty"` // represents Restore config
}

type Restore struct {
	SnapshotId string `json:"snapshotId,omitempty"`
	Name       string `json:"name,omitempty"`
	Database   string `json:"database,omitempty"`
	Status     string `json:"status,omitempty"`
}

// DataProtectionStatus defines the observed state of DataProtection
type DataProtectionStatus struct {
	Name        string `json:"name,omitempty"`
	Database    string `json:"database,omitempty"`
	OperationId string `json:"operationId"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DataProtection is the Schema for the dataprotections API
type DataProtection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataProtectionSpec   `json:"spec,omitempty"`
	Status DataProtectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataProtectionList contains a list of DataProtection
type DataProtectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataProtection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataProtection{}, &DataProtectionList{})
}
