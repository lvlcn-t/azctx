# azctx<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/azctx?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/azctx?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

- [About this component](#about-this-component)
- [Installation](#installation)
  - [Binary](#binary)
  - [Container Image](#container-image)
- [Usage](#usage)
  - [Config file](#config-file)
  - [Commands](#commands)
  - [Output formats](#output-formats)
  - [Typical workflow](#typical-workflow)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

## About this component

`azctx` is a CLI tool for managing Azure CLI contexts, modelled after
[`kubectx`](https://github.com/ahmetb/kubectx). It maintains its own
composable config file that maps named contexts to a tenant, a credential,
and an optional subscription. Running `azctx use <name>` switches the active
context and syncs the Azure CLI session by calling `az login` and
`az account set`.

## Installation

### Binary

Download the archive for your platform from the
[releases page](../../releases), extract the binary, and place it on
your `PATH`.

**Linux**:

```bash
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sL "https://github.com/lvlcn-t/azctx/releases/latest/download/azctx_linux_${ARCH}.tar.gz" \
  | tar -xz -C ~/.local/bin azctx
chmod +x ~/.local/bin/azctx
```

**macOS**:

```bash
ARCH=$(uname -m | sed 's/x86_64/amd64/')
curl -sL "https://github.com/lvlcn-t/azctx/releases/latest/download/azctx_darwin_${ARCH}.tar.gz" \
  | tar -xz -C ~/.local/bin azctx
chmod +x ~/.local/bin/azctx
```

**Windows** (PowerShell):

```powershell
$arch = $env:PROCESSOR_ARCHITECTURE.ToLower()
$url = "https://github.com/lvlcn-t/azctx/releases/latest/download/azctx_windows_$arch.tar.gz"
Invoke-WebRequest $url -OutFile azctx.tar.gz
tar -xf azctx.tar.gz azctx.exe
Move-Item azctx.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\azctx.exe"
```

Verify integrity by checking the downloaded archive against
`azctx_*_checksums.txt` published alongside each release.

### Container Image

Pre-built images are published to GitHub Container Registry on every release.
The image is based on `mcr.microsoft.com/azure-cli` so `az` is available and
`azctx use` works inside the container. It runs as the `nonroot` user
(uid 65532) that the base image ships for this purpose.

```bash
docker pull ghcr.io/lvlcn-t/azctx:latest
```

Available tags: `latest`, `vMAJOR`, `vMAJOR.MINOR`, and the full semver tag.

Mount your azctx config and Azure CLI state directories so the container can
read your contexts and persist the login session:

```bash
docker run --rm \
  -v ~/.azctx:/home/nonroot/.azctx \
  -v ~/.azure:/home/nonroot/.azure \
  ghcr.io/lvlcn-t/azctx:latest use dev-west
```

Both directories must be owned by or writable by uid 65532 on the host.
Alternatively, use the `AZCTX` and `AZURE_CONFIG_DIR` environment variables
to point the container at different paths:

```bash
docker run --rm \
  -e AZCTX=/config/azctx.yaml \
  -e AZURE_CONFIG_DIR=/config/azure \
  -v /my/config:/config \
  ghcr.io/lvlcn-t/azctx:latest use dev-west
```

## Usage

### Config file

`azctx` reads from `~/.azctx/config.yaml` by default. Set the `AZCTX`
environment variable to a colon-separated list of paths to load and merge
multiple files. Merged entries follow **first-wins** semantics; writes always
go to the first existing file in the list.

A config file has three sections — tenants, credentials, and contexts:

```yaml
tenants:
  - name: dev
    id: 00000000-0000-0000-0000-000000000000

credentials:
  - name: ci-sp
    type: service-principal
    client-id: 11111111-1111-1111-1111-111111111111
    client-secret: super-secret

contexts:
  - name: dev
    tenant: dev
    credential: ci-sp
    subscription: 22222222-2222-2222-2222-222222222222

current-context: dev
```

Supported credential types and their required fields:

| Type                | Required fields                                            |
| ------------------- | ---------------------------------------------------------- |
| `service-principal` | `client-id` + `client-secret` or `client-certificate-path` |
| `user`              | _(none — interactive login)_                               |
| `managed-identity`  | _(none)_                                                   |
| `oidc`              | `client-id` + `federated-token-file`                       |

### Commands

| Command                                      | Alias             | Description                                        |
| -------------------------------------------- | ----------------- | -------------------------------------------------- |
| `use NAME`                                   | `use-context`     | Switch active context and sync Azure CLI state     |
| `current`                                    | `current-context` | Show the active context name                       |
| `list`                                       | `get-contexts`    | List all contexts                                  |
| `get [NAME]`                                 | `get-context`     | Show details for one context (defaults to current) |
| `set-tenant NAME --id ID`                    |                   | Create or update a tenant entry                    |
| `set-credential NAME --type TYPE`            |                   | Create or update a credential entry                |
| `set-context NAME --tenant T --credential C` |                   | Create or update a context entry                   |
| `rename-context OLD NEW`                     |                   | Rename a context                                   |
| `delete-context NAME`                        | `unset-context`   | Remove a context from config                       |
| `view`                                       |                   | Display the merged config                          |

### Output formats

Commands that read config accept `-o text` (default), `-o table`, and
`-o json`. The `current` command also accepts `--verbose` / `-v` to print
full context details instead of just the name.

### Typical workflow

First, register the building blocks once:

```bash
azctx set-tenant dev --id 00000000-0000-0000-0000-000000000000

azctx set-credential ci-sp \
  --type service-principal \
  --client-id 11111111-1111-1111-1111-111111111111 \
  --client-secret super-secret
```

Then create a context that ties them together:

```bash
azctx set-context dev \
  --tenant dev \
  --credential ci-sp \
  --subscription 22222222-2222-2222-2222-222222222222
```

Switch to it — this calls `az login` and `az account set` automatically:

```bash
azctx use dev
```

Inspect what you have:

```bash
azctx current          # prints "dev"
azctx list -o table    # all contexts with status column
azctx get dev -o json
```

## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of
conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code
of conduct.

## Working Language

We decided to apply _English_ as the primary project language.

Consequently, all content will be made available primarily in English.
We also ask all interested people to use English as the preferred language to create issues,
in their code (comments, documentation, etc.) and when you send requests to us.
The application itself and all end-user facing content will be made available in other languages as needed.

## Support and Feedback

The following channels are available for discussions, feedback, and support requests:

| Type       | Channel                                                                                                                 |
| ---------- | ----------------------------------------------------------------------------------------------------------------------- |
| **Issues** | [![General Discussion](https://img.shields.io/github/issues/lvlcn-t/azctx?style=flat-square)](/../../issues/new/choose) |

## How to Contribute

Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project
structure, as well as additional contribution information, see our [Contribution Guidelines](./CONTRIBUTING.md). By
participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (c) 2024 lvlcn-t.

Licensed under the **MIT** (the "License"); you may not use this file except in compliance with
the License.

You may obtain a copy of the License at <https://www.mit.edu/~amini/LICENSE.md>.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "
AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](./LICENSE) for
the specific language governing permissions and limitations under the License.
