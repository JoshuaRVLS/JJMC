package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VersionsManager handles installing and listing server versions
type VersionsManager struct {
	manager *Manager
}

func NewVersionsManager(m *Manager) *VersionsManager {
	return &VersionsManager{manager: m}
}

// WriteCounter counts the number of bytes written to it.
type WriteCounter struct {
	Total   uint64
	Current uint64
	Manager *Manager
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Current += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	// Calculate percentage
	// Only broadcast every 1% or so to avoid spamming?
	// Or just broadcast everything and let frontend throttle if needed.
	// For 50MB file, 1% is 500KB.
	// Let's just broadcast straightforwardly.
	percent := float64(wc.Current) / float64(wc.Total) * 100
	msg := fmt.Sprintf("Downloading... %.2f%%", percent)
	wc.Manager.broadcast <- msg
}

func (v *VersionsManager) InstallFabric(version string) error {
	// 1. Download Fabric Installer
	installerUrl := "https://maven.fabricmc.net/net/fabricmc/fabric-installer/1.0.1/fabric-installer-1.0.1.jar"
	installerName := "fabric-installer.jar"

	// Ensure work dir exists (already done by instance creation, but good to be safe)
	workDir := v.manager.workDir
	installerPath := filepath.Join(workDir, installerName)

	v.manager.broadcast <- "Starting download: Fabric Installer"
	if err := v.downloadFileWithProgress(installerPath, installerUrl); err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 2. Run Installer to generate server.jar
	// java -jar fabric-installer.jar server -mcversion <version> -downloadMinecraft
	v.manager.broadcast <- "Running Fabric Installer..."
	cmd := exec.Command("java", "-jar", installerName, "server", "-mcversion", version, "-downloadMinecraft")
	cmd.Dir = workDir // Execute in the instance directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	v.manager.broadcast <- "Fabric Installer completed."

	// 3. Rename/Setup
	// Fabric installer usually creates fabric-server-launch.jar and server.jar

	fabricLaunchJar := filepath.Join(workDir, "fabric-server-launch.jar")
	fabricJar := filepath.Join(workDir, "fabric.jar")

	if _, err := os.Stat(fabricLaunchJar); err == nil {
		os.Rename(fabricLaunchJar, fabricJar)
	}

	// 4. Download Fabric API Mod
	// Dynamic fetching from Modrinth
	v.manager.broadcast <- "Fetching compatible Fabric API version..."
	fabricApiUrl, err := v.getFabricApiUrl(version)
	if err != nil {
		v.manager.broadcast <- fmt.Sprintf("Warning: Failed to find Fabric API for %s: %v", version, err)
		return fmt.Errorf("failed to resolve fabric-api version: %v", err)
	}

	modsDir := filepath.Join(workDir, "mods")
	if err := os.MkdirAll(modsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mods dir: %v", err)
	}

	// Extract filename from URL or just use generic name?
	// Modrinth URLs usually end in .jar
	fabricApiName := filepath.Base(fabricApiUrl)
	fabricApiPath := filepath.Join(modsDir, fabricApiName)

	v.manager.broadcast <- fmt.Sprintf("Downloading %s...", fabricApiName)
	if err := v.downloadFileWithProgress(fabricApiPath, fabricApiUrl); err != nil {
		return fmt.Errorf("failed to download fabric-api: %v", err)
	}

	return nil
}

func (v *VersionsManager) getFabricApiUrl(mcVersion string) (string, error) {
	// Project ID for Fabric API is "P7dR8mSH" (or just "fabric-api")
	// Query Modrinth API
	// https://api.modrinth.com/v2/project/fabric-api/version?game_versions=["<ver>"]&loaders=["fabric"]

	baseUrl := "https://api.modrinth.com/v2/project/fabric-api/version"
	params := url.Values{}
	params.Add("game_versions", fmt.Sprintf(`["%s"]`, mcVersion))
	params.Add("loaders", `["fabric"]`)

	reqUrl := fmt.Sprintf("%s?%s", baseUrl, params.Encode())

	resp, err := http.Get(reqUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("modrinth api returned %d", resp.StatusCode)
	}

	// Minimal struct for parsing
	type File struct {
		Url     string `json:"url"`
		Primary bool   `json:"primary"`
	}
	type Version struct {
		Files []File `json:"files"`
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no fabric-api versions found for mc %s", mcVersion)
	}

	// Get the first version (usually latest)
	for _, file := range versions[0].Files {
		if file.Primary {
			return file.Url, nil
		}
	}

	if len(versions[0].Files) > 0 {
		return versions[0].Files[0].Url, nil
	}

	return "", fmt.Errorf("no files found in latest fabric-api version")
}

