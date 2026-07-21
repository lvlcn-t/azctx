package cmd

import (
	"context"
	"testing"

	"github.com/lvlcn-t/azctx/az"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUseWiring verifies the cobra glue for use: a successful switch prints the
// confirmation message. Switch logic is covered in the login package.
func TestUseWiring(t *testing.T) {
	writeConfig(t, baseConfig())

	var mock *az.CLIMock
	mock = &az.CLIMock{
		WithTenantFunc:           func(string) az.CLI { return mock },
		WithCredentialFunc:       func(*config.Credential) az.CLI { return mock },
		WithSubscriptionFunc:     func(string) az.CLI { return mock },
		AllowNoSubscriptionsFunc: func(bool) az.CLI { return mock },
		WithFederatedTokenFunc:   func(string) az.CLI { return mock },
		LoginFunc:                func(context.Context) error { return nil },
	}

	command := &useCommand{
		loader: config.NewLoader(),
		manager: &login.Manager{
			NewClient: func(context.Context) (az.CLI, error) { return mock, nil },
			Writer:    config.NewWriter(),
		},
	}

	runCommand, stdout := newRunCmd()
	err := command.run(runCommand, []string{prodContext})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `Switched to context "prod".`)
}
