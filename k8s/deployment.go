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

const (
	// Port for Solr to respond on.
	Port = 8983
	// Core which stores data.
	Core = "core1"
)

// Deployment will marshal the Solr object into a Kubernetes Deployment.
func Deployment(solr *crd.Solr) (*v1beta1.Deployment, error) {
	var (
		name     = Name(solr)
		replicas = int32(1)
		history  = int32(2)
	)

	cpu, mem, err := sizeToResource(solr.Spec.Size)
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
					"addon": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"addon": name,
					},
					Namespace: solr.ObjectMeta.Namespace,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name:  "solr",
							Image: fmt.Sprintf("previousnext/solr:%s", solr.Spec.Version),
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: int32(Port),
								},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SOLR_HEAP",
									Value: fmt.Sprintf("%dm", mem.Value()),
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
									v1.ResourceCPU:    cpu,
									v1.ResourceMemory: mem,
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    cpu,
									v1.ResourceMemory: mem,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "data",
									MountPath: "/opt/solr/data",
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
func sizeToResource(size crd.SolrSpecSize) (resource.Quantity, resource.Quantity, error) {
	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("100m"), resource.MustParse("256Mi"), nil
	}

	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("500m"), resource.MustParse("512Mi"), nil
	}

	if size == crd.SolrSpecSizeSmall {
		return resource.MustParse("1000m"), resource.MustParse("1024Mi"), nil
	}

	return resource.MustParse("100m"), resource.MustParse("256Mi"), errors.New("cannot find size")
}
