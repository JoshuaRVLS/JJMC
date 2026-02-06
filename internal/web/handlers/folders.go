package handlers

import (
	"time"

	"jjmc/internal/database"
	"jjmc/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FolderHandler struct{}

func (h *FolderHandler) List(c *fiber.Ctx) error {
	var folders []models.Folder
	if err := database.DB.Find(&folders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list folders"})
	}
	return c.JSON(folders)
}

func (h *FolderHandler) Create(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if payload.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Folder name is required"})
	}

	folder := models.Folder{
		ID:        uuid.New().String(),
		Name:      payload.Name,
		CreatedAt: time.Now().Unix(),
	}

	if err := database.DB.Create(&folder).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create folder"})
	}

	return c.JSON(folder)
}

func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	// Transaction to delete folder and unset FolderID on instances
	tx := database.DB.Begin()

	if err := tx.Model(&models.InstanceModel{}).Where("folder_id = ?", id).Update("folder_id", "").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update instances"})
	}

	if err := tx.Delete(&models.Folder{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete folder"})
	}

	tx.Commit()
	return c.JSON(fiber.Map{"status": "deleted"})
}

func (h *FolderHandler) Rename(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := database.DB.Model(&models.Folder{}).Where("id = ?", id).Update("name", payload.Name).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to rename folder"})
	}

	return c.JSON(fiber.Map{"status": "updated"})
}
