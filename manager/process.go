package manager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (m *Manager) loadPid() {
	// Assumes caller holds lock
	data, err := os.ReadFile(fmt.Sprintf("%s/server.pid", m.workDir))
	if err == nil {
		pid, err := strconv.Atoi(string(data))
		if err == nil {
			if m.isPidRunning(pid) {
				m.pid = pid
				m.recoverLogs()
				m.startTailing()
			} else {
				// Cleanup stale pid
				os.Remove(fmt.Sprintf("%s/server.pid", m.workDir))
			}
		}
	}
}

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

func (m *Manager) isPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Send signal 0 to check if running
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// IsRunningUnsafe checks if running without locking.
// Caller MUST hold lock.
func (m *Manager) IsRunningUnsafe() bool {
	if m.cmd != nil && m.cmd.Process != nil && m.cmd.ProcessState == nil {
		return true
	}
	if m.pid > 0 && m.isPidRunning(m.pid) {
		return true
	}
	return false
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.IsRunningUnsafe()
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsRunningUnsafe() {
		return fmt.Errorf("server is already running")
	}

	// Stop tailing if it was running (shouldn't be, but sanity check)
	if m.tailCmd != nil {
		if m.tailCmd.Process != nil {
			m.tailCmd.Process.Kill()
		}
		m.tailCmd = nil
	}

	// Create log file
	logPath := fmt.Sprintf("%s/server.log", m.workDir)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	}

	// Build Config
	mem := m.maxMemory
	if mem <= 0 {
		mem = 2048
	}

	if m.startCommand != "" {
		// Replace variables in command
		cmdStr := m.startCommand
		cmdStr = strings.ReplaceAll(cmdStr, "${MAX_MEMORY}", strconv.Itoa(mem))
		cmdStr = strings.ReplaceAll(cmdStr, "${JAVA_ARGS}", m.javaArgs)

		// Use sh -c for complex commands
		m.cmd = exec.Command("sh", "-c", cmdStr)
	} else {
		var args []string
		args = append(args, fmt.Sprintf("-Xmx%dM", mem))
		args = append(args, fmt.Sprintf("-Xms%dM", mem))

		if m.javaArgs != "" {
			customArgs := strings.Fields(m.javaArgs)
			args = append(args, customArgs...)
		}

		args = append(args, "-jar", m.jarName, "nogui")

		m.cmd = exec.Command("java", args...)
	}

	m.cmd.Dir = m.workDir

	stdin, err := m.cmd.StdinPipe()
	if err != nil {
		return err
	}
	m.stdin = stdin

	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := m.cmd.Start(); err != nil {
		return err
	}

	m.pid = m.cmd.Process.Pid
	os.WriteFile(fmt.Sprintf("%s/server.pid", m.workDir), []byte(fmt.Sprintf("%d", m.pid)), 0644)

	// Stream output to both file and broadcast
	go m.streamOutput(stdout, logFile)
	go m.streamOutput(stderr, logFile)

	go func() {
		m.cmd.Wait()
		m.mu.Lock()
		m.cmd = nil
		m.pid = 0
		os.Remove(fmt.Sprintf("%s/server.pid", m.workDir))
		m.mu.Unlock()
		if logFile != nil {
			logFile.Close()
		}
		m.broadcast <- "Server stopped"
	}()

	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop tailing if active
	if m.tailCmd != nil {
		if m.tailCmd.Process != nil {
			m.tailCmd.Process.Kill()
		}
		m.tailCmd = nil
	}

	// If cmd exists, use it
	if m.cmd != nil && m.cmd.Process != nil {
		fmt.Fprintln(m.stdin, "stop")
		return nil
	}

	// If cmd missing but PID exists (orphaned), kill it
	if m.pid > 0 && m.isPidRunning(m.pid) {
		process, err := os.FindProcess(m.pid)
		if err == nil {
			// SIGTERM
			process.Signal(syscall.SIGTERM)
			m.pid = 0
			os.Remove(fmt.Sprintf("%s/server.pid", m.workDir))
			return nil
		}
	}

	return fmt.Errorf("server is not running")
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		// If it's not running, just Start
		if err.Error() == "server is not running" {
			return m.Start()
		}
		return err
	}

	// Wait for stop in background then Start
	go func() {
		// Wait up to 60 seconds for exit
		for i := 0; i < 60; i++ {
			if !m.IsRunning() {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if !m.IsRunning() {
			m.Start()
		} else {
			m.broadcast <- "Restart failed: server didn't stop in time"
		}
	}()

	return nil
}

func (m *Manager) WriteCommand(cmd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd == nil || m.stdin == nil {
		// Can't write to stdin of orphaned process?
		// We could potentially write to /proc/PID/fd/0? No.
		return fmt.Errorf("server is running in detached mode (backend restarted), cannot send command")
	}
	_, err := fmt.Fprintln(m.stdin, cmd)
	return err
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
