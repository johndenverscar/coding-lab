#!/bin/bash
set -e

CLUSTER_NAME="qbert-test"

echo "=== Step 1: Creating KIND cluster ==="
kind get clusters | grep "$CLUSTER_NAME" || kind create cluster --name "$CLUSTER_NAME"

echo ""
echo "=== Step 2: Building Docker image ==="
docker build -t qbert:test .

echo ""
echo "=== Step 3: Loading image into KIND cluster ==="
kind load docker-image qbert:test --name "$CLUSTER_NAME"

echo ""
echo "=== Step 4: Deploying qbert to cluster ==="
kubectl apply -f deploy/

echo ""
echo "=== Step 5: Waiting for deployment to be ready ==="
kubectl wait --for=condition=available deployment/qbert --timeout=60s

echo ""
echo "=== Step 6: Running integration tests ==="
go test -v ./test/integration_test.go

# Capture test exit code
TEST_EXIT_CODE=$?

echo ""
echo "=== Step 7: Cleaning up Kubernetes resources ==="
kubectl delete -f deploy/

echo ""
echo "=== Step 8: Deleting KIND cluster ==="
kind delete cluster --name "$CLUSTER_NAME"

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "✓ Integration tests PASSED"
else
  echo "✗ Integration tests FAILED"
fi

exit $TEST_EXIT_CODE