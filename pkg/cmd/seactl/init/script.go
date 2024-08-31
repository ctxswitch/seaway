// Copyright 2024 Seaway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package init

// Uggg. Quick and dirty way to compact the responsibility of the user to create these
// and add it to the cli.

//nolint:gochecknoglobals
var sharedScript = `
if [[ x$SEAWAY_S3_ACCESS_KEY == "x" ]]; then
  echo "SEAWAY_S3_ACCESS_KEY is not set"
  exit 1
fi

if [[ x$SEAWAY_S3_SECRET_KEY == "x" ]]; then
  echo "SEAWAY_S3_SECRET_KEY is not set"
  exit 1
fi

if [[ x$MINIO_ROOT_USER == "x" ]]; then
  echo "MINIO_ROOT_USER is not set"
  exit 1
fi

if [[ x$MINIO_ROOT_PASSWORD == "x" ]]; then
  echo "MINIO_ROOT_PASSWORD is not set"
  exit 1
fi

SEAWAY_SHARED_BASE_URL=https://github.com/ctxswitch/seaway/config/shared
SEAWAY_CNTL_BASE_URL=https://github.com/ctxswitch/seaway/config/seaway

echo "###############################################################"
echo "# Deploying cert-manager"
echo "###############################################################"
kustomize build "${SEAWAY_SHARED_BASE_URL}/cert-manager" | envsubst | kubectl ${CONTEXT} apply -f - && \
kubectl wait --for=condition=available --timeout=120s deploy -l app.kubernetes.io/group=cert-manager -n cert-manager

echo "###############################################################"
echo "# Deploying minio-operator"
echo "###############################################################"
kustomize build ${SEAWAY_SHARED_BASE_URL}/minio | envsubst | kubectl ${CONTEXT} apply -f - && \
kubectl wait --for=condition=available --timeout=120s deploy/minio-operator -n minio-operator

echo "###############################################################"
echo "# Deploying minio-tenant, registry, and other resources"
echo "###############################################################"
kustomize build ${SEAWAY_SHARED_BASE_URL}/overlays/local | envsubst | kubectl ${CONTEXT} apply -f -


if [[ ${ENABLE_TAILSCALE} == "true" ]]; then
  echo "###############################################################"
  echo "# Deploying tailscale"
  echo "###############################################################"
  kustomize build ${SEAWAY_SHARED_BASE_URL}/tailscale | envsubst | kubectl ${CONTEXT} apply -f -
fi

echo "###############################################################"
echo "# Deploying the seaway controller"
echo "###############################################################"
kustomize build ${SEAWAY_CNTL_BASE_URL}/base | envsubst | kubectl ${CONTEXT} apply -f - && \

echo
echo
echo "Finished deploying shared resources."
echo
`
