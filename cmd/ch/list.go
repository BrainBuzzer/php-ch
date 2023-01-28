package ch

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BrainBuzzer/php-ch/pkg"
	"github.com/spf13/cobra"
)

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List all available PHP versions",
	Long:  `List all available PHP versions`,
	Run: func(cmd *cobra.Command, args []string) {
		root := os.Getenv("PHP_HOME")
		vs, err := pkg.ListAllVersions()
		if err != nil {
			fmt.Println("Error while fetching versions: ", err)
		}

		for _, v := range vs {
			installdir := filepath.Join(root, "versions", v.String())
			if _, err := os.Stat(installdir); os.IsNotExist(err) {
				fmt.Println(v.String())
			} else {
				fmt.Println(v.String() + "* Installed")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCommand)
}
