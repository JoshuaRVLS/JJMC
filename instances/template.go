package instances

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"jjmc/database"
	"jjmc/models"
)

func (inst *Instance) InstallFromTemplate(tmpl models.Template) error {
	inst.Manager.Broadcast(fmt.Sprintf("Installing template: %s", tmpl.Name))

	for _, step := range tmpl.Install {
		switch step.Type {
		case "download":
			url, ok := step.Options["url"]
			if !ok {
				continue
			}
			target, ok := step.Options["target"]
			if !ok {
				target = filepath.Base(url)
			}
			targetPath := filepath.Join(inst.Directory, target)

			inst.Manager.Broadcast(fmt.Sprintf("Downloading %s...", target))
			// Use simple http get or reuse logic?
			// reusing helper if available or inline
			if err := downloadFile(targetPath, url); err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Failed: %v", err))
				return err
			}

			// If target is server.jar, ensure instance knows
			if target == "server.jar" {
				inst.JarFile = "server.jar"
				inst.Manager.SetJar("server.jar")
				// Update DB
				database.DB.Model(&models.InstanceModel{}).Where("id = ?", inst.ID).Update("jar_file", "server.jar")
			}
		}
	}

	inst.Manager.Broadcast("Installation complete.")
	return nil
}

func downloadFile(path string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
