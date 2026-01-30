package middleware

import (
	"jjmc/auth"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware creates a Fiber middleware for authentication
func AuthMiddleware(am *auth.AuthManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Always allow static resources if they are not API routes
		// But here we are protecting the API specifically
		if !isApiRoute(path) {
			return c.Next()
		}

		// Public API routes
		if path == "/api/auth/status" || path == "/api/auth/setup" || path == "/api/auth/login" {
			return c.Next()
		}

		// Check Setup
		if !am.IsSetup() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Setup required",
				"code":  "SETUP_REQUIRED",
			})
		}

		// Check Token
		token := c.Cookies("auth_token")
		if token == "" || !am.ValidateSession(token) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		return c.Next()
	}
}

func isApiRoute(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

// WebsocketAuthMiddleware tailored for WS upgrades if needed,
// though the main AuthMiddleware runs before upgrade usually.
// But fiber websocket middleware is special.
// For now, we assume the main middleware chain handles cookie check.
