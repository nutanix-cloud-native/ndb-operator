package util

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Returns the data for a key present in the given secret and namespace combination
// returns an error if either the secret or the data for the given key is not found
func GetDataFromSecret(ctx context.Context, client client.Client, name, namespace, key string) (data string, err error) {
	log := ctrllog.FromContext(ctx)
	log.Info("Entered util.GetDataFromSecret", "key", key, "Secret Name", name, "Secret Namespace", namespace)
	secret := &v1.Secret{}
	err = client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret)
	if err != nil {
		log.Error(err, "Error occured while trying to read secret", "Secret Name", name, "Secret Namespace", namespace)
	} else {
		if val, ok := secret.Data[key]; ok {
			data = string(val)
			log.Info("Returning from util.GetDataFromSecret")
		} else {
			err = fmt.Errorf(fmt.Sprintf("Key '%s' not present in the secret '%s' in namespace '%s'", key, name, namespace))
			log.Error(err, "Data not found in secret")
		}
	}
	return
}
