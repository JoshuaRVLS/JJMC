package instances

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ResolveVanillaVersion(version string) (string, error) {
	// 1. Fetch version manifest
	manifestUrl := "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	resp, err := http.Get(manifestUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch version manifest: %v", err)
	}
	defer resp.Body.Close()

	var manifest struct {
		Versions []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"versions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", fmt.Errorf("failed to decode version manifest: %v", err)
	}

	// 2. Find matching version
	var versionUrl string
	for _, v := range manifest.Versions {
		if v.ID == version {
			versionUrl = v.URL
			break
		}
	}

	if versionUrl == "" {
		return "", fmt.Errorf("version %s not found in manifest", version)
	}

	// 3. Fetch version package
	respArg, err := http.Get(versionUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch version package: %v", err)
	}
	defer respArg.Body.Close()

	var pkg struct {
		Downloads struct {
			Server struct {
				URL string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}

	if err := json.NewDecoder(respArg.Body).Decode(&pkg); err != nil {
		return "", fmt.Errorf("failed to decode version package: %v", err)
	}

	if pkg.Downloads.Server.URL == "" {
		return "", fmt.Errorf("server download url not found for version %s", version)
	}

	return pkg.Downloads.Server.URL, nil
}
