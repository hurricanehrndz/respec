package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chernand/respec/internal/config"
	"github.com/chernand/respec/internal/templates"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Render and install the /rsx:* prompt-templates and skill at user scope",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			rendered, err := templates.Render(cfg)
			if err != nil {
				return err
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			promptsDir := filepath.Join(home, ".pi", "agent", "prompts")
			skillsDir := filepath.Join(home, ".pi", "agent", "skills")

			// respec installs all prompts under the rsx: namespace; embedded
			// asset files cannot carry a colon (go:embed forbids it), so the
			// prefix is applied here.
			for name, content := range rendered.Prompts {
				dst := filepath.Join(promptsDir, "rsx:"+name)
				if err := writeFile(dst, content); err != nil {
					return err
				}
				fmt.Printf("installed prompt: %s\n", dst)
			}
			for rel, content := range rendered.Skills {
				dst := filepath.Join(skillsDir, filepath.FromSlash(rel))
				if err := writeFile(dst, content); err != nil {
					return err
				}
				fmt.Printf("installed skill:  %s\n", dst)
			}
			return nil
		},
	})
}

// writeFile creates parent dirs and writes content (idempotent overwrite).
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
