package quilt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"jjmc/pkg/downloader"
)

type FeedbackFunc func(string)

func Install(workDir, version string, feedback FeedbackFunc) error {
	installerUrl := "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/0.11.0/quilt-installer-0.11.0.jar"
	installerName := "quilt-installer.jar"
	installerPath := filepath.Join(workDir, installerName)

	dl := downloader.New()

	feedback("Starting download: Quilt Installer")
	err := dl.DownloadFile(downloader.DownloadOptions{
		Url:      installerUrl,
		DestPath: installerPath,
		OnProgress: func(current, total int64, percent float64) {
			feedback(fmt.Sprintf("Downloading... %.2f%%", percent))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to download installer: %v", err)
	}
	defer os.Remove(installerPath)

	feedback("Running Quilt Installer...")
	cmd := exec.Command("java", "-jar", installerName, "install", "server", version, "--download-server")
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installer failed: %v, output: %s", err, string(output))
	}
	feedback("Quilt Installer completed.")

	quiltLaunchJar := filepath.Join(workDir, "quilt-server-launch.jar")
	quiltJar := filepath.Join(workDir, "quilt.jar")

	if _, err := os.Stat(quiltLaunchJar); err == nil {
		os.Rename(quiltLaunchJar, quiltJar)
	}

	return nil
}
