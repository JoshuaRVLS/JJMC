package handlers

import (
	"fmt"
	"jjmc/services/java_manager"

	"github.com/gofiber/fiber/v2"
)

type JavaHandler struct {
	Manager *java_manager.JavaManager
}

func NewJavaHandler(manager *java_manager.JavaManager) *JavaHandler {
	return &JavaHandler{Manager: manager}
}

func (h *JavaHandler) ListInstalled(c *fiber.Ctx) error {
	runtimes, err := h.Manager.ListInstalled()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(runtimes)
}

func (h *JavaHandler) Install(c *fiber.Ctx) error {
	var body struct {
		Version int `json:"version"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	go func() {
		if err := h.Manager.DownloadVersion(body.Version); err != nil {
			fmt.Printf("Failed to download Java %d: %v\n", body.Version, err)
		} else {
			fmt.Printf("Successfully installed Java %d\n", body.Version)
		}
	}()

	return c.JSON(fiber.Map{"status": "installing"})
}

func (h *JavaHandler) Delete(c *fiber.Ctx) error {
	name := c.Params("name")
	if err := h.Manager.DeleteVersion(name); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(200)
}
