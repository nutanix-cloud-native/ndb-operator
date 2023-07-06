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

/*
GENERATED by operator-sdk
Changes added
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	NDB      NDB      `json:"ndb"`
	Instance Instance `json:"databaseInstance"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	IPAddress        string `json:"ipAddress"`
	Id               string `json:"id"`
	Status           string `json:"status"`
	DatabaseServerId string `json:"dbServerId"`
	Type             string `json:"type"`
}

// Database is the Schema for the databases API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName={"db","dbs"}
// +kubebuilder:printcolumn:name="IP Address",type=string,JSONPath=`.status.ipAddress`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.type`
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}

// These are required to have a deep copy, object interface implementation
// These are the structs for the Spec and Status

// Details of the NDB installation
type NDB struct {
	ClusterId string `json:"clusterId"`
	// Name of the secret holding the credentials for NDB (username and password)
	CredentialSecret string `json:"credentialSecret"`
	Server           string `json:"server"`
	// Skip server's certificate and hostname verification, default false
	SkipCertificateVerification bool `json:"skipCertificateVerification"`
}

// Database instance specific details
type Instance struct {
	// Name of the database instance, default "database_instance_name"
	DatabaseInstanceName string `json:"databaseInstanceName"`
	// Name(s) of the database(s) to be provisiond inside the database instance
	// default [ "database_one", "database_two", "database_three" ]
	DatabaseNames []string `json:"databaseNames"`
	// Name of the secret holding the credentials for the database instance (password and ssh key)
	CredentialSecret string `json:"credentialSecret"`
	// Size of the database instance, default 10, minimum 10
	Size int `json:"size"`
	// default UTC
	TimeZone string   `json:"timezone"`
	Type     string   `json:"type"`
	Profiles Profiles `json:"profiles"`
}

type Profiles struct {
	Software        Profile `json:"software"`
	Compute         Profile `json:"compute"`
	Network         Profile `json:"network"`
	DbParam         Profile `json:"dbParam"`
	DbParamInstance Profile `json:"dbParamInstance"`
}

type Profile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
