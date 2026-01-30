package instances

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"jjmc/database"
	"jjmc/manager"
	"jjmc/models"
	"jjmc/services"
)

type InstanceManager struct {
	instances   map[string]*Instance
	baseDir     string
	mu          sync.RWMutex
	TemplateMgr *services.TemplateManager
}

func NewInstanceManager(baseDir string, tm *services.TemplateManager) *InstanceManager {
	// Ensure base dir exists
	os.MkdirAll(baseDir, 0755)

	im := &InstanceManager{
		instances:   make(map[string]*Instance),
		baseDir:     baseDir,
		TemplateMgr: tm,
	}

	// Load instances from DB
	var dbModels []models.InstanceModel
	result := database.DB.Find(&dbModels)
	if result.Error != nil {
		fmt.Printf("Failed to load instances from DB: %v\n", result.Error)
	}

	for _, model := range dbModels {
		dir := filepath.Join(baseDir, model.ID)
		mgr := manager.NewManager()

		// Copy model data to instance struct
		instModel := model // Copy
		instance := &Instance{
			Instance: &models.Instance{
				ID:        instModel.ID,
				Name:      instModel.Name,
				Directory: dir,
				Type:      instModel.Type,
				Version:   instModel.Version,
				MaxMemory: instModel.MaxMemory,
				JavaArgs:  instModel.JavaArgs,
				JarFile:   instModel.JarFile,
			},
			Manager: mgr,
		}

		// Restore state
		instance.Manager.SetWorkDir(dir)
		if model.JarFile != "" {
			instance.Manager.SetJar(model.JarFile)
		} else {
			instance.Manager.SetJar("server.jar")
		}
		instance.Manager.SetMaxMemory(model.MaxMemory)
		instance.Manager.SetJavaArgs(model.JavaArgs)

		im.instances[model.ID] = instance
	}

	return im
}

func (im *InstanceManager) GetInstance(id string) (*Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	inst, ok := im.instances[id]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}

	// Update status before returning
	if inst.Manager.IsRunning() {
		inst.Status = "Online"
	} else {
		inst.Status = "Offline"
	}

	return inst, nil
}

func (im *InstanceManager) ListInstances() []*Instance {
	im.mu.RLock()
	defer im.mu.RUnlock()

	list := make([]*Instance, 0, len(im.instances))
	for _, inst := range im.instances {
		// Update status
		if inst.Manager.IsRunning() {
			inst.Status = "Online"
		} else {
			inst.Status = "Offline"
		}
		list = append(list, inst)
	}

	// Sort by Name, then ID for stability
	sort.Slice(list, func(i, j int) bool {
		if list[i].Name != list[j].Name {
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		}
		return list[i].ID < list[j].ID
	})

	return list
}

func (im *InstanceManager) UpdateSettings(id string, maxMemory int, javaArgs, jarFile string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	inst, ok := im.instances[id]
	if !ok {
		return fmt.Errorf("instance not found")
	}

	// Update DB
	err := database.DB.Model(&models.InstanceModel{}).Where("id = ?", id).Updates(models.InstanceModel{
		MaxMemory: maxMemory,
		JavaArgs:  javaArgs,
		JarFile:   jarFile,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}

	// Update Memory
	inst.MaxMemory = maxMemory
	inst.JavaArgs = javaArgs
	inst.JarFile = jarFile
	inst.Manager.SetMaxMemory(maxMemory)
	inst.Manager.SetJavaArgs(javaArgs)
	if jarFile != "" {
		inst.Manager.SetJar(jarFile)
	}

	return nil
}

func (inst *Instance) Reset(serverType, version string) error {
	// Reset logic
	// Lock Manager's mutex is internal. We should probably add a helper in manager if we need to lock externally?
	// inst.Manager.mu is not available if unexported.
	// Wait, Manager.mu is exported in my previous step? "mu sync.Mutex" -> unexported.
	// Good thing I saw this.
	// The original code accessed `inst.Manager.mu`.
	// I should export a method in Manager to stop safely or expose the mutex.
	// Or better, just call Stop() and handle errors.

	if inst.Manager.IsRunning() {
		if err := inst.Manager.Stop(); err != nil {
			return fmt.Errorf("failed to stop server: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	// List of files/directories to remove for a clean type switch
	toRemove := []string{
		"server.jar", // Assumed default jarName, might need to get from Instance config
		"libraries",
		"versions",
		"mods",
		"config",
		"plugins",
	}

	// Also remove the current jar if it's different
	if inst.JarFile != "" && inst.JarFile != "server.jar" {
		toRemove = append(toRemove, inst.JarFile)
	}

	for _, name := range toRemove {
		path := filepath.Join(inst.Directory, name)
		os.RemoveAll(path)
	}

	// Update Fields
	inst.Type = serverType
	inst.Version = version

	// Update DB
	err := database.DB.Model(&models.InstanceModel{}).Where("id = ?", inst.ID).Updates(map[string]interface{}{
		"type":    serverType,
		"version": version,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}

	return nil
}
