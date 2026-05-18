# Workload Identity Federation

Workload identity federation lets azctx authenticate to Azure using an
OIDC token instead of long-lived secrets. Azure validates the token
against a federated credential configured in your app registration
or user-assigned managed identity — no client secret required.

## Use cases

- **CI/CD without secrets** — pipelines authenticate using a
  platform-issued OIDC token (GitHub Actions, GitLab CI) instead of
  storing client secrets.
- **Local dev on restricted networks** — when direct Azure login is
  unavailable (e.g. locked-down tenants behind jump hosts), use an
  external IdP's OIDC flow to obtain a federated token locally.
- **Eliminate service principals** — federate directly against a
  user-assigned managed identity so no app registration or secret
  rotation is needed. Unlike app registrations, managed identities
  require no Global Admin or Privileged Role Administrator — anyone
  with Contributor on the resource group can create one, significantly
  lowering the privilege barrier.

## Token sources

azctx supports two token sources:

| Source   | Use case                              |
| -------- | ------------------------------------- |
| `oauth2` | Local dev — browser-based PKCE flow   |
| `file`   | CI/CD — token written by the platform |

## Prerequisites

Before using either source, configure a federated credential in
Microsoft Entra ID:

1. Open your app registration in the Azure portal.
2. Go to **Certificates & secrets** → **Federated credentials**.
3. Add a credential that matches the issuer, subject, and
  audience of your token.

For the oauth2 flow, the issuer is typically your IdP's issuer URL (e.g.
`https://login.microsoftonline.com/<tenant-id>/v2.0` for Entra ID).

For CI platforms (GitHub Actions, GitLab CI), use the
platform's OIDC issuer URL and the appropriate subject claim.

See [Microsoft's documentation on workload identity federation][wif-docs]
for detailed setup.

> [!TIP]
> **App registration vs. managed identity**
>
> The `azure.client-id` field can reference either an app registration
> or a user-assigned managed identity — both support federated
> credentials. Use a managed identity when you want to avoid creating
> service principals entirely.

## OAuth2 PKCE flow (local development)

azctx opens a browser to your identity provider's authorization
endpoint, completes the authorization code flow with PKCE (S256), and
uses the resulting ID token for `az login --federated-token`.

```yaml
credentials:
  - name: local-oidc
    credential:
      type: workload-identity
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
      token:
        source: oauth2
        oauth2:
          issuer: https://login.microsoftonline.com/<tenant-id>/v2.0
          client-id: 11111111-1111-1111-1111-111111111111
          scopes:
            - api://my-api/.default
          redirect-uri: http://localhost:8080/callback
```

### OAuth2 fields

| Field          | Required | Default                              |
| -------------- | -------- | ------------------------------------ |
| `issuer`       | yes      |                                      |
| `client-id`    | yes      |                                      |
| `scopes`       | yes      |                                      |
| `redirect-uri` | no       | `http://localhost:<random>/callback` |
| `pkce`         | no       | `auto` (S256 enabled)                |

`pkce` accepts `auto`, `enabled`, or `disabled`. Use `disabled` only if
your IdP does not support PKCE (rare).

### How the flow works

1. azctx starts a local HTTP server on the redirect URI port.
2. The browser opens the IdP's `/authorize` endpoint with PKCE
   challenge.
3. After you authenticate, the IdP redirects to localhost with an
   authorization code.
4. azctx exchanges the code for an ID token.
5. The token is passed to `az login --federated-token`.

The OAuth2 PKCE flow effectively acts as a local federated access
broker: you authenticate once to your external IdP and receive a scoped
token that Azure accepts for a specific managed identity, giving each
developer least-privilege access without any shared secret.

## File-based tokens (CI/CD)

In CI environments the platform injects an OIDC token into a file or
environment variable. Point azctx at that path:

```yaml
credentials:
  - name: github-ci
    credential:
      type: workload-identity
      azure:
        client-id: 11111111-1111-1111-1111-111111111111
      token:
        source: file
        file:
          path: /var/run/secrets/azure/tokens/federated-token
```

The `path` field supports environment variable expansion
(`${VARIABLE}` syntax).

### File fields

| Field  | Required | Description            |
| ------ | -------- | ---------------------- |
| `path` | yes      | Path to the token file |

## Troubleshooting

| Symptom                                                    | Cause                                                                  | Fix                                                                                                                                                                                                       |
| ---------------------------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AADSTS70021: No matching federated identity record found` | Federated credential in Entra doesn't match the token's issuer/subject | Verify issuer URL and subject claim in the app registration                                                                                                                                               |
| `token file not found`                                     | File path doesn't exist or variable isn't expanded                     | Check the path; ensure the CI platform populates the token before azctx runs                                                                                                                              |
| `oauth2: server returned empty id_token`                   | Scopes don't include an audience that yields an id_token               | Add `openid` to scopes or verify the API scope is correct                                                                                                                                                 |
| Browser doesn't open                                       | Non-interactive environment detected                                   | Use `file` source in CI; `oauth2` is for interactive use only                                                                                                                                             |
| `AADSTS700222: AAD-issued tokens may not be used...`       | Token was issued by Azure AD itself (e.g. another managed identity)    | The federated credential issuer must be external to Azure AD — cross-tenant UAMI-to-UAMI federation is [not supported][so-76561302]. Use an external IdP or a [multi-tenant app registration][arsenvlad]. |

[wif-docs]: https://learn.microsoft.com/en-us/entra/workload-id/workload-identity-federation
[so-76561302]: https://stackoverflow.com/questions/76561302
[arsenvlad]: https://github.com/arsenvlad/entra-cross-tenant-app-fic-managed-identity
