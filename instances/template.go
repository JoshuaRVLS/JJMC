package instances

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"jjmc/database"
	"jjmc/models"
)

func (inst *Instance) InstallFromTemplate(tmpl models.Template, version string) error {
	inst.Manager.Broadcast(fmt.Sprintf("Installing template: %s (Version %s)", tmpl.Name, version))

	// Resolve complex versions
	vars := map[string]string{
		"VERSION": version,
	}

	if tmpl.ID == "forge" {
		inst.Manager.Broadcast("Resolving Forge version...")
		forgeVer, err := ResolveForgeVersion(version)
		if err != nil {
			return err
		}
		vars["VERSION"] = forgeVer
		vars["MC_VERSION"] = version
		vars["FORGE_VERSION"] = forgeVer
		vars["FULL_VERSION"] = fmt.Sprintf("%s-%s", version, forgeVer)
	} else if tmpl.ID == "neoforge" {
		inst.Manager.Broadcast("Resolving NeoForge version...")
		neoVer, err := ResolveNeoForgeVersion(version)
		if err != nil {
			return err
		}
		vars["NEOFORGE_VERSION"] = neoVer
	}

	for _, step := range tmpl.Install {
		switch step.Type {
		case "command":
			cmdStr, ok := step.Options["command"]
			if !ok {
				continue
			}

			// Replace variables
			for k, v := range vars {
				cmdStr = strings.ReplaceAll(cmdStr, "${"+k+"}", v)
			}

			inst.Manager.Broadcast(fmt.Sprintf("Executing: %s", cmdStr))

			cmd := exec.Command("sh", "-c", cmdStr)
			cmd.Dir = inst.Directory

			output, err := cmd.CombinedOutput()
			if err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)))
				return fmt.Errorf("command failed: %s", string(output))
			}
			inst.Manager.Broadcast(fmt.Sprintf("Output: %s", string(output)))

		case "download":
			url, ok := step.Options["url"]
			if !ok {
				continue
			}

			// Replace variables
			for k, v := range vars {
				url = strings.ReplaceAll(url, "${"+k+"}", v)
			}

			target, ok := step.Options["target"]
			if !ok {
				target = filepath.Base(url)
			}
			targetPath := filepath.Join(inst.Directory, target)

			inst.Manager.Broadcast(fmt.Sprintf("Downloading %s...", target))
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
