REGISTRY = custompodautoscaler
NAME = operator
VERSION = latest

default:
	operator-sdk build "$(REGISTRY)/$(NAME):$(VERSION)"

generate:
	operator-sdk generate k8s
	operator-sdk generate openapi