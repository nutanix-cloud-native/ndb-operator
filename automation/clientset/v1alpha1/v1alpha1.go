package v1alpha1

import (
	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type V1alpha1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*V1alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: ndbv1alpha1.GroupVersion.Group, Version: ndbv1alpha1.GroupVersion.Version}
	config.APIPath = "/apis"
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)
	negotiatedSerializer := codecs.WithoutConversion()
	config.NegotiatedSerializer = negotiatedSerializer
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &V1alpha1Client{restClient: client}, nil
}

func (c *V1alpha1Client) Databases(namespace string) DatabaseInterface {
	if namespace == "" {
		namespace = "default"
	}
	return &DatabaseClient{
		restClient: c.restClient,
		namespace:  namespace,
	}
}

func (c *V1alpha1Client) NDBServers(namespace string) NDBServerInterface {
	if namespace == "" {
		namespace = "default"
	}
	return &NDBServerClient{
		restClient: c.restClient,
		namespace:  namespace,
	}
}
