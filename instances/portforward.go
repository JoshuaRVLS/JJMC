package instances

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
)

type TunnelConfig struct {
	Provider string `json:"provider"` // "playit" or "ngrok"
	Token    string `json:"token"`    // Secret key or auth token
}

type TunnelStatus struct {
	Running       bool         `json:"running"`
	Provider      string       `json:"provider"`
	PublicAddress string       `json:"public_address"`
	Log           string       `json:"log"`
	Config        TunnelConfig `json:"config"`
}

type TunnelManager struct {
	InstanceDir string
	Cmd         *exec.Cmd
	Config      TunnelConfig
	Status      TunnelStatus
	Mu          sync.Mutex
}

func NewTunnelManager(instanceDir string) *TunnelManager {
	tm := &TunnelManager{
		InstanceDir: instanceDir,
	}
	tm.LoadConfig()
	return tm
}

func (tm *TunnelManager) LoadConfig() {
	path := filepath.Join(tm.InstanceDir, "tunnel.json")
	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		json.Unmarshal(data, &tm.Config)
	}
}

func (tm *TunnelManager) SaveConfig() {
	path := filepath.Join(tm.InstanceDir, "tunnel.json")
	data, _ := json.MarshalIndent(tm.Config, "", "  ")
	os.WriteFile(path, data, 0644)
}

func (tm *TunnelManager) Start(provider, token string) error {
	tm.Mu.Lock()
	if tm.Status.Running {
		tm.Mu.Unlock()
		return fmt.Errorf("tunnel already running")
	}

	// Determine port from server.properties
	port := GetServerPort(tm.InstanceDir)
	portStr := fmt.Sprintf("%d", port)

	tm.Config.Provider = provider
	tm.Config.Token = token
	tm.SaveConfig()

	var cmd *exec.Cmd
	if provider == "playit" {
		// playit --secret <token>
		cmd = exec.Command("playit", "--secret", token)
	} else if provider == "ngrok" {
		// Check/Install ngrok
		ngrokPath := GetNgrokPath()

		if ngrokPath == "" {
			// Release lock while installing so GetStatus isn't blocked!
			tm.Mu.Unlock()

			logFunc := func(msg string) {
				tm.Mu.Lock()
				defer tm.Mu.Unlock()
				tm.Status.Log += msg + "\n"
			}

			logFunc("Ngrok not found. Installing...")
			if err := InstallNgrok(logFunc); err != nil {
				// Re-acquire lock to ensure safe return (though we return error immediately)
				return fmt.Errorf("failed to install ngrok: %v", err)
			}

			// Re-acquire lock for the rest of setup
			tm.Mu.Lock()

			ngrokPath = GetNgrokPath()
			if ngrokPath == "" {
				tm.Mu.Unlock()
				return fmt.Errorf("ngrok installed but not found")
			}
			tm.Status.Log += "Ngrok installed successfully.\n"
		}

		// ngrok tcp <port> --authtoken <token> --log=stdout
		tm.Status.Log += fmt.Sprintf("Forwarding local port %s\n", portStr)
		args := []string{"tcp", portStr, "--log=stdout"}
		if token != "" {
			args = append(args, "--authtoken", token)
		}
		cmd = exec.Command(ngrokPath, args...)
	} else {
		tm.Mu.Unlock()
		return fmt.Errorf("unknown provider: %s", provider)
	}

	// Capture output to find address
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		tm.Mu.Unlock()
		return err
	}

	tm.Cmd = cmd
	tm.Status.Running = true
	tm.Status.Provider = provider
	tm.Status.PublicAddress = "Starting..."
	// Don't clear log, just append separator
	tm.Status.Log += "\n--- Starting Tunnel ---\n"
	tm.Mu.Unlock() // Unlock here, goroutines will lock as needed

	// Monitor output in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			tm.Mu.Lock()
			tm.Status.Log += line + "\n"
			if len(tm.Status.Log) > 10000 { // Increased log size
				tm.Status.Log = tm.Status.Log[len(tm.Status.Log)-10000:]
			}
			// Parse address
			if provider == "ngrok" {
				// Standard TUI: Forwarding                    tcp://0.tcp.ngrok.io:12345 -> localhost:25565
				// Log format: url=tcp://0.tcp.ngrok.io:12345

				// Try log format first
				reLog := regexp.MustCompile(`url=(tcp://[^ ]+)`)
				if match := reLog.FindStringSubmatch(line); len(match) > 1 {
					tm.Status.PublicAddress = match[1]
				} else {
					// Try TUI format
					reTui := regexp.MustCompile(`(tcp://.*:\d+)`)
					if match := reTui.FindString(line); match != "" {
						tm.Status.PublicAddress = match
					}
				}
			} else if provider == "playit" {
				// Playit output varies, usually prints a claim URL or address
				// Looking for .gl.joinmc.link or similar
				if match, _ := regexp.MatchString(`.*\.gl\.joinmc\.link.*`, line); match {
					tm.Status.PublicAddress = line // Simplified for now
				}
			}
			tm.Mu.Unlock()
		}
	}()

	// Monitor stderr too
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			tm.Mu.Lock()
			tm.Status.Log += line + "\n"
			if len(tm.Status.Log) > 5000 {
				tm.Status.Log = tm.Status.Log[len(tm.Status.Log)-5000:]
			}
			tm.Mu.Unlock()
		}
	}()

	// Wait for exit
	go func() {
		cmd.Wait()
		tm.Mu.Lock()
		tm.Status.Running = false
		tm.Status.PublicAddress = ""
		tm.Cmd = nil
		tm.Mu.Unlock()
	}()

	return nil
}

func (tm *TunnelManager) Stop() error {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()

	if !tm.Status.Running || tm.Cmd == nil {
		return nil
	}

	if err := tm.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		tm.Cmd.Process.Kill()
	}

	// Wait logic handled by the goroutine above
	return nil
}

func (tm *TunnelManager) GetStatus() TunnelStatus {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()

	status := tm.Status
	// Always return the saved config so UI can populate it
	status.Config = tm.Config
	return status
}

func GetServerPort(instanceDir string) int {
	path := filepath.Join(instanceDir, "server.properties")
	file, err := os.Open(path)
	if err != nil {
		return 25565 // Default
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		// server-port=25565
		if match, _ := regexp.MatchString(`^server-port=(\d+)`, line); match {
			re := regexp.MustCompile(`^server-port=(\d+)`)
			parts := re.FindStringSubmatch(line)
			if len(parts) > 1 {
				var port int
				fmt.Sscanf(parts[1], "%d", &port)
				return port
			}
		}
	}
	return 25565
}
