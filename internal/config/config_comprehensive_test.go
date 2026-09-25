package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func validBaseConfig() Config {
	return Config{
		Server: Server{
			Address:      ":9009",
			BasePath:     "/api",
			MaxBodyBytes: 1048576,
		},
		Database: Database{
			Driver: "mysql",
			DSN:    "root:secret@tcp(localhost:3306)/basket",
		},
		Upstream: Upstream{
			Timeout:          10 * time.Second,
			MaxResponseBytes: 1048576,
		},
		Business: Business{
			MaxScrips:    20,
			AdminWorkers: 5,
			AdminQueue:   100,
			Timezone:     "Asia/Kolkata",
			MarketOpen:   "09:15",
			MarketClose:  "15:30",
		},
		Auth: Auth{
			Enabled:  true,
			Mode:     "oidc",
			Issuer:   "https://issuer.example.com",
			Audience: "basket",
		},
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	c := validBaseConfig()
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}

	// Also test BasePath == ""
	c.Server.BasePath = ""
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid config with empty BasePath, got: %v", err)
	}
}

const (
	errServerAndUpstreamRequired = "server address, body limit and upstream limits are required"
	errBusinessLimitsPositive    = "business limits must be positive"
	errDatabaseDriverRequired    = "database driver and dsn are required"
	errAuthIssuerAudienceReq     = "auth issuer and audience are required"
	errParsingTime               = "parsing time"
)

func TestConfigValidateServerAndUpstream(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
		errMsg string
	}{
		{
			name:   "empty server address",
			modify: func(c *Config) { c.Server.Address = "" },
			errMsg: errServerAndUpstreamRequired,
		},
		{
			name:   "zero max body bytes",
			modify: func(c *Config) { c.Server.MaxBodyBytes = 0 },
			errMsg: errServerAndUpstreamRequired,
		},
		{
			name:   "negative max body bytes",
			modify: func(c *Config) { c.Server.MaxBodyBytes = -1 },
			errMsg: errServerAndUpstreamRequired,
		},
		{
			name:   "zero upstream timeout",
			modify: func(c *Config) { c.Upstream.Timeout = 0 },
			errMsg: errServerAndUpstreamRequired,
		},
		{
			name:   "negative upstream response bytes",
			modify: func(c *Config) { c.Upstream.MaxResponseBytes = 0 },
			errMsg: errServerAndUpstreamRequired,
		},
		{
			name:   "base path without slash",
			modify: func(c *Config) { c.Server.BasePath = "api/v1" },
			errMsg: "base_path must start with /",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := validBaseConfig()
			tc.modify(&c)
			err := c.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Fatalf("expected error containing %q, got %v", tc.errMsg, err)
			}
		})
	}
}

func TestConfigValidateDatabaseAndBusiness(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
		errMsg string
	}{
		{
			name:   "empty db driver",
			modify: func(c *Config) { c.Database.Driver = "" },
			errMsg: errDatabaseDriverRequired,
		},
		{
			name:   "empty db dsn",
			modify: func(c *Config) { c.Database.DSN = "" },
			errMsg: errDatabaseDriverRequired,
		},
		{
			name:   "max scrips < 1",
			modify: func(c *Config) { c.Business.MaxScrips = 0 },
			errMsg: errBusinessLimitsPositive,
		},
		{
			name:   "admin workers < 1",
			modify: func(c *Config) { c.Business.AdminWorkers = 0 },
			errMsg: errBusinessLimitsPositive,
		},
		{
			name:   "admin queue < 1",
			modify: func(c *Config) { c.Business.AdminQueue = 0 },
			errMsg: errBusinessLimitsPositive,
		},
		{
			name:   "invalid timezone",
			modify: func(c *Config) { c.Business.Timezone = "Invalid/Nonexistent_Timezone" },
			errMsg: "unknown time zone",
		},
		{
			name:   "invalid market open time",
			modify: func(c *Config) { c.Business.MarketOpen = "25:99" },
			errMsg: errParsingTime,
		},
		{
			name:   "invalid market close format",
			modify: func(c *Config) { c.Business.MarketClose = "not-a-time" },
			errMsg: errParsingTime,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := validBaseConfig()
			tc.modify(&c)
			err := c.Validate()
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.errMsg)) {
				t.Fatalf("expected error containing %q, got %v", tc.errMsg, err)
			}
		})
	}
}

