package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/templates"
	"github.com/spf13/cobra"
)

func init() {
	var targetName string
	cmd := &cobra.Command{
		Use:   "install --target <pi|claude>",
		Short: "Render and install the prompt-templates and skill at user scope for an agent",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			target, ok := templates.Targets[targetName]
			if !ok {
				c.SilenceUsage = true
				return fmt.Errorf("unknown --target %q (valid: pi, claude)", targetName)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			feats := detectFeatures(c)
			rendered, err := templates.Render(cfg, target, feats)
			if err != nil {
				return err
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			return installFor(c, target, home, rendered)
		},
	}
	cmd.Flags().StringVar(&targetName, "target", "", "agent to install for: pi or claude (required)")
	_ = cmd.MarkFlagRequired("target")
	rootCmd.AddCommand(cmd)
}

// detectFeatures probes PATH for optional operator tooling the templates can
// reference, reporting what was (not) found so the operator knows to re-run
// install after installing a tool.
func detectFeatures(c *cobra.Command) templates.Features {
	feats := templates.Features{}
	if _, err := exec.LookPath("probe"); err == nil {
		feats.Probe = true
		_, _ = fmt.Fprintln(c.OutOrStdout(), "probe found on PATH: prompts include probe extract guidance")
	} else {
		_, _ = fmt.Fprintln(c.OutOrStdout(), "probe not found on PATH: prompts omit probe guidance (re-run install after installing it)")
	}
	return feats
}

// installFor writes the rendered prompts and skill into the target agent's
// user-scope directories.
func installFor(c *cobra.Command, target templates.Target, home string, rendered templates.Rendered) error {
	var promptDst func(name string) string
	var promptsDir, prefix, skillsDir string
	switch target {
	case templates.TargetPi:
		// pi installs prompts under the rsx: namespace; embedded asset files
		// cannot carry a colon (go:embed forbids it), so the prefix is
		// applied here.
		promptsDir = filepath.Join(home, ".pi", "agent", "prompts")
		prefix = "rsx:"
		skillsDir = filepath.Join(home, ".pi", "agent", "skills")
	case templates.TargetClaude:
		// Claude Code user-scope command names cannot contain a colon, so
		// the rsx namespace flattens to a hyphen: /rsx-research etc.
		promptsDir = filepath.Join(home, ".claude", "commands")
		prefix = "rsx-"
		skillsDir = filepath.Join(home, ".claude", "skills")
	default:
		return fmt.Errorf("no install layout for target %q", target.Name)
	}
	promptDst = func(name string) string { return filepath.Join(promptsDir, prefix+name) }

	for name, content := range rendered.Prompts {
		dst := promptDst(name)
		if err := writeFile(dst, content); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.OutOrStdout(), "installed prompt: %s\n", dst)
	}
	if err := pruneStalePrompts(c, promptsDir, prefix, rendered.Prompts); err != nil {
		return err
	}
	for rel, content := range rendered.Skills {
		dst := filepath.Join(skillsDir, filepath.FromSlash(rel))
		if err := writeFile(dst, content); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.OutOrStdout(), "installed skill:  %s\n", dst)
	}
	return nil
}

// pruneStalePrompts removes respec-owned prompts that are no longer rendered.
//
// Overwriting installed files is not enough when a prompt is retired: the old
// file keeps working as a slash command, pointing the operator at a workflow
// that no longer exists. Only the `rsx:`/`rsx-` namespace is touched, which
// respec owns by construction, and every removal is reported — re-running
// install recreates anything that should still be there.
func pruneStalePrompts(c *cobra.Command, dir, prefix string, current map[string]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		if _, kept := current[strings.TrimPrefix(e.Name(), prefix)]; kept {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.OutOrStdout(), "removed retired prompt: %s\n", path)
	}
	return nil
}

// writeFile creates parent dirs and writes content (idempotent overwrite).
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
