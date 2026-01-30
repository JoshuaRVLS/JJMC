package web

import (
	"jjmc/auth"
	"jjmc/instances"

	"github.com/gofiber/fiber/v2"
)

func RegisterBackupRoutes(router fiber.Router, authManager *auth.AuthManager, im *instances.InstanceManager) {
	g := router.Group("/api/instances/:id/backups")

	// List
	g.Get("/", func(c *fiber.Ctx) error {
		id := c.Params("id")
		backups, err := im.ListBackups(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(backups)
	})

	// Create
	g.Post("/", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := im.CreateBackup(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})

	// Restore
	g.Post("/:filename/restore", func(c *fiber.Ctx) error {
		id := c.Params("id")
		filename := c.Params("filename")
		if err := im.RestoreBackup(id, filename); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})

	// Delete
	g.Delete("/:filename", func(c *fiber.Ctx) error {
		id := c.Params("id")
		filename := c.Params("filename")
		if err := im.DeleteBackup(id, filename); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})
}
