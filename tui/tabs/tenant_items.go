package tabs

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
)

var (
	_ list.Item        = (*TenantItem)(nil)
	_ list.DefaultItem = (*TenantItem)(nil)
	_ details.Item     = (*TenantItem)(nil)
)

type TenantItem struct{ config.Tenant }

func tenantItems(s *config.Store) []list.Item {
	items := make([]list.Item, 0, len(s.Config.Tenants))
	for _, t := range s.Config.Tenants {
		i := TenantItem{t}
		items = append(items, &i)
	}
	return items
}

func (i *TenantItem) Title() string       { return i.Name }
func (i *TenantItem) Description() string { return i.Tenant.Details.ID }
func (i *TenantItem) FilterValue() string {
	return i.Name + " " + i.Tenant.Details.ID
}

func (i *TenantItem) Details() details.View {
	return details.View{
		Title: "Tenant: " + i.Name,
		Rows: []details.Row{
			{Label: labelName, Value: i.Name},
			{Label: "ID", Value: i.Tenant.Details.ID},
		},
	}
}

// tenantForm builds the create/edit form for a tenant. When item is a
// *TenantItem the fields are pre-filled for editing.
func tenantForm(item details.Item) form.Model {
	var name, id string
	title := "New tenant"
	if tenant, ok := item.(*TenantItem); ok {
		name = tenant.Name
		id = tenant.Tenant.Details.ID
		title = "Edit tenant"
	}

	return form.New(title, []form.Field{
		{Key: "name", Label: labelName, Placeholder: "corp", Value: name, Required: true},
		{Key: "id", Label: "ID", Placeholder: "00000000-0000-0000-0000-000000000000", Value: id, Required: true},
	})
}
