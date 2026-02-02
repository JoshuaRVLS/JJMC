package manager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

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

	webhookURL string

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

func (m *Manager) SetWebhookURL(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.webhookURL = url
}

func (m *Manager) sendWebhook(event string) {
	m.mu.Lock()
	url := m.webhookURL
	m.mu.Unlock()
	sendWebhookPayload(url, event)
}

func sendWebhookPayload(url, event string) {
	if url == "" {
		return
	}

	go func() {
		// Simple JSON payload for Discord
		payload := fmt.Sprintf(`{"content": "Server **%s**"}`, event)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(payload)))
		if err != nil {
			fmt.Printf("Failed to create webhook request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to send webhook: %v\n", err)
			return
		}
		defer resp.Body.Close()
	}()
}
