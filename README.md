K8s - Solr (CRD)
================

**Maintainer**: Nick Schuch

Provides a Custom Resource Definition for deploying and discovering Solr cores.

![Diagram](/docs/diagram.png "Diagram")

## Example

**Ask for a Solr instance**

`kubectl create -f solr.yaml`

```yaml
apiVersion: skpr.io/v1
kind: Solr
metadata:
  name: dev
  namespace: testing
spec:
  size: small
```

**Get a list of Solr instances**

```bash
kubectl -n testing get solrs
NAME      AGE
dev       22s
```

**Check the Deployments**

```bash
kubectl -n testing get deployments
NAME       DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
solr-dev   1         0         0            0           28s
```

**Check the PersistentVolumeClaims**

```bash
kubectl -n testing get pvc
NAME       STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
solr-dev   Pending                                       solr           39s
```

**Check the Service**

```bash
kubectl -n testing get svc
NAME       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
solr-dev   ClusterIP   242.0.121.157   <none>        8983/TCP   9m
```

## Development

### Getting started

For steps on getting started with Go:

https://golang.org/doc/install

To get a checkout of the project run the following commands:

```bash
# Make sure the parent directories exist.
mkdir -p $GOPATH/src/github.com/ShepBook

# Checkout the codebase.
git clone git@github.com:ShepBook/k8s-solr $GOPATH/src/github.com/ShepBook/k8s-solr

# Change into the project to run workflow commands.
cd $GOPATH/src/github.com/ShepBook/k8s-solr
```

### Documentation

See `/docs`

### Resources

* [Dave Cheney - Reproducible Builds](https://www.youtube.com/watch?v=c3dW80eO88I)
* [Bryan Cantril - Debugging under fire](https://www.youtube.com/watch?v=30jNsCVLpAE&t=2675s)
* [Sam Boyer - The New Era of Go Package Management](https://www.youtube.com/watch?v=5LtMb090AZI)
* [Kelsey Hightower - From development to production](https://www.youtube.com/watch?v=XL9CQobFB8I&t=787s)

### Tools

```bash
# Dependency management
go get -u github.com/golang/dep/cmd/dep

# Testing
go get -u github.com/golang/lint/golint

# Release management.
go get -u github.com/tcnksm/ghr

# Build
go get -u github.com/mitchellh/gox
```

### Workflow

**Testing**

```bash
make lint test
```

**Building**

```bash
make build
```

**Releasing**

```bash
make release
```
