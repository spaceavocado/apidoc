// Package cmd is a collection of CLI commands
// for running of this tool.
package cmd

import (
	"github.com/spaceavocado/apidoc/app"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// RootCmd is the main command running the apidoc tool
func RootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Short: "apidoc",
		Long:  "API Documentation Generator",
		Run: func(cmd *cobra.Command, args []string) {
			mainFile, err := cmd.PersistentFlags().GetString("main")
			endsRoot, err := cmd.PersistentFlags().GetString("endpoints")
			output, err := cmd.PersistentFlags().GetString("output")
			verbose, err := cmd.PersistentFlags().GetBool("verbose")
			if err != nil {
				log.Errorf("Invalid CLI flags, please use the -h flag to see all available options: %+v", err)
				return
			}

			app := app.New(app.Configuration{
				MainFile: mainFile,
				EndsRoot: endsRoot,
				Output:   output,
				Verbose:  verbose,
			})
			app.Start()
		},
	}

	// Flags
	rootCmd.PersistentFlags().StringP("main", "m", "main.go", "Main API documentation file")
	rootCmd.PersistentFlags().StringP("endpoints", "e", "./", "Root endpoints folder")
	rootCmd.PersistentFlags().StringP("output", "o", "docs/api", "Documentation output folder")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Show generation warnings")

	// Other commands
	rootCmd.AddCommand(versionCmd)

	return rootCmd
}
