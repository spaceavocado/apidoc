package cmd

import (
	"fmt"

	"github.com/spaceavocado/apidoc/app"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the APIDoc version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(app.Version)
	},
}
