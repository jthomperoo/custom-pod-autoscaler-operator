# Installation

## Cluster scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=0.2.0
curl -L "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.tar.gz" | tar xvz --to-command 'kubectl apply -f -'
```

## Namespace scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with namespaced scope on your cluster:
```
NAMESPACE=<INSERT_NAMESPACE_HERE>
VERSION=0.2.0
kubectl config set-context --current --namespace=${NAMESPACE}
curl -L "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/namespace.tar.gz" | tar xvz --to-command 'kubectl apply -f -'
```

## Manual install
If you want to customise your install, you can download either the cluster or namespace scoped config and edit it before applying to your kubernetes cluster:
### Cluster
```
VERSION=0.2.0
curl -OL "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.tar.gz"
```
### Namespace
```
VERSION=0.2.0
curl -OL "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/namespace.tar.gz"
```