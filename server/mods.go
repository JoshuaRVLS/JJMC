package server

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ModSearchResult reflects the structure returned by Modrinth Search API
type ModSearchResult struct {
	Hits []struct {
		ProjectID    string   `json:"project_id"`
		Slug         string   `json:"slug"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Categories   []string `json:"categories"`
		ClientSide   string   `json:"client_side"`
		ServerSide   string   `json:"server_side"`
		ProjectType  string   `json:"project_type"`
		Downloads    int      `json:"downloads"`
		IconUrl      string   `json:"icon_url"`
		Author       string   `json:"author"`
		Versions     []string `json:"versions"`
		Followers    int      `json:"followers"`
		DateCreated  string   `json:"date_created"`
		DateModified string   `json:"date_modified"`
		License      string   `json:"license"`
		Gallery      []string `json:"gallery"` // Simplified
	} `json:"hits"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total_hits"`
}

type ProjectVersion struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id"`
	VersionNumber string `json:"version_number"`
	Files         []struct {
		Url      string `json:"url"`
		Filename string `json:"filename"`
		Primary  bool   `json:"primary"`
	} `json:"files"`
	Dependencies []struct {
		VersionID      string `json:"version_id"`
		ProjectID      string `json:"project_id"`
		DependencyType string `json:"dependency_type"` // "required", "optional"
	} `json:"dependencies"`
}

// SearchMods queries Modrinth
func (inst *Instance) SearchMods(query string, isModpack bool, offset int, sort string, sides []string) ([]interface{}, error) {
	loader := inst.Type
	mcVersion := inst.Version

	ptype := "mod"
	if isModpack {
		ptype = "modpack"
	}

	u, _ := url.Parse("https://api.modrinth.com/v2/search")
	q := u.Query()
	q.Set("query", query)

	// Modrinth facets: [["facet:value"], ["facet:value"]] for AND logic
	var facetList []string
	if loader != "vanilla" && loader != "" {
		facetList = append(facetList, fmt.Sprintf(`["categories:%s"]`, loader))
	}
	facetList = append(facetList, fmt.Sprintf(`["versions:%s"]`, mcVersion))
	facetList = append(facetList, fmt.Sprintf(`["project_type:%s"]`, ptype))

	// Sides filter if provided (e.g. "client", "server")
	if len(sides) > 0 {
		var sideFacets []string
		for _, s := range sides {
			sideFacets = append(sideFacets, fmt.Sprintf(`"categories:%s"`, s))
		}
		facetList = append(facetList, fmt.Sprintf("[%s]", strings.Join(sideFacets, ",")))
	}

	q.Set("facets", fmt.Sprintf("[%s]", strings.Join(facetList, ",")))

	// Sorting
	if sort != "" {
		q.Set("index", sort) // relevance, downloads, follows, newest, updated
	} else if query == "" {
		q.Set("index", "downloads")
	} else {
		q.Set("index", "relevance")
	}

	q.Set("offset", fmt.Sprintf("%d", offset))
	q.Set("limit", "20")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth error (%d): %s", resp.StatusCode, string(body))
	}

	var result ModSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var hits []interface{}
	for _, h := range result.Hits {
		hits = append(hits, h)
	}
	return hits, nil
}

// InstallMod downloads the latest compatible version of a mod
func (inst *Instance) InstallMod(projectId string) error {
	ver, err := inst.getCompatibleVersion(projectId)
	if err != nil {
		return err
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

	modsDir := filepath.Join(inst.Directory, "mods")
	os.MkdirAll(modsDir, 0755)

	targetPath := filepath.Join(modsDir, fileName)
	inst.Manager.Broadcast(fmt.Sprintf("Downloading mod %s...", fileName))

	return inst.downloadFile(targetPath, fileUrl)
}

// InstallModpack downloads a modpack and installs its contents
func (inst *Instance) InstallModpack(projectId string) error {
	ver, err := inst.getCompatibleVersion(projectId)
	if err != nil {
		return err
	}

	// 1. Reset Mods
	modsDir := filepath.Join(inst.Directory, "mods")
	inst.Manager.Broadcast("Resetting mods directory...")
	os.RemoveAll(modsDir)
	os.MkdirAll(modsDir, 0755)

	// 2. Download .mrpack
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

	// 3. Read Index
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

	// 4. Download files
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

	// 5. Overrides
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

func (inst *Instance) getCompatibleVersion(projectId string) (*ProjectVersion, error) {
	loader := inst.Type
	mcVersion := inst.Version

	u, _ := url.Parse(fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectId))
	q := u.Query()
	q.Set("game_versions", fmt.Sprintf("[\"%s\"]", mcVersion))

	if loader != "vanilla" && loader != "" {
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

func (inst *Instance) GetInstalledMods() ([]string, error) {
	modsDir := filepath.Join(inst.Directory, "mods")
	files, err := os.ReadDir(modsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

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

	if len(hashes) == 0 {
		return []string{}, nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"hashes":    hashes,
		"algorithm": "sha1",
	})

	resp, err := http.Post("https://api.modrinth.com/v2/version_files", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth hash query failed")
	}

	var result map[string]struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	installed := make(map[string]bool)
	var ids []string
	for _, v := range result {
		if !installed[v.ProjectID] {
			installed[v.ProjectID] = true
			ids = append(ids, v.ProjectID)
		}
	}

	return ids, nil
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
