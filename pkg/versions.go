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
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/schollz/progressbar/v3"
)

var env = Environment{
	Root:    os.Getenv("PHP_HOME"),
	Symlink: os.Getenv("PHP_SYMLINK"),
}

func ListAllVersions() ([]*semver.Version, error) {
	// fetch all versions from https://windows.php.net/downloads/releases/archives/
	// print all versions
	resp, err := http.Get("https://windows.php.net/downloads/releases/archives/")
	if err != nil {
		fmt.Println("Error while fetching versions: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while fetching versions: ", err)
		return nil, err
	}

	versions := strings.Split(string(body), "php-")
	allVersions := make(map[string]bool)
	for _, version := range versions {
		if strings.Contains(version, "-Win32") && !strings.Contains(version, "pack") {
			allVersions[strings.Split(version, "-")[0]] = true
		}
	}

	// convert map[string]bool to []string
	var versionsList []string
	for version := range allVersions {
		versionsList = append(versionsList, version)
	}

	vs := make([]*semver.Version, len(versionsList))
	for i, r := range versionsList {
		v, err := semver.NewVersion(r)
		if err != nil {
			fmt.Println("Error while fetching versions: ", err)
			return nil, err
		}

		vs[i] = v
	}

	sort.Sort(semver.Collection(vs))

	return vs, nil
}

func ChangeVersion(version string) {
	// check if there is directory in C:\php\versions\version
	// if not, start download
	// if yes, run command "cmd /c mklink /D C:\php\current C:\php\versions\version"
	installdir := filepath.Join(env.Root, "versions", version)

	if _, err := os.Stat(installdir); os.IsNotExist(err) {
		fmt.Println("Version", version, "does not exist on local machine, checking if version is available on PHP Servers.")
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
	allVersions, err := ListAllVersions()
	if err != nil {
		panic(err)
	}

	isPresent := false
	for _, v := range allVersions {
		if v.Original() == version {
			isPresent = true
			break
		}
	}

	if isPresent {
		resp, err := http.Get("https://windows.php.net/downloads/releases/archives/")
		if err != nil {
			fmt.Println("Error while fetching versions: ", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while fetching versions: ", err)
		}

		versions := strings.Split(string(body), "php-"+version+"-Win32-")
		matched := false
		for _, v := range versions {
			// test regex for php-{version}-Win32-{something}-x64.zip
			part := strings.Split(v, ".zip")[0]
			if strings.Contains(part, "x64") {
				matched = true
				fmt.Println("Found x64 version: ", part)
				url := "https://windows.php.net/downloads/releases/archives/php-" + version + "-Win32-" + part + ".zip"
				fmt.Println("Downloading from: ", url)
				downloadAndInstall(url, version)
				break
			}
		}

		if !matched {
			fmt.Println("Could not find x64 version for PHP " + version)
		}

	}
}

func GetVersionPath(version string) string {
	return filepath.Join(env.Root, "versions", version)
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

	// edit php.ini and add the following lines
	// extension_dir = "ext"
	file, err := os.OpenFile(dst+"\\php.ini", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err = file.WriteString("extension_dir = \"ext\""); err != nil {
		panic(err)
	}

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
