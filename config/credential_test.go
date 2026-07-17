package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCredentialType(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    CredentialType
		wantErr string
	}{
		{name: "service principal", raw: "service-principal", want: CredentialTypeServicePrincipal},
		{name: "user", raw: "user", want: CredentialTypeUser},
		{name: "managed identity", raw: "managed-identity", want: CredentialTypeManagedIdentity},
		{name: "workload identity", raw: "workload-identity", want: CredentialTypeWorkloadIdentity},
		{name: "empty", raw: "", wantErr: "credential type is required"},
		{name: "unsupported", raw: "something-else", wantErr: "unsupported credential type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCredentialType(tt.raw)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCredentialValidate(t *testing.T) {
	tests := []struct {
		input   Credential
		name    string
		wantErr bool
	}{
		{
			name:  "service principal with secret",
			input: newCredential(t, "ci", CredentialTypeServicePrincipal),
		},
		{
			name:  "service principal with certificate",
			input: newCredential(t, "ci-cert", CredentialTypeServicePrincipal, withClientCert("/tmp/cert.pem")),
		},
		{
			name:    "service principal missing id",
			input:   newCredential(t, "ci-missing-id", CredentialTypeServicePrincipal, withClientID("")),
			wantErr: true,
		},
		{
			name:    "service principal missing auth material",
			input:   newCredential(t, "ci-missing-auth", CredentialTypeServicePrincipal, withClientSecret("")),
			wantErr: true,
		},
		{
			name:  "user",
			input: newCredential(t, "user", CredentialTypeUser),
		},
		{
			name:  "managed identity",
			input: newCredential(t, "mi", CredentialTypeManagedIdentity),
		},
		{
			name:  "workload identity",
			input: newCredential(t, "wi", CredentialTypeWorkloadIdentity),
		},
		{
			name:    "workload identity missing token source",
			input:   newCredential(t, "wi-missing-source", CredentialTypeWorkloadIdentity, withTokenSource("")),
			wantErr: true,
		},
		{
			name: "workload identity missing token file path",
			input: newCredential(
				t,
				"wi-missing-file",
				CredentialTypeWorkloadIdentity,
				withTokenSource(TokenSourceFile),
				withTokenFile(""),
			),
			wantErr: true,
		},
		{
			name: "workload identity with invalid token source",
			input: newCredential(
				t,
				"wi-invalid-source",
				CredentialTypeWorkloadIdentity,
				withTokenSource("invalid-source"),
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
