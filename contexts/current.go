package contexts

import "github.com/lvlcn-t/azctx/config"

// CurrentContextName returns the configured current-context name, or
// ErrCurrentContextUnset if it is not set.
func CurrentContextName(cfg *config.Config) (string, error) {
	if cfg == nil || cfg.CurrentContext == "" {
		return "", ErrCurrentContextUnset
	}

	return cfg.CurrentContext, nil
}
