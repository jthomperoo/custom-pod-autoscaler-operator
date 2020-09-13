[![Build](https://github.com/jthomperoo/custom-pod-autoscaler-operator/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler-operator/actions)
[![codecov](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler-operator/branch/master/graph/badge.svg)](https://codecov.io/gh/jthomperoo/custom-pod-autoscaler-operator)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/custom-pod-autoscaler-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler-operator)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler-operator)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

<p>This project is supported by:</p>
<p>
  <a href="https://www.digitalocean.com/">
    <img src="https://opensource.nyc3.cdn.digitaloceanspaces.com/attribution/assets/SVG/DO_Logo_horizontal_blue.svg" width="201px">
  </a>
</p>

# Custom Pod Autoscaler Operator
This is the operator for managing Custom Pod Autoscalers (CPA). This allows you to add
your own CPAs to the cluster to manage autoscaling deployments, enabling this is a
requirement before you can add your own CPAs.

The Custom Pod Autoscaler Operator is part of the
[Custom Pod Autoscaler Framework](https://custom-pod-autoscaler.readthedocs.io/en/stable/).

## Installation
### Quick start
Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=v1.0.2
HELM_CHART=custom-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

### Advanced
See the [install guide](INSTALL.md) to see more in depth installation options,
such as namespace specific installs and installation using kubectl.

## Usage
See the [usage guide](USAGE.md) to see some simple usage options. For more indepth
examples, check out the
[Custom Pod Autoscaler repo](https://github.com/jthomperoo/custom-pod-autoscaler).

## Developing

### Environment

Developing this project requires these dependencies:

* Go >= 1.13
* Golint
* [operator-sdk `v0.19.0`](https://github.com/operator-framework/operator-sdk) -
[install guide](https://sdk.operatorframework.io/docs/install-operator-sdk/)

### Commands

* `make` - builds the image for the operator
* `make lint` - lints the codebase
* `make generate` - generates boilerplate and YAML config for the operator