func (v *VersionsManager) InstallQuilt(version string) error {
	// 1. Download Quilt Installer
	// Using a recent version
	installerUrl := "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/0.11.0/quilt-installer-0.11.0.jar"
	installerName := "quilt-installer.jar"

	workDir := v.manager.workDir
	installerPath := filepath.Join(workDir, installerName)

	v.manager.broadcast <- "Starting download: Quilt Installer"
	if err := v.downloadFileWithProgress(installerPath, installerUrl); err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 2. Run Installer
	// java -jar quilt-installer.jar install server <minecraft-version> --download-server
	v.manager.broadcast <- "Running Quilt Installer..."
	cmd := exec.Command("java", "-jar", installerName, "install", "server", version, "--download-server")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	v.manager.broadcast <- "Quilt Installer completed."

	// 3. Rename/Setup
	// Quilt installer creates quilt-server-launch.jar
	quiltLaunchJar := filepath.Join(workDir, "quilt-server-launch.jar")
	quiltJar := filepath.Join(workDir, "quilt.jar")

	if _, err := os.Stat(quiltLaunchJar); err == nil {
		os.Rename(quiltLaunchJar, quiltJar)
	}

	return nil
}

func (v *VersionsManager) InstallForge(version string) error {
	// 1. Get Forge Version
	forgeVer, err := v.getForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get forge version: %v", err)
	}

	// 2. Construct URL
	// https://maven.minecraftforge.net/net/minecraftforge/forge/1.20.1-47.3.0/forge-1.20.1-47.3.0-installer.jar
	// Note: Some older versions might follow different naming, but for recent ones (1.17+) this is standard.
	// We might need to handle the naming carefully.
	// Usually: {mc}-{forge}/forge-{mc}-{forge}-installer.jar

	fileName := fmt.Sprintf("forge-%s-%s-installer.jar", version, forgeVer)
	url := fmt.Sprintf("https://maven.minecraftforge.net/net/minecraftforge/forge/%s-%s/%s", version, forgeVer, fileName)

	workDir := v.manager.workDir
	installerPath := filepath.Join(workDir, "forge-installer.jar")

	v.manager.broadcast <- fmt.Sprintf("Downloading Forge Installer %s...", forgeVer)
	if err := v.downloadFileWithProgress(installerPath, url); err != nil {
		return fmt.Errorf("failed to download forge installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 3. Run Installer
	// java -jar forge-installer.jar --installServer
	v.manager.broadcast <- "Running Forge Installer (this may take a while)..."
	cmd := exec.Command("java", "-jar", "forge-installer.jar", "--installServer")
	cmd.Dir = workDir

	// Capture output to log potential errors
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("forge installer failed: %v\nOutput: %s", err, string(output))
	}

	v.manager.broadcast <- "Forge Installer completed."

	// 4. Setup
	// Forge 1.17+ creates `run.sh` / `run.bat` and `user_jvm_args.txt`.
	// It relies on libraries in the `libraries` folder.
	// The main jar is usually not just "server.jar".
	// It usually provides a shim or we run the shell script.
	// However, JJMC likely expects a single JAR to run (based on current architecture).
	// Current architecture: `exec.Command("java", "-jar", inst.Jar, "nogui")` (assumed).
	// Forge 1.17+ servers are launched via `run.sh` or `java @user_jvm_args.txt -jar libraries/.../unix_args.txt ...` complex args.
	// OR they provide a `forge-1.x.x.jar` that acts as the server jar?
	// Actually recent Forge uses `run.sh` which calls `java @user_jvm_args.txt ...`.
	// We might need to change how we START the server for Forge.
	// For now, let's look for the forge jar that might be created.
	// Usually `forge-1.20.1-47.3.0-shim.jar` or similar?
	// Or maybe we treat "forge.jar" as the `run.sh` equivalent or we update the Start command logic later.
	// Let's look for a jar starting with "forge-" and ending in ".jar" that IS NOT the installer.

	// For now, I will scan for the generated jar and name it "forge.jar".
	return v.findAndRenameForgeJar(workDir)
}

func (v *VersionsManager) getForgeVersion(mcVersion string) (string, error) {
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

	// Try "recommended" first, then "latest"
	if ver, ok := promotions.Promos[mcVersion+"-recommended"]; ok {
		return ver, nil
	}
	if ver, ok := promotions.Promos[mcVersion+"-latest"]; ok {
		return ver, nil
	}

	return "", fmt.Errorf("no forge version found for %s", mcVersion)
}

