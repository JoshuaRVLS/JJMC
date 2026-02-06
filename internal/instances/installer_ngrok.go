package instances

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func GetToolsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cwd, ".tools")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetNgrokPath() string {

	if path, err := exec.LookPath("ngrok"); err == nil {
		return path
	}

	toolsDir, err := GetToolsDir()
	if err == nil {
		name := "ngrok"
		if runtime.GOOS == "windows" {
			name = "ngrok.exe"
		}
		localPath := filepath.Join(toolsDir, name)
		if _, err := os.Stat(localPath); err == nil {
			return localPath
		}
	}
	return ""
}

func InstallNgrok(logFunc func(string)) error {
	toolsDir, err := GetToolsDir()
	if err != nil {
		return err
	}

	var url string
	var ext string

	baseUrl := "https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable"

	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" {
			url = fmt.Sprintf("%s-linux-amd64.tgz", baseUrl)
			ext = ".tgz"
		} else if runtime.GOARCH == "arm64" {
			url = fmt.Sprintf("%s-linux-arm64.tgz", baseUrl)
			ext = ".tgz"
		} else {
			return fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			url = fmt.Sprintf("%s-darwin-amd64.zip", baseUrl)
			ext = ".zip"
		} else if runtime.GOARCH == "arm64" {
			url = fmt.Sprintf("%s-darwin-arm64.zip", baseUrl)
			ext = ".zip"
		} else {
			return fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			url = fmt.Sprintf("%s-windows-amd64.zip", baseUrl)
			ext = ".zip"
		} else {
			return fmt.Errorf("unsupported arch: %s", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	if logFunc != nil {
		logFunc(fmt.Sprintf("Downloading ngrok from %s...", url))
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "ngrok-download-*"+ext)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	tmpFile.Seek(0, 0)

	targetName := "ngrok"
	if runtime.GOOS == "windows" {
		targetName = "ngrok.exe"
	}
	targetPath := filepath.Join(toolsDir, targetName)

	if ext == ".tgz" {
		return untar(tmpFile, targetPath)
	} else if ext == ".zip" {

		stat, _ := tmpFile.Stat()
		return unzip(tmpFile, stat.Size(), targetPath)
	} else {
		return fmt.Errorf("unknown extraction format")
	}
}

func untar(r io.Reader, targetPath string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg && header.Name == "ngrok" {

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
			return os.Chmod(targetPath, 0755)
		}
	}
	return fmt.Errorf("ngrok binary not found in archive")
}

func unzip(r io.ReaderAt, size int64, targetPath string) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if f.Name == "ngrok" || f.Name == "ngrok.exe" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return err
			}
			return os.Chmod(targetPath, 0755)
		}
	}
	return fmt.Errorf("ngrok binary not found in archive")
}
