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
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
)

// Returns a request to delete a snapshot instance
func GenerateTakeSnapshotRequest(snapshot *ndbv1alpha1.Snapshot) (req *SnapshotRequest) {
	req = &SnapshotRequest{
		Name: snapshot.Spec.Name,
		SnapshotLcmConfig: SnapshotLcmConfig{
			SnapshotLCMConfigDetailed: SnapshotLcmConfigDetailed{
				ExpiryDetails: SnapshotExpiryDetails{
					ExpiryDateTimezone: snapshot.Spec.ExpiryDateTimezone,
					ExpireInDays:       snapshot.Spec.ExpireInDays,
				},
			},
		},
	}
	return
}
