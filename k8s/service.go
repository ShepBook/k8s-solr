package k8s

import (
	"github.com/ShepBook/k8s-solr/crd"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service will marshal the Solr object into a Kubernetes Service.
func Service(solr *crd.Solr) (*v1.Service, error) {
	name := Name(solr)

	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"addon": name,
			},
			Namespace: solr.ObjectMeta.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Port: int32(Port),
				},
			},
			Selector: map[string]string{
				"addon": name,
			},
			Type: "ClusterIP",
		},
	}, nil
}
