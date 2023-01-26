package pkg

import (
	"os"
	"strings"
)

type Environment struct {
	Root    string
	Symlink string
}

func RemoveFromPath(where string) {
	// check in path env to see if where is in the path
	// if yes, then remove it
	path := os.Getenv("PATH")
	path = strings.ReplaceAll(path, where+";", "")
	os.Setenv("PATH", path)
}
