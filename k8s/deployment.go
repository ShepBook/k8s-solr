package k8s

import (
	"errors"
	"fmt"

	"github.com/previousnext/k8s-solr/crd"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment will marshal the Solr object into a Kubernetes Deployment.
func Deployment(solr *crd.Solr) (*v1beta1.Deployment, error) {
	var (
		name     = Name(solr)
		replicas = int32(1)
		history  = int32(2)
	)

	// We default all our Solr cores to 5.x
	if solr.Spec.Version == "" {
		solr.Spec.Version = crd.SolrVersion5
	}

	cpuMin, cpuMax, mem, heap, err := sizeToResource(solr.Spec.Size)
	if err != nil {
		return nil, err
	}

	dply := &v1beta1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: solr.ObjectMeta.Namespace,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &history,
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"solr": solr.ObjectMeta.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"solr": solr.ObjectMeta.Name,
					},
					Namespace: solr.ObjectMeta.Namespace,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						// Our Solr containers run as the user "solr".
						// This container will ensure that the permissions are set.
						// Otherwise Solr will fail to boot in the first instance.
						v1.Container{
							Name:            "permissions",
							Image:           fmt.Sprintf("%s:init", Repository),
							ImagePullPolicy: "Always",
							Command: []string{
								"chown",
								"-R",
								"solr:solr",
								Data,
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "solr",
							Image: fmt.Sprintf("%s:%s", Repository, solr.Spec.Version),
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: int32(Port),
								},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SOLR_HEAP",
									Value: heap,
								},
								v1.EnvVar{
									Name:  "SOLR_CORE",
									Value: Core,
								},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									// https://cwiki.apache.org/confluence/display/solr/Ping
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt(Port),
									},
								},
								InitialDelaySeconds: 300,
								TimeoutSeconds:      10,
							},
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    cpuMin,
									v1.ResourceMemory: mem,
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    cpuMax,
									v1.ResourceMemory: mem,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "data",
									MountPath: Data,
								},
							},
							ImagePullPolicy: "Always",
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "data",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: name,
								},
							},
						},
					},
				},
			},
		},
	}

	return dply, nil
}

// Helper function to get sizing for the Solr deployment.
func sizeToResource(size crd.SolrSpecSize) (resource.Quantity, resource.Quantity, resource.Quantity, string, error) {
	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("100m"), resource.MustParse("1000m"), resource.MustParse("256Mi"), "256m", nil
	}

	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("500m"), resource.MustParse("1000m"), resource.MustParse("512Mi"), "512m", nil
	}

	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("1000m"), resource.MustParse("2000m"), resource.MustParse("1024Mi"), "1024m", nil
	}

	return resource.MustParse("100m"), resource.MustParse("1000m"), resource.MustParse("256Mi"), "256m", errors.New("cannot find size")
}
