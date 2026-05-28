package az

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lvlcn-t/azctx/semver"
)

// ensureInstalled validates that az is available in PATH.
func ensureInstalled() error {
	if _, err := exec.LookPath("az"); err != nil {
		return fmt.Errorf("%w in PATH. install it from %s", errCLIUnavailable, azInstallURL)
	}

	return nil
}

// az executes one Azure CLI command.
func az(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, "az", args...)

	var stderr bytes.Buffer
	command.Stdout = os.Stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return fmt.Errorf("az %s failed: %w: %s", redactArgs(args), err, stderrText)
		}

		return fmt.Errorf("az %s failed: %w", redactArgs(args), err)
	}

	return nil
}

// azVersion returns the installed Azure CLI version.
func azVersion(ctx context.Context) (semver.Version, error) {
	cmd := exec.CommandContext(ctx, "az", "version", "-o", "json")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run az version: %w: %s", err, stderr.String())
	}

	var result struct {
		AzureCLI string `json:"azure-cli"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("parse az version output: %w", err)
	}

	if result.AzureCLI == "" {
		return "", errors.New("az version output did not contain azure-cli version")
	}

	v := semver.Version(result.AzureCLI)
	if !v.IsValid() {
		return "", fmt.Errorf("invalid azure-cli version: %q", result.AzureCLI)
	}

	return v, nil
}
