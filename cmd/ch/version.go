package ch

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:     "set",
	Short:   "Set PHP Version",
	Aliases: []string{"v"},
	Long:    `This command checks if there is existing PHP installation for given version, if not, downloads the specified PHP version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.3")
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
}
