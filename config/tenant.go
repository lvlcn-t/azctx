package config

// Tenant represents a named Azure tenant definition.
type Tenant struct {
	Name    string        `yaml:"name" json:"name"`
	Details TenantDetails `yaml:"tenant" json:"tenant"`
}

// TenantDetails represents Azure tenant details.
type TenantDetails struct {
	ID string `yaml:"id" json:"id"`
}
