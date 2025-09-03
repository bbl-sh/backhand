// internal/docker/runner.go
package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func RunInContainer(image, command string, code, input []byte) (string, error) {
	// Create isolated temp directory
	tempDir, err := os.MkdirTemp("", "run-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure container user can traverse/read the bind mount
	if err := os.Chmod(tempDir, 0o755); err != nil {
		return "", fmt.Errorf("chmod temp dir failed: %w", err)
	}

	// Write files
	codePath := filepath.Join(tempDir, "solution.py")
	if err := os.WriteFile(codePath, code, 0o644); err != nil {
		return "", fmt.Errorf("failed to write code: %w", err)
	}
	inputPath := filepath.Join(tempDir, "input.txt")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		return "", fmt.Errorf("failed to write input: %w", err)
	}

	// Find docker
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return "", fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Build mount (add :Z on Linux to be safe with SELinux-enforcing hosts)
	mountSpec := tempDir + ":/app"
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/sys/fs/selinux"); err == nil {
			mountSpec += ":Z"
		}
	}

	// Diagnostics buffer (also stream to stderr)
	var diag bytes.Buffer
	diag.WriteString("diagnostics:\n")
	diag.WriteString("docker=" + dockerPath + "\n")
	diag.WriteString("DOCKER_HOST=" + os.Getenv("DOCKER_HOST") + "\n")

	// Helper to run a command and stream output
	runCmd := func(ctx context.Context, name string, args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Env = os.Environ()
		var buf bytes.Buffer
		tee := io.MultiWriter(&buf, os.Stderr)
		cmd.Stdout = tee
		cmd.Stderr = tee
		if err := cmd.Run(); err != nil {
			return buf.String(), fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
		}
		return buf.String(), nil
	}

	// Check docker connectivity early
	if out, err := runCmd(context.Background(), dockerPath, "version", "--format", "{{.Server.Version}}"); err == nil {
		diag.WriteString("docker_server_version=" + strings.TrimSpace(out) + "\n")
	} else {
		diag.WriteString("docker_version_error=" + err.Error() + "\n")
		return "", fmt.Errorf("docker not reachable from process; ensure this process has docker group membership and the daemon is running\n%s", diag.String())
	}

	if _, err := runCmd(context.Background(), dockerPath, "info", "--format", "OK"); err != nil {
		diag.WriteString("docker_info_error=" + err.Error() + "\n")
		return "", fmt.Errorf("docker daemon not accessible (likely permissions to docker.sock); restart the server after adding user to docker group\n%s", diag.String())
	} else {
		diag.WriteString("docker_info=OK\n")
	}

	// Pre-pull if needed with generous timeout
	if _, err := runCmd(context.Background(), dockerPath, "image", "inspect", image); err != nil {
		pctx, pcancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer pcancel()
		if out, perr := runCmd(pctx, dockerPath, "pull", image); perr != nil {
			if pctx.Err() == context.DeadlineExceeded {
				return "", fmt.Errorf("docker pull timed out (5m)\n%s\npull:\n%s", diag.String(), out)
			}
			return "", fmt.Errorf("docker pull failed: %v\n%s\npull:\n%s", perr, diag.String(), out)
		}
	}

	// Build docker run (always via a shell so redirection works)
	runArgs := []string{
		"run", "--rm",
		"-v", mountSpec,
		"-w", "/app",
		image,
		"sh", "-c", command,
	}

	// Provide stdin file for commands that read from stdin
	inF, _ := os.Open(inputPath)
	defer func() {
		if inF != nil {
			_ = inF.Close()
		}
	}()

	// Execute with longer timeout and stream logs
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, dockerPath, runArgs...)
	cmd.Env = os.Environ()
	if inF != nil {
		cmd.Stdin = inF
	}
	var outBuf bytes.Buffer
	tee := io.MultiWriter(&outBuf, os.Stderr)
	cmd.Stdout = tee
	cmd.Stderr = tee

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("execution timed out (2m)\n%s\nrun:\n%s", diag.String(), outBuf.String())
		}
		return "", fmt.Errorf("docker run failed: %v\n%s\nrun:\n%s", err, diag.String(), outBuf.String())
	}
	return strings.TrimSpace(outBuf.String()), nil
}
