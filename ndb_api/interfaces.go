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

package ndb_api

import (
	"context"
	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
)

// External Interfaces
// Used by other packages that make use of the ndb_api package
// Implementations defined in the packages that use this package
// For example - controller_adapters

type ProfileResolver interface {
	Resolve(ctx context.Context, allProfiles []ProfileResponse, filter func(p ProfileResponse) bool) (profile ProfileResponse, err error)
	GetName() string
	GetId() string
}

type ProfileResolvers map[string]ProfileResolver

type DatabaseInterface interface {
	IsClone() bool
	GetName() string
	GetDescription() string
	GetClusterId() string
	GetProfileResolvers() ProfileResolvers
	GetCredentialSecret() string
	GetTimeZone() string
	GetInstanceType() string
	GetInstanceDatabaseNames() string
	GetInstanceSize() int
	GetInstanceTMDetails() (string, string, string)
	GetTMScheduleForInstance() (Schedule, error)
	GetCloneSourceDBId() string
	GetCloneSnapshotId() string
	GetAdditionalArguments() map[string]string
	GetInstanceIsHighAvailability() bool
	GetInstanceNodes() []*v1alpha1.Node
}

// Internal Interfaces
// Used internally within the ndb_api package

type RequestAppender interface {
	// Function to add additional arguments to the Provisioning request
	appendProvisioningRequest(req *DatabaseProvisionRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseProvisionRequest, error)
	// Function to add additional arguments to the Cloning request
	appendCloningRequest(req *DatabaseCloneRequest, database DatabaseInterface, reqData map[string]interface{}) (*DatabaseCloneRequest, error)
}

// Implements RequestAppender
type MSSQLRequestAppender struct{}

// Implements RequestAppender
type MongoDbRequestAppender struct{}

// Implements RequestAppender
type PostgresRequestAppender struct{}

// Implements RequestAppender
type MySqlRequestAppender struct{}

// Implements RequestAppender
type PostgresHARequestAppender struct{}
