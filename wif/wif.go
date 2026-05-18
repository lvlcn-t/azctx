// Package wif provides federated identity token acquisition for Azure
// workload identity login.
package wif

import "context"

// Provider acquires a federated identity token for workload identity login.
//
//go:generate go tool moq -out wif_moq.go . Provider
type Provider interface {
	// AcquireToken acquires and returns a federated identity token.
	// The cached return value indicates whether the token was served from cache.
	// Options such as [WithForceRefresh] control caching behavior.
	AcquireToken(ctx context.Context, opts ...AcquireOption) (token string, cached bool, err error)
}
