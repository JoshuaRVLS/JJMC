package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *InstanceHandler) ListFiles(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}
	path := c.Query("path", ".")
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

	file, err := inst.ReadFileStream(path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer file.Close()

	return c.SendStream(file)
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

	if payload.Destination == "" {
		payload.Destination = "."

	}

	if err := inst.DecompressFile(payload.File, payload.Destination); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": "decompressed"})
}
