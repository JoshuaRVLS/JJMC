package instances

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PaperSource struct{}

func (s *PaperSource) Search(query string) ([]SourceResult, error) {

	return []SourceResult{
		{ID: "paper", Name: "Paper", Description: "High performance Spigot fork", Type: "server"},
		{ID: "velocity", Name: "Velocity", Description: "Minecraft proxy", Type: "proxy"},
		{ID: "waterfall", Name: "Waterfall", Description: "BungeeCord fork", Type: "proxy"},
		{ID: "folia", Name: "Folia", Description: "Regionized multithreading", Type: "server"},
	}, nil
}

func (s *PaperSource) GetLatestVersion() (string, error) {
	return s.getLatestVersion("paper")
}

func (s *PaperSource) getLatestVersion(project string) (string, error) {

	url := fmt.Sprintf("https://api.papermc.io/v2/projects/%s", project)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Versions []string `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Versions) == 0 {
		return "", fmt.Errorf("no versions found")
	}

	return result.Versions[len(result.Versions)-1], nil
}

func (s *PaperSource) GetVersion(versionId string) (*SourceVersion, error) {

	project := "paper"
	version := versionId

	url := fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds", project, version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paper api returned %d", resp.StatusCode)
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
		return nil, err
	}

	if len(result.Builds) == 0 {
		return nil, fmt.Errorf("no builds found for %s %s", project, version)
	}

	latest := result.Builds[len(result.Builds)-1]

	downloadUrl := fmt.Sprintf("https://api.papermc.io/v2/projects/%s/versions/%s/builds/%d/downloads/%s",
		project, version, latest.Build, latest.Downloads.Application.Name)

	return &SourceVersion{
		ID:      version,
		Name:    fmt.Sprintf("Paper-%s-%d", version, latest.Build),
		Version: version,
		Files: []SourceFile{
			{
				Url:      downloadUrl,
				Filename: latest.Downloads.Application.Name,
				Hash:     latest.Downloads.Application.Sha256,
				HashAlgo: "sha256",
				Primary:  true,
			},
		},
	}, nil
}
