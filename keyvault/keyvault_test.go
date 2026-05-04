package keyvault

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testVault = "my-vault"

const (
	testSecretURI = "keyvault://my-vault/secrets/my-secret" //nolint:gosec // test URI, not a real credential
	testCertURI   = "keyvault://my-vault/certificates/my-cert"
)

func TestIsReference(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{testSecretURI, true},
		{testCertURI, true},
		{testSecretURI + "/version", true},
		{"my-plain-secret", false},
		{"", false},
		{"keyvault:", false},
		{"keyvault://", true}, // prefix matches, but will fail parse
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, IsReference(tc.input))
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    Reference
		wantErr bool
	}{
		{
			name: "secret without version",
			uri:  testSecretURI,
			want: Reference{
				VaultName:  testVault,
				ObjectType: ObjectTypeSecrets,
				ObjectName: "my-secret",
			},
		},
		{
			name: "secret with version",
			uri:  "keyvault://my-vault/secrets/my-secret/abc123",
			want: Reference{
				VaultName:  testVault,
				ObjectType: ObjectTypeSecrets,
				ObjectName: "my-secret",
				Version:    "abc123",
			},
		},
		{
			name: "certificate without version",
			uri:  testCertURI,
			want: Reference{
				VaultName:  testVault,
				ObjectType: ObjectTypeCertificates,
				ObjectName: "my-cert",
			},
		},
		{
			name:    "invalid object type",
			uri:     "keyvault://my-vault/keys/my-key",
			wantErr: true,
		},
		{
			name:    "too few parts",
			uri:     "keyvault://my-vault/secrets",
			wantErr: true,
		},
		{
			name:    "too many parts",
			uri:     "keyvault://my-vault/secrets/name/version/extra",
			wantErr: true,
		},
		{
			name:    "not a keyvault URI",
			uri:     "https://vault.azure.net",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.uri)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReference_VaultURL(t *testing.T) {
	ref := Reference{VaultName: testVault}
	assert.Equal(t, "https://my-vault.vault.azure.net", ref.VaultURL())
}

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		mockValue string
		mockErr   error
		want      string
		wantErr   bool
	}{
		{
			name:      "resolves secret",
			uri:       testSecretURI,
			mockValue: "super-secret-value",
			want:      "super-secret-value",
		},
		{
			name:      "resolves certificate",
			uri:       testCertURI,
			mockValue: "-----BEGIN CERTIFICATE-----\nfoo\n-----END CERTIFICATE-----",
			want:      "-----BEGIN CERTIFICATE-----\nfoo\n-----END CERTIFICATE-----",
		},
		{
			name:    "client error propagates",
			uri:     "keyvault://my-vault/secrets/missing",
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "invalid URI",
			uri:     "not-a-keyvault-uri",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &ClientMock{
				GetSecretFunc: func(_ context.Context, _ Reference) (string, error) {
					return tc.mockValue, tc.mockErr
				},
			}

			resolver := NewResolver(mock)
			got, err := resolver.Resolve(context.Background(), tc.uri)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolver_ResolveCertificateBytes(t *testing.T) {
	pemData := "-----BEGIN CERTIFICATE-----\ndata\n-----END CERTIFICATE-----"

	mock := &ClientMock{
		GetSecretFunc: func(_ context.Context, ref Reference) (string, error) {
			assert.Equal(t, ObjectTypeCertificates, ref.ObjectType)
			return pemData, nil
		},
	}

	resolver := NewResolver(mock)
	got, err := resolver.ResolveCertificateBytes(context.Background(), testCertURI)

	require.NoError(t, err)
	assert.Equal(t, []byte(pemData), got)
}
