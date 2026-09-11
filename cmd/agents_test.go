package cmd

import (
	"strings"
	"testing"
)

// No installed CLI says nothing about the app's native worker capabilities.
func TestAgentsWithoutCLIsDoesNotRuleOutNativeSubagents(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	out, err := runRespecErr(t, "agents", "--json=false")
	if err != nil {
		t.Fatalf("agents: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Native subagents are not probed") {
		t.Errorf("agents must explain the probe's limit:\n%s", out)
	}
	if strings.Contains(out, "phases run in the orchestrator session") {
		t.Errorf("missing CLIs must not disable native workers:\n%s", out)
	}
}
