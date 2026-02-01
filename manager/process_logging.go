package manager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func (m *Manager) recoverLogs() {
	// Read last 100 lines of server.log
	logPath := fmt.Sprintf("%s/server.log", m.workDir)
	file, err := os.Open(logPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Take last 100
	start := 0
	if len(lines) > 100 {
		start = len(lines) - 100
	}
	m.logBuffer = lines[start:]
}

func (m *Manager) startTailing() {
	if m.tailCmd != nil {
		return
	}
	// tail -f -n 0 server.log
	logPath := fmt.Sprintf("%s/server.log", m.workDir)
	m.tailCmd = exec.Command("tail", "-f", "-n", "0", logPath)

	stdout, err := m.tailCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed to pipe tail:", err)
		return
	}

	if err := m.tailCmd.Start(); err != nil {
		fmt.Println("Failed to start tail:", err)
		return
	}

	go m.streamOutput(stdout, nil) // nil logFile because tail reads FROM file

	go func() {
		m.tailCmd.Wait()
		m.mu.Lock()
		m.tailCmd = nil
		m.mu.Unlock()
	}()
}

func (m *Manager) streamOutput(r io.Reader, logFile io.Writer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		if !m.silent {
			fmt.Println("Server:", text)
		}

		// Write to log file if available
		if logFile != nil {
			fmt.Fprintln(logFile, text)
		}

		// Buffer log
		m.mu.Lock()
		m.logBuffer = append(m.logBuffer, text)
		if len(m.logBuffer) > 100 {
			m.logBuffer = m.logBuffer[1:]
		}
		m.mu.Unlock()

		m.broadcast <- text
	}
}
