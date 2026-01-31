package instances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
)

func (v *VersionsManager) InstallPaper(version string) error {
	// PaperMC API v2
	// 1. Get latest build
	baseUrl := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds", version)
	resp, err := http.Get(baseUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch paper builds: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("paper api returned %d", resp.StatusCode)
	}

	var result struct {
		Builds []struct {
			Build     int `json:"build"`
			Downloads struct {
				Application struct {
					Name   string `json:"name"`
					Sha256 string `json:"sha256"`
				} `json:"application"`
			} `json:"downloads"`
		} `json:"builds"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode paper builds: %v", err)
	}

	if len(result.Builds) == 0 {
		return fmt.Errorf("no builds found for paper %s", version)
	}

	// Last build is usually the latest
	latestBuild := result.Builds[len(result.Builds)-1]
	buildNum := latestBuild.Build
	fileName := latestBuild.Downloads.Application.Name

	// 2. Download
	downloadUrl := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%d/downloads/%s",
		version, buildNum, fileName)

	workDir := v.manager.GetWorkDir()
	targetPath := filepath.Join(workDir, "server.jar")

	v.manager.Broadcast(fmt.Sprintf("Downloading Paper %s (Build %d)...", version, buildNum))
	if err := v.downloadFileWithProgress(targetPath, downloadUrl); err != nil {
		return fmt.Errorf("failed to download paper: %v", err)
	}

	v.manager.Broadcast("Paper installed successfully.")
	return nil
}
