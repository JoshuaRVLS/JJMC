package handlers

import (
	"jjmc/instances"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ModpackHandler struct{}

func NewModpackHandler() *ModpackHandler {
	return &ModpackHandler{}
}

func (h *ModpackHandler) Search(c *fiber.Ctx) error {
	query := c.Query("query")
	version := c.Query("version")
	loader := c.Query("loader")

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	sort := c.Query("sort", "")
	sides := c.Query("sides", "")

	sidesList := []string{}
	if sides != "" {
		sidesList = strings.Split(sides, ",")
	}

	// Always search for modpacks
	results, err := instances.SearchModrinth(query, "modpack", version, loader, offset, sort, sidesList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(results)
}
