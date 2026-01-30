package handlers

import (
	"fmt"
	"jjmc/instances"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type InstanceHandler struct {
	Manager *instances.InstanceManager
}

func NewInstanceHandler(im *instances.InstanceManager) *InstanceHandler {
	return &InstanceHandler{Manager: im}
}

func (h *InstanceHandler) List(c *fiber.Ctx) error {
	return c.JSON(h.Manager.ListInstances())
}

func (h *InstanceHandler) Get(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	return c.JSON(inst)
}

func (h *InstanceHandler) Create(c *fiber.Ctx) error {
	var payload struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Version string `json:"version"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	inst, err := h.Manager.CreateInstance(payload.ID, payload.Name, payload.Type, payload.Version)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(inst)
}

func (h *InstanceHandler) Import(c *fiber.Ctx) error {
	var payload struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		SourcePath string `json:"sourcePath"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	inst, err := h.Manager.ImportInstance(payload.ID, payload.Name, payload.SourcePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(inst)
}

func (h *InstanceHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.Manager.DeleteInstance(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *InstanceHandler) UpdateSettings(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload struct {
		MaxMemory int    `json:"maxMemory"`
		JavaArgs  string `json:"javaArgs"`
		JarFile   string `json:"jarFile"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.Manager.UpdateSettings(id, payload.MaxMemory, payload.JavaArgs, payload.JarFile); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}

func (h *InstanceHandler) ChangeType(c *fiber.Ctx) error {
	id := c.Params("id")
	inst, err := h.Manager.GetInstance(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		Type    string `json:"type"`
		Version string `json:"version"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	// 1. Reset Instance
	if err := inst.Reset(payload.Type, payload.Version); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to reset instance: %v", err)})
	}

	// 2. Install New Type
	if payload.Type == "custom" {
		return c.JSON(fiber.Map{"status": "changed", "Note": "Custom type requires manual jar upload"})
	}

	vm := instances.NewVersionsManager(inst.Manager)
	var installErr error
	var jarName string

	switch payload.Type {
	case "fabric":
		installErr = vm.InstallFabric(payload.Version)
		jarName = "fabric.jar"
	case "quilt":
		installErr = vm.InstallQuilt(payload.Version)
		jarName = "quilt.jar"
	case "forge":
		installErr = vm.InstallForge(payload.Version)
		jarName = "forge.jar"
	case "neoforge":
		installErr = vm.InstallNeoForge(payload.Version)
		jarName = "neoforge.jar"
	case "spigot":
		installErr = vm.InstallSpigot(payload.Version)
		jarName = "server.jar"
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported type for auto-install"})
	}

	if installErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Reset successful, but install failed: %v", installErr)})
	}

	inst.Manager.SetJar(jarName)
	return c.JSON(fiber.Map{"status": "changed"})
}

func (h *InstanceHandler) Start(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	if err := inst.Manager.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "started"})
}

func (h *InstanceHandler) Stop(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	if err := inst.Manager.Stop(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "stopped"})
}

func (h *InstanceHandler) Restart(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	if err := inst.Manager.Restart(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "restarting"})
}

func (h *InstanceHandler) Command(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	var payload struct {
		Command string `json:"command"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if err := inst.Manager.WriteCommand(payload.Command); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "sent"})
}

func (h *InstanceHandler) Install(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		Version string `json:"version"`
		Type    string `json:"type"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	vm := instances.NewVersionsManager(inst.Manager)

	if payload.Type == "fabric" {
		if err := vm.InstallFabric(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("fabric.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "fabric.jar"})
	}

	if payload.Type == "quilt" {
		if err := vm.InstallQuilt(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("quilt.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "quilt.jar"})
	}

	if payload.Type == "forge" {
		if err := vm.InstallForge(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("forge.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "forge.jar"})
	}

	if payload.Type == "neoforge" {
		if err := vm.InstallNeoForge(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("neoforge.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "neoforge.jar"})
	}

	if payload.Type == "spigot" {
		if err := vm.InstallSpigot(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("server.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "server.jar"})
	}

	return c.Status(400).JSON(fiber.Map{"error": "Unsupported version type"})
}

// Mods / Files handlers could be part of InstanceHandler or separate
// Let's implement Mods here for now.

func (h *InstanceHandler) SearchMods(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	query := c.Query("query")
	typeFilter := c.Query("type", "mod") // "mod" or "modpack"
	isModpack := typeFilter == "modpack"
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	sort := c.Query("sort", "")
	sides := c.Query("sides", "")
	sidesList := []string{}
	if sides != "" {
		sidesList = strings.Split(sides, ",")
	}

	results, err := inst.SearchMods(query, isModpack, offset, sort, sidesList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}

func (h *InstanceHandler) InstallMod(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"projectId"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := inst.InstallMod(payload.ProjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "installed"})
}

func (h *InstanceHandler) GetInstalledMods(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	ids, err := inst.GetInstalledMods()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ids)
}

func (h *InstanceHandler) UninstallMod(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"project_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := inst.UninstallMod(payload.ProjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "uninstalled"})
}

func (h *InstanceHandler) InstallModpack(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"projectId"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	go func() {
		if err := inst.InstallModpack(payload.ProjectID); err != nil {
			inst.Manager.Broadcast(fmt.Sprintf("Error installing modpack: %v", err))
		}
	}()

	return c.JSON(fiber.Map{"status": "installing"})
}

// Files
func (h *InstanceHandler) ListFiles(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	path := c.Query("path", ".") // Default to root
	files, err := inst.ListFiles(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(files)
}

func (h *InstanceHandler) ReadFile(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	path := c.Query("path")
	if path == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Path required"})
	}
	content, err := inst.ReadFile(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendString(string(content))
}

func (h *InstanceHandler) WriteFile(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := inst.WriteFile(payload.Path, []byte(payload.Content)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "saved"})
}

func (h *InstanceHandler) DeleteFile(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	path := c.Query("path")
	if path == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Path required"})
	}
	if err := inst.DeleteFile(path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *InstanceHandler) Mkdir(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	var payload struct {
		Path string `json:"path"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if err := inst.Mkdir(payload.Path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "created"})
}

func (h *InstanceHandler) Upload(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	path := c.Query("path", ".")
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid form"})
	}

	files := form.File["files"]
	for _, file := range files {
		if err := inst.HandleUpload(path, file); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to upload %s: %v", file.Filename, err)})
		}
	}

	return c.JSON(fiber.Map{"status": "uploaded", "count": len(files)})
}

func (h *InstanceHandler) Compress(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		Files       []string `json:"files"`
		Destination string   `json:"destination"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if len(payload.Files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No files selected"})
	}

	if err := inst.CompressFiles(payload.Files, payload.Destination); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "compressed"})
}

func (h *InstanceHandler) Decompress(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		File        string `json:"file"`
		Destination string `json:"destination"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	// Default dest to current dir if empty, or handle in backend
	if payload.Destination == "" {
		payload.Destination = "." // Extract relative to instance root, essentially "here" if logic handles it
		// But DecompressFile expects relative path. "." is fine.
		// Wait, if I am in "plugins/", I want to extract there?
		// Frontend should pass correct relative destination path (e.g. "plugins/")
	}

	if err := inst.DecompressFile(payload.File, payload.Destination); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "decompressed"})
}
