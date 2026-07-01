package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hurricanehrndz/respec/internal/config"
	"github.com/hurricanehrndz/respec/internal/site"
	"github.com/spf13/cobra"
)

// siteBuildDir returns (and materializes the scaffold into) the Hugo build
// directory under the user cache, wired to read content from the store.
func siteBuildDir() (buildDir, store string, err error) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", err
	}
	store = cfg.StorePath()
	if _, err := os.Stat(store); err != nil {
		return "", "", fmt.Errorf("store %q not accessible: %w", store, err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", "", err
	}
	buildDir = filepath.Join(cache, "respec", "site")
	if err := site.Materialize(buildDir); err != nil {
		return "", "", fmt.Errorf("materializing site scaffold: %w", err)
	}
	return buildDir, store, nil
}

// runHugo execs hugo with args, failing loud when hugo is not on PATH.
func runHugo(c *cobra.Command, args ...string) error {
	bin, err := exec.LookPath("hugo")
	if err != nil {
		return fmt.Errorf("hugo not found on PATH; respec render/serve require Hugo to be installed: %w", err)
	}
	h := exec.Command(bin, args...)
	h.Stdout = c.OutOrStdout()
	h.Stderr = c.ErrOrStderr()
	return h.Run()
}

func init() {
	var out string
	cmd := &cobra.Command{
		Use:   "render",
		Short: "Build the central store as a browsable Hugo site",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			buildDir, store, err := siteBuildDir()
			if err != nil {
				return err
			}
			if out == "" {
				out = filepath.Join(buildDir, "public")
			}
			return runHugo(c,
				"--source", buildDir,
				"--contentDir", store,
				"--destination", out,
				"--cleanDestinationDir",
			)
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output directory (default <cache>/respec/site/public)")
	rootCmd.AddCommand(cmd)
}
