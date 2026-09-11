// Package config loads and saves respec configuration from
// ~/.config/respec/config.yaml. All paths support ~ expansion and sane
// defaults are applied when keys are missing.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hurricanehrndz/respec/internal/agents"
	"gopkg.in/yaml.v3"
)

// Rules holds optional per-artifact instruction snippets injected into the
// rendered prompt-templates (R-13).
type Rules struct {
	Research  string `yaml:"research"`
	Spec      string `yaml:"spec"`
	Plan      string `yaml:"plan"`
	Implement string `yaml:"implement"`
}

// Agents holds optional standing defaults consulted during planning, never
// baked into installed skills. The operator chooses workers per effort and
// the plan records the final choices, independent of later config changes.
//
// Role fields are external CLI specs (`<harness>[:<model>][@<effort>]`),
// validated by internal/agents. Notes may suggest native model or budget
// preferences without selecting an external executor.
type Agents struct {
	Implementer string `yaml:"implementer"` // writes the code for a phase
	Reviewer    string `yaml:"reviewer"`    // independently reviews the phase diff
	Committer   string `yaml:"committer"`   // commits an accepted phase
	Notes       string `yaml:"notes"`       // free prose: budget stance, models to avoid
}

// Config is the on-disk respec configuration.
type Config struct {
	Store        string `yaml:"store"`         // central store path (R-1)
	ReflowWidth  int    `yaml:"reflow_width"`  // prose reflow width
	TemplatesDir string `yaml:"templates_dir"` // optional override dir (R-12); empty = embedded
	Context      string `yaml:"context"`       // optional shared context (R-13)
	Rules        Rules  `yaml:"rules"`         // optional per-artifact rules (R-13)
	Agents       Agents `yaml:"agents"`        // optional delegation preferences
}

// Defaults returns a Config populated with default values.
func Defaults() Config {
	return Config{
		Store:       "~/respec-store",
		ReflowWidth: 80,
	}
}

// Path returns the config file path (~/.config/respec/config.yaml).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "respec", "config.yaml"), nil
}

// Expand expands a leading ~ in a path to the user's home directory.
func Expand(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// Load reads the config file, applying defaults for missing keys. A missing
// file yields the defaults (not an error).
func Load() (Config, error) {
	cfg := Defaults()
	p, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", p, err)
	}
	applyDefaults(&cfg)
	return cfg, nil
}

// applyDefaults fills zero-valued required keys with defaults.
func applyDefaults(cfg *Config) {
	d := Defaults()
	if cfg.Store == "" {
		cfg.Store = d.Store
	}
	if cfg.ReflowWidth == 0 {
		cfg.ReflowWidth = d.ReflowWidth
	}
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// StorePath returns the configured store path with ~ expanded.
func (c Config) StorePath() string {
	return Expand(c.Store)
}

// Get returns the value of a dotted config key as a string.
func (c Config) Get(key string) (string, error) {
	switch key {
	case "store":
		return c.Store, nil
	case "reflow_width":
		return strconv.Itoa(c.ReflowWidth), nil
	case "templates_dir":
		return c.TemplatesDir, nil
	case "context":
		return c.Context, nil
	case "rules.research":
		return c.Rules.Research, nil
	case "rules.spec":
		return c.Rules.Spec, nil
	case "rules.plan":
		return c.Rules.Plan, nil
	case "rules.implement":
		return c.Rules.Implement, nil
	case "agents.implementer":
		return c.Agents.Implementer, nil
	case "agents.reviewer":
		return c.Agents.Reviewer, nil
	case "agents.committer":
		return c.Agents.Committer, nil
	case "agents.notes":
		return c.Agents.Notes, nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// Set assigns a dotted config key from a string value.
func (c *Config) Set(key, value string) error {
	switch key {
	case "store":
		c.Store = value
	case "reflow_width":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("reflow_width must be an integer: %w", err)
		}
		c.ReflowWidth = n
	case "templates_dir":
		c.TemplatesDir = value
	case "context":
		c.Context = value
	case "rules.research":
		c.Rules.Research = value
	case "rules.spec":
		c.Rules.Spec = value
	case "rules.plan":
		c.Rules.Plan = value
	case "rules.implement":
		c.Rules.Implement = value
	case "agents.implementer", "agents.reviewer", "agents.committer":
		// Validate on the way in: a bad spec stored here would otherwise fail
		// at delegation time, mid-phase, when it is most expensive to discover.
		if value != "" {
			if _, err := agents.ParseSpec(value); err != nil {
				return err
			}
		}
		switch key {
		case "agents.implementer":
			c.Agents.Implementer = value
		case "agents.reviewer":
			c.Agents.Reviewer = value
		case "agents.committer":
			c.Agents.Committer = value
		}
	case "agents.notes":
		c.Agents.Notes = value
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}
