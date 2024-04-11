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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"github.com/nutanix-cloud-native/ndb-operator/common/util"
	"github.com/nutanix-cloud-native/ndb-operator/ndb_client"
)

// SnapshotReconciler reconciles a Snapshot object
type SnapshotReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=snapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ndb.nutanix.com,resources=snapshots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Snapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	log.Info("<==============================Reconcile Started=============================>")
	// Fetch the snapshot from the namespace
	snapshot := &ndbv1alpha1.Snapshot{}
	err := r.Get(ctx, req.NamespacedName, snapshot)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("Snapshot not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Snapshot")
		return requeueOnErr(err)
	}

	log.Info("Snapshot Status: " + util.ToString(snapshot.Status))

	// Fetch the NDBServer resource from the namespace
	ndbServer := &ndbv1alpha1.NDBServer{}
	ndbNamespacedName := req.NamespacedName
	ndbNamespacedName.Name = "ndb"
	err = r.Get(ctx, ndbNamespacedName, ndbServer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("NDBServer resource not found. Ignoring since object must be deleted")
			return doNotRequeue()
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get NDB")
		return requeueOnErr(err)
	}

	NDBInfo := ndbServer.Spec
	username, password, caCert, err := getNDBCredentialsFromSecret(ctx, r.Client, NDBInfo.CredentialSecret, req.Namespace)
	if err != nil {
		r.recorder.Eventf(snapshot, "Warning", EVENT_INVALID_CREDENTIALS, "Error: %s", err.Error())
		return requeueOnErr(err)
	}
	if caCert == "" {
		log.Info("Ca-cert not found, falling back to host's HTTPs certs.")
	}
	ndbClient := ndb_client.NewNDBClient(username, password, NDBInfo.Server, caCert, NDBInfo.SkipCertificateVerification)

	return r.handleSync(ctx, snapshot, ndbClient, req, ndbServer)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//Create a new EventRecorder with the provided name
	r.recorder = mgr.GetEventRecorderFor("snapshot-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&ndbv1alpha1.Snapshot{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Endpoints{}).
		Complete(r)
}
