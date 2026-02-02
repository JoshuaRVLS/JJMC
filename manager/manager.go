package manager

import (
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type ConsoleClient interface {
	WriteMessage(messageType int, data []byte) error
}

type Manager struct {
	cmd     *exec.Cmd
	tailCmd *exec.Cmd
	stdin   io.WriteCloser
	clients map[ConsoleClient]bool

	// Stats
	StatsClients   map[*websocket.Conn]bool
	StatsBroadcast chan interface{}

	mu sync.Mutex

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

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager() *Manager {
	m := &Manager{
		clients:   make(map[ConsoleClient]bool),
		broadcast: make(chan string),
		jarName:   "server.jar",
		workDir:   ".",
		maxMemory: 2048,
		logBuffer: make([]string, 0, 100),

		StatsClients:   make(map[*websocket.Conn]bool),
		StatsBroadcast: make(chan interface{}),
	}
	go m.handleBroadcast()
	go m.handleStatsBroadcast()
	// Stats collection is now started when the server starts
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
