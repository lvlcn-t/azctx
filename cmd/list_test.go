package cmd

import (
	"testing"

	"github.com/lvlcn-t/azctx/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListText(t *testing.T) {
	writeConfig(t, baseConfig())

	stdout, _, err := execCmd(t, newListCmd(), "-o", "text")
	require.NoError(t, err)

	assert.Contains(t, stdout, "* dev\n")
	assert.Contains(t, stdout, "  prod\n")
}

func TestListEmpty(t *testing.T) {
	writeConfig(t, &config.Config{})

	stdout, _, err := execCmd(t, newListCmd(), "-o", "text")
	require.NoError(t, err)
	assert.Empty(t, stdout)
}
