# azctx<!-- omit from toc -->

<!-- markdownlint-disable MD033 MD013 -->
<p align="center">
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/azctx?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/azctx?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 MD013 -->

A CLI tool for managing Azure CLI contexts, modelled after
[`kubectx`](https://github.com/ahmetb/kubectx). It maintains a
composable config file that maps named contexts to a tenant, credential,
and optional subscription. `azctx use <name>` switches the active
context and syncs your Azure CLI session instantly:

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="./docs/demo.gif" width="900" alt="azctx demo">
</p>
<!-- markdownlint-enable MD033 -->

## Installation

### Homebrew

```bash
brew install lvlcn-t/tap/azctx
```

### go install

Requires Go 1.26+ and the `GOEXPERIMENT=jsonv2` flag:

```bash
GOEXPERIMENT=jsonv2 go install github.com/lvlcn-t/azctx@latest
```

### Manual

Download the archive for your platform from the
[releases page](../../releases), extract the binary, and place it on
your `PATH`.

**Linux / macOS**:

```bash
OS=$(uname -s)
ARCH=$(uname -m | sed 's/aarch64/arm64/')
curl -sL "https://github.com/lvlcn-t/azctx/releases/latest/download/azctx_${OS}_${ARCH}.tar.gz" \
  | tar -xz -C ~/.local/bin azctx
chmod +x ~/.local/bin/azctx
```

**Windows** (PowerShell):

```powershell
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }
$url = "https://github.com/lvlcn-t/azctx/releases/latest/download/azctx_Windows_$arch.zip"
Invoke-WebRequest $url -OutFile azctx.zip
Expand-Archive azctx.zip -DestinationPath .
Move-Item azctx.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\azctx.exe"
```

### Container Image

Pre-built images are published to GitHub Container Registry:

```bash
docker pull ghcr.io/lvlcn-t/azctx:latest
docker run --rm \
  -v ~/.config/azctx:/home/nonroot/.config/azctx \
  -v ~/.azure:/home/nonroot/.azure \
  ghcr.io/lvlcn-t/azctx:latest use dev-west
```

Available tags: `latest`, `vMAJOR`, `vMAJOR.MINOR`, and full semver.

## Quick start

Create a minimal config at `~/.config/azctx/config.yaml`:

```yaml
apiVersion: azctx.lvlcn-t.dev/v1alpha1
kind: Config

tenants:
  - name: dev
    id: 00000000-0000-0000-0000-000000000000

credentials:
  - name: personal
    credential:
      type: user

contexts:
  - name: dev
    context:
      tenant: dev
      credential: personal
```

Then switch to it:

```bash
azctx use dev        # opens browser for login, sets subscription
azctx current        # prints "dev"
azctx list -o table  # show all contexts
```

For a more detailed guide, see the [Usage documentation](docs/usage.md) and
the [example config](docs/reference/example.config.yaml).

## Documentation

| Guide                                          | Description                                                |
| ---------------------------------------------- | ---------------------------------------------------------- |
| [Configuration](docs/configuration.md)         | Config file format, credential types, Key Vault references |
| [Workload Identity](docs/workload-identity.md) | OIDC federation — OAuth2 PKCE and file-based tokens        |
| [Usage](docs/usage.md)                         | All commands, output formats, typical workflows            |
| [CLI Reference](docs/reference/)               | Auto-generated command documentation                       |
| [Contributing](CONTRIBUTING.md)                | Development setup, testing, PR guidelines                  |

## License

Copyright (c) 2026 lvlcn-t. Licensed under the
[MIT License](./LICENSE).

## Code of Conduct

This project has adopted the
[Contributor Covenant v2.1](CODE_OF_CONDUCT.md). All contributors must
abide by the code of conduct.
