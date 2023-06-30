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

func validateDatabaseCreate_NDBSpec(r *Database, allErrs field.ErrorList) field.ErrorList {
	databaselog.Info("validate validateDatabaseCreate_NDBSpec...")
	if r.Spec.NDB == (NDB{}) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("ndb"), r.Spec.NDB, "NDB field must not be null"))
	}

	if err := util.ValidateUUID(r.Spec.NDB.ClusterId); err != nil {
		databaselog.Info("ClusterId Validation", "error", err)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("ndb").Child("clusterId"), r.Spec.NDB.ClusterId, "ClusterId field must be a valid UUID"))
	}

	if r.Spec.NDB.CredentialSecret == "" {
		databaselog.Info("CredentialSecret must not be empty", "Credential Secret", r.Spec.NDB.CredentialSecret)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("ndb").Child("credentialSecret"), r.Spec.NDB.CredentialSecret, "CredentialSecret must not be empty"))
	}

	if err := util.ValidateURL(r.Spec.NDB.Server); err != nil {
		databaselog.Info("server must be a valid URL", "Server URL", r.Spec.NDB.Server)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("ndb").Child("server"), r.Spec.NDB.Server, "server must be a valid URL"))
	}

	databaselog.Info("allErrs", allErrs)

	return allErrs
}

func validateDatabaseCreate_NewDBSpec(r *Database, allErrs field.ErrorList) field.ErrorList {
	databaselog.Info("validate validateDatabaseCreate_NewDBSpec")

	if r.Spec.NDB.CredentialSecret == "" {
		databaselog.Info("credentialSecret must not be empty", "error", r.Spec.NDB.CredentialSecret)
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("instance").Child("credentialSecret"), r.Spec.NDB, "CredentialSecret must not be empty"))
	}

	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() error {
	databaselog.Info("validate create database", "name", r.Name)

	ndbSpecErrors := validateDatabaseCreate_NDBSpec(r, field.ErrorList{})
	dbSpecErrors := validateDatabaseCreate_NewDBSpec(r, field.ErrorList{})

	allErrs := append(ndbSpecErrors, dbSpecErrors...)

	databaselog.Info("Errors returned from the webhook", "errors", allErrs)

	return util.CombineFieldErrors(allErrs)
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
