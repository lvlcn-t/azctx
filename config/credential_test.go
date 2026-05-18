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
		name    string
		input   *Credential
		wantErr bool
	}{
		{
			name: "service principal with secret",
			input: &Credential{
				Name: ciName,
				Details: CredentialDetails{
					Type: CredentialTypeServicePrincipal,
					Azure: AzureCredential{
						ClientID:     clientIDVal,
						ClientSecret: "secret",
					},
				},
			},
		},
		{
			name: "service principal with certificate",
			input: &Credential{
				Name: ciName,
				Details: CredentialDetails{
					Type: CredentialTypeServicePrincipal,
					Azure: AzureCredential{
						ClientID:              clientIDVal,
						ClientCertificatePath: "/tmp/cert.pem",
					},
				},
			},
		},
		{
			name: "service principal missing id",
			input: &Credential{
				Name: ciName,
				Details: CredentialDetails{
					Type: CredentialTypeServicePrincipal,
					Azure: AzureCredential{
						ClientSecret: "secret",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "service principal missing auth material",
			input: &Credential{
				Name: ciName,
				Details: CredentialDetails{
					Type: CredentialTypeServicePrincipal,
					Azure: AzureCredential{
						ClientID: clientIDVal,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "user",
			input: &Credential{
				Name: devContext.Name,
				Details: CredentialDetails{
					Type: CredentialTypeUser,
				},
			},
		},
		{
			name: "managed identity",
			input: &Credential{
				Name: "mi",
				Details: CredentialDetails{
					Type: CredentialTypeManagedIdentity,
				},
			},
		},
		{
			name: "workload identity",
			input: &Credential{
				Name: "workload-identity",
				Details: CredentialDetails{
					Type: CredentialTypeWorkloadIdentity,
					Azure: AzureCredential{
						ClientID: clientIDVal,
					},
					Token: TokenDetails{
						Source: TokenSourceFile,
						File: &FileSource{
							Path: "/tmp/token",
						},
					},
				},
			},
		},
		{
			name: "workload identity missing token source",
			input: &Credential{
				Name: "workload-identity",
				Details: CredentialDetails{
					Type:  CredentialTypeWorkloadIdentity,
					Azure: AzureCredential{ClientID: clientIDVal},
				},
			},
			wantErr: true,
		},
		{
			name: "workload identity missing token file path",
			input: &Credential{
				Name: "workload-identity",
				Details: CredentialDetails{
					Type:  CredentialTypeWorkloadIdentity,
					Azure: AzureCredential{ClientID: clientIDVal},
					Token: TokenDetails{
						Source: TokenSourceFile,
						File:   &FileSource{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "workload identity with invalid token source",
			input: &Credential{
				Name: "workload-identity",
				Details: CredentialDetails{
					Type:  CredentialTypeWorkloadIdentity,
					Azure: AzureCredential{ClientID: clientIDVal},
					Token: TokenDetails{
						Source: "invalid-source",
					},
				},
			},
			wantErr: true,
		},
		{name: "nil credential", input: nil, wantErr: true},
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
