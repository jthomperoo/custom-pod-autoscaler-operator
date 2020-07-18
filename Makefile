REGISTRY = custompodautoscaler
NAME = operator
VERSION = latest

default: vendor
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

vendor:
	go mod vendor

# Run linting with golint
lint: vendor
	@echo "=============Linting============="
	go list -mod vendor ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status

# Run tests
test: vendor
	@echo "=============Running tests============="
	CGO_ENABLED=0 GOOS=linux go test -mod vendor ./... -cover -coverprofile unit_cover.out

# Build the docker image
docker: default
	docker build . -t $(REGISTRY)/$(NAME):$(VERSION)

# Generate code
generate:
	@echo "=============Generating YAML============="
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	$(CONTROLLER_GEN) $(CRD_OPTIONS) crd paths="./..." output:crd:artifacts:config=deploy/crd

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./main.go

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
