package web

import (
	"jjmc/auth"
	"jjmc/instances"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func RegisterBackupRoutes(router fiber.Router, authManager *auth.AuthManager, im *instances.InstanceManager) {
	g := router.Group("/api/instances/:id/backups")

	g.Get("/", func(c *fiber.Ctx) error {
		id := c.Params("id")
		backups, err := im.ListBackups(id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(backups)
	})

	g.Post("/", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := im.CreateBackup(id); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})

	g.Post("/:filename/restore", func(c *fiber.Ctx) error {
		id := c.Params("id")
		filename, err := url.QueryUnescape(c.Params("filename"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid filename encoding"})
		}

		if err := im.RestoreBackup(id, filename); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})

	g.Delete("/:filename", func(c *fiber.Ctx) error {
		id := c.Params("id")
		filename, err := url.QueryUnescape(c.Params("filename"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid filename encoding"})
		}

		if err := im.DeleteBackup(id, filename); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "success"})
	})
}
