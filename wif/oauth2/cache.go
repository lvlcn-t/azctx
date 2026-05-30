package oauth2

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json/v2"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	// cacheFile is the name of the file used to store ephemeral cached tokens on disk.
	// It uses a generic name to reduce exposure to filesystem
	// credential scraping by supply chain malware that targets predictable
	// filenames containing "token", "secret", or "credential".
	cacheFile = "state.json"
	// clockSkew is the amount of time subtracted from the token's expiry to account for clock skew.
	clockSkew = 60 * time.Second
	// dirPerms set restrictive permissions on the cache directory
	dirPerms fs.FileMode = 0o700
	// filePerms set restrictive permissions on the cache file
	filePerms fs.FileMode = 0o600
)

type cacheEntry struct {
	ExpiresAt time.Time `json:"expires_at"`
	IDToken   string    `json:"id_token"`
}

type cache struct {
	path string
	mu   sync.Mutex
}

func newCache(dir string) *cache {
	return &cache{path: filepath.Join(dir, cacheFile)}
}

// cacheKey produces a deterministic key from the token source parameters.
func cacheKey(issuer, clientID string, scopes []string) string {
	sorted := slices.Clone(scopes)
	slices.Sort(sorted)
	raw := issuer + "|" + clientID + "|" + strings.Join(sorted, " ")
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (c *cache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := c.readFile()
	if err != nil {
		return "", false
	}

	entry, ok := entries[key]
	if !ok {
		return "", false
	}

	if time.Now().After(entry.ExpiresAt.Add(-clockSkew)) {
		return "", false
	}

	return entry.IDToken, true
}

func (c *cache) put(key, idToken string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp, err := parseExp(idToken)
	if err != nil {
		return err
	}

	entries, _ := c.readFile()
	if entries == nil {
		entries = make(map[string]cacheEntry)
	}

	entries[key] = cacheEntry{IDToken: idToken, ExpiresAt: exp}
	return c.writeFile(entries)
}

func (c *cache) invalidate(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := c.readFile()
	if err != nil {
		return nil //nolint:nilerr // nothing to invalidate
	}

	delete(entries, key)
	return c.writeFile(entries)
}

func (c *cache) readFile() (map[string]cacheEntry, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, err
	}

	var entries map[string]cacheEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *cache) writeFile(entries map[string]cacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(c.path), dirPerms); err != nil {
		return err
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, filePerms)
}

// parseExp extracts the exp claim from a JWT without verifying the signature.
func parseExp(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("decode JWT payload: %w", err)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err = json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, fmt.Errorf("parse JWT claims: %w", err)
	}

	return time.Unix(claims.Exp, 0), nil
}
