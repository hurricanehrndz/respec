package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/templates"
	"github.com/spf13/cobra"
)

func init() {
	var targetName string
	cmd := &cobra.Command{
		Use:   "install --target <pi|prime-agent|claude|codex>",
		Short: "Render and install the respec skills at user scope for an agent",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			target, ok := templates.Targets[targetName]
			if !ok {
				c.SilenceUsage = true
				return fmt.Errorf("unknown --target %q (valid: pi, prime-agent, claude, codex)", targetName)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rendered, err := templates.Render(cfg, target)
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
	cmd.Flags().StringVar(&targetName, "target", "", "agent to install for: pi, prime-agent, claude, or codex (required)")
	_ = cmd.MarkFlagRequired("target")
	rootCmd.AddCommand(cmd)
}

// installFor writes the rendered skills into the target agent's user-scope
// skills directory and removes the prompt files the pre-migration install used
// to write.
func installFor(c *cobra.Command, target templates.Target, home string, rendered templates.Rendered) error {
	var skillsDir, promptDir, promptPrefix string
	switch target {
	case templates.TargetPi:
		skillsDir = filepath.Join(home, ".pi", "agent", "skills")
		promptDir = filepath.Join(home, ".pi", "agent", "prompts")
		promptPrefix = "rsx:"
	case templates.TargetPrimeAgent:
		skillsDir = filepath.Join(home, ".prime", "agent", "skills")
	case templates.TargetClaude:
		skillsDir = filepath.Join(home, ".claude", "skills")
		promptDir = filepath.Join(home, ".claude", "commands")
		promptPrefix = "rsx-"
	case templates.TargetCodex:
		codexHome := os.Getenv("CODEX_HOME")
		if codexHome == "" {
			codexHome = filepath.Join(home, ".codex")
		}
		skillsDir = filepath.Join(codexHome, "skills")
		promptDir = filepath.Join(codexHome, "prompts")
		promptPrefix = "rsx-"
	default:
		return fmt.Errorf("no install layout for target %q", target.Name)
	}
	for rel, content := range rendered.Skills {
		dst := filepath.Join(skillsDir, filepath.FromSlash(rel))
		if err := writeFile(dst, content); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.OutOrStdout(), "installed skill:  %s\n", dst)
	}
	if err := pruneStaleSkills(c, skillsDir, rendered.Skills); err != nil {
		return err
	}
	return pruneRetiredPrompts(c, promptDir, promptPrefix)
}

// pruneRetiredPrompts removes the prompt files respec installed before the
// skills migration. They live in a directory the skills-only install no longer
// writes, so without this they would linger as slash commands pointing at a
// workflow that no longer exists. Only the rsx: / rsx- names respec owns are
// touched, and a directory emptied by the removals is removed.
func pruneRetiredPrompts(c *cobra.Command, dir, prefix string) error {
	if dir == "" || prefix == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	removed := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil {
			return err
		}
		removed = true
		_, _ = fmt.Fprintf(c.OutOrStdout(), "removed retired prompt: %s\n", path)
	}
	if removed {
		_ = os.Remove(dir) // succeeds only when the directory is now empty
	}
	return nil
}

// pruneStaleSkills removes files respec no longer renders from the skill
// directories it owns, for the same reason prompts are pruned: an agent reads
// whatever is in the bundle, so a renamed or retired skill file would keep
// being loaded alongside its replacement. Only the top-level skill dirs named
// by the current render are touched, leaving other skills in the shared
// directory untouched.
func pruneStaleSkills(c *cobra.Command, skillsDir string, current map[string]string) error {
	roots := map[string]bool{}
	for rel := range current {
		roots[strings.SplitN(rel, "/", 2)[0]] = true
	}
	for root := range roots {
		var dirs []string
		err := filepath.WalkDir(filepath.Join(skillsDir, root), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				dirs = append(dirs, p)
				return nil
			}
			rel, err := filepath.Rel(skillsDir, p)
			if err != nil {
				return err
			}
			if _, kept := current[filepath.ToSlash(rel)]; kept {
				return nil
			}
			if err := os.Remove(p); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(c.OutOrStdout(), "removed retired skill file: %s\n", p)
			return nil
		})
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		// Deepest first, so a directory emptied by the removals above goes too;
		// os.Remove refuses non-empty dirs, which is exactly the wanted guard.
		for i := len(dirs) - 1; i >= 0; i-- {
			_ = os.Remove(dirs[i])
		}
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
