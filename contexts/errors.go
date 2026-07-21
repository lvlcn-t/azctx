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

	// ErrTenantExists indicates a create was attempted for an existing tenant.
	ErrTenantExists = errors.New("tenant already exists")
	// ErrTenantNotFound indicates an update or rename targeted a missing tenant.
	ErrTenantNotFound = errors.New("tenant not found")
	// ErrCredentialExists indicates a create was attempted for an existing credential.
	ErrCredentialExists = errors.New("credential already exists")
	// ErrCredentialNotFound indicates an update or rename targeted a missing credential.
	ErrCredentialNotFound = errors.New("credential not found")
	// ErrContextExists indicates a create was attempted for an existing context.
	ErrContextExists = errors.New("context already exists")
	// ErrContextNotFound indicates an update or rename targeted a missing context.
	ErrContextNotFound = errors.New("context not found")
)
