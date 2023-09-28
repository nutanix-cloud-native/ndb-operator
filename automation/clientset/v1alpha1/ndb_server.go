package v1alpha1

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Common Functionality used to interact with NDBServer CR using Kubernetes Client
type NDBServerInterface interface {
	List(opts metav1.ListOptions) (*ndbv1alpha1.NDBServerList, error)
	Get(name string, options metav1.GetOptions) (*ndbv1alpha1.NDBServer, error)
	Create(*ndbv1alpha1.NDBServer) (*ndbv1alpha1.NDBServer, error)
	Update(*ndbv1alpha1.NDBServer) (*ndbv1alpha1.NDBServer, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type NDBServerClient struct {
	restClient rest.Interface
	namespace  string
}

func (c *NDBServerClient) List(opts metav1.ListOptions) (*ndbv1alpha1.NDBServerList, error) {
	result := ndbv1alpha1.NDBServerList{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource("NDBServers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *NDBServerClient) Get(name string, opts metav1.GetOptions) (*ndbv1alpha1.NDBServer, error) {
	result := ndbv1alpha1.NDBServer{}
	err := c.restClient.
		Get().
		Namespace(c.namespace).
		Resource("NDBServers").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *NDBServerClient) Create(NDBServer *ndbv1alpha1.NDBServer) (*ndbv1alpha1.NDBServer, error) {
	result := ndbv1alpha1.NDBServer{}
	err := c.restClient.
		Post().
		Namespace(c.namespace).
		Resource("NDBServers").
		Body(NDBServer).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *NDBServerClient) Update(NDBServer *ndbv1alpha1.NDBServer) (*ndbv1alpha1.NDBServer, error) {
	result := ndbv1alpha1.NDBServer{}
	err := c.restClient.
		Put().
		Namespace(c.namespace).
		Resource("NDBServers").
		Name(NDBServer.Name).
		Body(NDBServer).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *NDBServerClient) Delete(name string, opts *metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(c.namespace).
		Resource("NDBServers").
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Error()
}

func (c *NDBServerClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.namespace).
		Resource("NDBServers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.TODO())
}
