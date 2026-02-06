package instances

import (
	"jjmc/internal/database"
	"jjmc/internal/models"
)

func (inst *Instance) Save() error {
	return database.DB.Model(&models.InstanceModel{}).Where("id = ?", inst.ID).Updates(models.InstanceModel{
		Type:      inst.Type,
		Version:   inst.Version,
		MaxMemory: inst.MaxMemory,
		JavaArgs:  inst.JavaArgs,
		JarFile:   inst.JarFile,
		JavaPath:  inst.JavaPath,
	}).Error
}
