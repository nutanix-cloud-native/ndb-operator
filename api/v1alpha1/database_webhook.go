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
	"regexp"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/api"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var databaselog = logf.Log.WithName("database-resource")

func (r *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ndb-nutanix-com-v1alpha1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=mdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Database{}

func instanceSpecDefaulterForCreate(r *Database) {

	if r.Spec.Instance.DatabaseInstanceName == "" {
		r.Spec.Instance.DatabaseInstanceName = "database_" + time.Now().Format("2006-Jan-02_03-04-05PM")
	}

	if len(r.Spec.Instance.DatabaseNames) == 0 {
		r.Spec.Instance.DatabaseNames = api.DefaultDatabaseNames
	}

	if r.Spec.Instance.TimeZone == "" {
		r.Spec.Instance.TimeZone = "UTC"
	}

	if r.Spec.Instance.Profiles.Compute.Id == "" && r.Spec.Instance.Profiles.Compute.Name == "" {
		r.Spec.Instance.Profiles.Compute.Name = common.PROFILE_DEFAULT_OOB_SMALL_COMPUTE
	}

	// time machine defaulter logic

	if r.Spec.Instance.TMInfo.Name == "" {
		r.Spec.Instance.TMInfo.Name = r.Spec.Instance.DatabaseInstanceName + "_TM"
	}

	if r.Spec.Instance.TMInfo.Description == "" {
		r.Spec.Instance.TMInfo.Name = "Time Machine For " + r.Spec.Instance.DatabaseInstanceName
	}

	if r.Spec.Instance.TMInfo.SnapshotsPerDay == 0 {
		r.Spec.Instance.TMInfo.SnapshotsPerDay = 1
	}

	if r.Spec.Instance.TMInfo.SLAName == "" {
		r.Spec.Instance.TMInfo.SLAName = common.SLA_NAME_NONE
	}

	if r.Spec.Instance.TMInfo.SLAName == "" {
		r.Spec.Instance.TMInfo.DailySnapshotTime = "04:00:00"
	}

	if r.Spec.Instance.TMInfo.LogCatchUpFrequency == 0 {
		r.Spec.Instance.TMInfo.LogCatchUpFrequency = 30
	}

	if r.Spec.Instance.TMInfo.WeeklySnapshotDay == "" {
		r.Spec.Instance.TMInfo.WeeklySnapshotDay = "FRIDAY"
	}

	if r.Spec.Instance.TMInfo.MonthlySnapshotDay == 0 {
		r.Spec.Instance.TMInfo.MonthlySnapshotDay = 15
	}

	if r.Spec.Instance.TMInfo.QuarterlySnapshotMonth == "" {
		r.Spec.Instance.TMInfo.QuarterlySnapshotMonth = "Jan"
	}

}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
	databaselog.Info("Entering Defaulter logic...")
	instanceSpecDefaulterForCreate(r)
	databaselog.Info("Exiting Defaulter logic...")
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ndb-nutanix-com-v1alpha1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=vdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Database{}

func ndbServerSpecValidatorForCreate(r *Database, allErrs field.ErrorList, ndbPath *field.Path) field.ErrorList {
	databaselog.Info("Entering ndbServerSpecValidatorForCreate...")

	// need to fix this and skipCertificateVerification after struct to pointer change
	if r.Spec.NDB == (NDB{}) {
		allErrs = append(allErrs, field.Invalid(ndbPath, r.Spec.NDB, "NDB spec field must not be empty"))
	}

	if err := util.ValidateUUID(r.Spec.NDB.ClusterId); err != nil {
		allErrs = append(allErrs, field.Invalid(ndbPath.Child("clusterId"), r.Spec.NDB.ClusterId, "ClusterId field must be a valid UUID"))
	}

	if r.Spec.NDB.CredentialSecret == "" {
		allErrs = append(allErrs, field.Invalid(ndbPath.Child("credentialSecret"), r.Spec.NDB.CredentialSecret, "CredentialSecret must not be empty"))
	}

	if err := util.ValidateURL(r.Spec.NDB.Server); err != nil {
		allErrs = append(allErrs, field.Invalid(ndbPath.Child("server"), r.Spec.NDB.Server, "Server must be a valid URL"))
	}

	databaselog.Info("Exiting ndbServerSpecValidatorForCreate...")
	return allErrs
}

func instanceSpecValidatorForCreate(r *Database, allErrs field.ErrorList, instancePath *field.Path) field.ErrorList {
	databaselog.Info("Entering instanceSpecValidatorForCreate...")

	if r.Spec.Instance.Size < 10 {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("size"), r.Spec.Instance.Size, "Initial Database size must be 10 GBs or more"))
	}

	if r.Spec.Instance.CredentialSecret == "" {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("credentialSecret"), r.Spec.Instance.CredentialSecret, "CredentialSecret must be provided"))
	}

	if _, isPresent := api.AllowedDatabaseTypes[r.Spec.Instance.Type]; !isPresent {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("type"), r.Spec.Instance.CredentialSecret, "A valid database type must be specified"))
	}

	if _, isPresent := api.ClosedSourceDatabaseTypes[r.Spec.Instance.Type]; isPresent {
		if r.Spec.Instance.Profiles == (Profiles{}) || r.Spec.Instance.Profiles.Software == (Profile{}) {
			allErrs = append(allErrs, field.Invalid(instancePath.Child("type"), r.Spec.Instance.CredentialSecret, "Software Profile must be provided for the closed-source database engines"))
		}
	}

	// validating time machine info
	tmPath := instancePath.Child("timeMachine")

	dailySnapshotTimeRegex := regexp.MustCompile(`^(2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9]$`)
	if isMatch := dailySnapshotTimeRegex.MatchString(r.Spec.Instance.TMInfo.DailySnapshotTime); !isMatch {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("dailySnapshotTime"), r.Spec.Instance.TMInfo.DailySnapshotTime, "Invalid time format for the daily snapshot time. Use the 24-hour format (HH:MM:SS)."))
	}

	if r.Spec.Instance.TMInfo.SnapshotsPerDay < 1 || r.Spec.Instance.TMInfo.SnapshotsPerDay > 6 {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("snapshotsPerDay"), r.Spec.Instance.TMInfo.SnapshotsPerDay, "Number of snapshots per day should be within 1 to 6"))
	}

	if _, isPresent := api.AllowedLogCatchupIntervals[r.Spec.Instance.TMInfo.LogCatchUpFrequency]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("logCatchUpFrequency"), r.Spec.Instance.TMInfo.LogCatchUpFrequency, "Log catchup frequency must have one of these values: {15, 30, 45, 60, 90, 120}"))
	}

	if _, isPresent := api.AllowedWeeklySnapshotDays[r.Spec.Instance.TMInfo.WeeklySnapshotDay]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("weeklySnapshotDay"), r.Spec.Instance.TMInfo.WeeklySnapshotDay, "Weekly snapshot day must have a valid value e.g. MONDAY"))
	}

	if r.Spec.Instance.TMInfo.MonthlySnapshotDay < 1 || r.Spec.Instance.TMInfo.MonthlySnapshotDay > 28 {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("monthlySnapshotDay"), r.Spec.Instance.TMInfo.MonthlySnapshotDay, "Monthly snapshot day value must be between 1 and 28"))
	}

	if _, isPresent := api.AllowedQuarterlySnapshotMonths[r.Spec.Instance.TMInfo.QuarterlySnapshotMonth]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("quarterlySnapshotMonth"), r.Spec.Instance.TMInfo.QuarterlySnapshotMonth, "Quarterly snapshot month must be one of {Jan, Feb, Mar}"))
	}

	databaselog.Info("Exiting instanceSpecValidatorForCreate...")
	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() error {
	databaselog.Info("Entering ValidateCreate...")

	ndbSpecErrors := ndbServerSpecValidatorForCreate(r, field.ErrorList{}, field.NewPath("spec").Child("ndb"))
	dbSpecErrors := instanceSpecValidatorForCreate(r, field.ErrorList{}, field.NewPath("spec").Child("instance"))

	allErrs := append(ndbSpecErrors, dbSpecErrors...)

	combined_err := util.CombineFieldErrors(allErrs)
	databaselog.Info("validate create database webhook response...", "combined_err", combined_err)

	databaselog.Info("Exiting ValidateCreate...")
	return combined_err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) error {
	databaselog.Info("validate update", "name", r.Name)

	// TODO: This method will be used to make fields immutable.
	// Here you can reject the updates to any fields. I think we should mark everything immutable by default.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() error {
	databaselog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
