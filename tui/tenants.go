package tui

import "github.com/lvlcn-t/azctx/config"

// TODO: Add edit capabilities for tenants.

// newTenantsTab creates a browse tab for tenants.
func newTenantsTab(cfg *config.Config, width, height int) browseTab {
	return newBrowseTab(buildTenantItems(cfg), width, height)
}
