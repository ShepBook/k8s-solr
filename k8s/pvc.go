package k8s

import (
	"github.com/previousnext/k8s-solr/crd"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaim will marshal the Solr object into a Kubernetes PersistentVolumeClaim.
func PersistentVolumeClaim(solr *crd.Solr) (*v1.PersistentVolumeClaim, error) {
	name := Name(solr)

	return &v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: solr.ObjectMeta.Namespace,
			Annotations: map[string]string{
				// This means that operators have to declare the storage.
				"volume.beta.kubernetes.io/storage-class": "solr",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("10Gi"),
				},
			},
		},
	}, nil
}
