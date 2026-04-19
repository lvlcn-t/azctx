# AGENTS.md

Agent guidance for the `azctx` repository — a Go CLI tool that manages
Azure CLI contexts (tenants, credentials, and subscriptions), modelled
after `kubectx`.

## Essential build facts

```bash
# Build binary to bin/azctx
make build

# Build and run (pass subcommands after target)
make dev use
make dev list

# Run tests (mirrors pre-commit)
GOEXPERIMENT=jsonv2 go test -race -count=1 -short ./...
```

`GOEXPERIMENT=jsonv2` **must** be set for build, test, vet, and lint —
it is exported in the Makefile but not globally in the shell. Tests and
`go vet` will silently behave differently without it.

## Pre-commit pipeline (run before committing)

```bash
pre-commit run --all-files
```

Hooks run in order: `go generate ./...` → `go mod tidy` →
`go test -race -count=1 -short` → `go vet` → `gofumpt` →
`golangci-lint --fix` → `helm-docs` → `gitleaks`.

Lint is auto-fixed by the hook (`--fix`). Format uses `gofumpt`, not
`gofmt` — run `gofumpt -l -w .` manually if needed.

## Code structure

| Path      | Purpose                                                |
| --------- | ------------------------------------------------------ |
| `main.go` | Entry point; injects version via ldflags               |
| `cmd/`    | Cobra command definitions (one file per command)       |
| `az/`     | Azure CLI integration (reads/writes `~/.azure/config`) |
| `config/` | azctx config loader, writer, and type definitions      |
| `output/` | Output formatting / printer                            |

The `az` package temporarily sets `login_experience_v2 = off` in the
Azure CLI config file before calling `az login`, then restores it.
The config path honours `$AZURE_CONFIG_DIR`.

## Config file behaviour

- Default path: `~/.config/azctx/config.yaml`
- Override with `$AZCTX` (colon-separated list of paths on Unix)
- Multiple paths are merged with **first-wins** semantics
- Writes always go to the first existing file, or the last listed path
  if none exist

Credential types: `service-principal`, `user`, `managed-identity`,
`oidc`. Each has required fields validated at runtime (see
`config/types.go:Credential.Validate`).

## Linter rules worth knowing

- No `init()` in application code (waived only for Cobra wiring in
  `cmd/root.go` — existing `//nolint:gochecknoinits` comment required)
- No `logrus` or `zap` — use `log/slog`
- Max 50 statements per function (`funlen`)
- `nolintlint` requires specific linter names; bare `//nolint` is an
  error

## CI

CI delegates to shared reusable workflows at `lvlcn-t/meta`. The
`.goreleaser-ci.yaml` config is used for snapshot builds in CI;
`.goreleaser.yaml` is for releases.
