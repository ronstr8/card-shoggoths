.PHONY: help run dev test dump-to-clipboard build push deploy undeploy clean tail-app-logs status self-signed-cert-issuer

# Extract version from Go source code
VERSION := $(shell grep 'const Version' cmd/card-shoggoths-server/version.go | sed 's/.*"\(.*\)".*/\1/')

# Variables for Kubernetes deployment
IMAGE_NAME = card-shoggoths
IMAGE_TAG ?= $(VERSION)
REGISTRY = localhost:5000
FULL_IMAGE = $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
NAMESPACE = poker
RELEASE_NAME = poker

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Development targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^(run|dev|test):.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Kubernetes targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]*k8s:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

run: ## Run server locally
	go run ./cmd/card-shoggoths-server

dev: ## Run with air for hot reload
	air

test: ## Run tests
	go test ./internal/... ./cmd/...

dump-to-clipboard:
	while read fn ; do \
		echo -ne "\\n\\n##### $$fn\\n\\n" ; \
		cat "$$fn" ; \
		echo ; \
	done < <( find -name Makefile -or -name go.mod -name '*.go' -or -name '*.html' -or -name '*.js' -or -name '*.css' ) | xclip

# Kubernetes targets
build: ## Build Docker image for Kubernetes
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(FULL_IMAGE)
	@echo "✓ Image built: $(FULL_IMAGE)"

push: build ## Push image to minikube registry
	@echo "Pushing image to minikube registry..."
	docker push $(FULL_IMAGE)
	@echo "✓ Image pushed to $(REGISTRY)"

deploy: push ## Deploy/upgrade Helm chart
	@echo "Deploying to Kubernetes..."
	helm upgrade --install $(RELEASE_NAME) ./helm \
		--namespace $(NAMESPACE) \
		--create-namespace \
		--set image.repository=$(REGISTRY)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)
	@echo "✓ Deployed successfully"

undeploy: ## Remove Helm release
	@echo "Uninstalling Helm release..."
	helm uninstall $(RELEASE_NAME) -n $(NAMESPACE)
	@echo "✓ Uninstalled"

clean: ## Remove local Docker images
	@echo "Cleaning up local images..."
	-docker rmi $(IMAGE_NAME):$(IMAGE_TAG) $(FULL_IMAGE)
	@echo "✓ Cleaned"

tail-app-logs: ## Tail pod logs
	@echo "Streaming logs from poker namespace..."
	kubectl logs -n $(NAMESPACE) -l app=card-shoggoths -f --tail=50

status: ## Show deployment status
	@echo "=== Deployment Status ==="
	@kubectl get all -n $(NAMESPACE)
	@echo ""
	@echo "=== Certificate Status ==="
	@kubectl get certificate -n $(NAMESPACE) 2>/dev/null || echo "No certificates found"
	@echo ""
	@echo "=== Ingress Status ==="
	@kubectl get ingress -n $(NAMESPACE) 2>/dev/null || echo "No ingress found"

self-signed-cert-issuer: ## Create self-signed ClusterIssuer for development
	@echo "Creating self-signed ClusterIssuer..."
	@kubectl apply -f - <<EOF
	apiVersion: cert-manager.io/v1
	kind: ClusterIssuer
	metadata:
	  name: letsencrypt-prod
	spec:
	  selfSigned: {}
	EOF
	@echo "✓ ClusterIssuer created"

