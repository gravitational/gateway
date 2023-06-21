# Teleport specific overrides for building and publishing off of our gateway-api 
# and envoy gateway forks.

# TODO(david): gateway-api use tag reference in release url.
GATEWAY_API_VERSION ?= "v0.6.2"
GATEWAY_RELEASE_URL ?= https://raw.githubusercontent.com/gravitational/gatweway-api/teleport/release/experimental-install.yaml

TAG = $(shell git describe --tags --dirty)
RELEASE_VERSION = ${TAG}

REGISTRY = public.ecr.aws/gravitational
IMAGE_NAME = envoy-gateway
IMAGE = ${REGISTRY}/${IMAGE_NAME}
OCI_REGISTRY = oci://public.ecr.aws/gravitational

.PHONY: teleport-generate
teleport-generate: generate manifests

.PHONY: teleport-test
teleport-test: ## Run tests using teleport specific configuration
teleport-test: test

.PHONY: teleport-verify
teleport-verify: ## Run lint and gen-check.
teleport-verify: gen-check lint

.PHONY: teleport-build
teleport-build: ## Run build using teleport specific configuration.
teleport-build: image.build.multiarch

.PHONY: teleport-push
teleport-push: ## Push the current build of envoy/gateway to teleport's registry.
teleport-push: push

.PHONY: teleport-helm-package
teleport-helm-package: ## Package envoy gateway helm chart with teleprot specific overrides.
teleport-helm-package: helm-package

.PHONY: teleport-helm-push
teleport-helm-push: ## Push envoy gateway helm chart to teleport's OCI registry.
teleport-helm-push: teleport-helm-package helm-push

.PHONY: teleport-dev-install
teleport-dev-install: ## Install the local changes to gateway the regional kind cluster.
teleport-dev-install: export CLUSTER_NAME := regional
teleport-dev-install: kube-install-image