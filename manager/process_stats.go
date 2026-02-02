package manager

import (
	"fmt"
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
	fmt.Println("Starting handleStatsBroadcast")
	for stats := range m.StatsBroadcast {
		m.mu.Lock()
		if len(m.StatsClients) > 0 {
			fmt.Printf("Broadcasting stats to %d clients\n", len(m.StatsClients))
		}
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

	var lastPid int
	var proc *process.Process

	for range ticker.C {
		if !m.IsRunning() {
			// Send zero stats if not running
			m.StatsBroadcast <- ProcessStats{
				CPU:    0,
				Memory: 0,
				Time:   time.Now().Unix(),
			}
			lastPid = 0
			proc = nil
			continue
		}

		m.mu.Lock()
		pid := m.pid
		m.mu.Unlock()

		if pid == 0 {
			continue
		}

		if pid != lastPid || proc == nil {
			p, err := process.NewProcess(int32(pid))
			if err != nil {
				fmt.Printf("Error creating process: %v\n", err)
				continue
			}
			proc = p
			lastPid = pid
		}

		cpu, _ := proc.Percent(0)
		memInfo, _ := proc.MemoryInfo()
		mem := uint64(0)
		if memInfo != nil {
			mem = memInfo.RSS
		}

		fmt.Printf("Stats: CPU=%.2f, Mem=%d\n", cpu, mem)

		m.StatsBroadcast <- ProcessStats{
			CPU:    cpu,
			Memory: mem,
			Time:   time.Now().Unix(),
		}
	}
}
