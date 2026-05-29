package tabs

import "github.com/lvlcn-t/azctx/tui/state"

var _ Tab = (*TenantsTab)(nil)

// TODO: Add edit capabilities for tenants.

type TenantsTab struct {
	browse
}

func tenantsTab(s *state.UI, l listBuilder) *TenantsTab { //nolint:gocritic // irrelevant on startup
	items := tenantItems(s.Config())
	return &TenantsTab{
		browse: newBrowse(l.WithItems(items...)),
	}
}
