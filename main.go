package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/ShepBook/k8s-solr/client"
	"github.com/ShepBook/k8s-solr/crd"
	"github.com/ShepBook/k8s-solr/k8s"
)

var (
	cliNamespace  = kingpin.Flag("namespace", "Which namespace to operate in.").Default(v1.NamespaceAll).Envar("K8S_NAMESPACE").String()
	cliKubernetes = kingpin.Flag("kubernetes", "Kubernetes apiserver endpoint (for local development)").String()
)

func main() {
	kingpin.Parse()

	log.Info("Starting Solr Controller")

	config, err := getConnection(*cliKubernetes)
	if err != nil {
		panic(err)
	}

	clientset, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	log.Infof("Installing %s", crd.FullCRDName)

	err = crd.Create(clientset)
	if err != nil {
		panic(err)
	}

	k8sclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	crdcs, scheme, err := crd.Client(config)
	if err != nil {
		panic(err)
	}

	crdclient := client.New(crdcs, scheme, *cliNamespace)

	log.Infof("Watching %s for changes", crd.FullCRDName)

	// Watch the Kubernetes API for changes to our CRD.
	_, controller := cache.NewInformer(
		crdclient.NewListWatch(),
		&crd.Solr{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				solr := obj.(*crd.Solr)

				log.Infof("Solr core %s/%s was added", solr.ObjectMeta.Namespace, solr.ObjectMeta.Name)

				if err := sync(k8sclient, solr); err != nil {
					log.Errorf("failed to create Solr: %s", err)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				solr := newObj.(*crd.Solr)

				log.Infof("Solr core %s/%s was updated", solr.ObjectMeta.Namespace, solr.ObjectMeta.Name)

				if err := sync(k8sclient, solr); err != nil {
					log.Errorf("failed to update Solr: %s", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				solr := obj.(*crd.Solr)

				log.Infof("Solr core %s/%s was deleted", solr.ObjectMeta.Namespace, solr.ObjectMeta.Name)

				if err := delete(k8sclient, solr); err != nil {
					log.Errorf("failed to delete Solr: %s", err)
				}
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Info("Shutdown signal received, exiting...")
	close(stop)
}

// Helper function to get a Kubernetes API config.
func getConnection(host string) (*rest.Config, error) {
	if host != "" {
		return &rest.Config{
			Host: host,
		}, nil
	}

	return rest.InClusterConfig()
}

// Helper function create or update a Solr deployment.
func sync(client *kubernetes.Clientset, solr *crd.Solr) error {
	dply, err := k8s.Deployment(solr)
	if err != nil {
		return errors.Wrap(err, "cannot build Deployment object")
	}

	svc, err := k8s.Service(solr)
	if err != nil {
		return errors.Wrap(err, "cannot build Service object")
	}

	pvc, err := k8s.PersistentVolumeClaim(solr)
	if err != nil {
		return errors.Wrap(err, "cannot build PersistentVolumeClaim object")
	}

	err = syncPersistentVolumeClaim(client, pvc)
	if err != nil {
		return errors.Wrap(err, "cannot sync PersistentVolumeClaim")
	}

	err = syncDeployment(client, dply)
	if err != nil {
		return errors.Wrap(err, "cannot sync Deployment")
	}

	ip, port, err := syncService(client, svc)
	if err != nil {
		return errors.Wrap(err, "cannot sync Service")
	}

	if solr.Spec.ConfigMap == "" {
		return nil
	}

	err = syncConfigMap(client, solr, ip, port, k8s.Core)
	if err != nil {
		return errors.Wrap(err, "cannot sync ConfigMap")
	}

	return nil
}

// Helper function to keep a Kubernetes PersistentVolumeClaim object in sync.
// https://kubernetes.io/docs/concepts/storage/persistent-volumes
func syncPersistentVolumeClaim(client *kubernetes.Clientset, pvc *v1.PersistentVolumeClaim) error {
	_, err := client.Core().PersistentVolumeClaims(pvc.ObjectMeta.Namespace).Create(pvc)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "failed to create PersistentVolumeClaim")
	}

	return nil
}

// Helper function to keep a Kubernetes Deployment object in sync.
// http://kubernetes.io/docs/user-guide/deployments
func syncDeployment(client *kubernetes.Clientset, dply *v1beta1.Deployment) error {
	_, err := client.Extensions().Deployments(dply.ObjectMeta.Namespace).Create(dply)
	if kerrors.IsAlreadyExists(err) {
		_, err := client.Extensions().Deployments(dply.ObjectMeta.Namespace).Update(dply)
		if err != nil {
			return errors.Wrap(err, "failed to update Deployment")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to create Deployment")
	}

	return nil
}

// Helper function to keep a Kubernetes Service object in sync.
// http://kubernetes.io/docs/user-guide/services
func syncService(client *kubernetes.Clientset, svc *v1.Service) (string, string, error) {
	_, err := client.Core().Services(svc.ObjectMeta.Namespace).Create(svc)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return "", "", errors.Wrap(err, "failed to create Service")
	}

	// Load the Service to get its current status. This way we can return the current connection details.
	endpoint, err := client.Core().Services(svc.ObjectMeta.Namespace).Get(svc.ObjectMeta.Name, meta_v1.GetOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get Service")
	}

	return endpoint.Spec.ClusterIP, strconv.Itoa(int(k8s.Port)), nil
}

// Helper function to keep a Kubernetes ConfigMap object in sync.
// This is the supported way for applications to get connection details to this service.
// http://kubernetes.io/docs/user-guide/configmap
func syncConfigMap(client *kubernetes.Clientset, solr *crd.Solr, ip, port, core string) error {
	cfg, err := client.Core().ConfigMaps(solr.ObjectMeta.Namespace).Get(solr.Spec.ConfigMap, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get ConfigMap")
	}

	if cfg.Data == nil {
		cfg.Data = make(map[string]string)
	}

	var (
		keyBase = fmt.Sprintf("solr.%s.%s", solr.ObjectMeta.Name, core)
		keyHost = fmt.Sprintf("%s.host", keyBase)
		keyPort = fmt.Sprintf("%s.port", keyBase)
		keyCore = fmt.Sprintf("%s.core", keyBase)
	)

	// Set our values for application discovery.
	cfg.Data[keyHost] = ip
	cfg.Data[keyPort] = port
	cfg.Data[keyCore] = core

	_, err = client.Core().ConfigMaps(solr.ObjectMeta.Namespace).Update(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to update ConfigMap")
	}

	return nil
}

// Helper function to delete a Solr deployment.
func delete(client *kubernetes.Clientset, solr *crd.Solr) error {
	name := k8s.Name(solr)

	// Delete the PersistentVolumeClaim.
	err := client.Core().PersistentVolumeClaims(solr.ObjectMeta.Namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete PersistentVolumeClaim")
	}

	// Delete the Service.
	err = client.Core().Services(solr.ObjectMeta.Namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete Service")
	}

	// Delete the Deployment.
	err = client.Extensions().Deployments(solr.ObjectMeta.Namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete Deployment")
	}

	return nil
}
