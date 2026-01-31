package instances

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (v *VersionsManager) InstallForge(version string) error {
	// 1. Get Forge Version
	forgeVer, err := v.getForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get forge version: %v", err)
	}

	// 2. Construct URL
	fileName := fmt.Sprintf("forge-%s-%s-installer.jar", version, forgeVer)
	url := fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/%s", version, forgeVer, fileName)

	workDir := v.manager.GetWorkDir()
	installerPath := filepath.Join(workDir, "forge-installer.jar")

	v.manager.Broadcast(fmt.Sprintf("Downloading Forge Installer %s...", forgeVer))
	if err := v.downloadFileWithProgress(installerPath, url); err != nil {
		return fmt.Errorf("failed to download forge installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 3. Run Installer
	// java -jar forge-installer.jar --installServer
	v.manager.Broadcast("Running Forge Installer (this may take a while)...")
	cmd := exec.Command("java", "-jar", "forge-installer.jar", "--installServer")
	cmd.Dir = workDir

	// Capture output to log potential errors
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("forge installer failed: %v\nOutput: %s", err, string(output))
	}

	v.manager.Broadcast("Forge Installer completed.")

	// 4. Setup
	return v.findAndRenameForgeJar(workDir)
}

func (v *VersionsManager) getForgeVersion(mcVersion string) (string, error) {
	resp, err := http.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var promotions struct {
		Promos map[string]string `json:"promos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promotions); err != nil {
		return "", fmt.Errorf("failed to decode promotions: %v", err)
	}

	// Try "recommended" first, then "latest"
	if ver, ok := promotions.Promos[mcVersion+"-recommended"]; ok {
		return ver, nil
	}
	if ver, ok := promotions.Promos[mcVersion+"-latest"]; ok {
		return ver, nil
	}

	return "", fmt.Errorf("no forge version found for %s", mcVersion)
}

func (v *VersionsManager) InstallNeoForge(version string) error {
	// 1. Get NeoForge Version
	neoVer, err := v.getNeoForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get neoforge version: %v", err)
	}

	// 2. Download
	fileName := fmt.Sprintf("neoforge-%s-installer.jar", neoVer)
	url := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/%s", neoVer, fileName)

	workDir := v.manager.GetWorkDir()
	installerPath := filepath.Join(workDir, "neoforge-installer.jar")

	v.manager.Broadcast(fmt.Sprintf("Downloading NeoForge %s...", neoVer))
	if err := v.downloadFileWithProgress(installerPath, url); err != nil {
		return fmt.Errorf("failed to download neoforge installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 3. Install
	v.manager.Broadcast("Running NeoForge Installer...")
	cmd := exec.Command("java", "-jar", "neoforge-installer.jar", "--installServer")
	cmd.Dir = workDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("neoforge installer failed: %v\nOutput: %s", err, string(output))
	}

	// 4. Rename
	return v.findAndRenameNeoForgeJar(workDir)
}

func (v *VersionsManager) getNeoForgeVersion(mcVersion string) (string, error) {
	var prefix string
	if len(mcVersion) > 2 && mcVersion[:2] == "1." {
		prefix = mcVersion[2:] // "21.1" or "20.4"
	} else {
		return "", fmt.Errorf("unsupported mc version format: %s", mcVersion)
	}

	// Fetch metadata
	resp, err := http.Get("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sBody := string(body)

	// Simple parsing using strings
	parts := strings.Split(sBody, "<version>")
	var bestVer string

	for i := 1; i < len(parts); i++ {
		// substring until </version>
		end := strings.Index(parts[i], "</version>")
		if end == -1 {
			continue
		}
		ver := parts[i][:end]

		// check if ver starts with prefix "21.1" etc.
		if strings.HasPrefix(ver, prefix) {
			bestVer = ver
		}
	}

	if bestVer != "" {
		return bestVer, nil
	}

	return "", fmt.Errorf("no neoforge version found for mc %s (prefix %s)", mcVersion, prefix)
}

func (v *VersionsManager) findAndRenameNeoForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() &&
			(len(name) > 8 && name[0:8] == "neoforge") &&
			filepath.Ext(name) == ".jar" &&
			name != "neoforge-installer.jar" &&
			name != "neoforge.jar" {
			return os.Rename(filepath.Join(workDir, name), filepath.Join(workDir, "neoforge.jar"))
		}
	}
	return fmt.Errorf("could not locate installed neoforge jar")
}

func (v *VersionsManager) findAndRenameForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() &&
			(len(name) > 5 && name[0:5] == "forge") &&
			filepath.Ext(name) == ".jar" &&
			name != "forge-installer.jar" &&
			name != "forge.jar" {

			// Found it
			return os.Rename(filepath.Join(workDir, name), filepath.Join(workDir, "forge.jar"))
		}
	}
	return fmt.Errorf("could not locate installed forge jar")
}
