package instances

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"jjmc/internal/database"
	"jjmc/internal/manager"
	"jjmc/internal/models"
)

func (im *InstanceManager) CreateInstance(id, name, serverType, version string) (*Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.instances[id]; exists {
		return nil, fmt.Errorf("instance with id %s already exists", id)
	}

	if im.TemplateMgr != nil {
		if tmpl, ok := im.TemplateMgr.GetTemplate(serverType); ok {

			model := models.InstanceModel{
				ID:           id,
				Name:         name,
				Type:         serverType,
				Version:      version,
				CreatedAt:    time.Now().Unix(),
				MaxMemory:    2048,
				StartCommand: tmpl.Run.Command,
			}
			if err := database.DB.Create(&model).Error; err != nil {
				return nil, fmt.Errorf("failed to save to db: %v", err)
			}

			dir := filepath.Join(im.baseDir, id)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)

			mgr := manager.NewManager()
			mgr.SetSilent(im.silent)
			instance := &Instance{
				Instance: &models.Instance{
					ID:           id,
					Name:         name,
					Directory:    dir,
					Type:         serverType,
					Version:      version,
					MaxMemory:    2048,
					JarFile:      "server.jar",
					StartCommand: tmpl.Run.Command,
				},
				Manager: mgr,
				Tunnel:  NewTunnelManager(dir),
			}
			instance.Manager.SetWorkDir(dir)
			instance.Manager.SetJar("server.jar")
			if tmpl.Run.Command != "" {
				instance.Manager.SetStartCommand(tmpl.Run.Command)
			}
			instance.Manager.SetMaxMemory(2048)

			im.instances[id] = instance

			go func() {
				if err := instance.InstallFromTemplate(tmpl, version); err != nil {
					fmt.Printf("Failed to install template for %s: %v\n", id, err)
					instance.Manager.Broadcast(fmt.Sprintf("Failed to install template: %v", err))
				}
			}()

			return instance, nil
		}
	}

	dir := filepath.Join(im.baseDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)

	model := models.InstanceModel{
		ID:        id,
		Name:      name,
		Type:      serverType,
		Version:   version,
		CreatedAt: time.Now().Unix(),
		MaxMemory: 2048,
	}
	if err := database.DB.Create(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to save to db: %v", err)
	}

	mgr := manager.NewManager()
	mgr.SetSilent(im.silent)
	instance := &Instance{
		Instance: &models.Instance{
			ID:        id,
			Name:      name,
			Directory: dir,
			Type:      serverType,
			Version:   version,
			MaxMemory: 2048,
			JarFile:   "server.jar",
		},
		Manager: mgr,
		Tunnel:  NewTunnelManager(dir),
	}

	instance.Manager.SetWorkDir(dir)
	instance.Manager.SetJar("server.jar")
	instance.Manager.SetMaxMemory(2048)

	im.instances[id] = instance
	return instance, nil
}

func (im *InstanceManager) DeleteInstance(id string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	inst, ok := im.instances[id]
	if !ok {
		return fmt.Errorf("instance not found")
	}

	if inst.Manager.IsRunning() {
		return fmt.Errorf("cannot delete running instance")
	}

	if err := database.DB.Delete(&models.InstanceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete from db: %v", err)
	}

	os.RemoveAll(inst.Directory)
	delete(im.instances, id)
	return nil
}

func (im *InstanceManager) ImportInstance(id, name, sourcePath string) (*Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.instances[id]; exists {
		return nil, fmt.Errorf("instance with id %s already exists", id)
	}

	dir := filepath.Join(im.baseDir, id)

	info, err := os.Stat(sourcePath)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid source path: %s", sourcePath)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	if err := copyDir(sourcePath, dir); err != nil {
		return nil, fmt.Errorf("failed to copy files: %v", err)
	}

	model := models.InstanceModel{
		ID:        id,
		Name:      name,
		Type:      "custom",
		Version:   "imported",
		CreatedAt: time.Now().Unix(),
		MaxMemory: 2048,
	}
	if err := database.DB.Create(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to save to db: %v", err)
	}

	mgr := manager.NewManager()
	mgr.SetSilent(im.silent)
	instance := &Instance{
		Instance: &models.Instance{
			ID:        id,
			Name:      name,
			Directory: dir,
			Type:      "custom",
			Version:   "imported",
			MaxMemory: 2048,
		},
		Manager: mgr,
		Tunnel:  NewTunnelManager(dir),
	}

	instance.Manager.SetWorkDir(dir)

	jars, _ := filepath.Glob(filepath.Join(dir, "*.jar"))
	jarName := "server.jar"
	if len(jars) > 0 {
		jarName = filepath.Base(jars[0])
	}
	instance.Manager.SetJar(jarName)
	instance.JarFile = jarName

	database.DB.Model(&models.InstanceModel{}).Where("id = ?", id).Update("jar_file", jarName)

	instance.Manager.SetMaxMemory(2048)

	im.instances[id] = instance
	return instance, nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
