package server

import (
	"fmt"
	"os"
	"path/filepath"
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
