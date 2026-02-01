package instances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func (v *VersionsManager) InstallFabric(version string) error {

	installerUrl := "https://maven.fabricmc.net/net/fabricmc/fabric-installer/1.0.1/fabric-installer-1.0.1.jar"
	installerName := "fabric-installer.jar"

	workDir := v.manager.GetWorkDir()

	installerPath := filepath.Join(workDir, installerName)

	v.manager.Broadcast("Starting download: Fabric Installer")
	if err := v.downloadFileWithProgress(installerPath, installerUrl); err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	v.manager.Broadcast("Running Fabric Installer...")
	cmd := exec.Command("java", "-jar", installerName, "server", "-mcversion", version, "-downloadMinecraft")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	v.manager.Broadcast("Fabric Installer completed.")

	fabricLaunchJar := filepath.Join(workDir, "fabric-server-launch.jar")
	fabricJar := filepath.Join(workDir, "fabric.jar")

	if _, err := os.Stat(fabricLaunchJar); err == nil {
		os.Rename(fabricLaunchJar, fabricJar)
	}

	v.manager.Broadcast("Fetching compatible Fabric API version...")
	fabricApiUrl, err := v.getFabricApiUrl(version)
	if err != nil {
		v.manager.Broadcast(fmt.Sprintf("Warning: Failed to find Fabric API for %s: %v", version, err))
		return fmt.Errorf("failed to resolve fabric-api version: %v", err)
	}

	modsDir := filepath.Join(workDir, "mods")
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods dir: %v", err)
	}

	fabricApiName := filepath.Base(fabricApiUrl)
	fabricApiPath := filepath.Join(modsDir, fabricApiName)

	v.manager.Broadcast(fmt.Sprintf("Downloading %s...", fabricApiName))
	if err := v.downloadFileWithProgress(fabricApiPath, fabricApiUrl); err != nil {
		return fmt.Errorf("failed to download fabric-api: %v", err)
	}

	return nil
}

func (v *VersionsManager) getFabricApiUrl(mcVersion string) (string, error) {
	baseUrl := "https://api.modrinth.com/v2/project/fabric-api/version"
	params := url.Values{}
	params.Add("game_versions", fmt.Sprintf(`["%s"]`, mcVersion))
	params.Add("loaders", `["fabric"]`)

	reqUrl := fmt.Sprintf("%s?%s", baseUrl, params.Encode())

	resp, err := http.Get(reqUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("modrinth api returned %d", resp.StatusCode)
	}

	type File struct {
		Url     string `json:"url"`
		Primary bool   `json:"primary"`
	}
	type Version struct {
		Files []File `json:"files"`
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no fabric-api versions found for mc %s", mcVersion)
	}

	for _, file := range versions[0].Files {
		if file.Primary {
			return file.Url, nil
		}
	}

	if len(versions[0].Files) > 0 {
		return versions[0].Files[0].Url, nil
	}

	return "", fmt.Errorf("no files found in latest fabric-api version")
}

func (v *VersionsManager) InstallQuilt(version string) error {

	installerUrl := "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/0.11.0/quilt-installer-0.11.0.jar"
	installerName := "quilt-installer.jar"

	workDir := v.manager.GetWorkDir()
	installerPath := filepath.Join(workDir, installerName)

	v.manager.Broadcast("Starting download: Quilt Installer")
	if err := v.downloadFileWithProgress(installerPath, installerUrl); err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	v.manager.Broadcast("Running Quilt Installer...")
	cmd := exec.Command("java", "-jar", installerName, "install", "server", version, "--download-server")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	v.manager.Broadcast("Quilt Installer completed.")

	quiltLaunchJar := filepath.Join(workDir, "quilt-server-launch.jar")
	quiltJar := filepath.Join(workDir, "quilt.jar")

	if _, err := os.Stat(quiltLaunchJar); err == nil {
		os.Rename(quiltLaunchJar, quiltJar)
	}

	return nil
}
