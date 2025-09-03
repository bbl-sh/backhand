// internal/docker/runner.go

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func RunInContainer(image, command string, code, input []byte) (string, error) {
	// Create isolated temp directory
	tempDir, err := os.MkdirTemp("", "run-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir")
	}
	defer os.RemoveAll(tempDir)

	// Write code to /app/solution.py (or whatever the command expects)
	codePath := filepath.Join(tempDir, "solution.py")
	if err := os.WriteFile(codePath, code, 0644); err != nil {
		return "", fmt.Errorf("failed to write code")
	}

	// Write input to /app/input.txt
	inputPath := filepath.Join(tempDir, "input.txt")
	if err := os.WriteFile(inputPath, input, 0644); err != nil {
		return "", fmt.Errorf("failed to write input")
	}

	// Parse command: split "sh -c 'python ...'" into args
	args := append([]string{"run", "--rm",
		"-v", tempDir + ":/app",
		"-w", "/app",
		image,
	}, shellSplit(command)...)

	// Set timeout (30 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run docker command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timed out")
		}
		return "", fmt.Errorf("run failed: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// Simple shell-like arg splitter (basic version)
// Handles: "python a.py < b.txt" → ["python", "a.py", "<", "b.txt"]
// Note: Shell redirection is handled by Docker, so this is safe
func shellSplit(command string) []string {
	return strings.Fields(command)
}
