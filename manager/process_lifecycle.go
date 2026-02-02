package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (m *Manager) isPidRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

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

	if m.tailCmd != nil {
		if m.tailCmd.Process != nil {
			m.tailCmd.Process.Kill()
		}
		m.tailCmd = nil
	}

	logPath := filepath.Join(m.workDir, "server.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
	}

	mem := m.maxMemory
	if mem <= 0 {
		mem = 2048
	}

	if m.startCommand != "" {

		cmdStr := m.startCommand
		cmdStr = strings.ReplaceAll(cmdStr, "${MAX_MEMORY}", strconv.Itoa(mem))
		cmdStr = strings.ReplaceAll(cmdStr, "${JAVA_ARGS}", m.javaArgs)

		if runtime.GOOS == "windows" {
			m.cmd = exec.Command("cmd", "/C", cmdStr)
		} else {
			m.cmd = exec.Command("sh", "-c", cmdStr)
		}
	} else {
		var args []string
		args = append(args, fmt.Sprintf("-Xmx%dM", mem))
		args = append(args, fmt.Sprintf("-Xms%dM", mem))

		if m.javaArgs != "" {
			customArgs := strings.Fields(m.javaArgs)
			args = append(args, customArgs...)
		}

		args = append(args, "-jar", m.jarName, "nogui")

		javaBin := m.javaPath
		if javaBin == "" {
			javaBin = "java"
		} else {
			if info, err := os.Stat(javaBin); err == nil && info.IsDir() {
				javaBin = filepath.Join(javaBin, "bin", "java")
			}
		}

		m.cmd = exec.Command(javaBin, args...)
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
	os.WriteFile(filepath.Join(m.workDir, "server.pid"), []byte(fmt.Sprintf("%d", m.pid)), 0644)

	go m.streamOutput(stdout, logFile)
	go m.streamOutput(stderr, logFile)

	// Start stats collection
	m.ctx, m.cancel = context.WithCancel(context.Background())
	go m.CollectStats(m.ctx)

	go func() {
		m.cmd.Wait()
		m.mu.Lock()
		m.cmd = nil
		m.pid = 0
		os.Remove(filepath.Join(m.workDir, "server.pid"))
		m.mu.Unlock()
		if logFile != nil {
			logFile.Close()
		}
		// Stop stats collection
		if m.cancel != nil {
			m.cancel()
		}
		m.broadcast <- "Server stopped"
	}()

	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tailCmd != nil {
		if m.tailCmd.Process != nil {
			m.tailCmd.Process.Kill()
		}
		m.tailCmd = nil
	}

	if m.cmd != nil && m.cmd.Process != nil {
		fmt.Fprintln(m.stdin, "stop")
		return nil
	}

	if m.pid > 0 && m.isPidRunning(m.pid) {
		process, err := os.FindProcess(m.pid)
		if err == nil {

			process.Signal(os.Interrupt)
			m.pid = 0
			os.Remove(filepath.Join(m.workDir, "server.pid"))
			return nil
		}
	}

	return fmt.Errorf("server is not running")
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {

		if err.Error() == "server is not running" {
			return m.Start()
		}
		return err
	}

	go func() {

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

		return fmt.Errorf("server is running in detached mode (backend restarted), cannot send command")
	}
	_, err := fmt.Fprintln(m.stdin, cmd)
	return err
}

func (m *Manager) loadPid() {

	data, err := os.ReadFile(filepath.Join(m.workDir, "server.pid"))
	if err == nil {
		pid, err := strconv.Atoi(string(data))
		if err == nil {
			if m.isPidRunning(pid) {
				m.pid = pid
				m.recoverLogs()
				m.startTailing()
			} else {

				os.Remove(filepath.Join(m.workDir, "server.pid"))
			}
		}
	}
}
