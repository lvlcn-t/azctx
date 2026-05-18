package oauth2

import (
	"context"
	"fmt"
	"os/exec"
)

// startCmd launches a command with the given arguments without waiting
// for it to complete.
func startCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // args are not user-controlled; they come from platform defaults or $BROWSER
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open URL in browser: %w", err)
	}
	return nil
}
