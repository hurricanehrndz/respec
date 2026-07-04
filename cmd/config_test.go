package cmd

import (
	"strings"
	"testing"
)

// `config print` must show the effective configuration — defaults applied — so
// the operator can inspect what respec will actually use even before a config
// file exists.
func TestConfigPrintsEffectiveConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // no config file: defaults apply

	out, err := runRespecErr(t, "config", "print")
	if err != nil {
		t.Fatalf("config print failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "store: ~/respec-store") || !strings.Contains(out, "reflow_width: 80") {
		t.Errorf("expected defaults in output, got:\n%s", out)
	}
}
