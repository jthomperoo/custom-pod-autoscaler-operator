REGISTRY = custompodautoscaler
NAME = operator
VERSION = latest

default:
	@echo "=============Building image============="
	operator-sdk build "$(REGISTRY)/$(NAME):$(VERSION)"

lint:
	@echo "=============Linting============="
	find . -name '*.go' | grep -v zz_generated | golint -set_exit_status


generate:
	@echo "=============Generating boilerplate/yaml============="
	operator-sdk generate k8s
	operator-sdk generate openapi