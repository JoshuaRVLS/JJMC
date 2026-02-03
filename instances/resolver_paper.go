package instances

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ResolvePaperBuild(project string, version string) (string, error) {
	// If version is "latest", we probably need to find the latest valid version first.
	// But assuming version is like "1.21" or "3.3.0-SNAPSHOT"

	// Velocity uses "3.3.0-SNAPSHOT" etc.
	if project == "velocity" && version == "latest" {
		// Hardcoded fallback or fetch versions?
		version = "3.4.0-SNAPSHOT" // Update as needed or fetch
	}

	url := fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds", project, version)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get builds for %s %s: %s", project, version, resp.Status)
	}

	var result struct {
		Builds []struct {
			Build int `json:"build"`
		} `json:"builds"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Builds) == 0 {
		return "", fmt.Errorf("no builds found for %s %s", project, version)
	}

	// Last build is usually latest
	latest := result.Builds[len(result.Builds)-1].Build
	return fmt.Sprintf("%d", latest), nil
}
