package contexts

import "errors"

var (
	// ErrCurrentContextUnset indicates that current-context is not configured.
	ErrCurrentContextUnset = errors.New("current-context is not set")
	// ErrContextNameRequired indicates a missing context name.
	ErrContextNameRequired = errors.New("context name must not be empty")
	// ErrTenantNameRequired indicates a missing tenant name.
	ErrTenantNameRequired = errors.New("tenant name must not be empty")
	// ErrTenantIDRequired indicates a missing tenant id.
	ErrTenantIDRequired = errors.New("tenant id must not be empty")
	// ErrCredentialNameRequired indicates a missing credential name.
	ErrCredentialNameRequired = errors.New("credential name must not be empty")
)
