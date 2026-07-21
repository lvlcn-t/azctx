package contexts

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentContextName(t *testing.T) {
	tests := []struct {
		wantErr error
		cfg     *config.Config
		name    string
		want    string
	}{
		{name: "set", cfg: &config.Config{CurrentContext: devContext}, want: devContext},
		{name: "unset", cfg: &config.Config{}, wantErr: ErrCurrentContextUnset},
		{name: "nil config", cfg: nil, wantErr: ErrCurrentContextUnset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CurrentContextName(tt.cfg)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
