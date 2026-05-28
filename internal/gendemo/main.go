// Package main generates docs/demo.gif by rendering demo.tape.tmpl with a
// temporary mock config and invoking vhs(1).
//
// The mock config is written to a temporary directory and injected into the
// VHS shell session via a hidden "export AZCTX=..." preamble so the env var
// never appears in the recorded GIF.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"
	"github.com/lvlcn-t/azctx/config"
	"go.yaml.in/yaml/v4"
)

//go:embed demo.tape.tmpl
var tapeTmpl string

// tapeData holds the values injected into the tape template.
type tapeData struct {
	Out     string // absolute path to the output GIF
	BinDir  string // temp directory containing the freshly-built azctx binary
	CfgPath string // path to the mock config.yaml
}

func main() {
	out := flag.String("out", "docs/demo.gif", "output path for the generated GIF")
	flag.Parse()

	ctx := context.Background()
	if err := run(ctx, *out); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, outPath string) error {
	tmpDir, err := os.MkdirTemp("", "azctx-demo-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err = writeMockConfig(cfgPath); err != nil {
		return fmt.Errorf("write mock config: %w", err)
	}

	if err = buildBinary(ctx, tmpDir); err != nil {
		return fmt.Errorf("build binary: %w", err)
	}

	tmpl, err := template.New("tape").Parse(tapeTmpl)
	if err != nil {
		return fmt.Errorf("parse tape template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, tapeData{
		Out:     outPath,
		BinDir:  tmpDir,
		CfgPath: cfgPath,
	}); err != nil {
		return fmt.Errorf("render tape template: %w", err)
	}

	tmpTape, err := os.CreateTemp("", "azctx-demo-*.tape")
	if err != nil {
		return fmt.Errorf("create temp tape: %w", err)
	}
	defer func() { _ = os.Remove(tmpTape.Name()) }()

	if _, err := buf.WriteTo(tmpTape); err != nil {
		_ = tmpTape.Close()
		return fmt.Errorf("write temp tape: %w", err)
	}
	if err := tmpTape.Close(); err != nil {
		return fmt.Errorf("close temp tape: %w", err)
	}

	cmd := exec.CommandContext(ctx, "vhs", tmpTape.Name()) // #nosec G204 // safe since we control tape path and contents
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs: %w", err)
	}

	return nil
}

// buildBinary compiles the current source tree and writes the binary to
// <dir>/azctx. GOEXPERIMENT=jsonv2 is set so the build matches the module's
// requirements.
func buildBinary(ctx context.Context, dir string) error {
	out := filepath.Join(dir, "azctx")
	cmd := exec.CommandContext(ctx, "go", "build", "-o", out, ".")
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeMockConfig(path string) error {
	cfg := config.Config{
		APIVersion:     config.APIVersion,
		Kind:           config.Kind,
		CurrentContext: "dev-west",
		Tenants: []config.Tenant{
			{Name: "dev", Details: config.TenantDetails{ID: uuid.NewString()}}, //nolint:goconst // not worth extracting
			{Name: "prod", Details: config.TenantDetails{ID: uuid.NewString()}},
		},
		Credentials: []config.Credential{
			{
				Name:    "personal",
				Details: config.CredentialDetails{Type: config.CredentialTypeUser},
			},
			{
				Name: "ci-sp", //nolint:goconst // not worth extracting
				Details: config.CredentialDetails{
					Type: config.CredentialTypeServicePrincipal,
					Azure: config.AzureCredential{
						ClientID:     uuid.NewString(),
						ClientSecret: "keyvault://prod-vault.vault.azure.net/secrets/ci-sp-secret", // #nosec G101 // fake secret reference for demo purposes
					},
				},
			},
		},
		Contexts: []config.Context{
			{
				Name: "dev-west",
				Details: config.ContextDetails{
					Tenant:       "dev",
					Credential:   "personal",
					Subscription: uuid.NewString(),
				},
			},
			{
				Name: "dev-east",
				Details: config.ContextDetails{
					Tenant:     "dev",
					Credential: "ci-sp",
				},
			},
			{
				Name: "prod-west",
				Details: config.ContextDetails{
					Tenant:       "prod",
					Credential:   "ci-sp",
					Subscription: uuid.NewString(),
				},
			},
		},
	}

	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal mock config: %w", err)
	}

	const fileMode = 0o600
	return os.WriteFile(path, raw, fileMode)
}
