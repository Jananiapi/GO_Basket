package app

import (
	"basket/internal/config"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
)

type Identity struct {
	UserID, Name, UCC, Token, Scope string
	Claims                          map[string]any
}
type identityKey struct{}

// userInfoCacheEntry stores a cached /userinfo response with a wall-clock expiry.
type userInfoCacheEntry struct {
	claims  map[string]any
	expires time.Time
}

// userInfoCache is a bounded, TTL-based cache for Keycloak /userinfo responses.
// Keys are SHA-256 hex digests of the access token — the raw token is never stored.
// Max entries prevent unbounded memory growth; oldest entries are evicted on overflow.
type userInfoCache struct {
	mu      sync.Mutex
	entries map[string]userInfoCacheEntry
	ttl     time.Duration
	maxSize int
}

func newUserInfoCache(ttl time.Duration, maxSize int) *userInfoCache {
	return &userInfoCache{
		entries: make(map[string]userInfoCacheEntry, maxSize),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

func (c *userInfoCache) get(tokenHash string) (map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[tokenHash]
	if !ok || time.Now().After(entry.expires) {
		if ok {
			delete(c.entries, tokenHash)
		}
		return nil, false
	}
	// Return a shallow copy so callers cannot mutate cached data.
	cp := make(map[string]any, len(entry.claims))
	for k, v := range entry.claims {
		cp[k] = v
	}
	return cp, true
}

func (c *userInfoCache) put(tokenHash string, claims map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Evict expired entries first if at capacity.
	if len(c.entries) >= c.maxSize {
		now := time.Now()
		for k, v := range c.entries {
			if now.After(v.expires) {
				delete(c.entries, k)
			}
		}
	}
	// If still at capacity, evict the earliest-expiring entry.
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestExp time.Time
		for k, v := range c.entries {
			if oldestKey == "" || v.expires.Before(oldestExp) {
				oldestKey = k
				oldestExp = v.expires
			}
		}
		delete(c.entries, oldestKey)
	}
	cp := make(map[string]any, len(claims))
	for k, v := range claims {
		cp[k] = v
	}
	c.entries[tokenHash] = userInfoCacheEntry{claims: cp, expires: time.Now().Add(c.ttl)}
}

// tokenHash returns a hex-encoded SHA-256 digest of the access token.
// This is used as the cache key so the raw token is never stored in memory.
func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

type Authenticator struct {
	config        config.Auth
	verifier      *oidc.IDTokenVerifier
	http          *http.Client
	userInfoURL   string
	userInfoCache *userInfoCache
}

func newAuth(ctx context.Context, c config.Config) (*Authenticator, error) {
	// Dedicated transport for auth HTTP calls — prevents DefaultTransport's
	// MaxIdleConnsPerHost=2 from starving Keycloak connections under load.
	authTransport := http.DefaultTransport.(*http.Transport).Clone()
	authTransport.MaxIdleConns = 20
	authTransport.MaxIdleConnsPerHost = 10
	authTransport.IdleConnTimeout = 90 * time.Second
	a := &Authenticator{
		config:        c.Auth,
		http:          &http.Client{Timeout: c.Upstream.Timeout, Transport: authTransport},
		userInfoURL:   c.Auth.UserInfoURL,
		userInfoCache: newUserInfoCache(60*time.Second, 10000),
	}
	if !c.Auth.Enabled || c.Auth.Mode == "hmac" {
		return a, nil
	}
	if c.Auth.Mode != "oidc" {
		return nil, fmt.Errorf("unsupported auth mode %q", c.Auth.Mode)
	}
	ctx = oidc.ClientContext(ctx, a.http)
	cfg := &oidc.Config{ClientID: c.Auth.Audience}
	if c.Auth.JWKSURL != "" {
		a.verifier = oidc.NewVerifier(c.Auth.Issuer, oidc.NewRemoteKeySet(ctx, c.Auth.JWKSURL), cfg)
	} else {
		p, e := oidc.NewProvider(ctx, c.Auth.Issuer)
		if e != nil {
			return nil, e
		}
		a.verifier = p.Verifier(cfg)
		if a.userInfoURL == "" {
			a.userInfoURL = p.UserInfoEndpoint()
		}
	}
	if c.Auth.UserInfoRequired && a.userInfoURL == "" {
		return nil, fmt.Errorf("userinfo_url is required with explicit JWKS when userinfo_required is true")
	}
	return a, nil
}

// added for sonarqube
const headerAuthorization = "Authorization"

/*
func (a *Authenticator) verify(r *http.Request) (Identity, error) {
	id := Identity{}
	if !a.config.Enabled {
		id.UserID = strings.ToUpper(a.config.DevUser)
		//added for sonarqube
		// id.Token = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		id.Token = strings.TrimPrefix(r.Header.Get(headerAuthorization), "Bearer ")
		return id, nil
	}
	//added for sonarqube
	// h := strings.Fields(r.Header.Get("Authorization"))
	h := strings.Fields(r.Header.Get(headerAuthorization))
	if len(h) != 2 || !strings.EqualFold(h[0], "Bearer") {
		return id, fmt.Errorf("missing bearer token")
	}
	id.Token = h[1]
	var claims map[string]any
	if a.config.Mode == "oidc" {
		t, e := a.verifier.Verify(r.Context(), id.Token)
		if e != nil {
			return id, e
		}
		if e = t.Claims(&claims); e != nil {
			return id, e
		}
		if a.config.UserInfoRequired {
			request, e := http.NewRequestWithContext(r.Context(), "GET", a.userInfoURL, nil)
			if e != nil {
				return id, e
			}
			//added for sonarqube
			// request.Header.Set("Authorization", "Bearer "+id.Token)
			request.Header.Set(headerAuthorization, "Bearer "+id.Token)
			response, e := a.http.Do(request)
			if e != nil {
				return id, e
			}
			defer response.Body.Close()
			if response.StatusCode != 200 {
				return id, fmt.Errorf("userinfo HTTP %d", response.StatusCode)
			}
			var info map[string]any
			if e = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&info); e != nil {
				return id, e
			}
			if str(info["sub"]) == "" || str(info["sub"]) != t.Subject {
				return id, fmt.Errorf("userinfo subject mismatch")
			}
			for k, v := range info {
				switch k {
				case "sub", "iss", "aud", "exp", "iat", "nbf":
				default:
					claims[k] = v
				}
			}
		}
	} else {
		c := jwt.MapClaims{}
		_, e := jwt.ParseWithClaims(id.Token, c, func(t *jwt.Token) (any, error) { return []byte(a.config.HMACSecret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(a.config.Issuer), jwt.WithAudience(a.config.Audience), jwt.WithExpirationRequired())
		if e != nil {
			return id, e
		}
		claims = map[string]any(c)
	}
	id.UserID = strings.ToUpper(str(claims[a.config.UserClaim]))
	id.UCC = str(claims["ucc"])
	id.Name = str(claims["name"])
	id.Scope = str(claims["scope"])
	id.Claims = claims
	if id.UserID == "" {
		return id, fmt.Errorf("missing user claim")
	}
	return id, nil
}
*/

// added for sonarqube
func (a *Authenticator) fetchUserInfo(ctx context.Context, token string, subject string, claims map[string]any) error {
	request, e := http.NewRequestWithContext(ctx, "GET", a.userInfoURL, nil)
	if e != nil {
		return e
	}
	request.Header.Set(headerAuthorization, "Bearer "+token)
	response, e := a.http.Do(request)
	if e != nil {
		return e
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return fmt.Errorf("userinfo HTTP %d", response.StatusCode)
	}
	var info map[string]any
	if e = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&info); e != nil {
		return e
	}
	if str(info["sub"]) == "" || str(info["sub"]) != subject {
		return fmt.Errorf("userinfo subject mismatch")
	}
	for k, v := range info {
		switch k {
		case "sub", "iss", "aud", "exp", "iat", "nbf":
		default:
			claims[k] = v
		}
	}
	return nil
}

