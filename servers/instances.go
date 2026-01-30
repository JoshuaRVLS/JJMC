package servers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"jjmc/database"
	"jjmc/models"
	"jjmc/services"
)

type Instance struct {
	*models.Instance
	Manager *Manager
}

type InstanceManager struct {
	instances   map[string]*Instance
	baseDir     string
	mu          sync.RWMutex
	TemplateMgr *services.TemplateManager
}

func NewInstanceManager(baseDir string, tm *services.TemplateManager) *InstanceManager {
	// Ensure base dir exists
	os.MkdirAll(baseDir, 0755)

	// Note: Database connection is handled in cmd/jjmc/main.go or similar initialization step usually,
	// but here we might rely on it being initialized already.
	// Original code called ConnectDB() in NewInstanceManager.
	// We should probably rely on the caller to initialize DB.

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
		manager := NewManager()

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
			Manager: manager,
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

	manager := NewManager()
	instance := &Instance{
		Instance: &models.Instance{
			ID:        id,
			Name:      name,
			Directory: dir,
			Type:      "custom",
			Version:   "imported",
			MaxMemory: 2048,
		},
		Manager: manager,
	}

	instance.Manager.SetWorkDir(dir)
	// Try to find server jar?
	jars, _ := filepath.Glob(filepath.Join(dir, "*.jar"))
	jarName := "server.jar"
	if len(jars) > 0 {
		jarName = filepath.Base(jars[0])
	}
	instance.Manager.SetJar(jarName)
	instance.JarFile = jarName

	// Update DB with detected jar
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

	// Check if type matches a template
	if im.TemplateMgr != nil {
		if tmpl, ok := im.TemplateMgr.GetTemplate(serverType); ok {
			// Pre-save to DB to establish ID/Dir
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

			dir := filepath.Join(im.baseDir, id)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)

			manager := NewManager()
			instance := &Instance{
				Instance: &models.Instance{
					ID:        id,
					Name:      name,
					Directory: dir,
					Type:      serverType,
					Version:   version,
					MaxMemory: 2048,
					JarFile:   "server.jar", // Default, might be updated by template
				},
				Manager: manager,
			}
			instance.Manager.SetWorkDir(dir)
			instance.Manager.SetJar("server.jar")
			instance.Manager.SetMaxMemory(2048)

			im.instances[id] = instance

			// Execute Template Install
			go func() {
				if err := instance.InstallFromTemplate(tmpl); err != nil {
					fmt.Printf("Failed to install template for %s: %v\n", id, err)
				}
			}()

			return instance, nil
		}
	}

	// Validate Template if serverType matches a template ID
	// For now, we assume serverType IS the template ID if it matches one.
	// Else we fallback to legacy behavior?
	// User wants refactor, so let's try to enforce templates.
	// But "fabric", "forge" are not templates yet, I only made "vanilla" and "paper".
	// So I should allow legacy types for now or create templates for them?
	// I'll assume legacy types are valid for now to avoid breaking everything immediately.

	dir := filepath.Join(im.baseDir, id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create eula.txt automatically
	os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true"), 0644)

	// Save to DB
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

	manager := NewManager()
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
		Manager: manager,
	}

	instance.Manager.SetWorkDir(dir)
	instance.Manager.SetJar("server.jar")
	instance.Manager.SetMaxMemory(2048)

	im.instances[id] = instance
	return instance, nil
}

func (inst *Instance) InstallFromTemplate(tmpl models.Template) error {
	inst.Manager.Broadcast(fmt.Sprintf("Installing template: %s", tmpl.Name))

	for _, step := range tmpl.Install {
		switch step.Type {
		case "download":
			url, ok := step.Options["url"]
			if !ok {
				continue
			}
			target, ok := step.Options["target"]
			if !ok {
				target = filepath.Base(url)
			}
			targetPath := filepath.Join(inst.Directory, target)

			inst.Manager.Broadcast(fmt.Sprintf("Downloading %s...", target))
			// Use simple http get or reuse logic?
			// reusing helper if available or inline
			if err := downloadFile(targetPath, url); err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Failed: %v", err))
				return err
			}

			// If target is server.jar, ensure instance knows
			if target == "server.jar" {
				inst.JarFile = "server.jar"
				inst.Manager.SetJar("server.jar")
				// Update DB
				database.DB.Model(&models.InstanceModel{}).Where("id = ?", inst.ID).Update("jar_file", "server.jar")
			}
		}
	}

	inst.Manager.Broadcast("Installation complete.")
	return nil
}

func downloadFile(path string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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
	// Lock Manager's mutex because we are changing state that might conflict with runtime
	// But resetting involves Stop().
	inst.Manager.mu.Lock()
	defer inst.Manager.mu.Unlock()

	// Stop if running
	if inst.Manager.isRunningUnsafe() {
		// Stop() acquires lock, so we must unlock or use internal stop
		// Wait, Stop() uses Lock.
		// So we shouldn't hold lock when calling Stop().
		// But we want atomicity.
		// Refactor: Reset() shouldn't hold lock during Stop()?
		// Or we implement StopUnsafe().
		// For now, let's Release lock, Stop, Acquire lock.
		inst.Manager.mu.Unlock()
		if err := inst.Manager.Stop(); err != nil {
			inst.Manager.mu.Lock()
			return fmt.Errorf("failed to stop server: %v", err)
		}
		time.Sleep(1 * time.Second)
		inst.Manager.mu.Lock()
	}

	// List of files/directories to remove for a clean type switch
	toRemove := []string{
		inst.Manager.jarName,
		"libraries",
		"versions",
		"mods",
		"config",
		"plugins",
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

	// Delete from DB
	if err := database.DB.Delete(&models.InstanceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete from db: %v", err)
	}

	// Remove files
	os.RemoveAll(inst.Directory)
	delete(im.instances, id)
	return nil
}
