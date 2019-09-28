# Custom Pod Autoscaler Operator [PRE-RELEASE]
This is the operator for managing Custom Pod Autoscalers (CPA).

## Installation
Run this to install the CPA definition and controller on your cluster:  
```
VERSION=0.1.0 
curl "https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/deploy.tar.gz" | tar xvz --to-command 'kubectl apply -f -'
```

## Developing

### Environement

Developing this project requires these dependencies:

* Go >= 1.13
* [operator-sdk](https://github.com/operator-framework/operator-sdk) - [install guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md)

### Commands

* make - builds the image for the operator
* make generate - generates boilerplate and YAML config for the operator