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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// LinkedDatabaseSpec defines the desired state of LinkedDatabase.
// +kubebuilder:validation:XValidation:rule="(has(self.sourceDatabaseId) && size(self.sourceDatabaseId) > 0) || (has(self.sourceDatabaseName) && size(self.sourceDatabaseName) > 0)",message="Either sourceDatabaseId or sourceDatabaseName must be provided"
type LinkedDatabaseSpec struct {
	// Name of the cluster-scoped NDBServer resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	NDBRef string `json:"ndbRef"`
	// UUID of the existing NDB database instance to add a linked database to.
	// Either sourceDatabaseId or sourceDatabaseName must be provided.
	// +optional
	// +kubebuilder:validation:MinLength=1
	SourceDatabaseId string `json:"sourceDatabaseId,omitempty"`
	// Name of the existing NDB database instance to add a linked database to.
	// Either sourceDatabaseId or sourceDatabaseName must be provided.
	// +optional
	// +kubebuilder:validation:MinLength=1
	SourceDatabaseName string `json:"sourceDatabaseName,omitempty"`
	// Name of the logical database to create on the existing database instance.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DatabaseName string `json:"databaseName"`
}

// LinkedDatabaseStatus defines the observed state of LinkedDatabase.
type LinkedDatabaseStatus struct {
	Status              string `json:"status,omitempty"`
	SourceDatabaseId    string `json:"sourceDatabaseId,omitempty"`
	LinkedDatabaseId    string `json:"linkedDatabaseId,omitempty"`
	CreationOperationId string `json:"creationOperationId,omitempty"`
	Message             string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName={"ldb","ldbs"}
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Source Database ID",type=string,JSONPath=`.status.sourceDatabaseId`
// +kubebuilder:printcolumn:name="Linked Database ID",type=string,JSONPath=`.status.linkedDatabaseId`
// +kubebuilder:printcolumn:name="Database Name",type=string,JSONPath=`.spec.databaseName`

// LinkedDatabase is the Schema for the linkeddatabases API.
type LinkedDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LinkedDatabaseSpec   `json:"spec,omitempty"`
	Status LinkedDatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LinkedDatabaseList contains a list of LinkedDatabase.
type LinkedDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LinkedDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LinkedDatabase{}, &LinkedDatabaseList{})
}
