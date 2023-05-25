// /*
// Copyright 2022-2023 Nutanix, Inc.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package ndb_api

import (
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/common"
)

// import (
// 	"context"
// 	"fmt"
// 	"reflect"
// 	"strings"

// 	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
// 	"github.com/nutanix-cloud-native/ndb-operator/common"
// 	"github.com/nutanix-cloud-native/ndb-operator/common/util"
// 	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
// )

// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //----------------------ProfileResolver--------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------

// // Profile type implements the ProfileResolver interface
// // type Profile struct {
// // 	v1alpha1.Profile
// // 	ProfileType string
// // }

// // func (p *Profile) GetName() (name string) {
// // 	name = p.Name
// // 	return
// // }

// // func (p *Profile) GetId() (id string) {
// // 	id = p.Id
// // 	return
// // }

// // func (inputProfile *Profile) Resolve(ctx context.Context, allProfiles []ProfileResponse, filter func(p ProfileResponse) bool) (profile ProfileResponse, err error) {
// // 	log := ctrllog.FromContext(ctx)
// // 	log.Info("Entered ndb_api_helpers.resolve", "input profile", inputProfile)

// // 	name, id := inputProfile.Name, inputProfile.Id
// // 	isNameProvided, isIdProvided := name != "", id != ""

// // 	var profileByName, profileById ProfileResponse

// // 	// resolve the profile based on provided input, return an error if not resolved
// // 	if isNameProvided {
// // 		profileByName, err = util.FindFirst(allProfiles, func(p ProfileResponse) bool { return p.Name == name })
// // 	}

// // 	if isIdProvided && err == nil {
// // 		profileById, err = util.FindFirst(allProfiles, func(p ProfileResponse) bool { return p.Id == id })
// // 	}

// // 	if err != nil {
// // 		log.Error(err, "could not resolve profile by id or name", "profile type", inputProfile.ProfileType, "id", id, "name", name)
// // 		return ProfileResponse{}, fmt.Errorf("could not resolve profile by id=%v or name=%v", id, name)
// // 	}

// // 	/*
// // 		1. if both name & id not provided => resolve the OOB profile
// // 		2. else if both name & id are provided => resolve by both & ensure that both resolved profiles are match
// // 		3. else if only id provided => resolve by id
// // 		4. else if only name provided => resolve by name
// // 		5. else => throw an error
// // 	*/
// // 	if !isNameProvided && !isIdProvided { // OOB
// // 		log.Info("Attempting to resolve the OOB profile, no id or name provided in the spec", "Profile", inputProfile.ProfileType)
// // 		oobProfile, err := util.FindFirst(allProfiles, filter)

// // 		if err != nil {
// // 			log.Error(err, "Error resolving OOB Profile", "type", inputProfile.ProfileType)
// // 			return ProfileResponse{}, fmt.Errorf("no OOB profile found of type=%v", inputProfile.ProfileType)
// // 		}
// // 		return oobProfile, nil

// // 	} else if isNameProvided && isIdProvided { // verify that both resolved profiles (by id and name) are one and the same
// // 		if !reflect.DeepEqual(profileById, profileByName) {
// // 			log.Error(err, "profile matching both the given name & id does not exist. Retry with correct inputs")
// // 			return ProfileResponse{}, fmt.Errorf("profiles returned by id & name resolve to different profiles")
// // 		}
// // 		return profileById, nil

// // 	} else if isIdProvided {
// // 		return profileById, nil

// // 	} else if isNameProvided {
// // 		return profileByName, nil
// // 	}

// // 	return ProfileResponse{}, fmt.Errorf("could not resolve the profile by Name or Id, err=%v", err)
// // }

var ComputeOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_COMPUTE && p.SystemProfile &&
		strings.EqualFold(p.Name, common.PROFILE_DEFAULT_OOB_SMALL_COMPUTE)
}

var SoftwareOOBProfileResolverForSingleInstance = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_SOFTWARE && p.SystemProfile && p.Topology == common.TOPOLOGY_SINGLE
}

var NetworkOOBProfileResolver = func(p ProfileResponse) bool {
	return p.Type == common.PROFILE_TYPE_NETWORK
}

var DbParamOOBProfileResolver = func(p ProfileResponse) bool {
	return p.SystemProfile && p.Type == common.PROFILE_TYPE_DATABASE_PARAMETER
}

// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //--------------------DatabaseActionArgs-------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------
// //---------------------------------------------------------------

// // MysqlActionArgs implements the DatabaseActionArgs interface
// type MysqlActionArgs struct{}

// func (m *MysqlActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ActionArgument {
// 	return []ActionArgument{
// 		{
// 			Name:  "listener_port",
// 			Value: "3306",
// 		},
// 	}
// }

// // PostgresActionArgs implements the DatabaseActionArgs interface
// type PostgresActionArgs struct{}

// func (p *PostgresActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ActionArgument {
// 	return []ActionArgument{
// 		{
// 			Name:  "proxy_read_port",
// 			Value: "5001",
// 		},
// 		{
// 			Name:  "listener_port",
// 			Value: "5432",
// 		},
// 		{
// 			Name:  "proxy_write_port",
// 			Value: "5000",
// 		},
// 		{
// 			Name:  "enable_synchronous_mode",
// 			Value: "false",
// 		},
// 		{
// 			Name:  "auto_tune_staging_drive",
// 			Value: "true",
// 		},
// 		{
// 			Name:  "backup_policy",
// 			Value: "primary_only",
// 		},
// 	}
// }

// // MongodbActionArgs implements the DatabaseActionArgs interface
// type MongodbActionArgs struct{}

// func (m *MongodbActionArgs) Get(dbSpec v1alpha1.DatabaseSpec) []ActionArgument {
// 	return []ActionArgument{
// 		{
// 			Name:  "listener_port",
// 			Value: "27017",
// 		},
// 		{
// 			Name:  "log_size",
// 			Value: "100",
// 		},
// 		{
// 			Name:  "journal_size",
// 			Value: "100",
// 		},
// 		{
// 			Name:  "restart_mongod",
// 			Value: "true",
// 		},
// 		{
// 			Name:  "working_dir",
// 			Value: "/tmp",
// 		},
// 		{
// 			Name:  "db_user",
// 			Value: "admin",
// 		},
// 		{
// 			Name:  "backup_policy",
// 			Value: "primary_only",
// 		},
// 	}
// }
