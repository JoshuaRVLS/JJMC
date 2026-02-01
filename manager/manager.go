package manager

import (
	"io"
	"os/exec"
	"sync"
)

type ConsoleClient interface {
	WriteMessage(messageType int, data []byte) error
}

type Manager struct {
	cmd     *exec.Cmd
	tailCmd *exec.Cmd
	stdin   io.WriteCloser
	clients map[ConsoleClient]bool
	mu      sync.Mutex

	workDir      string
	jarName      string
	startCommand string
	maxMemory    int
	javaArgs     string
	javaPath     string

	broadcast chan string
	logBuffer []string
	pid       int
	silent    bool
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

func (m *Manager) SetStartCommand(cmd string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCommand = cmd
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

func (m *Manager) SetJavaPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.javaPath = path
}

func (m *Manager) GetWorkDir() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.workDir
}

func (m *Manager) SetSilent(silent bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.silent = silent
}
