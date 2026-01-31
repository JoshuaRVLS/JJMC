package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h *InstanceHandler) SearchMods(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	query := c.Query("query")
	typeFilter := c.Query("type", "mod") // "mod" or "modpack"
	isModpack := typeFilter == "modpack"
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	sort := c.Query("sort", "")
	sides := c.Query("sides", "")
	sidesList := []string{}
	if sides != "" {
		sidesList = strings.Split(sides, ",")
	}

	results, err := inst.SearchMods(query, isModpack, offset, sort, sidesList)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}

func (h *InstanceHandler) InstallMod(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"projectId"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := inst.InstallMod(payload.ProjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "installed"})
}

func (h *InstanceHandler) GetInstalledMods(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	ids, err := inst.GetInstalledMods()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ids)
}

func (h *InstanceHandler) UninstallMod(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"project_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := inst.UninstallMod(payload.ProjectID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "uninstalled"})
}

func (h *InstanceHandler) InstallModpack(c *fiber.Ctx) error {
	inst, err := h.Manager.GetInstance(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	var payload struct {
		ProjectID string `json:"projectId"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	go func() {
		if err := inst.InstallModpack(payload.ProjectID); err != nil {
			inst.Manager.Broadcast(fmt.Sprintf("Error installing modpack: %v", err))
		}
	}()

	return c.JSON(fiber.Map{"status": "installing"})
}
