package cmd

import (
	"errors"

	"github.com/lvlcn-t/azctx/config"
)

// errCurrentContextUnset indicates that current-context is not configured.
var errCurrentContextUnset = errors.New("current-context is not set")

// mustCurrentContextName returns the configured current-context name.
func mustCurrentContextName(cfg *config.Config) (string, error) {
	if cfg == nil {
		return "", errCurrentContextUnset
	}

	if cfg.CurrentContext == "" {
		return "", errCurrentContextUnset
	}

	return cfg.CurrentContext, nil
}
