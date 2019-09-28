# Custom Pod Autoscaler Operator [ALPHA]

This is the operator for managing Custom Pod Autoscalers (CPA).

## Installation

Run these commands to install the operator on a Kubernetes cluster:  
Pull down configuration  
`git clone https://github.com/jthomperoo/custom-pod-autoscaler-operator && cd`  
Deploy Custom Resource Defintion for CPAs.  
`kubectl apply -f deploy/crds/`  
Deploy Operator deployment for managing CPAs.  
`kubectl apply -f deploy/`  

## Developing

### Environement

Developing this project requires these dependencies:

* Go >= 1.13
* [operator-sdk](https://github.com/operator-framework/operator-sdk) - [install guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md)

### Commands

* make - builds the image for the operator
* make generate - generates boilerplate and YAML config for the operator