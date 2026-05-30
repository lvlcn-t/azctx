package az

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactArgs(t *testing.T) {
	tests := []struct {
		name string
		want string
		args []string
	}{
		{
			name: "password redacted",
			args: []string{cmdLogin, flagServicePrincipal, flagUsername, "app-id", flagPassword, "s3cret", flagTenant, "t1"},
			want: "login --service-principal --username app-id --password [REDACTED] --tenant t1",
		},
		{
			name: "federated-token redacted",
			args: []string{cmdLogin, flagServicePrincipal, flagUsername, "app-id", flagFederatedToken, "eyJhbGci...", flagTenant, "t1"},
			want: "login --service-principal --username app-id --federated-token [REDACTED] --tenant t1",
		},
		{
			name: "no sensitive flags unchanged",
			args: []string{"account", "set", "--subscription", "sub-id"},
			want: "account set --subscription sub-id",
		},
		{
			name: "sensitive flag as last arg no panic",
			args: []string{cmdLogin, flagPassword},
			want: "login --password",
		},
		{
			name: "multiple sensitive flags",
			args: []string{flagPassword, "secret1", flagFederatedToken, "token1"},
			want: "--password [REDACTED] --federated-token [REDACTED]",
		},
		{
			name: "empty args",
			args: []string{},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, redactArgs(tc.args))
		})
	}
}
