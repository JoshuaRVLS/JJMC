package forge

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"jjmc/pkg/downloader"
)

type FeedbackFunc func(string)

func Install(workDir, version string, feedback FeedbackFunc) error {
	forgeVer, err := getForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get forge version: %v", err)
	}

	fileName := fmt.Sprintf("forge-%s-%s-installer.jar", version, forgeVer)
	url := fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/%s", version, forgeVer, fileName)

	installerPath := filepath.Join(workDir, "forge-installer.jar")

	dl := downloader.New()

	feedback(fmt.Sprintf("Downloading Forge Installer %s...", forgeVer))
	err = dl.DownloadFile(downloader.DownloadOptions{
		Url:      url,
		DestPath: installerPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download forge installer: %v", err)
	}
	defer os.Remove(installerPath)

	feedback("Running Forge Installer (this may take a while)...")
	cmd := exec.Command("java", "-jar", "forge-installer.jar", "--installServer")
	cmd.Dir = workDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("forge installer failed: %v\nOutput: %s", err, string(output))
	}

	feedback("Forge Installer completed.")

	return findAndRenameForgeJar(workDir)
}

func InstallNeo(workDir, version string, feedback FeedbackFunc) error {
	neoVer, err := getNeoForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get neoforge version: %v", err)
	}

	fileName := fmt.Sprintf("neoforge-%s-installer.jar", neoVer)
	url := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/%s", neoVer, fileName)

	installerPath := filepath.Join(workDir, "neoforge-installer.jar")
	dl := downloader.New()

	feedback(fmt.Sprintf("Downloading NeoForge %s...", neoVer))
	err = dl.DownloadFile(downloader.DownloadOptions{
		Url:      url,
		DestPath: installerPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download neoforge installer: %v", err)
	}
	defer os.Remove(installerPath)

	feedback("Running NeoForge Installer...")
	cmd := exec.Command("java", "-jar", "neoforge-installer.jar", "--installServer")
	cmd.Dir = workDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("neoforge installer failed: %v\nOutput: %s", err, string(output))
	}

	return findAndRenameNeoForgeJar(workDir)
}

func getForgeVersion(mcVersion string) (string, error) {
	resp, err := http.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var promotions struct {
		Promos map[string]string `json:"promos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promotions); err != nil {
		return "", fmt.Errorf("failed to decode promotions: %v", err)
	}

	if ver, ok := promotions.Promos[mcVersion+"-recommended"]; ok {
		return ver, nil
	}
	if ver, ok := promotions.Promos[mcVersion+"-latest"]; ok {
		return ver, nil
	}

	return "", fmt.Errorf("no forge version found for %s", mcVersion)
}

func getNeoForgeVersion(mcVersion string) (string, error) {
	var prefix string
	if len(mcVersion) > 2 && mcVersion[:2] == "1." {
		prefix = mcVersion[2:]
	} else {
		return "", fmt.Errorf("unsupported mc version format: %s", mcVersion)
	}

	resp, err := http.Get("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sBody := string(body)

	parts := strings.Split(sBody, "<version>")
	var bestVer string

	for i := 1; i < len(parts); i++ {
		end := strings.Index(parts[i], "</version>")
		if end == -1 {
			continue
		}
		ver := parts[i][:end]

		if strings.HasPrefix(ver, prefix) {
			bestVer = ver
		}
	}

	if bestVer != "" {
		return bestVer, nil
	}

	return "", fmt.Errorf("no neoforge version found for mc %s (prefix %s)", mcVersion, prefix)
}

func findAndRenameNeoForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() &&
			(len(name) > 8 && name[0:8] == "neoforge") &&
			filepath.Ext(name) == ".jar" &&
			name != "neoforge-installer.jar" &&
			name != "neoforge.jar" {
			return os.Rename(filepath.Join(workDir, name), filepath.Join(workDir, "neoforge.jar"))
		}
	}
	return fmt.Errorf("could not locate installed neoforge jar")
}

func findAndRenameForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() &&
			(len(name) > 5 && name[0:5] == "forge") &&
			filepath.Ext(name) == ".jar" &&
			name != "forge-installer.jar" &&
			name != "forge.jar" {

			return os.Rename(filepath.Join(workDir, name), filepath.Join(workDir, "forge.jar"))
		}
	}
	return fmt.Errorf("could not locate installed forge jar")
}
