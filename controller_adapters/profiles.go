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

package controller_adapters

import (
	"context"
	"fmt"
	"reflect"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Wrapper over api/v1alpha1.Profile
// required to provide implementation of the
// ProfileResolver interface defined in the package ndb_api
type Profile struct {
	v1alpha1.Profile
	ProfileType string
}

func (p *Profile) GetName() (name string) {
	name = p.Name
	return
}

func (p *Profile) GetId() (id string) {
	id = p.Id
	return
}

func (inputProfile *Profile) Resolve(ctx context.Context, allProfiles []ndb_api.ProfileResponse, filter func(p ndb_api.ProfileResponse) bool) (profile ndb_api.ProfileResponse, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered controller_adapters.resolve", "input profile", inputProfile)

	name, id := inputProfile.Name, inputProfile.Id
	isNameProvided, isIdProvided := name != "", id != ""

	var profileByName, profileById ndb_api.ProfileResponse

	// resolve the profile based on provided input, return an error if not resolved
	if isNameProvided {
		profileByName, err = util.FindFirst(allProfiles, func(p ndb_api.ProfileResponse) bool { return p.Type == inputProfile.ProfileType && p.Name == name })
		if err != nil {
			log.Error(err, "could not resolve profile by name", "profile type", inputProfile.ProfileType, "name", name)
			return ndb_api.ProfileResponse{}, fmt.Errorf("could not resolve profile using the provided name=%v", name)
		}
	}

	if isIdProvided {
		profileById, err = util.FindFirst(allProfiles, func(p ndb_api.ProfileResponse) bool { return p.Type == inputProfile.ProfileType && p.Id == id })
		if err != nil {
			log.Error(err, "could not resolve profile by id", "profile type", inputProfile.ProfileType, "id", id)
			return ndb_api.ProfileResponse{}, fmt.Errorf("could not resolve profile using the provided id=%v", id)
		}
	}

	/*
		1. if both name & id not provided => resolve the OOB profile
		2. else if both name & id are provided => resolve by both & ensure that both resolved profiles match
		3. else if only id provided => resolve by id
		4. else if only name provided => resolve by name
		5. else => throw an error
	*/
	if !isNameProvided && !isIdProvided { // OOB
		log.Info("Attempting to resolve the OOB profile, no id or name provided in the spec", "Profile", inputProfile.ProfileType)
		oobProfile, err := util.FindFirst(allProfiles, filter)

		if err != nil {
			log.Error(err, "Error resolving OOB Profile", "type", inputProfile.ProfileType)
			return ndb_api.ProfileResponse{}, fmt.Errorf("no OOB profile found of type=%v", inputProfile.ProfileType)
		}
		return oobProfile, nil

	} else if isNameProvided && isIdProvided { // verify that both resolved profiles (by id and name) are one and the same
		if !reflect.DeepEqual(profileById, profileByName) {
			log.Error(err, "profile matching both the given name & id does not exist. Retry with correct inputs")
			return ndb_api.ProfileResponse{}, fmt.Errorf("profiles returned by id & name resolve to different profiles")
		}
		return profileById, nil

	} else if isIdProvided {
		return profileById, nil

	} else if isNameProvided {
		return profileByName, nil
	}

	return ndb_api.ProfileResponse{}, fmt.Errorf("could not resolve the profile by Name or Id, err=%v", err)
}
