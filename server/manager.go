package server

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type Manager struct {
	cmd        *exec.Cmd
	tailCmd    *exec.Cmd // Command for tailing logs of detached process
	stdin      io.WriteCloser
	clients    map[*websocket.Conn]bool
	broadcast  chan string
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
	jarName    string
	workDir    string
	maxMemory  int
	javaArgs   string
	logBuffer  []string
	pid        int
}

func NewManager() *Manager {
	m := &Manager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan string),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		jarName:    "server.jar",
		workDir:    ".",
		maxMemory:  2048,
		logBuffer:  make([]string, 0, 100),
	}
	go m.run()
	return m
}

func (m *Manager) SetWorkDir(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workDir = dir
	m.loadPid()
}

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

	// Naive implementation: read all (ok for small files, bad for large)
	// For production, seek to end and read backwards.
	// Given we controlled the creation, let's assume it's manageable or rotate.
	// We'll just append what we find.
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

// isRunningUnsafe checks if running without locking.
// Caller MUST hold lock.
func (m *Manager) isRunningUnsafe() bool {
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
	return m.isRunningUnsafe()
}

func (m *Manager) SetJar(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jarName = name
}

func (m *Manager) SetMaxMemory(mem int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxMemory = mem
}

func (m *Manager) SetJavaArgs(args string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.javaArgs = args
}

func (m *Manager) Broadcast(msg string) {
	m.broadcast <- msg
}

func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			m.clients[client] = true
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Close()
			}
		case msg := <-m.broadcast:
			for client := range m.clients {
				client.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					client.Close()
					delete(m.clients, client)
				}
			}
		}
	}
}

func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunningUnsafe() {
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
	var args []string
	args = append(args, fmt.Sprintf("-Xmx%dM", mem))
	args = append(args, fmt.Sprintf("-Xms%dM", mem))

	if m.javaArgs != "" {
		customArgs := strings.Fields(m.javaArgs)
		args = append(args, customArgs...)
	}

	args = append(args, "-jar", m.jarName, "nogui")

	m.cmd = exec.Command("java", args...)
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
	return m.Stop()
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
		fmt.Println("Server:", text)

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

func (m *Manager) RegisterClient(c *websocket.Conn) {
	m.mu.Lock()
	// Send buffered logs
	for _, log := range m.logBuffer {
		c.WriteMessage(websocket.TextMessage, []byte(log))
	}
	m.mu.Unlock()
	m.register <- c
}

func (m *Manager) UnregisterClient(c *websocket.Conn) {
	m.unregister <- c
}
