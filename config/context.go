package config

// Context represents a named azctx context entry.
type Context struct {
	Name    string         `yaml:"name" json:"name"`
	Context ContextDetails `yaml:"context" json:"context"`
}

// ContextDetails represents the details of an azctx context.
type ContextDetails struct {
	Tenant               string `yaml:"tenant" json:"tenant"`
	Credential           string `yaml:"credential" json:"credential"`
	Subscription         string `yaml:"subscription,omitempty" json:"subscription,omitempty"`
	AllowNoSubscriptions bool   `yaml:"allow-no-subscriptions,omitempty" json:"allowNoSubscriptions,omitempty"`
}
