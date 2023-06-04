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
	"fmt"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// Fetches all the SLAs from the ndb and returns the NONE TM SLA.
// Returns an error if not found.
func GetNoneTimeMachineSLA(ctx context.Context, ndb_client *ndb_client.NDBClient) (sla SLAResponse, err error) {
	slas, err := GetAllSLAs(ctx, ndb_client)
	if err != nil {
		return
	}
	for _, s := range slas {
		if s.Name == common.SLA_NAME_NONE {
			sla = s
			return
		}
	}
	return sla, fmt.Errorf("NONE TimeMachine not found")
}

// Fetches all the SLAs from the ndb and returns the SLA matching the name
// Returns an error if not found.
func GetSLAByName(ctx context.Context, ndb_client *ndb_client.NDBClient, name string) (sla SLAResponse, err error) {
	slas, err := GetAllSLAs(ctx, ndb_client)
	if err != nil {
		return
	}
	sla, err = util.FindFirst(slas, func(s SLAResponse) bool { return s.Name == name })
	if err != nil {
		err = fmt.Errorf("SLA %s not found", name)
	}
	return
}