func TestConfigValidateAuth(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*Config)
		errMsg string
	}{
		{
			name: "auth enabled empty audience",
			modify: func(c *Config) {
				c.Auth.Enabled = true
				c.Auth.Audience = ""
			},
			errMsg: errAuthIssuerAudienceReq,
		},
		{
			name: "auth enabled empty issuer",
			modify: func(c *Config) {
				c.Auth.Enabled = true
				c.Auth.Issuer = ""
			},
			errMsg: errAuthIssuerAudienceReq,
		},
		{
			name: "auth enabled hmac empty secret",
			modify: func(c *Config) {
				c.Auth.Enabled = true
				c.Auth.Mode = "hmac"
				c.Auth.HMACSecret = ""
			},
			errMsg: "auth hmac_secret is required",
		},
		{
			name: "auth disabled empty dev user",
			modify: func(c *Config) {
				c.Auth.Enabled = false
				c.Auth.DevUser = ""
			},
			errMsg: "auth.dev_user is required when auth is disabled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := validBaseConfig()
			tc.modify(&c)
			err := c.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Fatalf("expected error containing %q, got %v", tc.errMsg, err)
			}
		})
	}

	// Valid HMAC
	hmacCfg := validBaseConfig()
	hmacCfg.Auth.Mode = "hmac"
	hmacCfg.Auth.HMACSecret = "super-secret"
	if err := hmacCfg.Validate(); err != nil {
		t.Fatalf("expected valid HMAC config, got %v", err)
	}

	// Valid DevUser
	devCfg := validBaseConfig()
	devCfg.Auth.Enabled = false
	devCfg.Auth.DevUser = "DEVELOPER1"
	if err := devCfg.Validate(); err != nil {
		t.Fatalf("expected valid disabled auth with DevUser, got %v", err)
	}
}

func TestConfigValidateNotifications(t *testing.T) {
	c := validBaseConfig()
	c.Modules.Notifications = true
	c.Notifications.BatchSize = 0
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "batch_size must be 1..500") {
		t.Fatalf("expected error for batch_size 0, got %v", err)
	}

	c.Notifications.BatchSize = 501
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "batch_size must be 1..500") {
		t.Fatalf("expected error for batch_size 501, got %v", err)
	}

	c.Notifications.BatchSize = 250
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid config for batch_size 250, got %v", err)
	}
}

func TestLoadDotEnvParsingAndPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Set an existing OS env var to test precedence
	t.Setenv("EXISTING_VAR_TEST", "original_os_val")

	content := `
# This is a comment
EMPTY_KEY=
   
export EXPORTED_VAR=exported_val
SINGLE_QUOTED='hello_single'
DOUBLE_QUOTED="hello_double"
SPACED_KEY = spaced_val
MALFORMED_LINE_NO_EQUALS
EXISTING_VAR_TEST=should_not_overwrite
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	loadDotEnv(envPath, "", envPath) // also test empty string and duplicate path

	if os.Getenv("EXPORTED_VAR") != "exported_val" {
		t.Errorf("expected 'exported_val', got %q", os.Getenv("EXPORTED_VAR"))
	}
	if os.Getenv("SINGLE_QUOTED") != "hello_single" {
		t.Errorf("expected 'hello_single', got %q", os.Getenv("SINGLE_QUOTED"))
	}
	if os.Getenv("DOUBLE_QUOTED") != "hello_double" {
		t.Errorf("expected 'hello_double', got %q", os.Getenv("DOUBLE_QUOTED"))
	}
	if os.Getenv("SPACED_KEY") != "spaced_val" {
		t.Errorf("expected 'spaced_val', got %q", os.Getenv("SPACED_KEY"))
	}
	if os.Getenv("EXISTING_VAR_TEST") != "original_os_val" {
		t.Errorf("expected OS env to take precedence ('original_os_val'), got %q", os.Getenv("EXISTING_VAR_TEST"))
	}
}

func TestLoadErrors(t *testing.T) {
	// Non-existent file
	if _, err := Load("non_existent_config_file.yaml"); err == nil {
		t.Fatal("expected error loading non-existent file, got nil")
	}

	// Malformed YAML
	tmpDir := t.TempDir()
	badYaml := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(badYaml, []byte("server: [unclosed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(badYaml); err == nil {
		t.Fatal("expected error for malformed YAML syntax, got nil")
	}
}
