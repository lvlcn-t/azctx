package tabs

import "github.com/lvlcn-t/azctx/tui/state"

var _ Tab = (*CredentialsTab)(nil)

type CredentialsTab struct {
	browseTab
}

func credentialsTab(s *state.UI, l listBuilder) *CredentialsTab { //nolint:gocritic // irrelevant on startup
	return &CredentialsTab{
		browseTab: newBrowseTab(s, credentialItems, l),
	}
}
