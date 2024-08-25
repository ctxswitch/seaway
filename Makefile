ENV ?= "dev"
SYSTEM ?= $(shell uname -s | tr '[:upper:]' '[:lower:]' )
ARCH ?= $(shell uname -m)

LOCALDEV_CLUSTER ?= "seaway"

CONTROLLER_TOOLS_VERSION ?= v0.14.0
KUSTOMIZE_VERSION ?= v5.4.2

KUBECTL ?= kubectl
LOCALBIN ?= $(shell pwd)/bin
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

CRD_OPTIONS ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true"
RBAC_OPTIONS ?= "rbac:roleName=seaway-system-role"
WEBHOOK_OPTIONS ?= "webhook"
OUTPUT_OPTIONS ?= "output:artifacts:config=config/base/crd"


# kube::codegen::gen_client \
#     --with-watch \
#     --output-dir "${SCRIPT_ROOT}/pkg/generated" \
#     --output-pkg "${THIS_PKG}/pkg/generated" \
#     --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#     "${SCRIPT_ROOT}/pkg/apis"

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
	GO_VERBOSE = ""
endif

###
### Generators
###
.PHONY: codegen
codegen: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."

.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) $(RBAC_OPTIONS) $(WEBHOOK_OPTIONS) paths="./pkg/..."

.PHONY: clients
clients: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CLIENT_OPTIONS) paths="./pkg/..."

.PHONY: generate
generate: codegen manifests

.PHONY: test
test:
	go test ./... $(GO_VERBOSE) $(GO_COVERPROFILE)

###
### Set up a local development environment
###
.PHONY: localdev
localdev: local-cluster local-cert-manager local-minio install

.PHONY: local-cluster
local-cluster:
	@if k3d cluster get $(LOCALDEV_CLUSTER) --no-headers >/dev/null 2>&1;  \
		then echo "Cluster exists, skipping creation"; \
		else k3d cluster create --config config/cluster/config.yaml --volume $(PWD):/app; \
		fi

.PHONY: local-cert-manager
local-cert-manager:
	@$(KUSTOMIZE) build config/cert-manager | envsubst | kubectl apply -f -
	@kubectl wait --for=condition=available --timeout=120s deploy -l app.kubernetes.io/group=cert-manager -n cert-manager

.PHONY: local-minio
local-minio:
	@$(KUSTOMIZE) build config/minio | envsubst | kubectl apply -f -
	kubectl wait --for=condition=available --timeout=120s deploy/minio-operator -n minio-operator 

###
### Build, install, run, and clean
###
.PHONY: install
install: $(KUSTOMIZE) generate
	@$(KUSTOMIZE) build config/overlays/$(ENV) | envsubst | kubectl apply -f -

.PHONY: uninstall
uninstall:
	kubectl delete -k config/overlays/$(ENV)

.PHONY: run
run:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n seaway-system -it pod/$(POD) -- bash -c "go run pkg/cmd/controller/*.go controller --log-level=5"

.PHONY: run-sync
run-sync:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	go run pkg/cmd/seactl/*.go sync

.PHONY: exec
exec:
	$(eval POD := $(shell kubectl get pods -n seaway-system -l app=seaway-controller -o=custom-columns=:metadata.name --no-headers))
	kubectl exec -n seaway-system -it pod/$(POD) -- bash

###
### Individual dep installs were copied out of kubebuilder testdata makefiles.
###
deps: $(CONTROLLER_GEN) $(KUSTOMIZE)

$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

$(CONTROLLER_GEN): $(LOCALBIN)
	@test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

$(KUSTOMIZE):
	@test -s $(KUSTOMIZE) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: clean
clean:
	@k3d cluster delete seaway 
