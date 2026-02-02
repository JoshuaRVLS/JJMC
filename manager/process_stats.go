package manager

import (
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessStats struct {
	CPU    float64 `json:"cpu"`
	Memory uint64  `json:"memory"`
	Time   int64   `json:"time"`
}

func (m *Manager) RegisterStatsClient(c *websocket.Conn) {
	m.mu.Lock()
	m.StatsClients[c] = true
	m.mu.Unlock()
}

func (m *Manager) UnregisterStatsClient(c *websocket.Conn) {
	m.mu.Lock()
	delete(m.StatsClients, c)
	m.mu.Unlock()
}

func (m *Manager) handleStatsBroadcast() {
	for stats := range m.StatsBroadcast {
		m.mu.Lock()
		for client := range m.StatsClients {
			if err := client.WriteJSON(stats); err != nil {
				client.Close()
				delete(m.StatsClients, client)
			}
		}
		m.mu.Unlock()
	}
}

func (m *Manager) CollectStats() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !m.IsRunning() {
			// Send zero stats if not running
			m.StatsBroadcast <- ProcessStats{
				CPU:    0,
				Memory: 0,
				Time:   time.Now().Unix(),
			}
			continue
		}

		m.mu.Lock()
		pid := m.pid
		m.mu.Unlock()

		if pid == 0 {
			continue
		}

		proc, err := process.NewProcess(int32(pid))
		if err != nil {
			continue
		}

		cpu, _ := proc.Percent(0)
		memInfo, _ := proc.MemoryInfo()
		mem := uint64(0)
		if memInfo != nil {
			mem = memInfo.RSS
		}

		m.StatsBroadcast <- ProcessStats{
			CPU:    cpu,
			Memory: mem,
			Time:   time.Now().Unix(),
		}
	}
}
