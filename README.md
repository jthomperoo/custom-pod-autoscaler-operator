# Custom Pod Autoscaler Operator
This is the operator for managing Custom Pod Autoscalers (CPA). This allows you to add your own CPAs to the cluster to manage autoscaling deployments, enabling this is a requirement before you can add your own CPAs.

## Installation
Run this to install the CPA definition and controller on your cluster:  
```
VERSION=0.1.0 
curl -L "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/deploy.tar.gz" | tar xvz --to-command 'kubectl apply -f -'
```

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