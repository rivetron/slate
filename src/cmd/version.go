package cmd

import (
	"github.com/rivetron/slate/src/utils"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Display version and build information for Slate",
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintBanner()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
