package oauth2

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeJWT(t *testing.T, exp time.Time) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload, err := json.Marshal(map[string]any{"exp": exp.Unix()})
	require.NoError(t, err)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	return header + "." + payloadB64 + ".signature"
}

func TestCache_PutGet(t *testing.T) {
	dir := t.TempDir()
	c := newCache(dir)

	token := makeJWT(t, time.Now().Add(time.Hour))
	key := cacheKey("https://issuer", "client-id", []string{"openid", "profile"})

	require.NoError(t, c.put(key, token))

	got, ok := c.get(key)
	assert.True(t, ok)
	assert.Equal(t, token, got)
}

func TestCache_Expired(t *testing.T) {
	dir := t.TempDir()
	c := newCache(dir)

	token := makeJWT(t, time.Now().Add(-time.Minute))
	key := cacheKey("https://issuer", "client-id", []string{"openid"})

	require.NoError(t, c.put(key, token))

	_, ok := c.get(key)
	assert.False(t, ok)
}

func TestCache_Invalidate(t *testing.T) {
	dir := t.TempDir()
	c := newCache(dir)

	token := makeJWT(t, time.Now().Add(time.Hour))
	key := cacheKey("https://issuer", "client-id", []string{"openid"})

	require.NoError(t, c.put(key, token))
	require.NoError(t, c.invalidate(key))

	_, ok := c.get(key)
	assert.False(t, ok)
}

func TestCacheKey_ScopeOrderIndependent(t *testing.T) {
	k1 := cacheKey("https://issuer", "cid", []string{"openid", "profile", "email"})
	k2 := cacheKey("https://issuer", "cid", []string{"email", "openid", "profile"})
	assert.Equal(t, k1, k2)
}

func TestCache_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	c := newCache(dir)

	token := makeJWT(t, time.Now().Add(time.Hour))
	require.NoError(t, c.put("key", token))

	info, err := os.Stat(filepath.Join(dir, cacheFile))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
