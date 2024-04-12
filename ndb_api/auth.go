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

package ndb_api

import (
	"context"
	"net/http"

	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Validates the auth credentials against the 'auth/validate' endpoint
// Returns the response along with an error (if any)
func AuthValidate(ctx context.Context, ndbClient ndb_client.NDBClientHTTPInterface) (authValidateResponse AuthValidateResponse, err error) {
	log := ctrllog.FromContext(ctx)
	if _, err = sendRequest(ctx, ndbClient, http.MethodGet, "auth/validate", nil, &authValidateResponse); err != nil {
		log.Error(err, "Error in AuthValidate")
		return
	}
	return
}
