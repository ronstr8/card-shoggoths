# Card Shoggoths Helm Chart

Deploy Card Shoggoths poker game to Kubernetes with this Helm chart.

## Quick Start

```bash
# Enable minikube registry
minikube addons enable registry

# Build, push, and deploy (uses version from Go source: 1.0.0)
make deploy

# View status
make status

# Stream logs
make tail-app-logs
```

## Prerequisites

- Kubernetes 1.19+
- Helm 3+
- Nginx Ingress Controller installed in your cluster
- minikube (for local development)
- cert-manager (for TLS certificates)

## Version Management

**Version is defined in:** `cmd/card-shoggoths-server/version.go`

The Makefile automatically extracts the version and uses it for Docker image tags. See [VERSION.md](../VERSION.md) for details.

```bash
# Check version
go run ./cmd/card-shoggoths-server --version

# Deploy uses this version automatically
make deploy  # Tags as localhost:5000/card-shoggoths:1.0.0
```

## Installation

### 1. Install cert-manager

```bash
# Verify cert-manager is installed
kubectl get pods -n cert-manager

# If not installed:
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

### 2. Create ClusterIssuer

**For development (self-signed):**

```bash
make self-signed-cert-issuer
```

**For production (Let's Encrypt):**

```bash
kubectl apply -f helm/letsencrypt-issuer.yaml
```

Update `helm/values.yaml` to set your email and configure DNS for HTTP-01 validation.

### 3. Deploy Application

```bash
# Build, push, and deploy
make deploy

# Check deployment status
make status
```

### 4. Configure DNS/Hosts

**For minikube (local development):**

```bash
# Get minikube IP
minikube ip

# Add to /etc/hosts (Linux/Mac) or C:\Windows\System32\drivers\etc\hosts (Windows)
echo "$(minikube ip) poker.fazigu.org" | sudo tee -a /etc/hosts
```

**For production:**

Point your DNS A record to your cluster's external IP.

### 5. Enable Ingress

```bash
minikube addons enable ingress
```

## Access the Application

- HTTP: <http://poker.fazigu.org>
- HTTPS: <https://poker.fazigu.org>

## Configuration

See [values.yaml](values.yaml) for all configurable parameters.

### Common Customizations

```yaml
# Custom image tag
image:
  tag: "1.1.0"

# Change pull policy (Always for dev, IfNotPresent for prod)
image:
  pullPolicy: Always

# Disable persistence (ephemeral storage)
persistence:
  enabled: false

# Use different cert-manager issuer
certManager:
  issuer:
    name: selfsigned-issuer
    kind: Issuer  # or ClusterIssuer

# Custom resource limits
resources:
  limits:
    cpu: 1000m
    memory: 512Mi
```

## Makefile Targets

Run `make help` to see all available targets.

### Kubernetes Deployment

| Target | Description |
|--------|-------------|
| `make build` | Build Docker image (version from Go source) |
| `make push` | Build and push to localhost:5000 registry |
| `make deploy` | Build, push, and deploy/upgrade Helm chart |
| `make undeploy` | Remove Helm release |
| `make clean` | Remove local Docker images |

### Monitoring

| Target | Description |
|--------|-------------|
| `make status` | Show all Kubernetes resources |
| `make tail-app-logs` | Stream pod logs |

### Utilities

| Target | Description |
|--------|-------------|
| `make self-signed-cert-issuer` | Create self-signed ClusterIssuer |
| `make help` | Show all available targets |

### Local Development

| Target | Description |
|--------|-------------|
| `make run` | Run server locally |
| `make dev` | Run with hot reload (air) |
| `make test` | Run tests |

## Uninstall

```bash
make undeploy
kubectl delete namespace poker
```

## Troubleshooting

### Check pod status

```bash
kubectl get pods -n poker
kubectl logs -n poker -l app=card-shoggoths --tail=50
```

### Check ingress

```bash
kubectl get ingress -n poker
kubectl describe ingress poker-card-shoggoths -n poker
```

### Test service directly

```bash
kubectl port-forward -n poker svc/poker-card-shoggoths 8080:8080
# Access at http://localhost:8080
```

### Image not updating

If pods aren't picking up new code:

```bash
# Ensure ImagePullPolicy is set to Always in values.yaml
# Force recreate pods
kubectl rollout restart deployment poker-card-shoggoths -n poker

# Or delete and redeploy
make undeploy
make deploy
```

### Certificate Issues

```bash
# Check Certificate resource
kubectl get certificate -n poker
kubectl describe certificate poker-card-shoggoths-tls -n poker

# Check orders and challenges
kubectl get orders,challenges -n poker
kubectl describe challenge -n poker

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager --tail=100

# Verify secret was created
kubectl get secret poker-tls -n poker -o yaml
```

**Common cert-manager issues:**

- **Self-check fails (hairpin NAT)**: Use self-signed issuer for local development
- **HTTP-01 validation fails**: Ensure port 80 is accessible from the internet for Let's Encrypt
- **Rate limiting**: Let's Encrypt has rate limits; use staging server for testing

### WebSocket Issues

If WebSocket connections fail:

```bash
# Verify ingress annotations
kubectl get ingress poker-card-shoggoths -n poker -o yaml | grep websocket

# Check nginx ingress config
kubectl get configmap -n ingress-nginx ingress-nginx-controller -o yaml
```

The ingress should have:

```yaml
annotations:
  nginx.ingress.kubernetes.io/websocket-services: card-shoggoths
```

## Development Workflow

```bash
# 1. Make code changes
vim cmd/card-shoggoths-server/main.go

# 2. Update version (if releasing)
vim cmd/card-shoggoths-server/version.go

# 3. Deploy
make deploy

# 4. Watch logs
make tail-app-logs

# 5. Test
curl https://poker.fazigu.org
```
