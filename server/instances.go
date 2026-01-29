package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Instance struct {
	*Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
	Type      string `json:"type"` // e.g. "fabric", "vanilla"
	Version   string `json:"version"`
	Status    string `json:"status"` // "Online" or "Offline"
	MaxMemory int    `json:"maxMemory"`
	JavaArgs  string `json:"javaArgs"`
}

type InstanceManager struct {
	instances map[string]*Instance
	baseDir   string
	mu        sync.RWMutex
}

func NewInstanceManager(baseDir string) *InstanceManager {
	// Ensure base dir exists
	os.MkdirAll(baseDir, 0755)

	// Connect to Database
	ConnectDB()

	im := &InstanceManager{
		instances: make(map[string]*Instance),
		baseDir:   baseDir,
	}

	// Load instances from DB
	var models []InstanceModel
	result := DB.Find(&models)
	if result.Error != nil {
		fmt.Printf("Failed to load instances from DB: %v\n", result.Error)
	}

	for _, model := range models {
		dir := filepath.Join(baseDir, model.ID)
		manager := NewManager()
		instance := &Instance{
			Manager:   manager,
			ID:        model.ID,
			Name:      model.Name,
			Directory: dir,
			Type:      model.Type,
			Version:   model.Version,
			MaxMemory: model.MaxMemory,
			JavaArgs:  model.JavaArgs,
		}
		// Restore state
		instance.Manager.SetWorkDir(dir)
		instance.Manager.SetJar("server.jar")
		instance.Manager.SetMaxMemory(model.MaxMemory)
		instance.Manager.SetJavaArgs(model.JavaArgs)

		im.instances[model.ID] = instance
	}

	return im
}

func (im *InstanceManager) ImportInstance(id, name, sourcePath string) (*Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.instances[id]; exists {
		return nil, fmt.Errorf("instance with id %s already exists", id)
	}

	dir := filepath.Join(im.baseDir, id)
	// Check if source path exists
	info, err := os.Stat(sourcePath)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid source path: %s", sourcePath)
	}

	// Create dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Copy files
	if err := copyDir(sourcePath, dir); err != nil {
		return nil, fmt.Errorf("failed to copy files: %v", err)
	}

	// Save to DB
	// Try to detect type/version? For now, default to "imported" / "custom"
	model := InstanceModel{
		ID:        id,
		Name:      name,
		Type:      "custom",
		Version:   "imported",
		CreatedAt: time.Now().Unix(),
		MaxMemory: 2048,
	}
	if err := DB.Create(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to save to db: %v", err)
	}

	manager := NewManager()
	instance := &Instance{
		Manager:   manager,
		ID:        id,
		Name:      name,
		Directory: dir,
		Type:      "custom",
		Version:   "imported",
		MaxMemory: 2048,
	}
	instance.Manager.SetWorkDir(dir)
	// Try to find server jar?
	jars, _ := filepath.Glob(filepath.Join(dir, "*.jar"))
	if len(jars) > 0 {
		instance.Manager.SetJar(filepath.Base(jars[0]))
	} else {
		instance.Manager.SetJar("server.jar")
	}
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

		// Copy file
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

func (im *InstanceManager) CreateInstance(id, name, serverType, version string) (*Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if _, exists := im.instances[id]; exists {
		return nil, fmt.Errorf("instance with id %s already exists", id)
	}

	dir := filepath.Join(im.baseDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create eula.txt automatically
	os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)

	// Save to DB
	model := InstanceModel{
		ID:        id,
		Name:      name,
		Type:      serverType,
		Version:   version,
		CreatedAt: time.Now().Unix(),
		MaxMemory: 2048,
	}
	if err := DB.Create(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to save to db: %v", err)
	}

	manager := NewManager()
	instance := &Instance{
		Manager:   manager,
		ID:        id,
		Name:      name,
		Directory: dir,
		Type:      serverType,
		Version:   version,
		MaxMemory: 2048,
	}
	instance.Manager.SetWorkDir(dir)
	instance.Manager.SetJar("server.jar") // Default
	instance.Manager.SetMaxMemory(2048)

	im.instances[id] = instance
	return instance, nil
}

func (im *InstanceManager) UpdateSettings(id string, maxMemory int, javaArgs string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	inst, ok := im.instances[id]
	if !ok {
		return fmt.Errorf("instance not found")
	}

	// Update DB
	err := DB.Model(&InstanceModel{}).Where("id = ?", id).Updates(InstanceModel{
		MaxMemory: maxMemory,
		JavaArgs:  javaArgs,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}

	// Update Memory
	inst.MaxMemory = maxMemory
	inst.JavaArgs = javaArgs
	inst.Manager.SetMaxMemory(maxMemory)
	inst.Manager.SetJavaArgs(javaArgs)

	return nil
}

func (inst *Instance) Reset(serverType, version string) error {
	inst.mu.Lock()
	defer inst.mu.Unlock()

	// Stop if running
	if inst.IsRunning() {
		if err := inst.Stop(); err != nil {
			return fmt.Errorf("failed to stop server: %v", err)
		}
		// Allow some time for cleanup if needed, though Stop() should be synchronous
		time.Sleep(1 * time.Second)
	}

	// List of files/directories to remove for a clean type switch
	// We keep world (saves), server.properties, eula.txt, ops, whitelist, etc.
	toRemove := []string{
		inst.jarName,
		"libraries",
		"versions",
		"mods",
		"config",
		"plugins", // in case switching from/to Spigot
		// Add others if needed
	}

	for _, name := range toRemove {
		path := filepath.Join(inst.Directory, name)
		os.RemoveAll(path)
	}

	// Update Fields
	inst.Type = serverType
	inst.Version = version

	// Update DB
	err := DB.Model(&InstanceModel{}).Where("id = ?", inst.ID).Updates(map[string]interface{}{
		"type":    serverType,
		"version": version,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update db: %v", err)
	}

	return nil
}

func (im *InstanceManager) GetInstance(id string) (*Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	inst, ok := im.instances[id]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}

	// Update status before returning
	if inst.IsRunning() {
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
		if inst.IsRunning() {
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

func (im *InstanceManager) DeleteInstance(id string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	inst, ok := im.instances[id]
	if !ok {
		return fmt.Errorf("instance not found")
	}

	if inst.IsRunning() {
		return fmt.Errorf("cannot delete running instance")
	}

	// Delete from DB
	if err := DB.Delete(&InstanceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete from db: %v", err)
	}

	// Remove files
	os.RemoveAll(inst.Directory)
	delete(im.instances, id)
	return nil
}
