// docker/docker.go
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

// RunInContainer runs code in a Docker container with /app/solution.py and /app/input.txt
func RunInContainer(image, command string, code, input []byte) (string, error) {

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "run-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir")
	}
	defer os.RemoveAll(tempDir)

	// Write solution.py
	if err := os.WriteFile(filepath.Join(tempDir, "solution.py"), code, 0644); err != nil {
		return "", fmt.Errorf("failed to write code")
	}

	// Write input.txt
	if err := os.WriteFile(filepath.Join(tempDir, "input.txt"), input, 0644); err != nil {
		return "", fmt.Errorf("failed to write input")
	}

	// Run Docker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx,
		"docker", "run", "--rm",
		"-v", tempDir+":/app",
		"-w", "/app",
		image,
		"sh", "-c", command,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("container timed out")
		}
		return "", fmt.Errorf("run failed: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
