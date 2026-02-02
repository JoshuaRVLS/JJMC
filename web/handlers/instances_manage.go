package handlers

import (
	"github.com/gofiber/fiber/v2"
)

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
		MaxMemory  int    `json:"maxMemory"`
		JavaArgs   string `json:"javaArgs"`
		JarFile    string `json:"jarFile"`
		JavaPath   string `json:"javaPath"`
		WebhookURL string `json:"webhookUrl"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.Manager.UpdateSettings(id, payload.MaxMemory, payload.JavaArgs, payload.JarFile, payload.JavaPath, payload.WebhookURL); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "updated"})
}
