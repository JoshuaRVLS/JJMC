package handlers

import (
	"jjmc/internal/models"
	"jjmc/internal/services/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	Scheduler *scheduler.Scheduler
}

func NewScheduleHandler(s *scheduler.Scheduler) *ScheduleHandler {
	return &ScheduleHandler{Scheduler: s}
}

func (h *ScheduleHandler) List(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	var schedules []models.Schedule
	if err := h.Scheduler.DB.Where("instance_id = ?", instanceID).Find(&schedules).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list schedules"})
	}
	return c.JSON(schedules)
}

func (h *ScheduleHandler) Create(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	var schedule models.Schedule
	if err := c.BodyParser(&schedule); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	schedule.ID = uuid.New().String()
	schedule.InstanceID = instanceID
	schedule.Enabled = true

	if err := h.Scheduler.DB.Create(&schedule).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create schedule"})
	}

	if err := h.Scheduler.AddJob(schedule); err != nil {
		// Rollback if cron add fails
		h.Scheduler.DB.Delete(&schedule)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid cron expression: " + err.Error()})
	}

	return c.JSON(schedule)
}

func (h *ScheduleHandler) Update(c *fiber.Ctx) error {
	scheduleID := c.Params("scheduleId")
	var payload models.Schedule
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	var schedule models.Schedule
	if err := h.Scheduler.DB.First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Schedule not found"})
	}

	schedule.Name = payload.Name
	schedule.CronExpression = payload.CronExpression
	schedule.Type = payload.Type
	schedule.Payload = payload.Payload
	schedule.Enabled = payload.Enabled

	if err := h.Scheduler.DB.Save(&schedule).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update schedule"})
	}

	if err := h.Scheduler.AddJob(schedule); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid cron expression: " + err.Error()})
	}

	return c.JSON(schedule)
}

func (h *ScheduleHandler) Delete(c *fiber.Ctx) error {
	scheduleID := c.Params("scheduleId")

	h.Scheduler.RemoveJob(scheduleID)

	if err := h.Scheduler.DB.Delete(&models.Schedule{}, "id = ?", scheduleID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete schedule"})
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}
