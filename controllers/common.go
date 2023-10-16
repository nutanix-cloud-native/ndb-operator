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

package controllers

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nutanix-cloud-native/ndb-operator/common"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	EVENT_INVALID_CREDENTIALS = "InvalidCredentials"

	EVENT_REQUEST_GENERATION         = "RequestGenerated"
	EVENT_REQUEST_GENERATION_FAILURE = "RequestGenerationFailed"
	EVENT_NDB_REQUEST_FAILED         = "NDBRequestFailed"

	EVENT_PROVISIONING_STARTED   = "ProvisioningStarted"
	EVENT_PROVISIONING_FAILED    = "ProvisioningFailed"
	EVENT_PROVISIONING_COMPLETED = "ProvisioningCompleted"

	EVENT_DEPROVISIONING_STARTED   = "DeprovisioningStarted"
	EVENT_DEPROVISIONING_FAILED    = "DeprovisioningFailed"
	EVENT_DEPROVISIONING_COMPLETED = "DeprovisioningCompleted"

	EVENT_CR_CREATED              = "CustomResourceCreated"
	EVENT_CR_DELETED              = "CustomResourceDeleted"
	EVENT_CR_STATUS_UPDATE_FAILED = "CustomResourceStatusUpdateFailed"

	EVENT_EXTERNAL_DELETE = "ExternalDeleteDetected"

	EVENT_RESOURCE_LOOKUP_ERROR = "ResourceLookupError"

	EVENT_SERVICE_SETUP_FAILED  = "ServiceSetupFailed"
	EVENT_ENDPOINT_SETUP_FAILED = "EndpointSetupFailed"

	EVENT_WAITING_FOR_NDB_RECONCILE = "WaitingForNDBReconcile"
	EVENT_WAITING_FOR_IP_ADDRESS    = "WaitingForIPAddress"
)

// doNotRequeue Finished processing. No need to put back on the reconcile queue.
func doNotRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// requeue Not finished processing. Put back on reconcile queue and continue.
func requeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// requeueOnErr Failed while processing. Put back on reconcile queue and try again.
func requeueOnErr(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// requeue after a timeout. Put back on reconcile queue after a timeout and continue.
func requeueWithTimeout(t int) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: time.Second * time.Duration(math.Abs(float64(t)))}, nil
}

// Returns the credentials(username, password and caCertificate) for NDB
// Returns an error if reading the secret containing credentials fails
func getNDBCredentialsFromSecret(ctx context.Context, k8sClient client.Client, name, namespace string) (username, password, caCert string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Reading secret", "Secret Name", name)
	secretDataMap, err := util.GetAllDataFromSecret(ctx, k8sClient, name, namespace)
	if err != nil {
		log.Error(err, "Error occured while fetching NDB secret", "Secret Name", name, "Namespace", namespace)
		return
	}
	username = string(secretDataMap[common.SECRET_DATA_KEY_USERNAME])
	password = string(secretDataMap[common.SECRET_DATA_KEY_PASSWORD])
	caCert = string(secretDataMap[common.SECRET_DATA_KEY_CA_CERTIFICATE])
	if username == "" || password == "" {
		errStatement := "NDB username or password is empty."
		err = fmt.Errorf("Empty credentials. " + errStatement)
		log.Error(err, errStatement, "username empty", username == "", "password empty", password == "")
		return
	}
	return
}
