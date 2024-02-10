REGISTRY = custompodautoscaler
NAME = operator
VERSION = latest

default: vendor_modules generate
	@echo "=============Building============="
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -o dist/$(NAME) main.go
	cp LICENSE dist/LICENSE

# Run linting with golint
lint: vendor_modules generate
	@echo "=============Linting============="
	go run honnef.co/go/tools/cmd/staticcheck ./...

format: vendor_modules
	@echo "=============Formatting============="
	gofmt -s -w .
	go mod tidy

# Run tests
test: vendor_modules generate
	@echo "=============Running tests============="
	go test -mod vendor ./... -cover -coverprofile unit_cover.out

# Build the docker image
docker: default
	docker build . -t $(REGISTRY)/$(NAME):$(VERSION)

# Generate code and manifests
generate: get_controller-gen
	@echo "=============Generating Golang and YAML============="
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
	controller-gen rbac:roleName=manager-role webhook crd paths="./..." output:crd:artifacts:config=helm/templates/crd

vendor_modules:
	go mod vendor

view_coverage:
	@echo "=============Loading coverage HTML============="
	go tool cover -html=unit_cover.out

get_controller-gen:
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
