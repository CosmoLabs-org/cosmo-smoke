package runner

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/CosmoLabs-org/cosmo-smoke/internal/schema"
)

// isDockerAvailable returns true if the docker daemon is reachable.
func isDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "docker", "info").Run() == nil
}

// CheckDockerContainerRunning checks if a named Docker container is running.
func CheckDockerContainerRunning(check *schema.DockerContainerCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{.State.Running}}", check.Name).Output()
	if err != nil {
		return AssertionResult{Type: "docker_container_running", Expected: check.Name, Actual: "container not found or docker unavailable: " + err.Error(), Passed: false}
	}
	running := strings.TrimSpace(string(out))
	if running != "true" {
		return AssertionResult{Type: "docker_container_running", Expected: "true", Actual: running, Passed: false}
	}
	return AssertionResult{Type: "docker_container_running", Expected: check.Name, Actual: "running", Passed: true}
}

// CheckDockerImageExists checks if a Docker image exists locally.
func CheckDockerImageExists(check *schema.DockerImageCheck) AssertionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := exec.CommandContext(ctx, "docker", "image", "inspect", check.Image).Run()
	if err != nil {
		return AssertionResult{Type: "docker_image_exists", Expected: check.Image, Actual: "image not found or docker unavailable: " + err.Error(), Passed: false}
	}
	return AssertionResult{Type: "docker_image_exists", Expected: check.Image, Actual: "exists", Passed: true}
}
