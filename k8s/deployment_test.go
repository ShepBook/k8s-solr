package k8s

import (
	"fmt"
	"testing"

	"github.com/previousnext/k8s-solr/crd"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIngress(t *testing.T) {
	solr := &crd.Solr{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: crd.SolrSpec{
			Size: crd.SolrSpecSizeSmall,
		},
	}

	have, err := Deployment(solr)
	assert.Nil(t, err)

	var (
		replicas = int32(1)
		history  = int32(2)
	)

	want := &v1beta1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "solr-foo",
			Namespace: "bar",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &history,
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"solr": "foo",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "solr-foo",
					Labels: map[string]string{
						"solr": "foo",
					},
					Namespace: solr.ObjectMeta.Namespace,
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						v1.Container{
							Name:            "permissions",
							Image:           "previousnext/solr:init",
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
							Image: fmt.Sprintf("previousnext/solr:5.x"),
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									ContainerPort: int32(Port),
								},
							},
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SOLR_HEAP",
									Value: "256m",
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
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1000m"),
									v1.ResourceMemory: resource.MustParse("256Mi"),
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
									ClaimName: "solr-foo",
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, want, have)
}
