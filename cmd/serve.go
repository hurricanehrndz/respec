package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	var bind string
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the central store as a Hugo site with live reload",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			buildDir, store, err := siteBuildDir()
			if err != nil {
				return err
			}
			return runHugo(c, "server",
				"--source", buildDir,
				"--contentDir", store,
				"--bind", bind,
				"--port", strconv.Itoa(port),
			)
		},
	}
	cmd.Flags().StringVar(&bind, "bind", "127.0.0.1", "interface to bind the server to")
	cmd.Flags().IntVar(&port, "port", 1313, "port to serve on")
	rootCmd.AddCommand(cmd)
}
