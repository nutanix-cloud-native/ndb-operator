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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_api"
)

var (
	MONTH_MAP = map[string]string{
		"":    "JANUARY", // For default case when no input is provided
		"Jan": "JANUARY",
		"Feb": "FEBRUARY",
		"Mar": "MARCH",
	}
)

// Wrapper over api/v1alpha1.Database
// required to provide implementation of the
// DatabaseInterface defined in the package ndb_api
type Database struct {
	v1alpha1.Database
}

func (d *Database) IsClone() bool {
	return d.Spec.IsClone
}

func (d *Database) GetName() string {
	if d.IsClone() {
		return d.Spec.Clone.Name
	}
	return d.Spec.Instance.Name
}

// Returns database description. If description is empty, creates a description
func (d *Database) GetDescription() string {
	var description string
	if d.IsClone() {
		description = d.Spec.Clone.Description
	} else {
		description = d.Spec.Instance.Description
	}
	if description == "" {
		return "Created by ndb-operator: " + d.GetName()
	}
	return description
}

func (d *Database) GetClusterId() string {
	if d.IsClone() {
		return d.Spec.Clone.ClusterId
	}
	return d.Spec.Instance.ClusterId
}

func (d *Database) GetProfileResolvers() ndb_api.ProfileResolvers {
	profileResolvers := make(ndb_api.ProfileResolvers)

	var profiles *v1alpha1.Profiles
	if d.IsClone() {
		profiles = d.Spec.Clone.Profiles
	} else {
		profiles = d.Spec.Instance.Profiles
	}

	profileResolvers[common.PROFILE_TYPE_COMPUTE] = &Profile{
		Profile:     profiles.Compute,
		ProfileType: common.PROFILE_TYPE_COMPUTE,
	}
	profileResolvers[common.PROFILE_TYPE_SOFTWARE] = &Profile{
		Profile:     profiles.Software,
		ProfileType: common.PROFILE_TYPE_SOFTWARE,
	}
	profileResolvers[common.PROFILE_TYPE_NETWORK] = &Profile{
		Profile:     profiles.Network,
		ProfileType: common.PROFILE_TYPE_NETWORK,
	}
	profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER] = &Profile{
		Profile:     profiles.DbParam,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}
	profileResolvers[common.PROFILE_TYPE_DATABASE_PARAMETER_INSTANCE] = &Profile{
		Profile:     profiles.DbParamInstance,
		ProfileType: common.PROFILE_TYPE_DATABASE_PARAMETER,
	}

	return profileResolvers
}

func (d *Database) GetCredentialSecret() string {
	if d.IsClone() {
		return d.Spec.Clone.CredentialSecret
	}
	return d.Spec.Instance.CredentialSecret
}

func (d *Database) GetTimeZone() string {
	if d.IsClone() {
		return d.Spec.Clone.TimeZone
	}
	return d.Spec.Instance.TimeZone
}

func (d *Database) GetAdditionalArguments() map[string]string {
	if d.IsClone() {
		return d.Spec.Clone.AdditionalArguments
	}
	return d.Spec.Instance.AdditionalArguments
}

func (d *Database) GetInstanceType() string {
	if d.IsClone() {
		return d.Spec.Clone.Type
	}
	return d.Spec.Instance.Type
}

func (d *Database) GetInstanceDatabaseNames() string {
	return strings.Join(d.Spec.Instance.DatabaseNames, ",")
}

func (d *Database) GetInstanceSize() int {
	return d.Spec.Instance.Size
}

func (d *Database) GetInstanceIsHighAvailability() bool {
	if d.IsClone() {
		// Clone doesn't currently support HA
		return false
	}
	return d.Spec.Instance.IsHighAvailability
}

func (d *Database) GetInstanceNodes() []*v1alpha1.Node {
	if d.IsClone() {
		return d.Spec.Instance.Nodes
	}
	return d.Spec.Instance.Nodes
}

// Returns basic details about the Time Machine if provided in the
// underlying database, else returns defaults like:
// TM Name: <db_instance_name>_TM
// TM Description: Time Machine for <db_instance_name>
// SLA Name: NONE
func (d *Database) GetInstanceTMDetails() (tmName, tmDescription, slaName string) {
	tmInfo := d.Spec.Instance.TMInfo

	tmName = tmInfo.Name
	tmDescription = tmInfo.Description
	slaName = tmInfo.SLAName

	if tmName == "" {
		tmName = d.GetName() + "_TM"
	}
	if tmDescription == "" {
		tmDescription = "Time Machine for " + d.GetName()
	}
	if slaName == "" {
		slaName = common.SLA_NAME_NONE
	}

	return
}

// Returns a schedule struct for the time machine.
func (d *Database) GetTMScheduleForInstance() (schedule ndb_api.Schedule, err error) {
	tmInfo := d.Spec.Instance.TMInfo

	hhmmss, err := time.Parse(time.TimeOnly, tmInfo.DailySnapshotTime)
	if err != nil {
		err = errors.Join(err, errors.New("error converting daily snapshot time"))
		return
	}
	hh, mm, ss := hhmmss.Hour(), hhmmss.Minute(), hhmmss.Second()

	quarterlySnapshotStartMonth, ok := MONTH_MAP[tmInfo.QuarterlySnapshotMonth]
	if !ok {
		err = fmt.Errorf("month %s not allowed for QuarterlySnapshotMonth", tmInfo.QuarterlySnapshotMonth)
		return
	}

	schedule = ndb_api.Schedule{
		SnapshotTimeOfDay: ndb_api.SnapshotTimeOfDay{
			Hours:   hh,
			Minutes: mm,
			Seconds: ss,
		},

		ContinuousSchedule: ndb_api.ContinuousSchedule{
			Enabled:           true,
			LogBackupInterval: tmInfo.LogCatchUpFrequency,
			SnapshotsPerDay:   tmInfo.SnapshotsPerDay,
		},

		WeeklySchedule: ndb_api.WeeklySchedule{
			Enabled:   true,
			DayOfWeek: tmInfo.WeeklySnapshotDay,
		},

		MonthlySchedule: ndb_api.MonthlySchedule{
			Enabled:    true,
			DayOfMonth: tmInfo.MonthlySnapshotDay,
		},

		QuarterlySchedule: ndb_api.QuarterlySchedule{
			Enabled:    true,
			StartMonth: quarterlySnapshotStartMonth,
			DayOfMonth: tmInfo.MonthlySnapshotDay,
		},

		YearlySchedule: ndb_api.YearlySchedule{
			Enabled:    false,
			DayOfMonth: 31,
			Month:      "DECEMBER",
		},
	}
	return
}

func (d *Database) GetCloneSourceDBId() string {
	return d.Spec.Clone.SourceDatabaseId
}

func (d *Database) GetCloneSnapshotId() string {
	return d.Spec.Clone.SnapshotId
}
