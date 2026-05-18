# Usage

azctx manages named contexts that map to an Azure tenant, credential,
and optional subscription. Switching contexts syncs the Azure CLI
session automatically.

## Commands

| Command                                      | Alias             | Description                                     |
| -------------------------------------------- | ----------------- | ----------------------------------------------- |
| `use NAME`                                   | `use-context`     | Switch context and sync Azure CLI               |
| `current`                                    | `current-context` | Show the active context name                    |
| `list`                                       | `get-contexts`    | List all contexts                               |
| `get [NAME]`                                 | `get-context`     | Show details for one context (default: current) |
| `set-tenant NAME --id ID`                    |                   | Create or update a tenant                       |
| `set-credential NAME --type TYPE`            |                   | Create or update a credential                   |
| `set-context NAME --tenant T --credential C` |                   | Create or update a context                      |
| `rename-context OLD NEW`                     |                   | Rename a context                                |
| `delete-context NAME`                        | `unset-context`   | Remove a context                                |
| `view`                                       |                   | Display the merged config                       |

For full flag details, run `azctx <command> --help` or see the
[CLI reference](reference/).

## Output formats

Commands that read config accept `-o` with one of:

- `text` — default, human-readable
- `table` — tabular with headers
- `json` — machine-readable JSON

The `current` command also accepts `--verbose` / `-v` to print full
context details instead of just the name.

## Typical workflow

### Register building blocks

Create the tenant and credential entries that your contexts will
reference:

```bash
azctx set-tenant dev --id 00000000-0000-0000-0000-000000000000

azctx set-credential ci-sp \
  --type service-principal \
  --client-id 11111111-1111-1111-1111-111111111111 \
  --client-secret super-secret
```

### Create a context

Tie a tenant and credential together, optionally pinning a
subscription:

```bash
azctx set-context dev-west \
  --tenant dev \
  --credential ci-sp \
  --subscription 22222222-2222-2222-2222-222222222222
```

### Switch and inspect

```bash
azctx use dev-west     # calls az login + az account set
azctx current          # prints "dev-west"
azctx list -o table    # all contexts with status column
azctx get dev-west -o json
```

## See also

- [Configuration](configuration.md) — config file format and
  credential types
- [Workload Identity](workload-identity.md) — OIDC federation setup
- [CLI reference](reference/) — auto-generated command docs
