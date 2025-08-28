package judge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golangXback/internal/config"
)

// RunSubmission runs the submitted code in Docker using the ProblemConfig.
// Returns stdout output (string) or an error (including stderr).
func RunSubmission(problemID int, code []byte) (string, error) {
	prb, ok := config.ProblemConfigs[problemID]
	if !ok {
		return "", fmt.Errorf("problem not found: %d", problemID)
	}

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "judge-"+fmt.Sprint(problemID)+"-")
	if err != nil {
		return "", fmt.Errorf("tmpdir: %w", err)
	}
	// clean up
	defer os.RemoveAll(tmpDir)

	// write files
	codePath := filepath.Join(tmpDir, "solution.py")
	if err := os.WriteFile(codePath, code, 0644); err != nil {
		return "", fmt.Errorf("write code: %w", err)
	}
	inputPath := filepath.Join(tmpDir, "input.txt")
	if err := os.WriteFile(inputPath, []byte(prb.Input), 0644); err != nil {
		return "", fmt.Errorf("write input: %w", err)
	}

	// Build docker command:
	// docker run --rm -v tmpDir:/app <image> sh -c "<command>"
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"-v", tmpDir+":/app",
		prb.Image,
		"sh", "-c", prb.Command,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// include stderr for debugging
		return "", fmt.Errorf("docker run failed: %w - %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetExpectedOutput returns expected output for a problem (or empty string)
func GetExpectedOutput(problemID int) string {
	if pr, ok := config.ProblemConfigs[problemID]; ok {
		return strings.TrimSpace(pr.ExpectedOutput)
	}
	return ""
}
