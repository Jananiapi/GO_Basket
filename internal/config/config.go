package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        Server        `yaml:"server"`
	Database      Database      `yaml:"database"`
	Modules       Modules       `yaml:"modules"`
	Auth          Auth          `yaml:"auth"`
	Cache         Cache         `yaml:"cache"`
	Upstream      Upstream      `yaml:"upstream"`
	Business      Business      `yaml:"business"`
	Scheduler     Scheduler     `yaml:"scheduler"`
	Notifications Notifications `yaml:"notifications"`
	Logging       Logging       `yaml:"logging"`
}
type Server struct {
	Address         string        `yaml:"address"`
	BasePath        string        `yaml:"base_path"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxBodyBytes    int64         `yaml:"max_body_bytes"`
	AllowedOrigins  []string      `yaml:"allowed_origins"`
}
type Database struct {
	Driver          string        `yaml:"driver"`
	DSN             string        `yaml:"dsn"`
	AutoMigrate     bool          `yaml:"auto_migrate"`
	MaxOpen         int           `yaml:"max_open"`
	MaxIdle         int           `yaml:"max_idle"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}
type Modules struct {
	Basket        bool `yaml:"basket"`
	Admin         bool `yaml:"admin"`
	Research      bool `yaml:"research"`
	Thematic      bool `yaml:"thematic"`
	Holdings      bool `yaml:"holdings"`
	Margin        bool `yaml:"margin"`
	Cache         bool `yaml:"cache"`
	Notifications bool `yaml:"notifications"`
	Scheduler     bool `yaml:"scheduler"`
	Token         bool `yaml:"token"`
}
type Auth struct {
	UserInfoRequired bool   `yaml:"userinfo_required"`
	UserInfoURL      string `yaml:"userinfo_url"`

	Enabled     bool     `yaml:"enabled"`
	Mode        string   `yaml:"mode"`
	Issuer      string   `yaml:"issuer"`
	Audience    string   `yaml:"audience"`
	JWKSURL     string   `yaml:"jwks_url"`
	HMACSecret  string   `yaml:"hmac_secret"`
	UserClaim   string   `yaml:"user_claim"`
	DevUser     string   `yaml:"dev_user"`
	AdminToken  string   `yaml:"admin_token"`
	PublicPaths []string `yaml:"public_paths"`
}
type Cache struct {
	Provider  string            `yaml:"provider"`
	File      string            `yaml:"file"`
	URL       string            `yaml:"url"`
	Token     string            `yaml:"token"`
	Cluster   string            `yaml:"cluster"`
	Addresses []string          `yaml:"addresses"`
	Maps      map[string]string `yaml:"maps"`
	Redis     RedisConfig       `yaml:"redis"`
}
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
type Upstream struct {
	CAFile                string `yaml:"ca_file"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`

	OrderURL         string        `yaml:"order_url"`
	SpanURL          string        `yaml:"span_url"`
	BasketMarginURL  string        `yaml:"basket_margin_url"`
	OrderBookURL     string        `yaml:"order_book_url"`
	Timeout          time.Duration `yaml:"timeout"`
	MaxResponseBytes int64         `yaml:"max_response_bytes"`
}
type Business struct {
	MaxScrips           int      `yaml:"max_scrips"`
	AdminExchanges      []string `yaml:"admin_exchanges"`
	Timezone            string   `yaml:"timezone"`
	MarketOpen          string   `yaml:"market_open"`
	MarketClose         string   `yaml:"market_close"`
	MarketDays          []int    `yaml:"market_days"`
	InvestBuffer        string   `yaml:"invest_buffer"`
	EquityMarginDivisor string   `yaml:"equity_margin_divisor"`
	SpanAccount         string   `yaml:"span_account"`
	SpanProduct         string   `yaml:"span_product"`
	ClosedResearchDays  int      `yaml:"closed_research_days"`
	ExecutionMessage    string   `yaml:"execution_message"`
	AdminWorkers        int      `yaml:"admin_workers"`
	AdminQueue          int      `yaml:"admin_queue"`
	ProductAliases      string   `yaml:"product_aliases"`
	LegacyNestFirstOnly bool     `yaml:"legacy_nest_first_only"`
}
type Scheduler struct {
	Cron           string `yaml:"cron"`
	DeleteScrips   bool   `yaml:"delete_scrips"`
	DeleteBaskets  bool   `yaml:"delete_baskets"`
	StartupCleanup bool   `yaml:"startup_cleanup"`
}
type Notifications struct {
	Provider        string `yaml:"provider"`
	URL             string `yaml:"url"`
	APIKey          string `yaml:"api_key"`
	CredentialsFile string `yaml:"credentials_file"`
	ProjectID       string `yaml:"project_id"`
	BatchSize       int    `yaml:"batch_size"`
}
type Logging struct {
	DatabaseConfig     *Database `yaml:"database_config"`
	AccessTablePattern string    `yaml:"access_table_pattern"`
	RestTablePattern   string    `yaml:"rest_table_pattern"`
	MaxBodyBytes       int       `yaml:"max_body_bytes"`

	Access                bool   `yaml:"access"`
	SQL                   bool   `yaml:"sql"`
	Database              bool   `yaml:"database"`
	ClickHouseURL         string `yaml:"clickhouse_url"`
	ClickHouseUser        string `yaml:"clickhouse_user"`
	ClickHousePassword    string `yaml:"clickhouse_password"`
	ClickHouseDatabase    string `yaml:"clickhouse_database"`
	ClickHouseAccessTable string `yaml:"clickhouse_access_table"`
	ClickHouseRestTable   string `yaml:"clickhouse_rest_table"`
}

