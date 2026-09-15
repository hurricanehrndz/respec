package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHome points HOME at a temp dir so config Load/Save touch a throwaway
// ~/.config/respec/config.yaml.
func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func TestLoadDefaultsWhenMissing(t *testing.T) {
	withTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Store != "~/respec-store" {
		t.Errorf("Store = %q, want ~/respec-store", cfg.Store)
	}
	if cfg.ReflowWidth != 80 {
		t.Errorf("ReflowWidth = %d, want 80", cfg.ReflowWidth)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempHome(t)
	in := Defaults()
	in.Store = "~/my-store"
	in.TemplatesDir = "/tmp/tpl"
	in.Context = "shared context"
	in.Rules.Plan = "always co-generate"
	if err := Save(in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.Store != in.Store || out.TemplatesDir != in.TemplatesDir ||
		out.Context != in.Context || out.Rules.Plan != in.Rules.Plan {
		t.Errorf("round-trip mismatch: got %+v want %+v", out, in)
	}
}

func TestSaveCreatesFile(t *testing.T) {
	home := withTempHome(t)
	if err := Save(Defaults()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	p := filepath.Join(home, ".config", "respec", "config.yaml")
	if _, err := os.Stat(p); err != nil {
		t.Errorf("config file not created at %s: %v", p, err)
	}
}

func TestExpand(t *testing.T) {
	home := withTempHome(t)
	t.Setenv("RESPEC_ROOT", filepath.Join(home, "custom-root"))
	cases := map[string]string{
		"~/respec-store":           filepath.Join(home, "respec-store"),
		"~":                        home,
		"$HOME/respec-store":       filepath.Join(home, "respec-store"),
		"${RESPEC_ROOT}/templates": filepath.Join(home, "custom-root", "templates"),
		"/abs/path":                "/abs/path",
		"relative":                 "relative",
	}
	for in, want := range cases {
		if got := Expand(in); got != want {
			t.Errorf("Expand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetSetRoundTrip(t *testing.T) {
	withTempHome(t)
	cfg := Defaults()
	if err := cfg.Set("reflow_width", "100"); err != nil {
		t.Fatalf("Set reflow_width: %v", err)
	}
	if err := cfg.Set("rules.research", "explore first"); err != nil {
		t.Fatalf("Set rules.research: %v", err)
	}
	if v, _ := cfg.Get("reflow_width"); v != "100" {
		t.Errorf("Get reflow_width = %q, want 100", v)
	}
	if v, _ := cfg.Get("rules.research"); v != "explore first" {
		t.Errorf("Get rules.research = %q, want 'explore first'", v)
	}
	if err := cfg.Set("reflow_width", "notanint"); err == nil {
		t.Error("Set reflow_width with non-int should error")
	}
	if _, err := cfg.Get("bogus"); err == nil {
		t.Error("Get of unknown key should error")
	}
}
