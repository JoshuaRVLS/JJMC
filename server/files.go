package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"`
}

// ListFiles lists files in a given relative path inside the instance directory
func (i *Instance) ListFiles(relPath string) ([]FileInfo, error) {
	// sanitize path
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path")
	}

	// If root ./ or just empty
	targetDir := filepath.Join(i.Directory, cleanPath)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().UnixMilli(),
		})
	}

	// Sort: Directories first, then alphabetical
	sort.Slice(files, func(a, b int) bool {
		if files[a].IsDir && !files[b].IsDir {
			return true
		}
		if !files[a].IsDir && files[b].IsDir {
			return false
		}
		return files[a].Name < files[b].Name
	})

	return files, nil
}

func (i *Instance) ReadFile(relPath string) ([]byte, error) {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(i.Directory, cleanPath)
	return os.ReadFile(targetPath)
}

func (i *Instance) WriteFile(relPath string, data []byte) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(i.Directory, cleanPath)
	return os.WriteFile(targetPath, data, 0644)
}

func (i *Instance) DeleteFile(relPath string) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(i.Directory, cleanPath)
	return os.RemoveAll(targetPath)
}

func (i *Instance) Mkdir(relPath string) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}
	targetPath := filepath.Join(i.Directory, cleanPath)
	return os.MkdirAll(targetPath, 0755)
}

func (i *Instance) HandleUpload(relPath string, file *multipart.FileHeader) error {
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path")
	}

	targetDir := filepath.Join(i.Directory, cleanPath)
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
