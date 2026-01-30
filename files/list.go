package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// List lists files in a given relative path inside the root directory
func List(rootDir, relPath string) ([]FileInfo, error) {
	// sanitize path
	cleanPath := filepath.Clean(relPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path")
	}

	// If root ./ or just empty
	targetDir := filepath.Join(rootDir, cleanPath)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	files := make([]FileInfo, 0)
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
