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

// GetToolsDir returns the absolute path to the .tools directory
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

// GetNgrokPath returns the path to the ngrok binary, or empty string if not found
func GetNgrokPath() string {
	// 1. Check PATH
	if path, err := exec.LookPath("ngrok"); err == nil {
		return path
	}
	// Actually exec.LookPath is better called by the caller or we duplicate logic.
	// But let's check our local tools dir first or second.

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

	// https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v3-stable-linux-amd64.tgz
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

	// Download
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

	// Create temp file
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

	// Rewind
	tmpFile.Seek(0, 0)

	// Extract
	targetName := "ngrok"
	if runtime.GOOS == "windows" {
		targetName = "ngrok.exe"
	}
	targetPath := filepath.Join(toolsDir, targetName)

	if ext == ".tgz" {
		return untar(tmpFile, targetPath)
	} else if ext == ".zip" {
		// For zip we need a ReaderAt, os.File implements it
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
			// Found it
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
