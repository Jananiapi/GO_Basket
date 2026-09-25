package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthentication(t *testing.T) {
	a := testApp(t)
	a.auth.config.Enabled = true
	a.auth.config.Mode = "hmac"
	a.auth.config.HMACSecret = "fixture-secret"
	a.auth.config.Issuer = "issuer"
	a.auth.config.Audience = "basket"
	a.auth.config.UserClaim = "preferred_username"
	sign := func(exp time.Time, secret string) string {
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iss": "issuer", "aud": "basket", "exp": exp.Unix(), "preferred_username": "user1"})
		s, e := tok.SignedString([]byte(secret))
		if e != nil {
			t.Fatal(e)
		}
		return s
	}
	for _, tc := range []struct {
		token  string
		status int
	}{{"", 401}, {sign(time.Now().Add(time.Hour), "bad"), 401}, {sign(time.Now().Add(-time.Hour), "fixture-secret"), 401}, {sign(time.Now().Add(time.Hour), "fixture-secret"), 200}} {
		r := httptest.NewRequest("GET", "/basketorder/get", nil)
		if tc.token != "" {
			r.Header.Set("Authorization", "Bearer "+tc.token)
		}
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Errorf("expected %d got %d", tc.status, w.Code)
		}
	}
}
func TestMalformedAndNullRequests(t *testing.T) {
	a := testApp(t)
	for _, body := range []string{"null", "{", "{} {}", `{"basketId":1,"scrips":null}`} {
		_, out := request(t, a, "POST", "/basketorder/add/scrips", body)
		if out.(map[string]any)["status"] != "Not ok" {
			t.Fatal(out)
		}
	}
	r := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
}
func TestLoggingRedaction(t *testing.T) {
	s := redact(`{"apiKey":"secret","scrips":[{"token":"123"}],"nested":{"password":"secret"}}`)
	if strings.Contains(s, "secret") || !strings.Contains(s, "123") {
		t.Fatal(s)
	}
}

func TestOIDCUserInfo(t *testing.T) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		t.Fatal(e)
	}
	var issuer string
	var userInfoHits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			jsonWrite(w, 200, M{"issuer": issuer, "jwks_uri": issuer + "/jwks", "userinfo_endpoint": issuer + "/userinfo", "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/jwks":
			jsonWrite(w, 200, M{"keys": []M{{"kty": "RSA", "kid": "fixture", "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": "AQAB"}}})
		case "/userinfo":
			userInfoHits++
			if r.Header.Get("Authorization") == "" {
				t.Error("userinfo missing authorization")
			}
			jsonWrite(w, 200, M{"sub": "subject", "preferred_username": "user1", "ucc": "ucc", "name": "User"})
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	issuer = server.URL
	a := testApp(t)
	c := a.Config
	c.Auth.Enabled = true
	c.Auth.Mode = "oidc"
	c.Auth.Issuer = issuer
	c.Auth.Audience = "basket"
	c.Auth.UserInfoRequired = true
	c.Auth.JWKSURL = ""
	auth, e := newAuth(context.Background(), c)
	if e != nil {
		t.Fatal(e)
	}
	a.auth = auth
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"iss": issuer, "aud": "basket", "sub": "subject", "exp": time.Now().Add(time.Hour).Unix()})
	token.Header["kid"] = "fixture"
	signed, e := token.SignedString(key)
	if e != nil {
		t.Fatal(e)
	}
	r := httptest.NewRequest("GET", "/basketorder/get", nil)
	r.Header.Set("Authorization", "Bearer "+signed)
	id, e := a.auth.verify(r)
	if e != nil || id.UserID != "USER1" || id.UCC != "ucc" {
		t.Fatal(id.UserID, e)
	}
	if userInfoHits != 1 {
		t.Fatalf("expected 1 userinfo call, got %d", userInfoHits)
	}

	// Second request with same token must hit cache - userInfoHits should remain 1
	id2, e2 := a.auth.verify(r)
	if e2 != nil || id2.UserID != "USER1" || id2.UCC != "ucc" {
		t.Fatal(id2.UserID, e2)
	}
	if userInfoHits != 1 {
		t.Fatalf("expected cached userinfo (hits=1), got %d hits", userInfoHits)
	}
}

func TestUserInfoCacheUnit(t *testing.T) {
	cache := newUserInfoCache(50*time.Millisecond, 2)
	k1 := tokenHash("token-1")
	k2 := tokenHash("token-2")
	k3 := tokenHash("token-3")

	cache.put(k1, map[string]any{"user": "u1"})
	val, ok := cache.get(k1)
	if !ok || val["user"] != "u1" {
		t.Fatalf("expected u1, got %v", val)
	}

	// Test cache mutation isolation
	val["user"] = "modified"
	val2, _ := cache.get(k1)
	if val2["user"] != "u1" {
		t.Fatalf("cache should be isolated from caller mutation")
	}

	// Test capacity eviction
	cache.put(k2, map[string]any{"user": "u2"})
	cache.put(k3, map[string]any{"user": "u3"}) // should evict one
	if _, ok := cache.get(k3); !ok {
		t.Fatalf("k3 should be present")
	}

	// Test TTL expiration
	time.Sleep(60 * time.Millisecond)
	if _, ok := cache.get(k3); ok {
		t.Fatalf("k3 should have expired")
	}
}
