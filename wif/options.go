package wif

// AcquireOption configures the behavior of [Provider.AcquireToken].
type AcquireOption func(*AcquireOptions)

// AcquireOptions holds the resolved option values for [Provider.AcquireToken].
type AcquireOptions struct {
	ForceRefresh bool
}

// ApplyOptions applies the given options to a default [AcquireOptions].
func ApplyOptions(opts []AcquireOption) AcquireOptions {
	var o AcquireOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// WithForceRefresh bypasses any cached token and forces a fresh
// acquisition flow.
func WithForceRefresh() AcquireOption {
	return func(o *AcquireOptions) {
		o.ForceRefresh = true
	}
}
