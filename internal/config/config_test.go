package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tarotclub/go-tpl/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// Change to a temp directory so no config.yaml is found.
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(tmp)

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.App.Name != "go-tpl" {
		t.Errorf("App.Name = %q; want %q", cfg.App.Name, "go-tpl")
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %q; want %q", cfg.Log.Level, "info")
	}
}

func TestLoadFromFile(t *testing.T) {
	content := `
app:
  name: myapp
  version: 1.2.3
log:
  level: debug
`
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.App.Name != "myapp" {
		t.Errorf("App.Name = %q; want %q", cfg.App.Name, "myapp")
	}
	if cfg.App.Version != "1.2.3" {
		t.Errorf("App.Version = %q; want %q", cfg.App.Version, "1.2.3")
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q; want %q", cfg.Log.Level, "debug")
	}
}
