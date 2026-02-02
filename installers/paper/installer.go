package paper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"jjmc/pkg/downloader"
)

type FeedbackFunc func(string)

func Install(workDir, version string, feedback FeedbackFunc) error {
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

	latestBuild := result.Builds[len(result.Builds)-1]
	buildNum := latestBuild.Build
	fileName := latestBuild.Downloads.Application.Name

	downloadUrl := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s/builds/%d/downloads/%s",
		version, buildNum, fileName)

	targetPath := filepath.Join(workDir, "server.jar")
	dl := downloader.New()

	feedback(fmt.Sprintf("Downloading Paper %s (Build %d)...", version, buildNum))
	err = dl.DownloadFile(downloader.DownloadOptions{
		Url:      downloadUrl,
		DestPath: targetPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download paper: %v", err)
	}

	feedback("Paper installed successfully.")
	return nil
}
