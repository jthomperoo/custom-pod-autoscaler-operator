# Installation

## Helm

### Cluster scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=v0.6.0
helm install https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

### Namespace scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with namespaced scope on your cluster:
```
NAMESPACE=<INSERT_NAMESPACE_HERE>
VERSION=v0.6.0
helm install --namespace=${NAMESPACE} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
``````

## Kubectl

### Cluster scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=v0.6.0
kubectl apply -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.yaml
```

### Namespace scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with namespaced scope on your cluster:
```
NAMESPACE=<INSERT_NAMESPACE_HERE>
VERSION=v0.6.0
kubectl config set-context --current --namespace=${NAMESPACE}
kubectl apply -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/namespaced.yaml
``````
