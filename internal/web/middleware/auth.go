package middleware

import (
	"jjmc/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(am *auth.AuthManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		if !isApiRoute(path) {
			return c.Next()
		}

		if path == "/api/auth/status" || path == "/api/auth/setup" || path == "/api/auth/login" {
			return c.Next()
		}

		if !am.IsSetup() {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Setup required",
				"code":  "SETUP_REQUIRED",
			})
		}

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
