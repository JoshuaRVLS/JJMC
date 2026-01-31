package instances

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

func (im *InstanceManager) GetBackupDir(instanceID string) string {
	return filepath.Join("backups", instanceID)
}

func (im *InstanceManager) CreateBackup(instanceID string) error {
	inst, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	backupDir := im.GetBackupDir(instanceID)
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("%s_%s.zip", inst.Name, timestamp)
	target := filepath.Join(backupDir, filename)

	return archiver.ZipDirectory(inst.Directory, target)
}

func (im *InstanceManager) ListBackups(instanceID string) ([]Backup, error) {
	backupDir := im.GetBackupDir(instanceID)
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

func (im *InstanceManager) RestoreBackup(instanceID, backupName string) error {
	inst, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	backupPath := filepath.Join(im.GetBackupDir(instanceID), backupName)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found")
	}

	if inst.IsRunning() {
		return fmt.Errorf("instance must be offline to restore backup")
	}

	return archiver.Unzip(backupPath, filepath.Dir(inst.Directory))
}

func (im *InstanceManager) DeleteBackup(instanceID, backupName string) error {
	backupPath := filepath.Join(im.GetBackupDir(instanceID), backupName)
	return os.Remove(backupPath)
}
