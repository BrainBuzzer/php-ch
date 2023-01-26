package ch

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/BrainBuzzer/php-ch/pkg"
	"github.com/spf13/cobra"
)

var env = pkg.Environment{
	Root:    "",
	Symlink: "",
}

var setupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Setup PHP-CH",
	Long:  `This command checks if there is existing PHP installation and if not, it installs the latest PHP version.`,
	Run: func(cmd *cobra.Command, args []string) {
		env.Root = os.Getenv("PHP_HOME")
		env.Symlink = os.Getenv("PHP_SYMLINK")

		// execute the php.exe -v command and extract the version
		_, err := exec.Command("php", "-v").Output()
		if err != nil {
			fmt.Println("No PHP installation found. Installing the latest PHP.")
			pkg.ChangeVersion("8.2")
		}

		// check if the exec path of php is the same as the PHP_HOME env variable
		// if not, then remove the exec path of php from the PATH env variable
		where, err := exec.LookPath("php")
		if err != nil {
			fmt.Println("PHP installation not found.")
			os.Exit(1)
		}

		if where != env.Root+"\\php.exe" {
			fmt.Println("PHP installation found in a different location. Removing the PHP installation from the PATH.")
			pkg.RemoveFromPath(where)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCommand)
}

// func moveFilesRecursively(src string, dest string, skip string) {
// 	files, err := os.ReadDir(src)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	for _, f := range files {
// 		if f.IsDir() && f.Name() != skip {
// 			os.MkdirAll(dest+"\\"+f.Name(), os.ModePerm)
// 			moveFilesRecursively(src+"\\"+f.Name(), dest+"\\"+f.Name(), skip)
// 			os.Remove(src + "\\" + f.Name())
// 		}
// 		os.Rename(src+"\\"+f.Name(), dest+"\\"+f.Name())
// 	}
// }
