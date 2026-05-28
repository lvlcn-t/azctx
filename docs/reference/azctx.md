## azctx

A CLI tool for managing Azure contexts.

### Synopsis

azctx is a CLI tool for managing Azure contexts.
It behaves similarly to kubectl config and manages its own composable config.

A context maps to:
	- a tenant (required)
	- a credential (required)
	- a subscription (optional)

When you run 'azctx use', azctx updates the current context in config and
also syncs the Azure CLI session by calling 'az login' and 'az account set'.


```
azctx [flags]
```

### Examples

```
  # Create or update the building blocks
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
  azctx get dev-west -o json
```

### Options

```
  -h, --help   help for azctx
```

### SEE ALSO

* [azctx current](azctx_current.md)	 - Show the current active context
* [azctx delete-context](azctx_delete-context.md)	 - Delete a context from azctx config
* [azctx get](azctx_get.md)	 - Get information about a context
* [azctx list](azctx_list.md)	 - List all available Azure contexts
* [azctx rename-context](azctx_rename-context.md)	 - Rename a context in azctx config
* [azctx set-context](azctx_set-context.md)	 - Set a context entry in azctx config
* [azctx set-credential](azctx_set-credential.md)	 - Set a credential entry in azctx config
* [azctx set-tenant](azctx_set-tenant.md)	 - Set a tenant entry in azctx config
* [azctx use](azctx_use.md)	 - Set the active Azure context
* [azctx view](azctx_view.md)	 - Display merged azctx config

