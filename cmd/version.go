package cmd

import (
	"fmt"
	"sshpky/pkg/common"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of sshpky",
	Long:  `All software has versions. This is sshpky's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(common.GetVersionInfo().String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
