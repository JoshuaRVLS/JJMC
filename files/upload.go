package files

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func HandleUpload(rootDir, relPath string, file *multipart.FileHeader) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}

	targetDir := filepath.Join(rootDir, cleanPath)
	// Ensure directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	targetPath := filepath.Join(targetDir, file.Filename)

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}
