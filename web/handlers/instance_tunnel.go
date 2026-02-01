package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func (h *InstanceHandler) GetTunnelStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	inst, err := h.Manager.GetInstance(id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}

	return c.JSON(inst.Tunnel.GetStatus())
}

var tunnelStartReq struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

func (h *InstanceHandler) StartTunnel(c *fiber.Ctx) error {
	id := c.Params("id")
	inst, err := h.Manager.GetInstance(id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}

	if err := c.BodyParser(&tunnelStartReq); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	if err := inst.Tunnel.Start(tunnelStartReq.Provider, tunnelStartReq.Token); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(inst.Tunnel.GetStatus())
}

func (h *InstanceHandler) StopTunnel(c *fiber.Ctx) error {
	id := c.Params("id")
	inst, err := h.Manager.GetInstance(id)
	if err != nil {
		return c.Status(404).SendString("Instance not found")
	}

	if err := inst.Tunnel.Stop(); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(inst.Tunnel.GetStatus())
}
