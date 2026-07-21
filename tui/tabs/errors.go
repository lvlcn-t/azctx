package tabs

import "errors"

// errReferenceUnknown indicates a form value referenced a tenant or credential
// that does not exist.
var errReferenceUnknown = errors.New("does not exist")
