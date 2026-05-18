package oauth2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestS256Challenge(t *testing.T) {
	// RFC 7636 Appendix B test vector:
	// verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	// expected challenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	const expected = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	got := s256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")
	assert.Equal(t, expected, got)
}

func TestRandomState(t *testing.T) {
	const iterations = 100
	seen := make(map[string]struct{}, iterations)

	for range iterations {
		s, err := randomState()
		require.NoError(t, err)
		assert.Len(t, s, 43)
		_, err = base64.RawURLEncoding.DecodeString(s)
		assert.NoError(t, err, "expected valid base64url encoding")
		seen[s] = struct{}{}
	}

	assert.Len(t, seen, iterations, "expected all values to be unique")
}

func TestRandomVerifier(t *testing.T) {
	const iterations = 100
	seen := make(map[string]struct{}, iterations)

	for range iterations {
		v, err := randomVerifier()
		require.NoError(t, err)
		assert.Len(t, v, 43)
		_, err = base64.RawURLEncoding.DecodeString(v)
		assert.NoError(t, err, "expected valid base64url encoding")
		seen[v] = struct{}{}
	}

	assert.Len(t, seen, iterations, "expected all values to be unique")
}

func TestTrySend(t *testing.T) {
	t.Run("delivers to empty channel", func(t *testing.T) {
		ch := make(chan string, 1)
		trySend(ch, "hello")
		assert.Equal(t, "hello", <-ch)
	})

	t.Run("does not block on full channel", func(t *testing.T) {
		ch := make(chan string, 1)
		ch <- "first"
		assert.NotPanics(t, func() {
			trySend(ch, "second")
		})
		assert.Equal(t, "first", <-ch)
	})
}

func TestAwaitCallback(t *testing.T) {
	const validState = "test-state-123"

	tests := []struct {
		name     string
		query    url.Values
		wantCode string
		wantErr  bool
	}{
		{
			name:     "valid code",
			query:    url.Values{"state": {validState}, "code": {"auth-code-abc"}},
			wantCode: "auth-code-abc",
		},
		{
			name:    "invalid state",
			query:   url.Values{"state": {"wrong"}, "code": {"auth-code-abc"}},
			wantErr: true,
		},
		{
			name:    "auth error from provider",
			query:   url.Values{"state": {validState}, "error": {"access_denied"}, "error_description": {"user declined"}},
			wantErr: true,
		},
		{
			name:    "missing code",
			query:   url.Values{"state": {validState}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestProvider(t)
			serverErr := p.startServer()
			t.Cleanup(func() { _ = p.shutdown(t.Context()) })

			callbackURL, err := url.Parse(p.cfg.RedirectURL)
			require.NoError(t, err)
			callbackURL.RawQuery = tt.query.Encode()

			// Make the callback request in a goroutine since awaitCallback blocks.
			ctx := t.Context()
			type result struct {
				code string
				err  error
			}
			resultCh := make(chan result, 1)
			go func() {
				code, err := p.awaitCallback(ctx, validState, serverErr)
				resultCh <- result{code, err}
			}()

			// Poll until awaitCallback has registered its handler (not 404).
			for {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, callbackURL.String(), http.NoBody)
				require.NoError(t, err)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				resp.Body.Close()
				if resp.StatusCode != http.StatusNotFound {
					break
				}
				runtime.Gosched()
			}

			r := <-resultCh
			if tt.wantErr {
				require.Error(t, r.err)
				return
			}
			require.NoError(t, r.err)
			assert.Equal(t, tt.wantCode, r.code)
		})
	}
}

func TestAwaitCallback_ContextCanceled(t *testing.T) {
	p := newTestProvider(t)
	serverErr := p.startServer()
	t.Cleanup(func() { _ = p.shutdown(t.Context()) })

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := p.awaitCallback(ctx, "some-state", serverErr)
	require.Error(t, err)
}

func TestAwaitCallback_ServerDied(t *testing.T) {
	p := newTestProvider(t)
	// Close the listener to force Serve to return an error.
	p.listener.Close()

	serverErr := p.startServer()

	ctx := t.Context()
	_, err := p.awaitCallback(ctx, "some-state", serverErr)
	require.Error(t, err)
}

func TestExchangeForIDToken(t *testing.T) {
	tests := []struct {
		name       string
		response   map[string]any
		statusCode int
		wantToken  string
		wantErr    bool
	}{
		{
			name: "success",
			response: map[string]any{
				"access_token": "access-xyz",
				"token_type":   "Bearer",
				"id_token":     "eyJ.test.sig",
			},
			statusCode: http.StatusOK,
			wantToken:  "eyJ.test.sig",
		},
		{
			name: "missing id_token",
			response: map[string]any{
				"access_token": "access-xyz",
				"token_type":   "Bearer",
			},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "exchange error",
			response:   map[string]any{"error": "invalid_grant"},
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				err := json.NewEncoder(w).Encode(tt.response)
				require.NoError(t, err)
			}))
			t.Cleanup(srv.Close)

			p := &Provider{
				cfg: oauth2.Config{
					ClientID: "test-client",
					Endpoint: oauth2.Endpoint{
						TokenURL: srv.URL + "/token",
					},
				},
			}

			got, err := p.exchangeForIDToken(t.Context(), "test-code")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantToken, got)
		})
	}
}

func TestExchangeForIDToken_SendsPKCEVerifier(t *testing.T) {
	const verifier = "test-verifier-value"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, verifier, r.PostForm.Get("code_verifier"))

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-xyz",
			"token_type":   "Bearer",
			"id_token":     "eyJ.test.sig",
		})
		require.NoError(t, err)
	}))
	t.Cleanup(srv.Close)

	p := &Provider{
		cfg: oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				TokenURL: srv.URL + "/token",
			},
		},
		codeVerifier: verifier,
	}

	got, err := p.exchangeForIDToken(t.Context(), "test-code")
	require.NoError(t, err)
	assert.Equal(t, "eyJ.test.sig", got)
}

func TestStartServerAndShutdown(t *testing.T) {
	p := newTestProvider(t)
	serverErr := p.startServer()

	// Verify server is accepting requests.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, p.cfg.RedirectURL, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode) // no handler registered yet

	// Shutdown should succeed.
	require.NoError(t, p.shutdown(t.Context()))

	// Server error channel should receive nil (graceful shutdown).
	select {
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server error channel")
	}
}

// newTestProvider constructs a Provider with a local listener and mux,
// bypassing OIDC discovery. Suitable for testing callback and exchange logic.
func newTestProvider(t *testing.T) *Provider {
	t.Helper()

	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := ln.Addr().String()
	redirectURL, err := url.Parse("http://" + addr + "/callback")
	require.NoError(t, err)
	oauth2URL, err := url.Parse("http://" + addr + "/token")
	require.NoError(t, err)

	mux := http.NewServeMux()
	return &Provider{
		listener: ln,
		mux:      mux,
		callback: http.Server{
			Handler:           mux,
			Addr:              addr,
			ReadHeaderTimeout: maxReadHeaderTimeout,
		},
		cfg: oauth2.Config{
			ClientID:    "test-client",
			RedirectURL: redirectURL.String(),
			Endpoint: oauth2.Endpoint{
				TokenURL: oauth2URL.String(),
			},
		},
		provider:     nil,
		codeVerifier: "",
	}
}
