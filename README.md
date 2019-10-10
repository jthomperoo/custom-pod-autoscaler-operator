![](https://github.com/jthomperoo/custom-pod-autoscaler-operator/workflows/main/badge.svg)
# Custom Pod Autoscaler Operator
This is the operator for managing Custom Pod Autoscalers (CPA). This allows you to add your own CPAs to the cluster to manage autoscaling deployments, enabling this is a requirement before you can add your own CPAs.

## Installation
### Quick start
Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=0.2.0
curl -L "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/cluster.tar.gz" | tar xvz --to-command 'kubectl apply -f -'
```
### Advanced
See the [install guide](INSTALL.md) to see more in depth installation options, such as namespace specific installs.

## Developing

### Environment

Developing this project requires these dependencies:

* Go >= 1.13
* Golint
* [operator-sdk](https://github.com/operator-framework/operator-sdk) - [install guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md)

### Commands

* `make` - builds the image for the operator
* `make lint` - lints the codebase
* `make generate` - generates boilerplate and YAML config for the operator