package handlers

import (
	"jjmc/internal/instances"

	"github.com/gofiber/fiber/v2"
)

type NetworkRequest struct {
	Name           string `json:"name"`
	ProxyType      string `json:"proxyType"`
	BackendType    string `json:"backendType"`
	BackendVersion string `json:"backendVersion"`
}

func RegisterNetworkRoutes(router fiber.Router, im *instances.InstanceManager) {
	g := router.Group("/api/network")

	g.Post("/create", func(c *fiber.Ctx) error {
		var req NetworkRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if req.Name == "" || req.ProxyType == "" || req.BackendType == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
		}

		// Run in background because it takes time
		go func() {
			err := im.CreateNetwork(req.Name, req.ProxyType, req.BackendType, req.BackendVersion)
			if err != nil {
				// Log error? We don't have good async error handling for this yet
			}
		}()

		// Find a way to notify user or just return success and let them see instances appear
		return c.JSON(fiber.Map{"status": "success", "message": "Network creation started"})
	})
}
