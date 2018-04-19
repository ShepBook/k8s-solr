package client

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/ShepBook/k8s-solr/crd"
)

func New(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *crdclient {
	return &crdclient{
		cl:     cl,
		ns:     namespace,
		plural: crd.Plural,
		codec:  runtime.NewParameterCodec(scheme),
	}
}

type crdclient struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

func (f *crdclient) Create(obj *crd.Solr) (*crd.Solr, error) {
	var result crd.Solr
	err := f.cl.Post().Namespace(f.ns).Resource(f.plural).Body(obj).Do().Into(&result)
	return &result, err
}

func (f *crdclient) Update(obj *crd.Solr) (*crd.Solr, error) {
	var result crd.Solr
	err := f.cl.Put().Namespace(f.ns).Resource(f.plural).Body(obj).Do().Into(&result)
	return &result, err
}

func (f *crdclient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().Namespace(f.ns).Resource(f.plural).Name(name).Body(options).Do().Error()
}

func (f *crdclient) Get(name string) (*crd.Solr, error) {
	var result crd.Solr
	err := f.cl.Get().Namespace(f.ns).Resource(f.plural).Name(name).Do().Into(&result)
	return &result, err
}

func (f *crdclient) List(opts meta_v1.ListOptions) (*crd.SolrList, error) {
	var result crd.SolrList
	err := f.cl.Get().Namespace(f.ns).Resource(f.plural).VersionedParams(&opts, f.codec).Do().Into(&result)
	return &result, err
}

// Create a new List watch for our TPR
func (f *crdclient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}
