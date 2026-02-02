package instances

import (
	"fmt"
	"path/filepath"

	"jjmc/backup"
)

// Re-export Backup type for compatibility
type Backup = backup.Backup

func (im *InstanceManager) GetBackupDir(instanceID string) string {
	return filepath.Join("backups", instanceID)
}

func (im *InstanceManager) CreateBackup(instanceID string) error {
	inst, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	backupDir := im.GetBackupDir(instanceID)
	return backup.Create(inst.Directory, backupDir, inst.Name)
}

func (im *InstanceManager) ListBackups(instanceID string) ([]Backup, error) {
	backupDir := im.GetBackupDir(instanceID)
	return backup.List(backupDir)
}

func (im *InstanceManager) RestoreBackup(instanceID, backupName string) error {
	inst, err := im.GetInstance(instanceID)
	if err != nil {
		return err
	}

	if inst.IsRunning() {
		return fmt.Errorf("instance must be offline to restore backup")
	}

	backupPath := filepath.Join(im.GetBackupDir(instanceID), backupName)
	return backup.Restore(backupPath, filepath.Dir(inst.Directory))
}

func (im *InstanceManager) DeleteBackup(instanceID, backupName string) error {
	backupPath := filepath.Join(im.GetBackupDir(instanceID), backupName)
	return backup.Delete(backupPath)
}
