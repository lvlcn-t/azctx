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

	"golang.org/x/mod/semver"
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

// version represents an Azure CLI version.
type version string

// String returns the version as a string.
func (v version) String() string {
	return string(v)
}

// semver returns the version in semver format, ensuring it has a "v" prefix.
func (v version) semver() string {
	if strings.HasPrefix(string(v), "v") {
		return string(v)
	}
	return "v" + string(v)
}

// AtLeast reports whether the version is greater than or equal to another version.
func (v version) AtLeast(other version) bool {
	return semver.Compare(v.semver(), other.semver()) >= 0
}

// supportsScopedLogin reports whether the Azure CLI version supports scoped login with tenant and credential.
// All versions after 2.86.0 support this feature, which allows us to avoid calling az account set for those versions.
func (v version) supportsScopedLogin() bool {
	return v.AtLeast(version("2.86.0"))
}

// azVersion returns the installed Azure CLI version.
func azVersion(ctx context.Context) (version, error) {
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

	v := version(result.AzureCLI)
	if !semver.IsValid(v.semver()) {
		return "", fmt.Errorf("invalid azure-cli version: %q", result.AzureCLI)
	}

	return v, nil
}
