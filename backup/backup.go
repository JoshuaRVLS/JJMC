package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"jjmc/pkg/archiver"
)

type Backup struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

func GetBackupDir(instanceDir string) string {
	return filepath.Join("backups", filepath.Base(instanceDir))
}

// Create creates a zip backup of the instance directory
func Create(instanceDir string, backupDir string, instanceName string) error {
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.zip", instanceName, timestamp)
	target := filepath.Join(backupDir, filename)

	return archiver.ZipDirectory(instanceDir, target)
}

func List(backupDir string) ([]Backup, error) {
	entries, err := os.ReadDir(backupDir)
	if os.IsNotExist(err) {
		return []Backup{}, nil
	}
	if err != nil {
		return nil, err
	}

	backups := []Backup{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".zip" {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, Backup{
				Name:      entry.Name(),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			})
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

func Restore(backupPath string, restoreDir string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found")
	}

	return archiver.Unzip(backupPath, restoreDir)
}

func Delete(backupPath string) error {
	return os.Remove(backupPath)
}
