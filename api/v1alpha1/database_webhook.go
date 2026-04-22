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
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
// Note: databaselog is only used by helper methods in webhook_helpers.go that don't receive a context.
// Webhook methods (Default, ValidateCreate, etc.) should use logf.FromContext(ctx) for request-aware logging.
var databaselog = logf.Log.WithName("database-resource")

// +kubebuilder:object:generate=false
// DatabaseCustomDefaulter injects ConfigMap defaults into the Database CR before validation.
// This ensures validation sees the fully populated CR and we don't need to relax validation.
// Client is a client.Reader (not client.Client) so that a direct API reader can be injected,
// avoiding a cluster-wide ConfigMap watch just to serve occasional webhook admission calls.
type DatabaseCustomDefaulter struct {
	Client client.Reader
}

var _ admission.CustomDefaulter = &DatabaseCustomDefaulter{}

func (r *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// In controller-runtime v0.21.0+, you must explicitly set the defaulter and validator
	// The For() method alone does not automatically detect these interfaces
	// Use GetAPIReader (direct, non-cached client) instead of GetClient (cache-backed).
	// The webhook only reads a ConfigMap on each admission event, so caching adds no value.
	// A cache-backed client would set up a cluster-wide ConfigMap watch, which requires the
	// 'watch' verb; GetAPIReader only needs 'get', which we already have.
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithDefaulter(&DatabaseCustomDefaulter{Client: mgr.GetAPIReader()}).
		WithValidator(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ndb-nutanix-com-v1alpha1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=mdatabase.kb.io,admissionReviewVersions=v1

// Default implements admission.CustomDefaulter. Fetches ConfigMap if defaultsConfigMapRef is set,
// applies defaults to the CR, then runs standard defaulter. Validation runs on the fully populated CR.
func (d *DatabaseCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)
	log.Info("Entering Default()...")

	db := obj.(*Database)

	// Apply ConfigMap defaults first (if specified)
	if db.Spec.DefaultsConfigMapRef != "" {
		defaults, err := FetchConfigMapDefaults(ctx, d.Client, db.Namespace, db.Spec.DefaultsConfigMapRef)
		if err != nil {
			log.Info("Could not fetch ConfigMap, proceeding without ConfigMap defaults", "configMapName", db.Spec.DefaultsConfigMapRef, "error", err)
		} else if len(defaults) > 0 {
			ApplyDefaultsFromConfigMap(ctx, db, defaults)
		}
	}

	// Run standard defaulter (description, databaseNames, timezone fallback, etc.)
	getDatabaseWebhookHandler(db).defaulter(&db.Spec)

	log.Info("Exiting Default()!")
	return nil
}

// +kubebuilder:webhook:path=/validate-ndb-nutanix-com-v1alpha1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=ndb.nutanix.com,resources=databases,verbs=create;update,versions=v1alpha1,name=vdatabase.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &Database{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Database) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Entering ValidateCreate...")

	db := obj.(*Database)
	errors := &field.ErrorList{}
	var path string
	if db.Spec.IsClone {
		path = "Clone"
	} else {
		path = "Instance"
	}

	// Defaulter webhook injects ConfigMap values before validation, so we always validate strictly
	getDatabaseWebhookHandler(db).validateCreate(&db.Spec, errors, field.NewPath("spec").Child(path))

	combined_err := util.CombineFieldErrors(*errors)

	log.Info("ValidateCreate webhook response...", "combined_err", combined_err)

	log.Info("Exiting ValidateCreate!")

	return nil, combined_err
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("validate update", "name", newObj.(*Database).Name)

	// TODO: This method will be used to make fields immutable.
	// Here you can reject the updates to any fields. I think we should mark everything immutable by default.
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *Database) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("validate delete", "name", obj.(*Database).Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

/* Checks if configured additional arguments are valid or not and returns the corresponding additional arguments. If error is nil valid, else invalid */
func additionalArgumentsValidationCheck(isClone bool, dbType string, specifiedAdditionalArguments map[string]string) error {
	// Empty additionalArguments is always valid
	if specifiedAdditionalArguments == nil {
		return nil
	}

	allowedAdditionalArguments, err := util.GetAllowedAdditionalArguments(isClone, dbType)

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
