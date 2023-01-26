package pkg

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

var env = Environment{
	Root:    os.Getenv("PHP_HOME"),
	Symlink: os.Getenv("PHP_SYMLINK"),
}

func ChangeVersion(version string) {
	// check if there is directory in C:\php\versions\version
	// if not, start download
	// if yes, run command "cmd /c mklink /D C:\php\current C:\php\versions\version"
	installdir := filepath.Join(env.Root, "versions", version)

	if _, err := os.Stat(installdir); os.IsNotExist(err) {
		fmt.Println("Version", version, "does not exist")
		fmt.Println("Downloading " + version + "...")
		DownloadVersion(version)
	} else {
		if _, err := os.Stat(env.Symlink); !os.IsNotExist(err) {
			elevatedRun(env.Root, "cmd", "/C", "rmdir", env.Symlink)
		}

		ok, err := elevatedRun(env.Root, "cmd", "/C", "mklink", "/D", env.Symlink, installdir)
		if err != nil {
			fmt.Println("Error while changing version: ", err)
		}

		if !ok {
			fmt.Println("Error while changing version: ", errors.New("could not run the script as admin"))
		}

		fmt.Println("PHP " + version + " installed successfully.")
	}
}

func DownloadVersion(version string) {
	// check if system is 32 or 64 bit
	// download the correct version
	// extract the zip file
	// move all content from where to C:\php\versions\version
	// open php.ini from C:\php\versions\version and add the following lines
	// extension_dir = "ext"

	switch version {
	case "8.2":
		downloadAndInstall("https://windows.php.net/downloads/releases/php-8.2.1-Win32-vs16-x64.zip", version)
	case "8.1":
		downloadAndInstall("https://windows.php.net/downloads/releases/php-8.1.14-Win32-vs16-x64.zip", version)
	case "8.0":
		downloadAndInstall("https://windows.php.net/downloads/releases/php-8.0.27-Win32-vs16-x64.zip", version)
	case "7.4":
		downloadAndInstall("https://windows.php.net/downloads/releases/php-7.4.33-Win32-vc15-x64.zip", version)
	default:
		fmt.Println("Version", version, "does not exist")
	}
}

func downloadAndInstall(url string, version string) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	downloadPath := filepath.Join(env.Root, "php-"+version+".zip")

	f, _ := os.OpenFile(downloadPath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	dst := filepath.Join(env.Root, "versions", version)
	if file, err := os.Stat(dst); os.IsNotExist(err) || !file.IsDir() {
		os.Mkdir(dst, 0755)
	}

	archive, err := zip.OpenReader(downloadPath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	os.Remove(downloadPath)

	// copy php.ini from C:\php\versions\version\php.ini-development to C:\php\versions\version\php.ini
	os.Rename(dst+"\\php.ini-development", dst+"\\php.ini")

	// change version to the one that was just downloaded
	ChangeVersion(version)
}

func elevatedRun(root string, name string, arg ...string) (bool, error) {
	ok, err := run("cmd", append([]string{"/C", name}, arg...)...)
	if err != nil {
		ok, err = run(filepath.Join(root, "elevate.cmd"), append([]string{"cmd", "/C", name}, arg...)...)
	}

	return ok, err
}

func run(name string, arg ...string) (bool, error) {
	c := exec.Command(name, arg...)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		return false, errors.New(fmt.Sprint(err) + ": " + stderr.String())
	}

	return true, nil
}
