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
	} else if tmpl.ID == "velocity" || tmpl.ID == "paper" {
		// Resolve PaperMC project builds
		inst.Manager.Broadcast(fmt.Sprintf("Resolving %s build...", tmpl.ID))
		build, err := ResolvePaperBuild(tmpl.ID, version)
		if err != nil {
			return err
		}
		vars["BUILD"] = build
	} else if tmpl.ID == "vanilla" {
		inst.Manager.Broadcast("Resolving Vanilla version...")
		url, err := ResolveVanillaVersion(version)
		if err != nil {
			return err
		}
		vars["URL"] = url
	}

	for _, step := range tmpl.Install {
		switch step.Type {
		case "command":
			cmdStr, ok := step.Options["command"]
			if !ok {
				continue
			}

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

			for k, v := range vars {
				url = strings.ReplaceAll(url, "${"+k+"}", v)
			}

			// If we have a resolved URL for vanilla, use it
			if val, ok := vars["URL"]; ok && tmpl.ID == "vanilla" {
				url = val
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

			if target == "server.jar" {
				inst.JarFile = "server.jar"
				inst.Manager.SetJar("server.jar")

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
