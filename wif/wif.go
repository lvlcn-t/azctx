// Package wif provides federated identity token acquisition for Azure
// workload identity login.
package wif

import (
	"context"
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif/file"
	"github.com/lvlcn-t/azctx/wif/oauth2"
)

// Compile-time interface satisfaction checks.
var (
	_ Provider = (*oauth2.Provider)(nil)
	_ Provider = (*file.Provider)(nil)
)

// Provider acquires a federated identity token for workload identity login.
//
//go:generate go tool moq -out wif_moq.go . Provider
type Provider interface {
	// AcquireToken acquires and returns a federated identity token.
	AcquireToken(ctx context.Context) (string, error)
}

// NewProvider creates a token provider based on the configured token source.
func NewProvider(ctx context.Context, cfg config.TokenDetails) (Provider, error) {
	switch cfg.Source {
	case config.TokenSourceFile:
		return file.NewProvider(cfg.File)
	case config.TokenSourceOAuth2:
		return oauth2.NewProvider(ctx, cfg.OAuth2)
	default:
		return nil, fmt.Errorf("unsupported token source: %q", cfg.Source)
	}
}
