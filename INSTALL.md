# Installation

## Helm

### Cluster scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with
cluster-wide scope on your cluster:

```
VERSION=v1.1.1
HELM_CHART=custom-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

### Namespace scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with
namespaced scope on your cluster:

```
NAMESPACE=<INSERT_NAMESPACE_HERE>
VERSION=v1.1.1
HELM_CHART=custom-pod-autoscaler-operator
helm install --set mode=namespaced --namespace=${NAMESPACE}  ${HELM_CHART} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

## Kubectl

### Cluster scoped install

> Installation this using kubectl method only supports installation into the
> 'default' namespace when doing a cluster wide install.
> This is due to a limitation in how ClusterRoleBindings link to
> ServiceAccounts, to deploy to any namespace please use the helm deployment
> method.

Run this to install the Operator and Custom Pod Autoscaler definition with
cluster-wide scope on your cluster:

```
VERSION=v1.1.1
kubectl apply -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.yaml
```

### Namespace scoped install
Run this to install the Operator and Custom Pod Autoscaler definition with
namespaced scope on your cluster:

```
NAMESPACE=<INSERT_NAMESPACE_HERE>
VERSION=v1.1.1
kubectl config set-context --current --namespace=${NAMESPACE}
kubectl apply -f https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/namespaced.yaml
```
