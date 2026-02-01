package instances

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (inst *Instance) InstallModpack(projectId string) error {
	ver, err := inst.getCompatibleVersion(projectId)
	if err != nil {
		return err
	}

	modsDir := filepath.Join(inst.Directory, "mods")
	inst.Manager.Broadcast("Resetting mods directory...")
	os.RemoveAll(modsDir)
	os.MkdirAll(modsDir, 0755)

	var mrpackUrl string
	for _, f := range ver.Files {
		if strings.HasSuffix(f.Filename, ".mrpack") {
			mrpackUrl = f.Url
			break
		}
	}
	if mrpackUrl == "" {
		return fmt.Errorf("no .mrpack file found for version")
	}

	packPath := filepath.Join(inst.Directory, "modpack.mrpack")
	inst.Manager.Broadcast("Downloading modpack...")
	if err := inst.downloadFile(packPath, mrpackUrl); err != nil {
		return err
	}
	defer os.Remove(packPath)

	inst.Manager.Broadcast("Parsing modpack...")
	r, err := zip.OpenReader(packPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var indexFile *zip.File
	for _, f := range r.File {
		if f.Name == "modrinth.index.json" {
			indexFile = f
			break
		}
	}
	if indexFile == nil {
		return fmt.Errorf("invalid modpack: missing modrinth.index.json")
	}

	rc, err := indexFile.Open()
	if err != nil {
		return err
	}

	var index struct {
		Files []struct {
			Path      string            `json:"path"`
			Hashes    map[string]string `json:"hashes"`
			Downloads []string          `json:"downloads"`
		} `json:"files"`
	}

	if err := json.NewDecoder(rc).Decode(&index); err != nil {
		rc.Close()
		return err
	}
	rc.Close()

	totalFiles := len(index.Files)
	for i, f := range index.Files {
		if len(f.Downloads) == 0 {
			continue
		}

		localPath := filepath.Join(inst.Directory, f.Path)
		os.MkdirAll(filepath.Dir(localPath), 0755)

		inst.Manager.Broadcast(fmt.Sprintf("Downloading file %d/%d: %s", i+1, totalFiles, filepath.Base(f.Path)))
		if err := inst.downloadFile(localPath, f.Downloads[0]); err != nil {
			inst.Manager.Broadcast(fmt.Sprintf("Failed to download %s: %v", f.Path, err))
		}
	}

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "overrides/") {
			relPath := strings.TrimPrefix(f.Name, "overrides/")
			if relPath == "" || strings.HasSuffix(relPath, "/") {
				continue
			}
			target := filepath.Join(inst.Directory, relPath)
			os.MkdirAll(filepath.Dir(target), 0755)

			src, _ := f.Open()
			dst, _ := os.Create(target)
			io.Copy(dst, src)
			src.Close()
			dst.Close()
		}
	}

	inst.Manager.Broadcast("Modpack installed successfully.")
	return nil
}
