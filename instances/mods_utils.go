package instances

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ProjectVersion struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id"`
	VersionNumber string `json:"version_number"`
	DatePublished string `json:"date_published"`
	Files         []struct {
		Url      string `json:"url"`
		Filename string `json:"filename"`
		Primary  bool   `json:"primary"`
	} `json:"files"`
	Dependencies []struct {
		VersionID      string `json:"version_id"`
		ProjectID      string `json:"project_id"`
		DependencyType string `json:"dependency_type"`
	} `json:"dependencies"`
}

func (inst *Instance) GetModVersions(projectId string, resourceType string) ([]interface{}, error) {
	if resourceType == "plugin" {
		client := NewSpigetClient()
		var id int
		if _, err := fmt.Sscanf(projectId, "%d", &id); err != nil {
			return nil, fmt.Errorf("invalid spiget resource id: %s", projectId)
		}
		versions, err := client.GetResourceVersions(id)
		if err != nil {
			return nil, err
		}
		var result []interface{}
		for _, v := range versions {
			result = append(result, map[string]interface{}{
				"id":             fmt.Sprintf("%d", v.ID),
				"name":           v.Name,
				"version_number": v.Name,
				"date_published": v.Date * 1000,
			})
		}
		return result, nil
	}

	loader := inst.Type
	mcVersion := inst.Version

	u, _ := url.Parse(fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectId))
	q := u.Query()

	q.Set("game_versions", fmt.Sprintf("[\"%s\"]", mcVersion))

	if loader != "vanilla" && loader != "" && loader != "spigot" && loader != "paper" {
		q.Set("loaders", fmt.Sprintf("[\"%s\"]", loader))
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth api error: %d", resp.StatusCode)
	}

	var versions []ProjectVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	var result []interface{}
	for _, v := range versions {
		result = append(result, map[string]interface{}{
			"id":             v.ID,
			"name":           v.VersionNumber,
			"version_number": v.VersionNumber,
			"date_published": v.DatePublished,
			"files":          v.Files,
		})
	}
	return result, nil
}

func (inst *Instance) getCompatibleVersion(projectId string) (*ProjectVersion, error) {
	loader := inst.Type
	mcVersion := inst.Version

	u, _ := url.Parse(fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectId))
	q := u.Query()
	q.Set("game_versions", fmt.Sprintf("[\"%s\"]", mcVersion))

	q.Set("game_versions", fmt.Sprintf("[\"%s\"]", mcVersion))

	if loader != "vanilla" && loader != "" && loader != "spigot" && loader != "paper" && loader != "bukkit" {
		q.Set("loaders", fmt.Sprintf("[\"%s\"]", loader))
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth version error (%d): %s", resp.StatusCode, string(body))
	}

	var versions []ProjectVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no compatible versions found")
	}

	return &versions[0], nil
}

func (inst *Instance) downloadFile(path string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download error (%d): %s", resp.StatusCode, url)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
