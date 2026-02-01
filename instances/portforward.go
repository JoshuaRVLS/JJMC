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
	Provider string `json:"provider"`
	Token    string `json:"token"`
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

	port := GetServerPort(tm.InstanceDir)
	portStr := fmt.Sprintf("%d", port)

	tm.Config.Provider = provider
	tm.Config.Token = token
	tm.SaveConfig()

	var cmd *exec.Cmd
	if provider == "ngrok" {

		ngrokPath := GetNgrokPath()

		if ngrokPath == "" {

			tm.Mu.Unlock()

			logFunc := func(msg string) {
				tm.Mu.Lock()
				defer tm.Mu.Unlock()
				tm.Status.Log += msg + "\n"
			}

			logFunc("Ngrok not found. Installing...")
			if err := InstallNgrok(logFunc); err != nil {

				return fmt.Errorf("failed to install ngrok: %v", err)
			}

			tm.Mu.Lock()

			ngrokPath = GetNgrokPath()
			if ngrokPath == "" {
				tm.Mu.Unlock()
				return fmt.Errorf("ngrok installed but not found")
			}
			tm.Status.Log += "Ngrok installed successfully.\n"
		}

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

	tm.Status.Log += "\n--- Starting Tunnel ---\n"
	tm.Mu.Unlock()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			tm.Mu.Lock()
			tm.Status.Log += line + "\n"
			if len(tm.Status.Log) > 10000 {
				tm.Status.Log = tm.Status.Log[len(tm.Status.Log)-10000:]
			}

			if provider == "ngrok" {

				reLog := regexp.MustCompile(`url=(tcp://[^ ]+)`)
				if match := reLog.FindStringSubmatch(line); len(match) > 1 {
					tm.Status.PublicAddress = match[1]
				} else {

					reTui := regexp.MustCompile(`(tcp://.*:\d+)`)
					if match := reTui.FindString(line); match != "" {
						tm.Status.PublicAddress = match
					}
				}
			}
			tm.Mu.Unlock()
		}
	}()

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

	return nil
}

func (tm *TunnelManager) GetStatus() TunnelStatus {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()

	status := tm.Status

	status.Config = tm.Config
	return status
}

func GetServerPort(instanceDir string) int {
	path := filepath.Join(instanceDir, "server.properties")
	file, err := os.Open(path)
	if err != nil {
		return 25565
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}

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
