package config

import (
	"fmt"
	"strings"
)

const keyVaultScheme = "keyvault://"

// maxKeyVaultURIParts is the max number of path segments in a keyvault URI.
const maxKeyVaultURIParts = 4

// isKeyVaultRef reports whether a value is a keyvault:// URI reference.
func isKeyVaultRef(value string) bool {
	return strings.HasPrefix(value, keyVaultScheme)
}

// validateKeyVaultURI validates the structure of a keyvault:// URI without
// resolving it.
func validateKeyVaultURI(uri string) error {
	path := strings.TrimPrefix(uri, keyVaultScheme)
	parts := strings.Split(path, "/")

	if len(parts) < 3 || len(parts) > maxKeyVaultURIParts {
		return fmt.Errorf(
			"expected keyvault://<vault>/<secrets|certificates>/<name>[/<version>], got %q",
			uri,
		)
	}

	objectType := parts[1]
	if objectType != "secrets" && objectType != "certificates" {
		return fmt.Errorf(
			"object type must be \"secrets\" or \"certificates\", got %q",
			objectType,
		)
	}

	return nil
}