// added for sonarqube
func (a *Authenticator) verifyOIDC(ctx context.Context, token string) (map[string]any, error) {
	var claims map[string]any
	t, e := a.verifier.Verify(ctx, token)
	if e != nil {
		return nil, e
	}
	if e = t.Claims(&claims); e != nil {
		return nil, e
	}
	if a.config.UserInfoRequired {
		th := tokenHash(token)
		if cached, ok := a.userInfoCache.get(th); ok {
			// Cache hit — merge cached userinfo claims into JWT claims.
			for k, v := range cached {
				claims[k] = v
			}
		} else {
			// Cache miss — call Keycloak /userinfo and cache the result.
			if e := a.fetchUserInfo(ctx, token, t.Subject, claims); e != nil {
				return nil, e
			}
			// Cache only the non-standard claims that fetchUserInfo merged.
			infoOnly := make(map[string]any)
			for k, v := range claims {
				switch k {
				case "sub", "iss", "aud", "exp", "iat", "nbf":
				default:
					infoOnly[k] = v
				}
			}
			a.userInfoCache.put(th, infoOnly)
		}
	}
	return claims, nil
}

// added for sonarqube
func (a *Authenticator) verifyHMAC(token string) (map[string]any, error) {
	c := jwt.MapClaims{}
	_, e := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (any, error) { return []byte(a.config.HMACSecret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(a.config.Issuer), jwt.WithAudience(a.config.Audience), jwt.WithExpirationRequired())
	if e != nil {
		return nil, e
	}
	return map[string]any(c), nil
}

// added for sonarqube
func (a *Authenticator) verify(r *http.Request) (Identity, error) {
	id := Identity{}
	if !a.config.Enabled {
		id.UserID = strings.ToUpper(a.config.DevUser)
		id.Token = strings.TrimPrefix(r.Header.Get(headerAuthorization), "Bearer ")
		return id, nil
	}
	h := strings.Fields(r.Header.Get(headerAuthorization))
	if len(h) != 2 || !strings.EqualFold(h[0], "Bearer") {
		return id, fmt.Errorf("missing bearer token")
	}
	id.Token = h[1]
	var claims map[string]any
	var err error
	if a.config.Mode == "oidc" {
		claims, err = a.verifyOIDC(r.Context(), id.Token)
	} else {
		claims, err = a.verifyHMAC(id.Token)
	}
	if err != nil {
		return id, err
	}
	id.UserID = strings.ToUpper(str(claims[a.config.UserClaim]))
	id.UCC = str(claims["ucc"])
	id.Name = str(claims["name"])
	id.Scope = str(claims["scope"])
	id.Claims = claims
	if id.UserID == "" {
		return id, fmt.Errorf("missing user claim")
	}
	return id, nil
}
func identity(r *http.Request) Identity {
	v, _ := r.Context().Value(identityKey{}).(Identity)
	return v
}
func adminAuthorized(got, want string) bool {
	return want != "" && subtle.ConstantTimeCompare([]byte(strings.ToLower(got)), []byte(strings.ToLower(want))) == 1
}
