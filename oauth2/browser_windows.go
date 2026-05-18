//go:build windows

package oauth2

import (
	"context"
	"os"
)

func openBrowser(ctx context.Context, rawURL string) error {
	if browser := os.Getenv("BROWSER"); browser != "" {
		return startCmd(ctx, browser, rawURL)
	}
	return startCmd(ctx, "rundll32", "url.dll,FileProtocolHandler", rawURL)
}
