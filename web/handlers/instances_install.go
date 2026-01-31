package handlers

import (
	"fmt"
	"jjmc/instances"

	"github.com/gofiber/fiber/v2"
)

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
	case "bukkit":
		installErr = vm.InstallCraftBukkit(payload.Version)
		jarName = "server.jar"
	case "paper":
		installErr = vm.InstallPaper(payload.Version)
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

	if payload.Type == "bukkit" {
		if err := vm.InstallCraftBukkit(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("server.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "server.jar"})
	}

	if payload.Type == "paper" {
		if err := vm.InstallPaper(payload.Version); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		inst.Manager.SetJar("server.jar")
		return c.JSON(fiber.Map{"status": "installed", "jar": "server.jar"})
	}

	return c.Status(400).JSON(fiber.Map{"error": "Unsupported version type"})
}
