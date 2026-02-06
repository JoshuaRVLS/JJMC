package fabric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"jjmc/pkg/downloader"
)

type FeedbackFunc func(string)

func Install(workDir, version string, feedback FeedbackFunc) error {
	installerUrl := "https://maven.fabricmc.net/net/fabricmc/fabric-installer/1.1.1/fabric-installer-1.1.1.jar"
	installerName := "fabric-installer.jar"
	installerPath := filepath.Join(workDir, installerName)

	dl := downloader.New()

	feedback("Starting download: Fabric Installer")
	err := dl.DownloadFile(downloader.DownloadOptions{
		Url:      installerUrl,
		DestPath: installerPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	feedback("Running Fabric Installer...")
	cmd := exec.Command("java", "-jar", installerName, "server", "-mcversion", version, "-downloadMinecraft")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	feedback("Fabric Installer completed.")

	fabricLaunchJar := filepath.Join(workDir, "fabric-server-launch.jar")
	fabricJar := filepath.Join(workDir, "fabric.jar")

	if _, err := os.Stat(fabricLaunchJar); err == nil {
		os.Rename(fabricLaunchJar, fabricJar)
	}

	feedback("Fetching compatible Fabric API version...")
	fabricApiUrl, err := getFabricApiUrl(version)
	if err != nil {
		feedback(fmt.Sprintf("Warning: Failed to find Fabric API for %s: %v", version, err))
		// We don't fail the entire install if API fetch fails, matching original logic?
		// Original logic returned error: return fmt.Errorf("failed to resolve fabric-api version: %v", err)
		return fmt.Errorf("failed to resolve fabric-api version: %v", err)
	}

	modsDir := filepath.Join(workDir, "mods")
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods dir: %v", err)
	}

	fabricApiName := filepath.Base(fabricApiUrl)
	fabricApiPath := filepath.Join(modsDir, fabricApiName)

	feedback(fmt.Sprintf("Downloading %s...", fabricApiName))
	err = dl.DownloadFile(downloader.DownloadOptions{
		Url:      fabricApiUrl,
		DestPath: fabricApiPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download fabric-api: %v", err)
	}

	return nil
}

func getFabricApiUrl(mcVersion string) (string, error) {
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
