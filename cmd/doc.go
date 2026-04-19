// Package cmd contains the command-line interface (CLI) commands for the azctx CLI.
// It defines the root command and any subcommands, along with their associated flags and functionality.
package cmd

const (
	// Name is the name of the azctx CLI tool.
	Name = "azctx"
	// Version is the version of the azctx CLI tool. It is set at build time using ldflags.
	// Defaults to "dev" if not set.
	Version = "dev"
	// ShortDescription is a brief description of the azctx CLI tool.
	ShortDescription = "A CLI tool for managing Azure contexts."
	// Description is a detailed description of the azctx CLI tool.
	Description = `azctx is a CLI tool for managing Azure contexts.
It behaves similarly to kubectl config and manages its own composable config.

A context maps to:
	- a tenant (required)
	- a credential (required)
	- a subscription (optional)

When you run 'azctx use', azctx updates the current context in config and
also syncs the Azure CLI session by calling 'az login' and 'az account set'.
`
	// Example provides usage examples for the azctx CLI tool.
	Example = `  # Create or update the building blocks
  azctx set-tenant dev --id 00000000-0000-0000-0000-000000000000
  azctx set-credential ci-sp \
    --type service-principal \
    --client-id 11111111-1111-1111-1111-111111111111 \
    --client-secret super-secret

  # Create a context (tenant + credential + optional subscription)
  azctx set-context dev-west \
    --tenant dev \
    --credential ci-sp \
    --subscription 22222222-2222-2222-2222-222222222222

  # Switch and inspect contexts
  azctx use dev-west
  azctx current
  azctx list -o table
  azctx get dev-west -o json`
)
