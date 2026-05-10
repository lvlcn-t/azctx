package az

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/lvlcn-t/azctx/keyvault"
	"github.com/spf13/afero"
)

// resolver returns the keyvault resolver, creating it lazily on first use.
func (c *client) resolver() (*keyvault.Resolver, error) {
	if c.kvResolver != nil {
		return c.kvResolver, nil
	}

	kvClient, err := keyvault.NewAzureClient()
	if err != nil {
		return nil, err
	}

	c.kvResolver = keyvault.NewResolver(kvClient)

	return c.kvResolver, nil
}

// sensitiveFlags lists az CLI flags whose values must not appear in error messages.
var sensitiveFlags = map[string]struct{}{
	flagPassword:       {},
	flagFederatedToken: {},
}

// redactArgs returns a joined argument string with sensitive flag values replaced.
func redactArgs(args []string) string {
	redacted := make([]string, len(args))
	copy(redacted, args)

	for i, arg := range redacted {
		if _, ok := sensitiveFlags[arg]; ok && i+1 < len(redacted) {
			redacted[i+1] = "[REDACTED]"
		}
	}

	return strings.Join(redacted, " ")
}

// writeTempCert writes PEM bytes to a temporary file with restricted
// permissions and returns the file path.
func writeTempCert(pem []byte) (path string, err error) {
	f, err := afero.TempFile(fsys, "", "azctx-cert-*.pem")
	if err != nil {
		return "", err
	}

	name := f.Name()

	closed := false
	closeFile := func() error {
		if closed {
			return nil
		}

		// We're tracking the closed state before calling Close() because Close
		// must be treated as a one-shot operation. The io.Closer contract says
		// behavior after the first Close is undefined unless the implementation
		// documents otherwise:
		// https://pkg.go.dev/io#Closer
		//
		// On POSIX-like systems, close errors may be reported after the file
		// descriptor has already been released, so retrying Close can be unsafe:
		// the descriptor number may have been reused and a retry could close
		// something unrelated. Linux documents this explicitly:
		// https://man7.org/linux/man-pages/man2/close.2.html
		closed = true
		if cErr := f.Close(); cErr != nil {
			return fmt.Errorf("closing temp cert file %q: %w", name, cErr)
		}

		return nil
	}

	cleanup := func(cause error) error {
		cErr := closeFile()
		return errors.Join(cause, cErr, removeTempCert(name))
	}

	keep := false
	defer func() {
		if !keep {
			err = cleanup(err)
			path = ""
		}
	}()

	const certFileMode fs.FileMode = 0o600
	if err = fsys.Chmod(name, certFileMode); err != nil {
		return "", fmt.Errorf("setting permissions on temp cert file %q: %w", name, err)
	}

	if _, err = f.Write(pem); err != nil {
		return "", fmt.Errorf("writing to temp cert file %q: %w", name, err)
	}

	if err = closeFile(); err != nil {
		return "", err
	}

	keep = true
	return name, nil
}

// removeTempCert attempts to remove the temporary certificate file at the given path.
func removeTempCert(path string) error {
	if err := fsys.Remove(path); err != nil {
		return fmt.Errorf(
			"%w: failed to remove temp cert file %q: %w",
			errTempCertMayRemain,
			path,
			err,
		)
	}

	return nil
}
