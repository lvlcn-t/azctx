package file

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *config.FileSource
		wantErr bool
	}{
		{
			name: "valid config",
			cfg:  &config.FileSource{Path: "/some/path"},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name:    "empty path",
			cfg:     &config.FileSource{Path: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewProvider(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}

func TestAcquireToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid token",
			content:   "eyJ.token.sig\n",
			wantToken: "eyJ.token.sig",
		},
		{
			name:      "token with whitespace",
			content:   "  eyJ.token.sig  \n",
			wantToken: "eyJ.token.sig",
		},
		{
			name:    "empty file",
			content: "  \n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			require.NoError(t, afero.WriteFile(fs, "/token", []byte(tt.content), 0o600))

			p := &Provider{path: "/token", fsys: fs}

			got, _, err := p.AcquireToken(t.Context())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantToken, got)
		})
	}
}

func TestAcquireToken_FileNotFound(t *testing.T) {
	t.Parallel()

	p := &Provider{path: "/nonexistent/token", fsys: afero.NewMemMapFs()}

	_, _, err := p.AcquireToken(t.Context())
	require.Error(t, err)
}
