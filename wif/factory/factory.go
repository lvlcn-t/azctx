// Package factory creates [wif.Provider] instances based on credential
// configuration.
package factory

import (
	"context"
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
	"github.com/lvlcn-t/azctx/wif/file"
	"github.com/lvlcn-t/azctx/wif/oauth2"
)

// Compile-time interface satisfaction checks.
var (
	_ wif.Provider  = (*oauth2.Provider)(nil)
	_ wif.Provider  = (*file.Provider)(nil)
	_ TokenProvider = NewTokenProvider
)

type TokenProvider func(ctx context.Context, cfg config.TokenDetails, cacheDir string) (wif.Provider, error)

// NewTokenProvider creates a token provider based on the configured token source.
// The cacheDir is used by the OAuth2 provider to store cached tokens.
func NewTokenProvider(ctx context.Context, cfg config.TokenDetails, cacheDir string) (wif.Provider, error) {
	switch cfg.Source {
	case config.TokenSourceFile:
		return file.NewProvider(cfg.File)
	case config.TokenSourceOAuth2:
		return oauth2.NewProvider(ctx, cfg.OAuth2, cacheDir)
	default:
		return nil, fmt.Errorf("unsupported token source: %q", cfg.Source)
	}
}
