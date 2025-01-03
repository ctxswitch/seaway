ENV ?= "local"
SYSTEM ?= $(shell uname -s | tr '[:upper:]' '[:lower:]' )
ARCH ?= $(shell uname -m)
ifeq ($(ARCH), "aarch64")
	ARCH = "arm64"
endif

LOCALDEV_CLUSTER ?= "seaway"

CONTROLLER_TOOLS_VERSION ?= v0.16.1
KUSTOMIZE_VERSION ?= v5.4.2
GOLANGCI_LINT_VERSION ?= v1.60.3
ADDLICENSE_VERSION ?= v1.0.0

KUBECTL ?= kubectl
LOCALBIN ?= $(shell pwd)/bin
TARGETDIR ?= $(shell pwd)/dist
SEACTL_RELEASE_TARGET ?= $(TARGETDIR)/seactl-$(SYSTEM)-$(ARCH).tar.gz
SEACTL_BIN ?= seactl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
ADDLICENSE = $(LOCALBIN)/addlicense

CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"
RBAC_OPTIONS ?= "rbac:roleName=seaway-system-role"
WEBHOOK_OPTIONS ?= "webhook"
OUTPUT_OPTIONS ?= output:crd:dir=config/seaway/crd output:webhook:dir=config/seaway/webhook output:rbac:dir=config/seaway/rbac
VERSION ?= $(shell git describe --tags --always --dirty)

COVERAGE ?= 1
ifeq ($(COVERAGE), 1)
	GO_COVERPROFILE = "-coverprofile=cover.out"
else
	GO_COVERPROFILE = ""
endif

VERBOSE ?= 0
ifeq ($(VERBOSE), 1)
	GO_VERBOSE = "-v"
else
	GO_VERBOSE =
endif

###
### Generators
###
.PHONY: codegen
codegen: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."

.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) paths="./pkg/..." $(CRD_OPTIONS) $(RBAC_OPTIONS) $(WEBHOOK_OPTIONS) $(OUTPUT_OPTIONS)

.PHONY: installgen
installgen:
	go run hack/generator.go config/seaway pkg/cmd/seactl/install

.PHONY: generate
generate: codegen manifests installgen

###
### Set up a local development environment
###
.PHONY: localdev
localdev: localdev-cluster localdev-shared

.PHONY: localdev-cluster
localdev-cluster:
	@if k3d cluster get $(LOCALDEV_CLUSTER) --no-headers >/dev/null 2>&1;  \
		then echo "Cluster exists, skipping creation"; \
		else k3d cluster create --config config/k3d/config.yaml --volume $(PWD):/app; \
		fi

.PHONY: localdev-shared
localdev-shared:
	@$(KUSTOMIZE) build config/shared/cert-manager | envsubst | $(KUBECTL) apply -f -
	@$(KUBECTL) wait --for=condition=available --timeout=120s deploy -l app.kubernetes.io/group=cert-manager -n cert-manager
	@$(KUSTOMIZE) build config/shared/minio | envsubst | $(KUBECTL) apply -f -
	@$(KUBECTL) wait --for=condition=available --timeout=120s deploy/minio-operator -n minio-operator
	@$(KUSTOMIZE) build config/shared/overlays/$(ENV) | envsubst | $(KUBECTL) apply -f -

.PHONY: localdev-seaway
localdev-seaway:
	@$(KUSTOMIZE) build config/seaway/overlays/$(ENV) | envsubst | $(KUBECTL) apply -f -

.PHONY: clean-localdev-shared
clean-localdev-shared:
	@$(KUBECTL) delete -k config/shared/overlays/$(ENV)

###
### Build, install, run, and clean
###
.PHONY: install
install: $(KUSTOMIZE) generate
	@$(KUSTOMIZE) build config/seaway/crd | kubectl apply -f -

.PHONY: uninstall
uninstall:
	kubectl delete -k config/overlays/$(ENV)

###
### Tests/Utils
###
.PHONY: test
test:
	go test ./... $(GO_VERBOSE) $(GO_COVERPROFILE)

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run --fix

.PHONY: license
license: $(ADDLICENSE)
	@find . -name '*.go' | xargs $(ADDLICENSE) -c "Seaway Authors" -y 2024 -l apache

.PHONY: run
run:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -it -n seaway-system pod/$(POD) -- bash -c "go run pkg/cmd/seaway/*.go controller --log-level=5"

.PHONY: run-uploader
run-uploader:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -it -n seaway-system pod/$(POD) -- bash -c "go run pkg/cmd/seaway/*.go uploader --log-level=5"

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n seaway-system -it pod/$(POD) -- bash

###
### Builds
###
$(TARGETDIR):
	mkdir -p $(TARGETDIR)

.PHONY: build
build: $(TARGETDIR)
	CGO_ENABLED=0 go build -trimpath --ldflags "-s -w -X ctx.sh/seaway/pkg/build.Version=$(VERSION)" -o $(TARGETDIR)/seactl ./pkg/cmd/seactl
	CGO_ENABLED=0 go build -trimpath --ldflags "-s -w -X ctx.sh/seaway/pkg/build.Version=$(VERSION)" -o $(TARGETDIR)/seaway ./pkg/cmd/seaway

.PHONY: build-seactl-release
build-seactl-release: $(TARGETDIR) $(SEACTL_RELEASE_TARGET)

$(SEACTL_RELEASE_TARGET):
	GOOS=$(SYSTEM) GOARCH=$(ARCH) CGO_ENABLED=0 go build -trimpath --ldflags "-s -w -X ctx.sh/seaway/pkg/build.Version=$(VERSION)" -o $(TARGETDIR)/$(SEACTL_BIN) ./pkg/cmd/seactl && \
		tar -C $(TARGETDIR) -zcvf $@ $(SEACTL_BIN) && \
		rm -f $(TARGETDIR)/$(SEACTL_BIN)

###
### Individual dep installs were copied out of kubebuilder testdata makefiles.
###
deps: $(CONTROLLER_GEN) $(KUSTOMIZE) $(GOLANGCI_LINT) $(ADDLICENSE)

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

$(CONTROLLER_GEN): $(LOCALBIN)
	@test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

$(KUSTOMIZE):
	@test -s $(KUSTOMIZE) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

$(GOLANGCI_LINT): $(LOCALBIN)
	@test -s $(GOLANGCI_LINT) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

$(ADDLICENSE): $(LOCALBIN)
	@test -s $(ADDLICENSE) || \
  GOBIN=$(LOCALBIN) go install github.com/google/addlicense@$(ADDLICENSE_VERSION)

.PHONY: clean
clean:
	@k3d cluster delete seaway 
