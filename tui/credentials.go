package tui

import "github.com/lvlcn-t/azctx/config"

// TODO: Add edit capabilities for credentials.

// newCredentialsTab creates a browse tab for credentials.
func newCredentialsTab(cfg *config.Config, width, height int) browseTab {
	return newBrowseTab(buildCredentialItems(cfg), width, height)
}
