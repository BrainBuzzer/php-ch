package ch

import (
	"fmt"
	"os/exec"

	"github.com/BrainBuzzer/php-ch/pkg"
	"github.com/spf13/cobra"
)

var iniCommand = &cobra.Command{
	Use:   "ini",
	Short: "Edit php.ini file",
	Long:  `Edit php.ini file`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		} else if len(args) == 1 {
			switch args[0] {
			case "7.4", "8.0", "8.1", "8.2":
				return nil
			default:
				return fmt.Errorf("invalid version: %s", args[0])
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		exepath := "C:\\Windows\\system32\\notepad.exe"

		defaultIniPath := "C:\\Program Files\\php\\php.ini"
		if len(args) == 1 {
			// open php.ini of specific version
			iniPath := pkg.GetVersionPath(args[0]) + "\\php.ini"
			err := exec.Command(exepath, iniPath).Run()
			if err != nil {
				fmt.Println(err)
			}
		}

		// open default php.ini
		err := exec.Command(exepath, defaultIniPath).Run()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(iniCommand)
}
