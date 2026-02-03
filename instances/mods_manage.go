package instances

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (inst *Instance) InstallMod(projectId string, resourceType string, versionId string) error {
	if resourceType == "plugin" {
		client := NewSpigetClient()
		var id int
		if _, err := fmt.Sscanf(projectId, "%d", &id); err != nil {
			return fmt.Errorf("invalid spiget resource id: %s", projectId)
		}

		res, err := client.GetResourceDetails(id)
		if err != nil {
			return fmt.Errorf("failed to get resource details: %v", err)
		}

		safeName := strings.ReplaceAll(res.Name, " ", "_")
		safeName = strings.Map(func(r rune) rune {
			if strings.ContainsRune(`\/:*?"<>|`, r) {
				return -1
			}
			return r
		}, safeName)

		fileName := fmt.Sprintf("%s.jar", safeName)
		pluginsDir := filepath.Join(inst.Directory, "plugins")
		os.MkdirAll(pluginsDir, 0755)

		targetPath := filepath.Join(pluginsDir, fileName)

		downloadUrl := client.GetDownloadURL(id)

		if versionId != "" {
			var vid int
			if _, err := fmt.Sscanf(versionId, "%d", &vid); err == nil {
				downloadUrl = client.GetVersionDownloadURL(id, vid)
			}
		}

		inst.Manager.Broadcast(fmt.Sprintf("Downloading plugin %s...", fileName))

		err = inst.downloadFile(targetPath, downloadUrl)
		if err != nil {
			return err
		}

		metaPath := filepath.Join(inst.Directory, "installed_plugins.json")
		var plugins []InstalledPlugin
		if data, err := os.ReadFile(metaPath); err == nil {
			json.Unmarshal(data, &plugins)
		}

		found := false
		for i, p := range plugins {
			if p.ID == projectId {
				plugins[i].Filename = fileName
				plugins[i].Name = res.Name
				found = true
				break
			}
		}
		if !found {
			plugins = append(plugins, InstalledPlugin{
				ID:       projectId,
				Name:     res.Name,
				Filename: fileName,
			})
		}

		data, _ := json.MarshalIndent(plugins, "", "  ")
		return os.WriteFile(metaPath, data, 0644)
	}

	var ver *ProjectVersion
	var err error

	if versionId != "" {

		u := fmt.Sprintf("https://api.modrinth.com/v2/version/%s", versionId)
		resp, err := http.Get(u)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get version %s", versionId)
		}
		ver = &ProjectVersion{}
		if err := json.NewDecoder(resp.Body).Decode(ver); err != nil {
			return err
		}
	} else {

		ver, err = inst.getCompatibleVersion(projectId)
		if err != nil {
			return err
		}
	}

	var fileUrl, fileName string
	for _, f := range ver.Files {
		if f.Primary {
			fileUrl = f.Url
			fileName = f.Filename
			break
		}
	}
	if fileUrl == "" && len(ver.Files) > 0 {
		fileUrl = ver.Files[0].Url
		fileName = ver.Files[0].Filename
	}
	if fileUrl == "" {
		return fmt.Errorf("no files found for version %s", ver.ID)
	}

	// Check for Velocity instance type to determine directory
	targetDir := "mods"
	if inst.Type == "velocity" {
		targetDir = "plugins"
	}

	modsDir := filepath.Join(inst.Directory, targetDir)
	os.MkdirAll(modsDir, 0755)

	targetPath := filepath.Join(modsDir, fileName)
	inst.Manager.Broadcast(fmt.Sprintf("Downloading mod %s...", fileName))

	if err := inst.downloadFile(targetPath, fileUrl); err != nil {
		return err
	}

	// Handle Dependencies
	for _, dep := range ver.Dependencies {
		if dep.DependencyType == "required" && dep.ProjectID != "" {
			inst.Manager.Broadcast(fmt.Sprintf("Installing dependency %s...", dep.ProjectID))
			// Recursively install dependency. We don't specify versionId to let it find compatible one.
			// Or if versionID is present in dependency, utilize it?
			// Modrinth dependency usually has version_id or project_id.
			// If version_id is set, use it.
			if err := inst.InstallMod(dep.ProjectID, "mod", dep.VersionID); err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Failed to install dependency %s: %v", dep.ProjectID, err))
				// We don't hard fail on dependency failure, just warn? or hard fail?
				// Let's log it but continue for now to avoid total failure chain.
				fmt.Printf("Dependency error: %v\n", err)
			}
		}
	}

	return nil
}

func (inst *Instance) UninstallMod(projectId string, resourceType string) error {
	if resourceType == "plugin" {
		metaPath := filepath.Join(inst.Directory, "installed_plugins.json")
		var plugins []InstalledPlugin
		if data, err := os.ReadFile(metaPath); err == nil {
			json.Unmarshal(data, &plugins)
		}

		var newPlugins []InstalledPlugin
		var filename string
		for _, p := range plugins {
			if p.ID == projectId {
				filename = p.Filename
			} else {
				newPlugins = append(newPlugins, p)
			}
		}

		if filename != "" {

			os.Remove(filepath.Join(inst.Directory, "plugins", filename))

			data, _ := json.MarshalIndent(newPlugins, "", "  ")
			os.WriteFile(metaPath, data, 0644)
			return nil
		}
		return fmt.Errorf("plugin not found in metadata")
	}

	return fmt.Errorf("uninstall not supported for this type yet")
}

func (inst *Instance) GetInstalledMods() ([]string, error) {
	var ids []string

	metaPath := filepath.Join(inst.Directory, "installed_plugins.json")
	if _, err := os.Stat(metaPath); err == nil {
		var plugins []InstalledPlugin
		if data, err := os.ReadFile(metaPath); err == nil {
			json.Unmarshal(data, &plugins)
			for _, p := range plugins {
				ids = append(ids, p.ID)
			}
		}
	}

	modsDir := filepath.Join(inst.Directory, "mods")
	files, err := os.ReadDir(modsDir)
	if err == nil {
		var hashes []string
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jar") {
				continue
			}

			path := filepath.Join(modsDir, f.Name())
			h, err := hashFile(path)
			if err == nil {
				hashes = append(hashes, h)
			}
		}

		if len(hashes) > 0 {
			payload, _ := json.Marshal(map[string]interface{}{
				"hashes":    hashes,
				"algorithm": "sha1",
			})

			resp, err := http.Post("https://api.modrinth.com/v2/version_files", "application/json", bytes.NewBuffer(payload))
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var result map[string]struct {
						ProjectID string `json:"project_id"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
						installed := make(map[string]bool)

						for _, id := range ids {
							installed[id] = true
						}

						for _, v := range result {
							if !installed[v.ProjectID] {
								installed[v.ProjectID] = true
								ids = append(ids, v.ProjectID)
							}
						}
					}
				}
			}
		}
	}

	return ids, nil
}
