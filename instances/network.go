package instances

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (im *InstanceManager) CreateNetwork(name string, proxyType string, backendType string, backendVersion string) error {
	// 1. Generate forwarding secret
	secretBytes := make([]byte, 16)
	rand.Read(secretBytes)
	forwardingSecret := hex.EncodeToString(secretBytes)

	// 2. Determine versions
	// Let's use "latest" for velocity
	proxyVersion := "3.3.0-SNAPSHOT" // Velocity version
	if backendVersion == "" {
		backendVersion = "1.21" // Default fallback
	}

	// 3. Create Proxy Instance
	proxyId := fmt.Sprintf("%s-proxy", strings.ToLower(name))
	proxyInst, err := im.CreateInstance(proxyId, fmt.Sprintf("%s Proxy", name), "velocity", proxyVersion)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	// 4. Configure Proxy (velocity.toml)
	// We need to wait for instance directory creation, which CreateInstance does.
	// But we need to write config. logic for velocity.toml is complex.
	// For MVP, we'll append forwarding-secret to a template or just write a basic one.

	// 5. Create Backend 1 (Lobby)
	lobbyId := fmt.Sprintf("%s-lobby", strings.ToLower(name))
	lobbyInst, err := im.CreateInstance(lobbyId, fmt.Sprintf("%s Lobby", name), backendType, backendVersion)
	if err != nil {
		return fmt.Errorf("failed to create lobby: %v", err)
	}
	// Configure port (e.g. 25566)
	// Apply forwarding-secret (paper-global.yml or spigot.yml?)
	// For modern Paper, it's config/paper-global.yml -> velocity -> secret

	// 6. Create Backend 2 (Survival)
	survivalId := fmt.Sprintf("%s-survival", strings.ToLower(name))
	survivalInst, err := im.CreateInstance(survivalId, fmt.Sprintf("%s Survival", name), backendType, backendVersion)
	if err != nil {
		return fmt.Errorf("failed to create survival: %v", err)
	}
	// Configure port (e.g. 25567)

	// Update ports to avoid conflict.
	// We need a Port Manager really.
	// For MVP, let's just pick random ports or increment?
	// im.CreateInstance defaults to 25565 in server.properties?
	// We need to edit server.properties.

	// Install ViaVersion, ViaBackwards, ViaRewind
	// P1OZGk5p = ViaVersion
	// lz8f2WQI = ViaBackwards
	// 5aaWibGx = ViaRewind

	// We pass empty versionId to get latest compatible
	proxyInst.InstallMod("P1OZGk5p", "mod", "") // ViaVersion
	proxyInst.InstallMod("lz8f2WQI", "mod", "") // ViaBackwards
	proxyInst.InstallMod("5aaWibGx", "mod", "") // ViaRewind

	setupVelocity(proxyInst.Directory, forwardingSecret, []string{lobbyId, survivalId})
	setupBackend(lobbyInst.Directory, 25566, forwardingSecret, backendVersion)
	setupBackend(survivalInst.Directory, 25567, forwardingSecret, backendVersion)

	return nil
}

func setupVelocity(dir string, secret string, backends []string) {
	// Write velocity.toml
	// This is very simplified. In real app, we'd parse TOML.
	content := fmt.Sprintf(`
[servers]
lobby = "127.0.0.1:25566"
survival = "127.0.0.1:25567"
try = ["lobby"]

[advanced]
forwarding-secret = "%s"
forwarding-mode = "MODERN"
`, secret)

	os.WriteFile(filepath.Join(dir, "velocity.toml"), []byte(content), 0644)
}

func setupBackend(dir string, port int, secret string, version string) {
	// 1. server.properties
	propsPath := filepath.Join(dir, "server.properties")
	props := fmt.Sprintf("server-port=%d\nonline-mode=false\n", port)
	os.WriteFile(propsPath, []byte(props), 0644)

	// 2. Paper Config
	// Check version for config location
	// Simple interaction: 1.19+ uses config/paper-global.yml
	// 1.13-1.18 uses paper.yml
	// We assume version string starts with 1.X

	isModernPaper := false
	var major, minor int
	fmt.Sscanf(version, "%d.%d", &major, &minor)

	if major == 1 && minor >= 19 {
		isModernPaper = true
	} else if version == "latest" || version == "" {
		isModernPaper = true // Assume latest is modern
	}

	if isModernPaper {
		// 1.19+
		configDir := filepath.Join(dir, "config")
		os.MkdirAll(configDir, 0755)

		paperConfig := fmt.Sprintf(`
proxies:
  velocity:
    enabled: true
    online-mode: true
    secret: "%s"
`, secret)
		os.WriteFile(filepath.Join(configDir, "paper-global.yml"), []byte(paperConfig), 0644)
	} else {
		// 1.18 and below (Legacy Paper Config)
		// paper.yml at root
		paperConfig := fmt.Sprintf(`
settings:
  velocity-support:
    enabled: true
    online-mode: true
    secret: "%s"
`, secret)
		os.WriteFile(filepath.Join(dir, "paper.yml"), []byte(paperConfig), 0644)
	}
}
