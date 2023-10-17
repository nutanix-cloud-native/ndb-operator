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
)

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
	GetInstanceTMSchedule() (Schedule, error)
	GetCloneSourceDBId() string
	GetCloneSnapshotId() string
	GetAdditionalArguments() map[string]string
}
