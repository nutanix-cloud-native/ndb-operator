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
	"strings"

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

// +kubebuilder:webhook:path=/mutate-ndb-nutanix-com-v1alpha1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=mdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Database{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
	databaselog.Info("Entering Default()...")

	getDatabaseWebhookHandler(r).defaulter(&r.Spec)

	databaselog.Info("Exiting Default()!")
}

// +kubebuilder:webhook:path=/validate-ndb-nutanix-com-v1alpha1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=vdatabase.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Database{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() (admission.Warnings, error) {
	databaselog.Info("Entering ValidateCreate...")

	errors := &field.ErrorList{}
	var path string
	if r.Spec.IsClone {
		path = "Clone"
	} else {
		path = "Instance"
	}

	getDatabaseWebhookHandler(r).validateCreate(&r.Spec, errors, field.NewPath("spec").Child(path))

	combined_err := util.CombineFieldErrors(*errors)

	databaselog.Info("ValidateCreate webhook response...", "combined_err", combined_err)

	databaselog.Info("Exiting ValidateCreate!")

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

/* Checks if configured additional arguments are valid or not and returns the corresponding additional arguments. If error is nil valid, else invalid */
func additionalArgumentsValidationCheck(isClone bool, dbType string, isHA bool, specifiedAdditionalArguments map[string]string) error {
	// Empty additionalArguments is always valid
	if specifiedAdditionalArguments == nil {
		return nil
	}

	allowedAdditionalArguments, err := util.GetAllowedAdditionalArguments(isClone, dbType, isHA)

	// Invalid type returns error
	if err != nil {
		return err
	}

	// Checking if arguments are valid
	invalidArgs := []string{}
	for name := range specifiedAdditionalArguments {
		if _, isPresent := allowedAdditionalArguments[name]; !isPresent {
			invalidArgs = append(invalidArgs, name)
		}
	}

	if len(invalidArgs) == 0 {
		return nil
	} else {
		return fmt.Errorf(
			"additional arguments validation for type: %s failed! The following args are invalid: %s. These are the allowed args: %s",
			dbType,
			strings.Join(invalidArgs, ", "),
			reflect.ValueOf(allowedAdditionalArguments).MapKeys())
	}
}
