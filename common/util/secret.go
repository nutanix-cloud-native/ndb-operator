package util

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns all the data in the given secret and namespace combination
// returns an error if the secret is not found
func GetAllDataFromSecret(ctx context.Context, client client.Client, name, namespace string) (data map[string][]byte, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered util.GetAllDataFromSecret", "Secret Name", name, "Secret Namespace", namespace)
	secret := &v1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret)
	if err != nil {
		log.Error(err, "Error occured while trying to read secret", "Secret Name", name, "Secret Namespace", namespace)
	} else {
		data = secret.Data
		log.Info("Returning from util.GetAllDataFromSecret")
	}
	return
}
