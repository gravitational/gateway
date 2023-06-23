# Teleport specific overrides for building and publishing off of our gateway-api 
# and envoy gateway forks.

TAG ?= $(shell git describe --tags --dirty --always)
RELEASE_VERSION ?= ${TAG}
CHART_VERSION ?= ${RELEASE_VERSION}

REGISTRY ?= public.ecr.aws/gravitational
IMAGE_NAME ?= envoy-gateway
IMAGE ?= ${REGISTRY}/${IMAGE_NAME}
OCI_REGISTRY ?= oci://public.ecr.aws/gravitational

.PHONY: teleport-generate
teleport-generate: generate manifests

.PHONY: teleport-test
teleport-test: ## Run tests using teleport specific configuration
teleport-test: test

.PHONY: teleport-verify
teleport-verify: ## Run lint and gen-check.
teleport-verify: gen-check lint

teleport-go-build: go.build

.PHONY: teleport-build
teleport-build: ## Run build using teleport specific configuration.
teleport-build: teleport-go-build image.build

.PHONY: teleport-push
teleport-push: ## Push the current build of envoy/gateway to teleport's registry.
teleport-push: teleport-build push

.PHONY: teleport-build-multiarch
teleport-build-multiarch: go.build.multiarch image-multiarch

teleport-helm-%: RELEASE_VERSION

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