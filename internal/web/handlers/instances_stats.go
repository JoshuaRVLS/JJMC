package handlers

import (
	"log"

	"github.com/gofiber/contrib/websocket"
)

func (h *InstanceHandler) EnsureStatsCollection(id string) error {
	// inst is not needed unless we want to checking something specific on it.
	if _, err := h.Manager.GetInstance(id); err != nil {
		return err
	}
	// Ensuring collection is running is handled by instance lifecycle,
	// but we might want to trigger it if not running?
	// The Manager.CollectStats() logic is missing a trigger to start.
	// We should probably start it in Manager.NewManager or Manager.Start if safe.
	// But Manager.Start is for the PROCESS.
	// CollectStats() should probably run always if we want monitoring even when "stopped" (will return 0).
	// We'll address that in main/instance loader.
	return nil
}

func (h *InstanceHandler) StatsWebSocket(c *websocket.Conn) {
	id := c.Params("id")
	inst, err := h.Manager.GetInstance(id)
	if err != nil {
		log.Printf("Stats WS: Instance not found: %s", id)
		c.Close()
		return
	}

	inst.Manager.RegisterStatsClient(c)
	defer inst.Manager.UnregisterStatsClient(c)

	// Keep connection open
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
