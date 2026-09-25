package app

import (
	"basket/internal/config"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAppLifecycleStartAndClose(t *testing.T) {
	c, err := config.Load("../../config.yaml")
	if err != nil {
		c = config.Config{}
	}
	c.Auth.Enabled = false
	c.Auth.DevUser = "USER1"
	c.Business.Timezone = "UTC"
	c.Business.AdminWorkers = 1
	c.Business.AdminQueue = 10
	c.Modules.Scheduler = true
	c.Modules.Cache = true
	c.Modules.Admin = true
	c.Scheduler.StartupCleanup = true
	c.Scheduler.DeleteScrips = true
	c.Scheduler.DeleteBaskets = true
	c.Scheduler.Cron = "*/5 * * * * *"

	db, err := gorm.Open(sqlite.Open("file:test_lifecycle?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	cache := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
	a, err := New(context.Background(), c, db, cache)
	if err != nil {
		t.Fatalf("New app failed: %v", err)
	}

	if err := a.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Close(ctx); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestAppLoggingAccessAndRest(t *testing.T) {
	a := testApp(t)
	a.Config.Logging.Database = true
	a.Config.Logging.MaxBodyBytes = 1024
	a.Config.Logging.AccessTablePattern = "tbl_access_log"
	a.Config.Logging.RestTablePattern = "tbl_rest_log"
	a.LogsDB = a.DB

	// Ensure tables exist for audit logs
	_ = a.DB.Exec("CREATE TABLE IF NOT EXISTS tbl_access_log (id INTEGER PRIMARY KEY, uri TEXT, method TEXT, req_body TEXT, res_body TEXT, status INTEGER, in_time DATETIME, out_time DATETIME, lag_time INTEGER, user_id TEXT, ucc TEXT, req_id TEXT, source TEXT, vendor TEXT, module TEXT, device_ip TEXT, user_agent TEXT, domain TEXT, content_type TEXT, session TEXT)")
	_ = a.DB.Exec("CREATE TABLE IF NOT EXISTS tbl_rest_log (id INTEGER PRIMARY KEY, url TEXT, method TEXT, req_body TEXT, res_body TEXT, in_time DATETIME, out_time DATETIME, total_time TEXT, module TEXT, user_id TEXT)")

	ctx := context.Background()
	req := httptest.NewRequest("GET", "/api/v1/test", strings.NewReader(`{"sample":"body"}`))
	cw := &captureWriter{ResponseWriter: httptest.NewRecorder(), body: []byte(`{"status":"ok"}`), status: 200}

	a.accessLog(req, route{}, time.Now(), `{"sample":"body"}`, cw)
	a.restLog(ctx, "https://api.example.com/upstream?token=secret", `{"req":"data"}`, `{"res":"data"}`, time.Now())
}

func TestCacheProvidersHttpAndFile(t *testing.T) {
	// 1. httpCache mock
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode("cached_value")
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	hCache := &httpCache{
		url:    server.URL,
		token:  "test_token",
		client: server.Client(),
	}

	ctx := context.Background()
	var val string
	if err := hCache.Get(ctx, "test_map", "test_key", &val); err != nil {
		t.Fatalf("httpCache.Get failed: %v", err)
	}
	if val != "cached_value" {
		t.Fatalf("expected cached_value, got %s", val)
	}
	if err := hCache.Put(ctx, "test_map", "test_key", "new_val"); err != nil {
		t.Fatalf("httpCache.Put failed: %v", err)
	}
	if err := hCache.Delete(ctx, "test_map", "test_key"); err != nil {
		t.Fatalf("httpCache.Delete failed: %v", err)
	}
	if err := hCache.Close(ctx); err != nil {
		t.Fatalf("httpCache.Close failed: %v", err)
	}

	// 2. NewCache branches
	cfg := config.Config{}
	cfg.Cache.Provider = "http"
	if _, err := NewCache(ctx, cfg); err == nil {
		t.Fatal("expected error when cache.url is empty for http provider")
	}

	cfg.Cache.URL = server.URL
	cInstance, err := NewCache(ctx, cfg)
	if err != nil {
		t.Fatalf("NewCache http failed: %v", err)
	}
	_ = cInstance.Close(ctx)

	// file cache with valid temp file
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "cache.json")
	sampleData := map[string]map[string]json.RawMessage{
		"map1": {"key1": json.RawMessage(`"val1"`)},
	}
	b, _ := json.Marshal(sampleData)
	if err := os.WriteFile(cacheFile, b, 0600); err != nil {
		t.Fatal(err)
	}

	cfg.Cache.Provider = "file"
	cfg.Cache.File = cacheFile
	fileCacheInstance, err := NewCache(ctx, cfg)
	if err != nil {
		t.Fatalf("NewCache file failed: %v", err)
	}
	_ = fileCacheInstance.Close(ctx)

	cfg.Cache.Provider = "unknown_provider"
	if _, err := NewCache(ctx, cfg); err == nil {
		t.Fatal("expected error for unknown cache provider")
	}
}

func TestOpenDatabaseUnsupported(t *testing.T) {
	cfg := config.Config{}
	cfg.Database.Driver = "unsupported_driver"
	_, err := OpenDatabase(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported db driver")
	}
}

func TestAppNewConfigurationBranches(t *testing.T) {
	ctx := context.Background()
	db, _ := gorm.Open(sqlite.Open("file:test_new_branches?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	cache := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}

	// 1. Invalid timezone
	c := config.Config{}
	c.Business.Timezone = "Invalid/Zone_Name"
	if _, err := New(ctx, c, db, cache); err == nil {
		t.Fatal("expected error for invalid timezone")
	}

	// 2. Non-existent CAFile
	c = config.Config{}
	c.Business.Timezone = "UTC"
	c.Upstream.CAFile = filepath.Join(t.TempDir(), "non_existent.pem")
	if _, err := New(ctx, c, db, cache); err == nil {
		t.Fatal("expected error for non-existent CA file")
	}

	// 3. Invalid PEM content in CAFile
	badPemFile := filepath.Join(t.TempDir(), "bad.pem")
	_ = os.WriteFile(badPemFile, []byte("NOT-A-PEM-CERTIFICATE"), 0600)
	c.Upstream.CAFile = badPemFile
	if _, err := New(ctx, c, db, cache); err == nil {
		t.Fatal("expected error for invalid PEM data")
	}

	// 4. Bad logging database config
	c.Upstream.CAFile = ""
	c.Logging.Database = true
	c.Logging.DatabaseConfig = &config.Database{Driver: "unsupported_driver"}
	if _, err := New(ctx, c, db, cache); err == nil {
		t.Fatal("expected error for invalid logging database driver")
	}
}

func TestAppHandlerEndpoints(t *testing.T) {
	a := testApp(t)
	a.Config.Server.AllowedOrigins = []string{"https://example.com"}
	a.Config.Modules.Token = true
	a.Config.Modules.Cache = true
	a.Config.Modules.Thematic = true
	a.Config.Logging.Access = true

	handler := a.Handler()

	// 1. CORS Preflight OPTIONS
	reqOpt := httptest.NewRequest("OPTIONS", "/token", nil)
	reqOpt.Header.Set("Origin", "https://example.com")
	recOpt := httptest.NewRecorder()
	handler.ServeHTTP(recOpt, reqOpt)
	if recOpt.Code != 204 {
		t.Fatalf("expected 204 for OPTIONS, got %d", recOpt.Code)
	}
	if recOpt.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("missing Access-Control-Allow-Origin header")
	}

	// 2. GET /token
	reqToken := httptest.NewRequest("GET", "/token", nil)
	recToken := httptest.NewRecorder()
	handler.ServeHTTP(recToken, reqToken)
	if recToken.Code != 200 || !strings.Contains(recToken.Body.String(), "username") {
		t.Fatalf("expected 200 with username, got %d: %s", recToken.Code, recToken.Body.String())
	}

	// 3. GET /token/logout
	reqLogout := httptest.NewRequest("GET", "/token/logout", nil)
	recLogout := httptest.NewRecorder()
	handler.ServeHTTP(recLogout, reqLogout)
	if recLogout.Code != 200 || !strings.Contains(recLogout.Body.String(), "You are logged out") {
		t.Fatalf("expected 200 with 'You are logged out', got %d", recLogout.Code)
	}

	// 4. GET /thematic/basket/get/category/header
	reqCat := httptest.NewRequest("GET", "/thematic/basket/get/category/header", nil)
	recCat := httptest.NewRecorder()
	handler.ServeHTTP(recCat, reqCat)
	if recCat.Code != 204 {
		t.Fatalf("expected 204 for category header, got %d", recCat.Code)
	}

	// 5. POST /cache/delete/expiry
	reqCache := httptest.NewRequest("POST", "/cache/delete/expiry", nil)
	recCache := httptest.NewRecorder()
	handler.ServeHTTP(recCache, reqCache)
	if recCache.Code != 200 {
		t.Fatalf("expected 200 for cache expiry, got %d", recCache.Code)
	}

	// 6. Unauthenticated request to protected endpoint (list vs non-list)
	a.Config.Auth.Enabled = true
	a.Config.Auth.Mode = "hmac"
	a.Config.Auth.HMACSecret = "secret"
	a.auth.config.Enabled = true
	a.auth.config.Mode = "hmac"
	a.auth.config.HMACSecret = "secret"

	// List endpoint (e.g. /thematic/basket/invest)
	reqList := httptest.NewRequest("POST", "/thematic/basket/invest", nil)
	recList := httptest.NewRecorder()
	handler.ServeHTTP(recList, reqList)
	if recList.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated list route, got %d", recList.Code)
	}

	// Non-list endpoint
	reqNonList := httptest.NewRequest("POST", "/basketorder/create", nil)
	recNonList := httptest.NewRecorder()
	handler.ServeHTTP(recNonList, reqNonList)
	if recNonList.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated non-list route, got %d", recNonList.Code)
	}
}
