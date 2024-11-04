CLUSTER_CONFIG_URL ?= https://raw.githubusercontent.com/ctxswitch/seaway/main/config/k3d/config.yaml
SEAWAY_SHARED_URL_BASE ?= https://github.com/ctxswitch/seaway/config/shared

init:
	pip install -r requirements.txt

venv:
	python3 -m venv .venv

run:
	source .venv/bin/activate && python3 -m app

.PHONY: k3d
k3d:
	curl -LSso /tmp/cluster.yaml $(CLUSTER_CONFIG_URL) && k3d cluster create --config /tmp/cluster.yaml

.PHONY: seaway
seaway:
	source .creds
	kustomize build $(SEAWAY_SHARED_URL_BASE)/cert-manager | envsubst | kubectl --context k3d-seaway apply -f -
	kubectl wait --for=condition=available --timeout=120s deploy -l app.kubernetes.io/group=cert-manager -n cert-manager
	kustomize build $(SEAWAY_SHARED_URL_BASE)/minio | envsubst | kubectl --context k3d-seaway apply -f -
	kubectl wait --for=condition=available --timeout=120s deploy/minio-operator -n minio-operator
	kustomize build $(SEAWAY_SHARED_URL_BASE)/overlays/local | envsubst | kubectl --context k3d-seaway apply -f -

.PHONY: clean
clean:
	kubectl --context k3d-seaway delete -k $(SEAWAY_SHARED_URL_BASE)/overlays/local
	kubectl --context k3d-seaway delete -k $(SEAWAY_SHARED_URL_BASE)/minio
	kubectl --context k3d-seaway delete -k $(SEAWAY_SHARED_URL_BASE)/cert-manager


