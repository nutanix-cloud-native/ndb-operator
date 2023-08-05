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
	if err != nil || username == "" || password == "" {
		var errStatement string
		if err != nil {
			errStatement = "An error occured while fetching the NDB Secrets"
		} else {
			errStatement = "NDB username or password cannot be empty"
			err = fmt.Errorf("empty credentials")
		}
		log.Error(err, errStatement, "username empty", username == "", "password empty", password == "")
		return
	}
	return
}
