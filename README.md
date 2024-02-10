[![Build](https://github.com/jthomperoo/custom-pod-autoscaler-operator/workflows/main/badge.svg)](https://github.com/jthomperoo/custom-pod-autoscaler-operator/actions)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/jthomperoo/custom-pod-autoscaler-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/jthomperoo/custom-pod-autoscaler-operator)](https://goreportcard.com/report/github.com/jthomperoo/custom-pod-autoscaler-operator)
[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

# Custom Pod Autoscaler Operator

This is the operator for managing [Custom Pod Autoscalers](https://github.com/jthomperoo/custom-pod-autoscaler) (CPA).
This allows you to add your own CPAs to the cluster to manage autoscaling deployments, enabling this is a requirement
before you can add your own CPAs.

## Installation

See the [install guide](INSTALL.md) to see more in depth installation options, such as namespace specific installs and
installation using kubectl.

### Quick start

Run this to install the Operator and Custom Pod Autoscaler definition with cluster-wide scope on your cluster:
```
VERSION=v1.4.1
HELM_CHART=custom-pod-autoscaler-operator
helm install ${HELM_CHART} https://github.com/jthomperoo/custom-pod-autoscaler-operator/releases/download/${VERSION}/custom-pod-autoscaler-operator-${VERSION}.tgz
```

## Usage

See the [usage guide](USAGE.md) to see some simple usage options. For more indepth examples, check out the
[Custom Pod Autoscaler repo](https://github.com/jthomperoo/custom-pod-autoscaler).

## Developing

Developing this project requires these dependencies:

* [Go](https://golang.org/doc/install) == `1.18`

See the [contributing guide](CONTRIBUTING.md) for more information about how you can develop and contribute to this
project.

## Commands

* `make` - builds the operator binary.
* `make docker` - build the docker image for the operator.
* `make lint` - lints the codebase.
* `make format` - formats the codebase, must be run to pass the CI.
* `make test` - runs the Go tests.
* `make generate` - generates boilerplate and YAML config for the operator.
* `make view_coverage` - opens up any generated coverage reports in the browser.
