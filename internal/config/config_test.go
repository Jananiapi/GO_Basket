package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLValidationAndSecretExpansion(t *testing.T) {
	raw, e := os.ReadFile("../../config.yaml")
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("BASKET_DB_DSN", "user:p#ss:word@tcp(localhost)/basket?parseTime=true")
	t.Setenv("BASKET_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("BASKET_OIDC_AUDIENCE", "basket")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if e = os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	c, e := Load(path)
	if e != nil {
		t.Fatal(e)
	}
	if c.Database.Driver != "mysql" || c.Database.DSN != os.Getenv("BASKET_DB_DSN") {
		t.Fatal("defaults or DSN changed")
	}
	bad := strings.Replace(string(raw), "  max_scrips: 20", "  max_scrips_typo: 20", 1)
	os.WriteFile(path, []byte(bad), 0600)
	if _, e = Load(path); e == nil {
		t.Fatal("unknown config field accepted")
	}
}
