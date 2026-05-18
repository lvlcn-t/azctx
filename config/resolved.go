package config

import "fmt"

// ResolvedContext is a fully dereferenced context with its tenant and
// credential resolved from the config.
type ResolvedContext struct {
	Name                 string
	Tenant               Tenant
	Credential           Credential
	Subscription         string
	AllowNoSubscriptions bool
}

// Resolve looks up a context by name and resolves its tenant and
// credential references, returning a validated [ResolvedContext].
func (s *Store) Resolve(name string) (ResolvedContext, error) {
	ctx, found := s.Config.ContextByName(name)
	if !found {
		return ResolvedContext{}, fmt.Errorf("context %q not found", name)
	}

	tenant, found := s.Config.TenantByName(ctx.Details.Tenant)
	if !found {
		return ResolvedContext{}, fmt.Errorf(
			"tenant %q not found for context %q",
			ctx.Details.Tenant, ctx.Name,
		)
	}

	credential, found := s.Config.CredentialByName(ctx.Details.Credential)
	if !found {
		return ResolvedContext{}, fmt.Errorf(
			"credential %q not found for context %q",
			ctx.Details.Credential, ctx.Name,
		)
	}

	if tenant.Details.ID == "" {
		return ResolvedContext{}, fmt.Errorf(
			"tenant %q is missing id", tenant.Name,
		)
	}

	return ResolvedContext{
		Name:                 ctx.Name,
		Tenant:               tenant,
		Credential:           credential,
		Subscription:         ctx.Details.Subscription,
		AllowNoSubscriptions: ctx.Details.AllowNoSubscriptions,
	}, nil
}
