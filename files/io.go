package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Read(rootDir, relPath string) ([]byte, error) {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(rootDir, cleanPath)
	return os.ReadFile(targetPath)
}

func Write(rootDir, relPath string, data []byte) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(rootDir, cleanPath)
	return os.WriteFile(targetPath, data, 0644)
}

func Delete(rootDir, relPath string) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(rootDir, cleanPath)
	return os.RemoveAll(targetPath)
}

func Mkdir(rootDir, relPath string) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(rootDir, cleanPath)
	return os.MkdirAll(targetPath, 0755)
}
