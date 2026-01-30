package manager

import (
	"io"
	"os/exec"
	"sync"
)

// Client interface for console consumers
type ConsoleClient interface {
	WriteMessage(messageType int, data []byte) error
}

// Manager controls a single server instance process
type Manager struct {
	cmd     *exec.Cmd
	tailCmd *exec.Cmd // Command for tailing logs of detached process
	stdin   io.WriteCloser
	clients map[ConsoleClient]bool
	mu      sync.Mutex

	// Configuration
	workDir   string
	jarName   string
	maxMemory int
	javaArgs  string

	broadcast chan string
	logBuffer []string
	pid       int
}

func NewManager() *Manager {
	m := &Manager{
		clients:   make(map[ConsoleClient]bool),
		broadcast: make(chan string),
		jarName:   "server.jar",
		workDir:   ".",
		maxMemory: 2048,
		logBuffer: make([]string, 0, 100),
	}
	go m.handleBroadcast()
	return m
}

func (m *Manager) SetWorkDir(dir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workDir = dir
	m.loadPid()
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

func (m *Manager) GetWorkDir() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.workDir
}
