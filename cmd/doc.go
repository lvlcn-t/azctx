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
It allows users to easily switch between different Azure subscriptions and manage their resources.
Features:
	- List and switch between Azure subscriptions
	- Save and load context configurations
	- Interactive mode for easy navigation
	- Support for multiple Azure environments
`
	// Example provides usage examples for the azctx CLI tool.
	Example = ``
)
