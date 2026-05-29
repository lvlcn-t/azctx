package tabs

import "github.com/lvlcn-t/azctx/tui/state"

var _ Tab = (*CredentialsTab)(nil)

// TODO: Add edit capabilities for tenants.

type CredentialsTab struct {
	browseTab
}

func credentialsTab(s *state.UI, l listBuilder) *CredentialsTab { //nolint:gocritic // irrelevant on startup
	items := credentialItems(s.Config())
	return &CredentialsTab{
		browseTab: newBrowseTab(l.WithItems(items...)),
	}
}
