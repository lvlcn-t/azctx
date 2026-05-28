package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

type testConfig struct {
	t   testing.TB
	cfg Config
}

func newTestConfig(t testing.TB) *testConfig {
	t.Helper()
	return &testConfig{
		t: t,
		cfg: Config{
			APIVersion: APIVersion,
			Kind:       Kind,
		},
	}
}

func (t *testConfig) withTenant(name string) *testConfig {
	t.t.Helper()
	t.cfg.Tenants = append(t.cfg.Tenants, Tenant{Name: name, Details: TenantDetails{ID: newUUID()}})
	return t
}

func (t *testConfig) withUserCredential(name string) *testConfig {
	t.t.Helper()
	t.cfg.Credentials = append(t.cfg.Credentials, newCredential(t.t, name, CredentialTypeUser))
	return t
}

func (t *testConfig) withSPCredential(name string, opts ...credentialOption) *testConfig {
	t.t.Helper()
	t.cfg.Credentials = append(t.cfg.Credentials, newCredential(t.t, name, CredentialTypeServicePrincipal, opts...))
	return t
}

func (t *testConfig) withMICredential(name string, opts ...credentialOption) *testConfig {
	t.t.Helper()
	t.cfg.Credentials = append(t.cfg.Credentials, newCredential(t.t, name, CredentialTypeManagedIdentity, opts...))
	return t
}

func (t *testConfig) withWIFCredential(name string, opts ...credentialOption) *testConfig {
	t.t.Helper()
	t.cfg.Credentials = append(t.cfg.Credentials, newCredential(t.t, name, CredentialTypeWorkloadIdentity, opts...))
	return t
}

func (t *testConfig) withContext(name, tenant, credential string, opts ...contextOption) *testConfig {
	t.t.Helper()
	t.cfg.Contexts = append(t.cfg.Contexts, newContext(t.t, name, tenant, credential, opts...))
	return t
}

func (t *testConfig) withCurrentContext(name string) *testConfig {
	t.t.Helper()
	t.cfg.CurrentContext = name
	return t
}

func (t *testConfig) build() Config {
	t.t.Helper()
	return t.cfg
}

func (t *testConfig) yaml() []byte {
	t.t.Helper()

	bytes, err := yaml.Marshal(t.cfg)
	require.NoError(t.t, err)

	return bytes
}

func (t *testConfig) write(fs afero.Fs, path string) *testConfig {
	t.t.Helper()

	const (
		dirMode  = 0o700
		fileMode = 0o600
	)
	require.NoError(t.t, fs.MkdirAll(filepath.Dir(path), dirMode))
	require.NoError(t.t, afero.WriteFile(fs, path, t.yaml(), fileMode))
	return t
}

func newUUID() string {
	return uuid.NewString()
}

func newTenant(t testing.TB, name string) Tenant {
	t.Helper()

	return Tenant{
		Name: name,
		Details: TenantDetails{
			ID: newUUID(),
		},
	}
}

type credentialOption func(*CredentialDetails)

func withClientID(id string) credentialOption {
	return func(details *CredentialDetails) {
		details.Azure.ClientID = id
	}
}

func withClientSecret(s string) credentialOption {
	return func(details *CredentialDetails) {
		details.Azure.ClientSecret = s
	}
}

func withClientCert(path string) credentialOption {
	return func(details *CredentialDetails) {
		details.Azure.ClientCertificatePath = path
	}
}

func withTokenSource(source TokenSource) credentialOption {
	return func(details *CredentialDetails) {
		details.Token.Source = source
	}
}

func withTokenFile(path string) credentialOption {
	return func(details *CredentialDetails) {
		details.Token.File = &FileSource{Path: path}
	}
}

func newCredential(t testing.TB, name string, typ CredentialType, opts ...credentialOption) Credential {
	t.Helper()

	d := CredentialDetails{
		Type: typ,
	}

	switch typ {
	case CredentialTypeUser, CredentialTypeManagedIdentity:
		d.Azure = AzureCredential{}
	case CredentialTypeServicePrincipal:
		d.Azure = AzureCredential{
			ClientID:     newUUID(),
			ClientSecret: newUUID(),
		}
	case CredentialTypeWorkloadIdentity:
		d.Azure = AzureCredential{ClientID: newUUID()}
		d.Token = TokenDetails{
			Source: TokenSourceFile,
			File:   &FileSource{Path: fmt.Sprintf("/testdata/%s.token", name)},
		}
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&d)
		}
	}

	return Credential{
		Name:    name,
		Details: d,
	}
}

type contextOption func(*ContextDetails)

func withSubscription(subscription string) contextOption {
	return func(details *ContextDetails) {
		details.Subscription = subscription
	}
}

func newContext(t testing.TB, name, tenant, credential string, opts ...contextOption) Context {
	t.Helper()

	d := ContextDetails{Tenant: tenant, Credential: credential}
	for _, opt := range opts {
		if opt != nil {
			opt(&d)
		}
	}

	return Context{
		Name:    name,
		Details: d,
	}
}
