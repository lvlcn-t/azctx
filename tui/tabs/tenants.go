package tabs

import "github.com/lvlcn-t/azctx/tui/state"

var _ Tab = (*TenantsTab)(nil)

type TenantsTab struct {
	browseTab
}

func tenantsTab(s *state.UI, l listBuilder) *TenantsTab { //nolint:gocritic // irrelevant on startup
	return &TenantsTab{
		browseTab: newCRUDBrowseTab(s, tenantItems, l),
	}
}
