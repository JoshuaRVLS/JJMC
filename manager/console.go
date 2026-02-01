package manager

import (
	"io"
	"strings"
	"time"
)

func (m *Manager) handleBroadcast() {
	for msg := range m.broadcast {
		m.mu.Lock()

		m.logBuffer = append(m.logBuffer, msg)
		if len(m.logBuffer) > 100 {
			m.logBuffer = m.logBuffer[1:]
		}

		for client := range m.clients {

			if err := client.WriteMessage(1, []byte(msg)); err != nil {

				delete(m.clients, client)

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

type executeCapture struct {
	output chan string
	done   chan struct{}
}

func (e *executeCapture) WriteMessage(messageType int, data []byte) error {
	select {
	case e.output <- string(data):
	case <-e.done:
	}
	return nil
}

func (m *Manager) ExecuteCommand(cmd string, timeout time.Duration) (string, error) {
	if err := m.WriteCommand(cmd); err != nil {
		return "", err
	}

	capture := &executeCapture{
		output: make(chan string, 100),
		done:   make(chan struct{}),
	}

	m.RegisterClient(capture)
	defer m.UnregisterClient(capture)

	var output []string
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	<-timer.C
	close(capture.done)

Loop:
	for {
		select {
		case line := <-capture.output:
			output = append(output, line)
		default:
			break Loop
		}
	}

	return strings.Join(output, "\n"), nil
}
