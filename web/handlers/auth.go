package handlers

import (
	"time"

	"jjmc/auth"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Manager *auth.AuthManager
}

func NewAuthHandler(am *auth.AuthManager) *AuthHandler {
	return &AuthHandler{Manager: am}
}

func (h *AuthHandler) GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"isSetup":       h.Manager.IsSetup(),
		"authenticated": h.Manager.ValidateSession(c.Cookies("auth_token")),
	})
}

func (h *AuthHandler) Setup(c *fiber.Ctx) error {
	if h.Manager.IsSetup() {
		return c.Status(400).JSON(fiber.Map{"error": "Already setup"})
	}
	var payload struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}
	if len(payload.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}
	if err := h.Manager.SetPassword(payload.Password); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Auto login
	token := h.Manager.CreateSession()
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{"status": "setup_complete"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var payload struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if !h.Manager.VerifyPassword(payload.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
	}

	token := h.Manager.CreateSession()
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{"status": "logged_in"})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	token := c.Cookies("auth_token")
	if token != "" {
		h.Manager.RevokeSession(token)
	}
	c.ClearCookie("auth_token")
	return c.JSON(fiber.Map{"status": "logged_out"})
}
