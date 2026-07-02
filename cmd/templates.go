package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/templates"
	"github.com/spf13/cobra"
)

func init() {
	parent := &cobra.Command{
		Use:   "templates",
		Short: "Inspect and customize the /rsx:* prompt and skill templates",
	}
	parent.AddCommand(templatesListCmd(), templatesEjectCmd())
	rootCmd.AddCommand(parent)
}

type templateInfo struct {
	Name   string `json:"name"`
	Key    string `json:"key"`
	Skill  bool   `json:"skill"`
	Source string `json:"source"`
}

func templatesListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates and whether each is embedded or overridden",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			files, err := templates.Resolve(cfg)
			if err != nil {
				return err
			}
			if asJSON {
				infos := make([]templateInfo, 0, len(files))
				for _, f := range files {
					infos = append(infos, templateInfo{Name: f.Name, Key: f.Key, Skill: f.IsSkill, Source: f.Source})
				}
				return json.NewEncoder(c.OutOrStdout()).Encode(infos)
			}
			var b strings.Builder
			for _, f := range files {
				source := f.Source
				if source != templates.EmbeddedSource {
					source = "override: " + source
				}
				fmt.Fprintf(&b, "%-28s [%s]\n", f.Name, source)
			}
			_, err = fmt.Fprint(c.OutOrStdout(), b.String())
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	return cmd
}

func templatesEjectCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "eject [name]",
		Short: "Copy embedded default template(s) into templates_dir for editing",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.TemplatesDir == "" {
				c.SilenceUsage = true
				return fmt.Errorf("templates_dir is not set (run: respec config set templates_dir <dir>)")
			}
			dst := config.Expand(cfg.TemplatesDir)

			embeddedFiles, err := templates.Embedded()
			if err != nil {
				return err
			}
			var selected []templates.TemplateFile
			if len(args) == 1 {
				for _, f := range embeddedFiles {
					if matchesTemplate(f, args[0]) {
						selected = append(selected, f)
					}
				}
				if len(selected) == 0 {
					c.SilenceUsage = true
					return fmt.Errorf("no embedded template matches %q", args[0])
				}
			} else {
				selected = embeddedFiles
			}

			var b strings.Builder
			for _, f := range selected {
				out := filepath.Join(dst, filepath.FromSlash(f.Name))
				if _, err := os.Stat(out); err == nil && !force {
					c.SilenceUsage = true
					return fmt.Errorf("%s already exists; re-run with --force to overwrite", out)
				}
			}
			for _, f := range selected {
				raw, err := f.ReadRaw()
				if err != nil {
					return err
				}
				out := filepath.Join(dst, filepath.FromSlash(f.Name))
				if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(out, raw, 0o644); err != nil {
					return err
				}
				fmt.Fprintf(&b, "ejected %s\n", out)
			}
			_, err = fmt.Fprint(c.OutOrStdout(), b.String())
			return err
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files in templates_dir")
	return cmd
}

// matchesTemplate reports whether arg names a template file, accepting the
// logical name, the install key, or a prompt base name without .md.
func matchesTemplate(f templates.TemplateFile, arg string) bool {
	return arg == f.Name ||
		arg == f.Key ||
		arg == strings.TrimSuffix(f.Key, ".md")
}