func (v *VersionsManager) InstallNeoForge(version string) error {
	// 1. Get NeoForge Version
	neoVer, err := v.getNeoForgeVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get neoforge version: %v", err)
	}

	// 2. Download
	// https://maven.neoforged.net/releases/net/neoforged/neoforge/21.1.42/neoforge-21.1.42-installer.jar
	fileName := fmt.Sprintf("neoforge-%s-installer.jar", neoVer)
	url := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/%s", neoVer, fileName)

	workDir := v.manager.workDir
	installerPath := filepath.Join(workDir, "neoforge-installer.jar")

	v.manager.broadcast <- fmt.Sprintf("Downloading NeoForge %s...", neoVer)
	if err := v.downloadFileWithProgress(installerPath, url); err != nil {
		return fmt.Errorf("failed to download neoforge installer: %v", err)
	}
	defer os.Remove(installerPath)

	// 3. Install
	v.manager.broadcast <- "Running NeoForge Installer..."
	cmd := exec.Command("java", "-jar", "neoforge-installer.jar", "--installServer")
	cmd.Dir = workDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("neoforge installer failed: %v\nOutput: %s", err, string(output))
	}

	// 4. Rename
	// Similar scanning logic
	return v.findAndRenameNeoForgeJar(workDir)
}

func (v *VersionsManager) getNeoForgeVersion(mcVersion string) (string, error) {
	// Map MC version to NeoForge prefix
	// 1.21.1 -> 21.1
	// 1.20.4 -> 20.4
	// 1.20.1 -> 20.1 ? (Need to verify)
	// Generally parsing the string: remove "1.", match the rest.

	var prefix string
	if len(mcVersion) > 2 && mcVersion[:2] == "1." {
		prefix = mcVersion[2:] // "21.1" or "20.4"
	} else {
		return "", fmt.Errorf("unsupported mc version format: %s", mcVersion)
	}

	// Fetch metadata
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

	// Match versions like <version>21.1.42</version>
	// We want the highest one that starts with our prefix.
	// Example regex: <version>(21\.1\.\d+)</version>
	// But prefix is variable.

	// Simple parsing using strings
	parts := strings.Split(sBody, "<version>")
	var bestVer string

	for i := 1; i < len(parts); i++ {
		// substring until </version>
		end := strings.Index(parts[i], "</version>")
		if end == -1 {
			continue
		}
		ver := parts[i][:end]

		// check if ver starts with prefix "21.1" etc.
		if strings.HasPrefix(ver, prefix) {
			// Found a candidate.
			// Since maven metadata is usually ordered or we want the last one seen?
			// "Latest" tag also exists but we want specific match for MC version.
			// Assuming metadata is sorted or we just take the last one seen.
			// So we just keep updating bestVer.
			bestVer = ver
		}
	}

	if bestVer != "" {
		return bestVer, nil
	}

	return "", fmt.Errorf("no neoforge version found for mc %s (prefix %s)", mcVersion, prefix)
}

func (v *VersionsManager) findAndRenameNeoForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		// neoforge-21.1.42-shim.jar ? or just neoforge-...
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

func (v *VersionsManager) findAndRenameForgeJar(workDir string) error {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		// Look for forge-1.x.x.jar (but not installer)
		// Usually: forge-1.20.1-47.3.0.jar or forge-1.20.1-47.3.0-shim.jar
		if !entry.IsDir() &&
			(len(name) > 5 && name[0:5] == "forge") &&
			filepath.Ext(name) == ".jar" &&
			name != "forge-installer.jar" &&
			name != "forge.jar" { // Don't rename if already renamed

			// Found it
			return os.Rename(filepath.Join(workDir, name), filepath.Join(workDir, "forge.jar"))
		}
	}
	return fmt.Errorf("could not locate installed forge jar")
}

func (v *VersionsManager) downloadFileWithProgress(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	counter := &WriteCounter{
		Total:   uint64(resp.ContentLength),
		Manager: v.manager,
	}

	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		return err
	}
	return nil
}

func (v *VersionsManager) InstallSpigot(version string) error {
	// Direct download from cdn.getbukkit.org as requested by user
	downloadUrl := fmt.Sprintf("https://cdn.getbukkit.org/spigot/spigot-%s.jar", version)

	workDir := v.manager.workDir
	targetJarPath := filepath.Join(workDir, "server.jar")

	v.manager.broadcast <- fmt.Sprintf("Downloading Spigot %s...", version)

	// We download directly to spigot.jar.
	// Note: using downloadFileWithProgress to show progress.
	if err := v.downloadFileWithProgress(targetJarPath, downloadUrl); err != nil {
		return fmt.Errorf("failed to download Spigot: %v. Please check if version %s exists on getbukkit.org", err, version)
	}

	v.manager.broadcast <- "Spigot installed successfully."
	return nil
}
