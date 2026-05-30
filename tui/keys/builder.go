package keys

import (
	"cmp"

	"github.com/charmbracelet/bubbles/key"
)

type Builder struct {
	help    string
	primary key.Binding
	aliases []key.Binding
}

func New(k key.Binding) Builder {
	return Builder{
		primary: k,
	}
}

func (k Builder) WithAliases(aliases ...key.Binding) Builder { //nolint:gocritic // irrelevant on startup
	k.aliases = aliases
	return k
}

func (k Builder) DefaultHelp() Builder { //nolint:gocritic // irrelevant on startup
	k.help = cmp.Or(k.help, k.primary.Help().Desc)
	return k
}

func (k Builder) WithHelp(help string) Builder { //nolint:gocritic // irrelevant on startup
	k.help = help
	return k
}

func (k Builder) Bind() key.Binding { //nolint:gocritic // irrelevant on startup
	keys := k.primary.Keys()
	for _, alias := range k.aliases {
		keys = append(keys, alias.Keys()...)
	}

	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(k.primary.Help().Key, k.help),
	)
}
