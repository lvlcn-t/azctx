package tabs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExistsValidator(t *testing.T) {
	validate := existsValidator([]string{"corp", "platform"})

	require.NoError(t, validate("corp"))
	require.ErrorIs(t, validate("ghost"), errReferenceUnknown)
}

func TestCredentialTypeValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "user", value: "user", wantErr: false},
		{name: "service-principal", value: "service-principal", wantErr: false},
		{name: "workload-identity", value: "workload-identity", wantErr: false},
		{name: "unknown", value: "bogus", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := credentialTypeValidator(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
