package manager

import (
	"io"
)

func (m *Manager) handleBroadcast() {
	for msg := range m.broadcast {
		m.mu.Lock()
		// Update log buffer
		m.logBuffer = append(m.logBuffer, msg)
		if len(m.logBuffer) > 100 {
			m.logBuffer = m.logBuffer[1:]
		}

		for client := range m.clients {
			// 1 = TextMessage.
			if err := client.WriteMessage(1, []byte(msg)); err != nil {
				// If client write fails, assume connection is dead
				delete(m.clients, client)
				// check type assertion if we want to be nice
				if closer, ok := client.(io.Closer); ok {
					closer.Close()
				}
			}
		}
		m.mu.Unlock()
	}
}

func (m *Manager) RegisterClient(client ConsoleClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client] = true

	// Send recent logs
	for _, line := range m.logBuffer {
		client.WriteMessage(1, []byte(line))
	}
}

func (m *Manager) UnregisterClient(client ConsoleClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, client)
	if closer, ok := client.(io.Closer); ok {
		closer.Close()
	}
}

func (m *Manager) Broadcast(msg string) {
	m.broadcast <- msg
}
