package oauth2

import (
	"cmp"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/wif"
	"golang.org/x/oauth2"
)

const (
	maxReadHeaderTimeout = 5 * time.Second
	exchangeTimeout      = 30 * time.Second
	shutdownTimeout      = 5 * time.Second
)

// Provider performs the OAuth2 authorization code flow with OIDC discovery
// to obtain an id_token for Azure workload identity federation.
type Provider struct {
	cfg          oauth2.Config
	listener     net.Listener
	mux          *http.ServeMux
	provider     *oidc.Provider
	cache        *cache
	codeVerifier string
	cacheKey     string
	callback     http.Server
}

// NewProvider initializes an OIDC provider via discovery and prepares a
// local callback listener for the authorization code flow. PKCE (S256)
// is enabled by default unless the config explicitly sets pkce to "disabled".
// The cacheDir is used to store cached id_tokens on disk.
func NewProvider(ctx context.Context, cfg *config.OAuth2Source, cacheDir string) (*Provider, error) {
	p, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("initialize OIDC provider: %w", err)
	}

	var verifier string
	if cfg.PKCE.IsEnabled() {
		verifier, err = randomVerifier()
		if err != nil {
			return nil, fmt.Errorf("generate PKCE verifier: %w", err)
		}
	}

	// Default to an OS-assigned port on the loopback interface. RFC 8252 §7.3
	// (https://datatracker.ietf.org/doc/html/rfc8252#section-7.3) specifies that
	// authorization servers MUST allow any port for loopback redirect URIs, making
	// this safe for compliant IdPs such as Entra ID. For IdPs that require an exact
	// redirect URI match, users can set redirect-uri in their config.
	redirectURI := cmp.Or(cfg.RedirectURI, "http://127.0.0.1:0/callback")
	u, err := url.Parse(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("parse redirect URI: %w", err)
	}

	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", u.Host)
	if err != nil {
		return nil, fmt.Errorf("listen for OAuth2 callback: %w", err)
	}
	u.Host = ln.Addr().String()

	mux := http.NewServeMux()
	return &Provider{
		provider: p,
		listener: ln,
		mux:      mux,
		callback: http.Server{
			Handler:           mux,
			Addr:              u.Host,
			ReadHeaderTimeout: maxReadHeaderTimeout,
		},
		cfg: oauth2.Config{
			ClientID:    cfg.ClientID,
			Endpoint:    p.Endpoint(),
			Scopes:      cfg.Scopes,
			RedirectURL: u.String(),
		},
		codeVerifier: verifier,
		cache:        newCache(cacheDir),
		cacheKey:     cacheKey(cfg.Issuer, cfg.ClientID, cfg.Scopes),
	}, nil
}

// AcquireToken performs the OAuth2 authorization code flow and returns
// the id_token from the token response. The id_token is a signed JWT
// suitable for use as a federated credential with Azure workload
// identity login (az login --federated-token).
//
// Cached tokens are returned when available unless [wif.WithForceRefresh]
// is passed.
func (p *Provider) AcquireToken(ctx context.Context, opts ...wif.AcquireOption) (idToken string, cached bool, err error) {
	o := wif.ApplyOptions(opts)

	if !o.ForceRefresh {
		if tok, ok := p.cache.get(p.cacheKey); ok {
			return tok, true, nil
		}
	}

	state, err := randomState()
	if err != nil {
		return "", false, fmt.Errorf("generate OAuth2 state: %w", err)
	}

	p.callback.BaseContext = func(net.Listener) context.Context { return ctx }
	serverErr := p.startServer()
	defer func() {
		if sErr := p.shutdown(ctx); sErr != nil {
			err = errors.Join(err, fmt.Errorf("shutdown callback server: %w", sErr))
		}
	}()

	p.promptAuthorization(ctx, state)

	code, err := p.awaitCallback(ctx, state, serverErr)
	if err != nil {
		return "", false, err
	}

	idToken, err = p.exchangeForIDToken(ctx, code)
	if err != nil {
		return "", false, err
	}

	_ = p.cache.put(p.cacheKey, idToken)
	return idToken, false, nil
}

// startServer starts the HTTP callback server in a background goroutine
// and returns a channel that receives the server's terminal error (nil
// on graceful shutdown).
func (p *Provider) startServer() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		if err := p.callback.Serve(p.listener); errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
		} else {
			errCh <- err
		}
	}()
	return errCh
}

// shutdown gracefully stops the callback server.
func (p *Provider) shutdown(ctx context.Context) error {
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	return p.callback.Shutdown(c)
}

// promptAuthorization builds the authorization URL, attempts to open it
// in the user's browser, and prints it to stdout.
func (p *Provider) promptAuthorization(ctx context.Context, state string) {
	opts := []oauth2.AuthCodeOption{}
	if p.codeVerifier != "" {
		challenge := s256Challenge(p.codeVerifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	authURL := p.cfg.AuthCodeURL(state, opts...)

	if err := openBrowser(ctx, authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser automatically: %v\n\n", err)
	}
	fmt.Fprintf(os.Stdout, "Open this URL to authorize:\n\n%s\n\n", authURL)
}

// awaitCallback registers the callback handler and blocks until an
// authorization code is received, an error occurs, the server stops, or
// the context is canceled.
func (p *Provider) awaitCallback(ctx context.Context, state string, serverErr <-chan error) (string, error) {
	redirectURL, err := url.Parse(p.cfg.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("parse configured redirect URL: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	p.mux.HandleFunc(redirectURL.Path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if gotState := query.Get("state"); gotState != state {
			trySend(errCh, fmt.Errorf("invalid OAuth2 state"))
			http.Error(w, "Invalid OAuth2 state. You can close this window.", http.StatusBadRequest)
			return
		}

		if errMsg := query.Get("error"); errMsg != "" {
			description := query.Get("error_description")
			trySend(errCh, fmt.Errorf("authorization error: %s - %s", errMsg, description))
			http.Error(w, "Authorization failed. You can close this window.", http.StatusBadRequest)
			return
		}

		code := query.Get("code")
		if code == "" {
			trySend(errCh, fmt.Errorf("authorization code not found in callback"))
			http.Error(w, "Authorization code not found. You can close this window.", http.StatusBadRequest)
			return
		}

		trySend(codeCh, code)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
	})

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case err := <-serverErr:
		return "", fmt.Errorf("callback server stopped unexpectedly: %w", err)
	case <-ctx.Done():
		return "", fmt.Errorf("authorization canceled: %w", ctx.Err())
	}
}

// exchangeForIDToken exchanges the authorization code for a token
// response and extracts the id_token claim.
func (p *Provider) exchangeForIDToken(ctx context.Context, code string) (string, error) {
	exchangeCtx, cancel := context.WithTimeout(ctx, exchangeTimeout)
	defer cancel()

	opts := []oauth2.AuthCodeOption{}
	if p.codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", p.codeVerifier))
	}

	token, err := p.cfg.Exchange(exchangeCtx, code, opts...)
	if err != nil {
		return "", fmt.Errorf("exchange authorization code: %w", err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok || idToken == "" {
		return "", fmt.Errorf("token response did not contain an id_token")
	}

	return idToken, nil
}

// randomState generates a cryptographically random state parameter.
func randomState() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// randomVerifier generates a cryptographically random PKCE code verifier.
func randomVerifier() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// s256Challenge computes the S256 code challenge for the given verifier.
func s256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// trySend sends a value on the channel without blocking.
func trySend[T any](ch chan<- T, v T) {
	select {
	case ch <- v:
	default:
	}
}
