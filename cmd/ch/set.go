package ch

import (
	"fmt"

	"github.com/BrainBuzzer/php-ch/pkg"
	"github.com/spf13/cobra"
)

var setCommand = &cobra.Command{
	Use:   "set",
	Short: "Set PHP Version",
	Long:  `This command checks if there is existing PHP installation for given version, if not, downloads the specified PHP version.`,
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		fmt.Println("Changing version to", version)
		pkg.ChangeVersion(version)
	},
}

func init() {
	rootCmd.AddCommand(setCommand)
}
