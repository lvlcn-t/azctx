package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/tui/styles"
)

type listBuilder struct {
	delegate  list.ItemDelegate
	title     string
	items     []list.Item
	width     int
	height    int
	statusBar bool
	help      bool
	filtering bool
}

func newList(width, height int) listBuilder {
	return listBuilder{
		width:    width,
		height:   height,
		delegate: styles.NewAzureDelegate(),
	}
}

func (b listBuilder) WithItems(i ...list.Item) listBuilder { //nolint:gocritic // irrelevant on startup
	b.items = append(b.items, i...)
	return b
}

func (b listBuilder) WithDelegate(d list.ItemDelegate) listBuilder { //nolint:gocritic // irrelevant on startup
	b.delegate = d
	return b
}

func (b listBuilder) WithTitle(title string) listBuilder { //nolint:gocritic // irrelevant on startup
	b.title = title
	return b
}

func (b listBuilder) ShowStatusBar(show bool) listBuilder { //nolint:gocritic // irrelevant on startup
	b.statusBar = show
	return b
}

func (b listBuilder) ShowHelp(show bool) listBuilder { //nolint:gocritic // irrelevant on startup
	b.help = show
	return b
}

func (b listBuilder) EnableFiltering(enabled bool) listBuilder { //nolint:gocritic // irrelevant on startup
	b.filtering = enabled
	return b
}

func (b listBuilder) Build() list.Model { //nolint:gocritic // irrelevant on startup
	l := list.New(b.items, b.delegate, b.width, b.height)
	l.Title = b.title
	l.SetShowTitle(b.title != "")
	l.SetShowStatusBar(b.statusBar)
	l.SetShowHelp(b.help)
	l.SetFilteringEnabled(b.filtering)
	return l
}
