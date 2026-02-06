package instances

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (v *VersionsManager) InstallSpigot(version string) error {
	return v.runBuildTools(version, "spigot")
}

func (v *VersionsManager) InstallCraftBukkit(version string) error {
	return v.runBuildTools(version, "craftbukkit")
}

func (v *VersionsManager) runBuildTools(version string, serverType string) error {
	workDir := v.manager.GetWorkDir()
	buildDir := filepath.Join(workDir, "build-tools")

	useDocker := false
	if path, err := exec.LookPath("docker"); err == nil && path != "" {
		useDocker = true
	}

	if !useDocker {
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git is required for BuildTools but was not found in PATH")
		}
	}

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build dir: %v", err)
	}
	defer os.RemoveAll(buildDir)

	buildToolsUrl := "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar"
	buildToolsPath := filepath.Join(buildDir, "BuildTools.jar")

	v.manager.Broadcast("Downloading BuildTools...")
	if err := v.downloadFileWithProgress(buildToolsPath, buildToolsUrl); err != nil {
		return fmt.Errorf("failed to download BuildTools: %v", err)
	}

	var cmd *exec.Cmd

	if useDocker {
		v.manager.Broadcast(fmt.Sprintf("Running BuildTools for %s %s (using Docker)...", serverType, version))

		absBuildDir, err := filepath.Abs(buildDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		dockerImage := "maven:3.9-eclipse-temurin-21"

		cmd = exec.Command("docker", "run", "--rm",
			"-v", fmt.Sprintf("%s:/data", absBuildDir),
			"-w", "/data",
			dockerImage,
			"java", "-jar", "BuildTools.jar", "--rev", version)
	} else {
		v.manager.Broadcast(fmt.Sprintf("Running BuildTools for %s %s (Local)...", serverType, version))

		cmd = exec.Command("java", "-jar", "BuildTools.jar", "--rev", version)
		cmd.Dir = buildDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("BuildTools failed: %v\nOutput: %s", err, string(output))
	}

	v.manager.Broadcast("BuildTools finished successfully.")

	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("failed to read build dir: %v", err)
	}

	var foundJar string

	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasPrefix(name, serverType+"-") && strings.HasSuffix(name, ".jar") {
			foundJar = name
			break
		}
	}

	if foundJar == "" {

		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && strings.HasPrefix(name, serverType) && strings.HasSuffix(name, ".jar") {
				foundJar = name
				break
			}
		}
	}

	if foundJar == "" {
		return fmt.Errorf("could not locate installed server jar in %s", buildDir)
	}

	srcPath := filepath.Join(buildDir, foundJar)
	destPath := filepath.Join(workDir, "server.jar")

	os.Remove(destPath)

	if err := os.Rename(srcPath, destPath); err != nil {
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move server jar: %v", err)
		}
	}

	v.manager.Broadcast(fmt.Sprintf("%s installed as server.jar", foundJar))
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
