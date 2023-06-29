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
	"errors"
	"strings"

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

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-ndb-nutanix-com-v1alpha1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=mdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Database{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
	databaselog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-ndb-nutanix-com-v1alpha1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=vdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Database{}

func validateDatabaseCreateNDBSpec(r *Database, allErrs field.ErrorList) field.ErrorList {
	if r.Spec.NDB == (NDB{}) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec").Child("ndb"), r.Spec.NDB, "NDB field must not be null"))
	}

	err := util.IsValidUUID(r.Spec.NDB.ClusterId)

	if err != nil {
		databaselog.Error(err, "database validation error", "error", err)
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec").Child("ndb").Child("clusterId"), r.Spec.NDB, "clusterId field must be a valid UUID"))
	}

	return allErrs
}

func CombineFieldErrors(fieldErrors field.ErrorList) error {

	if len(fieldErrors) == 0 {
		return nil
	}

	var errorStrings []string
	for _, fe := range fieldErrors {
		errorStrings = append(errorStrings, fe.Error())
	}
	return errors.New(strings.Join(errorStrings, "; "))
}

func validateDatabaseCreateNewDatabaseSpec(r *Database, allErrs field.ErrorList) field.ErrorList {
	if err := util.IsValidUUID(r.Spec.NDB.ClusterId); err != nil {
		databaselog.Error(err, "database validation error", "error", err)
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spc").Child("instance").Child("credentialSecret"), r.Spec.NDB, "CredentialSecret field must not be null"))
	}

	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() error {
	databaselog.Info("validate create database", "name", r.Name)
	allErrs := field.ErrorList{}

	validateDatabaseCreateNDBSpec(r, allErrs)

	databaselog.Info("fix field errors from validateDatabaseCreateNDBSpec ", "allErrs", allErrs)
	validateDatabaseCreateNewDatabaseSpec(r, allErrs)

	databaselog.Info("fix field errors from validateDatabaseCreateNewDatabaseSpec", "allErrs", allErrs)
	return CombineFieldErrors(allErrs)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) error {
	databaselog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() error {
	databaselog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