/*
func loadDotEnv(paths ...string) {
	seen := make(map[string]bool)
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err == nil {
			if seen[abs] {
				continue
			}
			seen[abs] = true
		}
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if key == "" {
				continue
			}
			if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
				val = val[1 : len(val)-1]
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
		_ = f.Close()
	}
}
*/

// added for sonarqube
func parseEnvLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return "", "", false
	}
	if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
		val = val[1 : len(val)-1]
	}
	return key, val, true
}

// added for sonarqube
func processDotEnvFile(p string, seen map[string]bool) {
	if p == "" {
		return
	}
	abs, err := filepath.Abs(p)
	if err == nil {
		if seen[abs] {
			return
		}
		seen[abs] = true
	}
	f, err := os.Open(p)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if key, val, ok := parseEnvLine(scanner.Text()); ok {
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
}

// added for sonarqube
func loadDotEnv(paths ...string) {
	seen := make(map[string]bool)
	for _, p := range paths {
		processDotEnvFile(p, seen)
	}
}

func Load(path string) (Config, error) {
	loadDotEnv(filepath.Join(filepath.Dir(path), ".env"), ".env")
	var c Config
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	// Expand scalar values after parsing: secrets cannot inject YAML structure.
	var node yaml.Node
	if e = yaml.Unmarshal(b, &node); e != nil {
		return c, e
	}
	var expand func(*yaml.Node)
	expand = func(n *yaml.Node) {
		if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
			n.Value = os.ExpandEnv(n.Value)
		}
		for _, v := range n.Content {
			expand(v)
		}
	}
	expand(&node)
	b, e = yaml.Marshal(&node)
	if e != nil {
		return c, e
	}
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if e = d.Decode(&c); e != nil {
		return c, e
	}
	return c, c.Validate()
}

/*
func (c Config) Validate() error {
	if c.Server.Address == "" || c.Server.MaxBodyBytes <= 0 || c.Upstream.Timeout <= 0 || c.Upstream.MaxResponseBytes <= 0 {
		return fmt.Errorf("server address, body limit and upstream limits are required")
	}
	if !strings.HasPrefix(c.Server.BasePath, "/") && c.Server.BasePath != "" {
		return fmt.Errorf("base_path must start with /")
	}
	if c.Database.Driver == "" || c.Database.DSN == "" {
		return fmt.Errorf("database driver and dsn are required")
	}
	if c.Business.MaxScrips < 1 || c.Business.AdminWorkers < 1 || c.Business.AdminQueue < 1 {
		return fmt.Errorf("business limits must be positive")
	}
	if _, e := time.LoadLocation(c.Business.Timezone); e != nil {
		return e
	}
	for _, s := range []string{c.Business.MarketOpen, c.Business.MarketClose} {
		if _, e := time.Parse("15:04", s); e != nil {
			return e
		}
	}
	if c.Auth.Enabled && (c.Auth.Audience == "" || c.Auth.Issuer == "") {
		return fmt.Errorf("auth issuer and audience are required")
	}
	if c.Auth.Enabled && c.Auth.Mode == "hmac" && c.Auth.HMACSecret == "" {
		return fmt.Errorf("auth hmac_secret is required")
	}
	if !c.Auth.Enabled && c.Auth.DevUser == "" {
		return fmt.Errorf("auth.dev_user is required when auth is disabled")
	}
	if c.Modules.Notifications && (c.Notifications.BatchSize < 1 || c.Notifications.BatchSize > 500) {
		return fmt.Errorf("notification batch_size must be 1..500")
	}
	return nil
}
*/

// added for sonarqube
func (c Config) validateServerAndUpstream() error {
	if c.Server.Address == "" || c.Server.MaxBodyBytes <= 0 || c.Upstream.Timeout <= 0 || c.Upstream.MaxResponseBytes <= 0 {
		return fmt.Errorf("server address, body limit and upstream limits are required")
	}
	if !strings.HasPrefix(c.Server.BasePath, "/") && c.Server.BasePath != "" {
		return fmt.Errorf("base_path must start with /")
	}
	if c.Database.Driver == "" || c.Database.DSN == "" {
		return fmt.Errorf("database driver and dsn are required")
	}
	return nil
}

// added for sonarqube
func (c Config) validateBusiness() error {
	if c.Business.MaxScrips < 1 || c.Business.AdminWorkers < 1 || c.Business.AdminQueue < 1 {
		return fmt.Errorf("business limits must be positive")
	}
	if _, e := time.LoadLocation(c.Business.Timezone); e != nil {
		return e
	}
	for _, s := range []string{c.Business.MarketOpen, c.Business.MarketClose} {
		if _, e := time.Parse("15:04", s); e != nil {
			return e
		}
	}
	return nil
}

// added for sonarqube
func (c Config) validateAuth() error {
	if c.Auth.Enabled && (c.Auth.Audience == "" || c.Auth.Issuer == "") {
		return fmt.Errorf("auth issuer and audience are required")
	}
	if c.Auth.Enabled && c.Auth.Mode == "hmac" && c.Auth.HMACSecret == "" {
		return fmt.Errorf("auth hmac_secret is required")
	}
	if !c.Auth.Enabled && c.Auth.DevUser == "" {
		return fmt.Errorf("auth.dev_user is required when auth is disabled")
	}
	return nil
}

// added for sonarqube
func (c Config) validateModules() error {
	if c.Modules.Notifications && (c.Notifications.BatchSize < 1 || c.Notifications.BatchSize > 500) {
		return fmt.Errorf("notification batch_size must be 1..500")
	}
	return nil
}

// added for sonarqube
func (c Config) Validate() error {
	if err := c.validateServerAndUpstream(); err != nil {
		return err
	}
	if err := c.validateBusiness(); err != nil {
		return err
	}
	if err := c.validateAuth(); err != nil {
		return err
	}
	return c.validateModules()
}
