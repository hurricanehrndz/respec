package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	var skillsDir string
	switch target {
	case templates.TargetPi:
		// pi installs prompts under the rsx: namespace; embedded asset files
		// cannot carry a colon (go:embed forbids it), so the prefix is
		// applied here.
		promptsDir := filepath.Join(home, ".pi", "agent", "prompts")
		promptDst = func(name string) string { return filepath.Join(promptsDir, "rsx:"+name) }
		skillsDir = filepath.Join(home, ".pi", "agent", "skills")
	case templates.TargetClaude:
		// Claude Code user-scope command names cannot contain a colon, so
		// the rsx namespace flattens to a hyphen: /rsx-research etc.
		commandsDir := filepath.Join(home, ".claude", "commands")
		promptDst = func(name string) string { return filepath.Join(commandsDir, "rsx-"+name) }
		skillsDir = filepath.Join(home, ".claude", "skills")
	default:
		return fmt.Errorf("no install layout for target %q", target.Name)
	}

	for name, content := range rendered.Prompts {
		dst := promptDst(name)
		if err := writeFile(dst, content); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.OutOrStdout(), "installed prompt: %s\n", dst)
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

// writeFile creates parent dirs and writes content (idempotent overwrite).
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
