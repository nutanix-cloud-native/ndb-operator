package v1alpha1

import (
	"context"

	ndbv1alpha1 "github.com/nutanix-cloud-native/ndb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type DatabaseInterface interface {
	List(opts metav1.ListOptions) (*ndbv1alpha1.DatabaseList, error)
	Get(name string, options metav1.GetOptions) (*ndbv1alpha1.Database, error)
	Create(*ndbv1alpha1.Database) (*ndbv1alpha1.Database, error)
	Update(*ndbv1alpha1.Database) (*ndbv1alpha1.Database, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type DatabaseClient struct {
	restClient rest.Interface
	ns         string
}

func (c *DatabaseClient) List(opts metav1.ListOptions) (*ndbv1alpha1.DatabaseList, error) {
	result := ndbv1alpha1.DatabaseList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("Databases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *DatabaseClient) Get(name string, opts metav1.GetOptions) (*ndbv1alpha1.Database, error) {
	result := ndbv1alpha1.Database{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("Databases").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *DatabaseClient) Create(database *ndbv1alpha1.Database) (*ndbv1alpha1.Database, error) {
	result := ndbv1alpha1.Database{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("Databases").
		Body(database).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *DatabaseClient) Update(database *ndbv1alpha1.Database) (*ndbv1alpha1.Database, error) {
	result := ndbv1alpha1.Database{}
	err := c.restClient.
		Put().
		Namespace(c.ns).
		Resource("Databases").
		Name(database.Name).
		Body(database).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *DatabaseClient) Delete(name string, opts *metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(c.ns).
		Resource("Databases").
		Name(name).
		VersionedParams(opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Error()
}

func (c *DatabaseClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("Databases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(context.TODO())
}
