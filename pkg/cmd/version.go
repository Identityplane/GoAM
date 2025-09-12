package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoAM",
	Long:  `All software has versions. This is GoAM's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoAM v0.1.0 -- HEAD")
	},
}
