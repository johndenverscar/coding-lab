# qbert

A lightweight Kubernetes API server that provides HTTP endpoints to query and modify deployment replica counts.

## Prerequisites

- Go 1.25+
- Docker
- kubectl
- Local Kubernetes cluster (Colima, KIND, or similar)

## Project Structure

```
qbert/
├── cmd/
│   └── server/
│       └── main.go          # HTTP server entry point
├── pkg/
│   ├── api/
│   │   └── handler.go       # HTTP handlers
│   └── k8s/
│       ├── client.go        # Kubernetes client wrapper
│       └── client_test.go   # Unit tests
├── deploy/
│   └── deployment.yaml      # Kubernetes manifests
├── Dockerfile               # Multi-stage Docker build
├── go.mod                   # Go module definition
└── go.sum                   # Go module checksums
```

## Running Locally

### Option 1: Run directly with Go

```bash
# Install dependencies
go mod download

# Run the server (requires local kubeconfig)
go run cmd/server/main.go
```

The server will start on port 8080 and use your local `~/.kube/config` for authentication.

### Option 2: Run tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/k8s
```

## Building and Deploying to Kubernetes

### 1. Build Docker Image

```bash
docker build -t qbert:latest .
```

### 2. Deploy to Kubernetes

```bash
# Apply all manifests (ServiceAccount, RBAC, Deployment, Service)
kubectl apply -f deploy/deployment.yaml

# Verify deployment
kubectl get pods -l app=qbert
kubectl get svc qbert
```

### 3. Check Logs

```bash
# View application logs
kubectl logs -l app=qbert

# Follow logs in real-time
kubectl logs -f -l app=qbert
```

### 4. Update After Code Changes

```bash
# Rebuild and restart
docker build -t qbert:latest .
kubectl rollout restart deployment qbert

# Watch rollout status
kubectl rollout status deployment qbert
```

## API Endpoints

### 1. Health Check

Check if the service is running.

**Endpoint:** `GET /health`

**Example:**
```bash
curl http://localhost:30080/health
```

**Response:**
```
ok
```

### 2. Get Deployment Replica Count

Retrieve the replica count for a specific deployment.

**Endpoint:** `GET /deployments/{namespace}/{name}/replicas`

**Parameters:**
- `namespace` - Kubernetes namespace (e.g., `default`, `kube-system`)
- `name` - Deployment name (e.g., `qbert`, `coredns`)

**Example - Success:**
```bash
curl http://localhost:30080/deployments/default/qbert/replicas
```

**Response:**
```json
{
  "replicas": 1
}
```

**Example - Not Found:**
```bash
curl http://localhost:30080/deployments/default/nonexistent/replicas
```

**Response:**
```json
{
  "error": "deployment not found"
}
```

**Example - Query Different Namespace:**
```bash
curl http://localhost:30080/deployments/kube-system/coredns/replicas
```

### 3. Set Deployment Replica Count

Update the replica count for a specific deployment.

**Endpoint:** `PUT /deployments/{namespace}/{name}/replicas`

**Parameters:**
- `namespace` - Kubernetes namespace (e.g., `default`, `kube-system`)
- `name` - Deployment name (e.g., `nginx`, `qbert`)

**Request Body:**
```json
{
  "replicas": 5
}
```

**Example - Success:**
```bash
curl -X PUT -H "Content-Type: application/json" \
  -d '{"replicas": 5}' \
  http://localhost:30080/deployments/default/nginx/replicas
```

**Response:**
```json
{
  "message": "replica count updated successfully"
}
```

**Example - Scale to Zero:**
```bash
curl -X PUT -H "Content-Type: application/json" \
  -d '{"replicas": 0}' \
  http://localhost:30080/deployments/default/nginx/replicas
```

**Example - Invalid Input (Negative Replicas):**
```bash
curl -X PUT -H "Content-Type: application/json" \
  -d '{"replicas": -1}' \
  http://localhost:30080/deployments/default/nginx/replicas
```

**Response:**
```json
{
  "error": "replica count cannot be negative"
}
```

**Example - Deployment Not Found:**
```bash
curl -X PUT -H "Content-Type: application/json" \
  -d '{"replicas": 5}' \
  http://localhost:30080/deployments/default/nonexistent/replicas
```

**Response:**
```json
{
  "error": "deployment not found"
}
```

**Verify the Change:**
```bash
# Check the deployment status
kubectl get deployment nginx

# Or query via API
curl http://localhost:30080/deployments/default/nginx/replicas
```

## RBAC Permissions

The qbert service requires specific Kubernetes permissions to function:

**ServiceAccount**: `qbert`
**ClusterRole Permissions**:
- **apiGroups**: `apps`
- **resources**: `deployments`
- **verbs**: `get`, `list`, `update`, `patch`

These permissions allow qbert to:
- Read deployment information across all namespaces (`get`, `list`)
- Modify deployment replica counts (`update`, `patch`)

The RBAC configuration is included in [deploy/deployment.yaml](deploy/deployment.yaml) and is automatically applied when you deploy the service.

## Accessing the Service

### Local Development (NodePort)

The service is exposed as a NodePort on port 30080:

```bash
# Direct access (no port-forward needed)
curl http://localhost:30080/health
curl http://localhost:30080/deployments/default/qbert/replicas

# Set replica count
curl -X PUT -H "Content-Type: application/json" \
  -d '{"replicas": 3}' \
  http://localhost:30080/deployments/default/qbert/replicas
```