package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

// createTestDeployment creates a simple nginx deployment for testing
func createTestDeployment(namespace, name string, replicas int32) error {
	// First, ensure the namespace exists
	if err := ensureNamespace(namespace); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	manifest := fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
`, name, namespace, replicas, name, name)

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w, output: %s", err, string(output))
	}
	return nil
}

// ensureNamespace creates a namespace if it doesn't exist
func ensureNamespace(namespace string) error {
	// Check if namespace exists
	cmd := exec.Command("kubectl", "get", "namespace", namespace)
	if err := cmd.Run(); err == nil {
		// Namespace already exists
		return nil
	}

	// Create the namespace
	cmd = exec.Command("kubectl", "create", "namespace", namespace)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w, output: %s", err, string(output))
	}
	return nil
}

// deleteTestDeployment removes a test deployment
func deleteTestDeployment(namespace, name string) error {
	cmd := exec.Command("kubectl", "delete", "deployment", name, "-n", namespace, "--ignore-not-found=true")
	return cmd.Run()
}

// deleteNamespace removes a namespace
func deleteNamespace(namespace string) error {
	cmd := exec.Command("kubectl", "delete", "namespace", namespace, "--ignore-not-found=true")
	return cmd.Run()
}

// waitForDeployment waits for a deployment to be ready
func waitForDeployment(namespace, name string, timeout time.Duration) error {
	cmd := exec.Command("kubectl", "wait",
		"--for=condition=available",
		fmt.Sprintf("deployment/%s", name),
		"-n", namespace,
		fmt.Sprintf("--timeout=%s", timeout))
	return cmd.Run()
}

// getDeploymentReplicas gets current replica count via kubectl
func getDeploymentReplicas(namespace, name string) (int32, error) {
	cmd := exec.Command("kubectl", "get", "deployment", name, "-n", namespace,
		"-o", "jsonpath={.spec.replicas}")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var replicas int32
	fmt.Sscanf(string(out), "%d", &replicas)
	return replicas, nil
}

// waitForDeploymentReplicas waits for a deployment to reach a specific replica count
func waitForDeploymentReplicas(namespace, name string, expectedReplicas int32, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		currentReplicas, err := getDeploymentReplicas(namespace, name)
		if err != nil {
			// If we can't get the replicas, wait and try again
			time.Sleep(1 * time.Second)
			continue
		}

		if currentReplicas == expectedReplicas {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	// Timeout reached
	currentReplicas, _ := getDeploymentReplicas(namespace, name)
	return fmt.Errorf("timeout waiting for deployment %s/%s to reach %d replicas (current: %d)",
		namespace, name, expectedReplicas, currentReplicas)
}
