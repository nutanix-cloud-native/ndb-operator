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
	"fmt"
	"reflect"
	"regexp"

	"github.com/nutanix-cloud-native/ndb-operator/api"
	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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

func instanceSpecDefaulterForCreate(instance *Instance) {

	if instance.Description == "" {
		instance.Description = "Database provisioned by ndb-operator: " + instance.DatabaseInstanceName
	}

	if len(instance.DatabaseNames) == 0 {
		instance.DatabaseNames = api.DefaultDatabaseNames
	}

	if instance.TimeZone == "" {
		instance.TimeZone = common.TIMEZONE_UTC
	}

	// initialize Profiles field, if it's nil

	if instance.Profiles == nil {
		databaselog.Info("Initialzing empty Profiles ...")
		instance.Profiles = &(Profiles{})
	}

	// Profiles defaulting logic
	if instance.Profiles.Compute.Id == "" && instance.Profiles.Compute.Name == "" {
		instance.Profiles.Compute.Name = common.PROFILE_DEFAULT_OOB_SMALL_COMPUTE
	}

	// initialize TM field, if it's nil
	if instance.TMInfo == nil {
		databaselog.Info("Initialzing empty TMInfo...")
		instance.TMInfo = &(DBTimeMachineInfo{})
	}

	// TM defaulting logic
	if instance.TMInfo.SnapshotsPerDay == 0 {
		instance.TMInfo.SnapshotsPerDay = 1
	}

	if instance.TMInfo.SLAName == "" {
		instance.TMInfo.SLAName = common.SLA_NAME_NONE
	}

	if instance.TMInfo.DailySnapshotTime == "" {
		instance.TMInfo.DailySnapshotTime = "04:00:00"
	}

	if instance.TMInfo.LogCatchUpFrequency == 0 {
		instance.TMInfo.LogCatchUpFrequency = 30
	}

	if instance.TMInfo.WeeklySnapshotDay == "" {
		instance.TMInfo.WeeklySnapshotDay = "FRIDAY"
	}

	if instance.TMInfo.MonthlySnapshotDay == 0 {
		instance.TMInfo.MonthlySnapshotDay = 15
	}

	if instance.TMInfo.QuarterlySnapshotMonth == "" {
		instance.TMInfo.QuarterlySnapshotMonth = "Jan"
	}

}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
	databaselog.Info("Entering Defaulter logic...")
	instanceSpecDefaulterForCreate(&r.Spec.Instance)
	databaselog.Info("Exiting Defaulter logic...")
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ndb-nutanix-com-v1alpha1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=vdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Database{}

func instanceSpecValidatorForCreate(instance *Instance, allErrs field.ErrorList, instancePath *field.Path) field.ErrorList {
	databaselog.Info("Entering instanceSpecValidatorForCreate...")

	databaselog.Info("Logging the Instance details inside validator method", "databaseInstance", instance)

	// need to assert using a regex
	if instance.DatabaseInstanceName == "" {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("databaseInstanceName"), instance.DatabaseInstanceName, "A valid Database Instance Name must be specified"))
	}

	if instance.ClusterId == "" {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("clusterId"), instance.ClusterId, "ClusterId field must be a valid UUID"))
	}

	if instance.Size < 10 {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("size"), instance.Size, "Initial Database size must be specified with a value 10 GBs or more"))
	}

	if instance.CredentialSecret == "" {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("credentialSecret"), instance.CredentialSecret, "CredentialSecret must be provided in the Instance Spec"))
	}

	if _, isPresent := api.AllowedDatabaseTypes[instance.Type]; !isPresent {
		allErrs = append(allErrs, field.Invalid(instancePath.Child("type"), instance.Type,
			fmt.Sprintf("A valid database type must be specified. Valid values are: %s", reflect.ValueOf(api.AllowedDatabaseTypes).MapKeys()),
		))
	}

	if _, isPresent := api.ClosedSourceDatabaseTypes[instance.Type]; isPresent {
		if instance.Profiles == &(Profiles{}) || instance.Profiles.Software == (Profile{}) {
			allErrs = append(allErrs, field.Invalid(instancePath.Child("profiles").Child("software"), instance.Profiles.Software, "Software Profile must be provided for the closed-source database engines"))
		}
	}

	// validating time machine info
	tmPath := instancePath.Child("timeMachine")

	dailySnapshotTimeRegex := regexp.MustCompile(`^(2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9]$`)
	if isMatch := dailySnapshotTimeRegex.MatchString(instance.TMInfo.DailySnapshotTime); !isMatch {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("dailySnapshotTime"), instance.TMInfo.DailySnapshotTime, "Invalid time format for the daily snapshot time. Use the 24-hour format (HH:MM:SS)."))
	}

	if instance.TMInfo.SnapshotsPerDay < 1 || instance.TMInfo.SnapshotsPerDay > 6 {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("snapshotsPerDay"), instance.TMInfo.SnapshotsPerDay, "Number of snapshots per day should be within 1 to 6"))
	}

	if _, isPresent := api.AllowedLogCatchupFrequencyInMinutes[instance.TMInfo.LogCatchUpFrequency]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("logCatchUpFrequency"), instance.TMInfo.LogCatchUpFrequency,
			fmt.Sprintf("Log catchup frequency must be specified. Valid values are: %s", reflect.ValueOf(api.AllowedLogCatchupFrequencyInMinutes).MapKeys()),
		))
	}

	if _, isPresent := api.AllowedWeeklySnapshotDays[instance.TMInfo.WeeklySnapshotDay]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("weeklySnapshotDay"), instance.TMInfo.WeeklySnapshotDay,
			fmt.Sprintf("Weekly Snapshot day must be specified. Valid values are: %s", reflect.ValueOf(api.AllowedWeeklySnapshotDays).MapKeys()),
		))
	}

	if instance.TMInfo.MonthlySnapshotDay < 1 || instance.TMInfo.MonthlySnapshotDay > 28 {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("monthlySnapshotDay"), instance.TMInfo.MonthlySnapshotDay, "Monthly snapshot day value must be between 1 and 28"))
	}

	if _, isPresent := api.AllowedQuarterlySnapshotMonths[instance.TMInfo.QuarterlySnapshotMonth]; !isPresent {
		allErrs = append(allErrs, field.Invalid(tmPath.Child("quarterlySnapshotMonth"), instance.TMInfo.QuarterlySnapshotMonth,
			fmt.Sprintf("Quarterly snapshot month must be specified. Valid values are: %s", reflect.ValueOf(api.AllowedQuarterlySnapshotMonths).MapKeys()),
		))
	}

	databaselog.Info("Exiting instanceSpecValidatorForCreate...")
	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() (admission.Warnings, error) {
	databaselog.Info("Entering ValidateCreate...")

	// ndbSpecErrors := ndbServerSpecValidatorForCreate(&r.Spec.NDB, field.ErrorList{}, field.NewPath("spec").Child("ndb"))
	dbSpecErrors := instanceSpecValidatorForCreate(&r.Spec.Instance, field.ErrorList{}, field.NewPath("spec").Child("databaseInstance"))

	// allErrs := append(ndbSpecErrors, dbSpecErrors...)
	allErrs := dbSpecErrors

	combined_err := util.CombineFieldErrors(allErrs)

	databaselog.Info("validate create database webhook response...", "combined_err", combined_err)

	databaselog.Info("Exiting ValidateCreate...")
	return nil, combined_err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	databaselog.Info("validate update", "name", r.Name)

	// TODO: This method will be used to make fields immutable.
	// Here you can reject the updates to any fields. I think we should mark everything immutable by default.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() (admission.Warnings, error) {
	databaselog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
