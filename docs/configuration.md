# Configuration

azctx reads a YAML config file that defines tenants, credentials, and
contexts. A context ties a tenant to a credential and an optional
subscription — running `azctx use` activates that combination.

## Config file location

Default path: `~/.config/azctx/config.yaml`

If `XDG_CONFIG_HOME` is set, azctx uses `$XDG_CONFIG_HOME/azctx/config.yaml`
instead.

Override with the `AZCTX` environment variable. On Unix systems this is
a colon-separated list of paths; on Windows use semicolons. When
multiple paths are given:

- Entries are merged with **first-wins** semantics (the first file that
  defines a name wins).
- Writes always go to the first existing file in the list, or the last
  path if none exist yet.

```bash
export AZCTX="$HOME/.config/azctx/work.yaml:$HOME/.config/azctx/personal.yaml"
```

## File structure

A config file has three top-level lists and one scalar:

```yaml
tenants:
  - name: dev
    id: 00000000-0000-0000-0000-000000000000

credentials:
  - name: ci-sp
    credential:
      type: service-principal
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
        client-secret: super-secret

contexts:
  - name: dev
    context:
      tenant: dev
      credential: ci-sp
      subscription: 22222222-2222-2222-2222-222222222222

current-context: dev
```

## Credential types

| Type                | When to use                                     |
| ------------------- | ----------------------------------------------- |
| `service-principal` | Automation, CI with a client secret or cert     |
| `user`              | Interactive local dev (browser login)           |
| `managed-identity`  | Running on Azure (VMs, ACI, AKS pods)           |
| `workload-identity` | OIDC federation — CI or local dev via PKCE flow |

### `service-principal`

Required: `azure.client-id` plus one of `azure.client-secret` or
`azure.client-certificate-path`.

```yaml
credentials:
  - name: ci-sp
    credential:
      type: service-principal
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
        client-secret: super-secret
```

Both `client-secret` and `client-certificate-path` accept
[Key Vault references](#key-vault-secret-references).

### `user`

No additional fields required — azctx triggers interactive browser
login.

```yaml
credentials:
  - name: personal
    credential:
      type: user
      azure: {}
```

### `managed-identity`

For system-assigned identity, no additional fields are required. For
user-assigned identity, set `azure.client-id`:

```yaml
credentials:
  # System-assigned
  - name: sami-example
    credential:
      type: managed-identity
      azure: {}

  # User-assigned
  - name: uami-example
    credential:
      type: managed-identity
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
```

### `workload-identity`

Requires `azure.client-id` and a `token` block. See the dedicated
[Workload Identity guide](workload-identity.md) for full details.

```yaml
credentials:
  - name: oidc-local
    credential:
      type: workload-identity
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
      token:
        source: oauth2
        oauth2:
          issuer: https://login.microsoftonline.com/<tenant-id>/v2.0
          client-id: 11111111-1111-1111-1111-111111111111
          scopes: [api://my-api/.default]
```

## Key Vault secret references

`client-secret` and `client-certificate-path` on a `service-principal`
credential accept a `keyvault://` URI. azctx fetches the secret at
login time so nothing sensitive is stored on disk.

```yaml
azure:
  client-secret: "keyvault://my-vault/secrets/ci-sp-secret"
  # or:
  client-certificate-path: "keyvault://my-vault/certificates/ci-cert"
```

URI format: `keyvault://<vault>/<secrets|certificates>/<name>[/<version>]`

<!-- markdownlint-disable MD028 -->

> [!IMPORTANT]
> Resolution uses [`DefaultAzureCredential`][dac], so an ambient
> credential must already be active (a user session, managed identity,
> or environment variables). Using the same service principal whose
> secret lives in Key Vault as the only credential is unsupported.

> [!NOTE]
> For certificates, azctx writes the fetched PEM to a temporary file
> (mode `0600`) for the duration of `az login`, then deletes it.

<!-- markdownlint-enable MD028 -->

[dac]: https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/local-development-dev-accounts?tabs=azure-portal%2Csign-in-azure-cli
