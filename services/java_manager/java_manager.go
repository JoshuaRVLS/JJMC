package java_manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type JavaRuntime struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Path     string `json:"path"`
	Status   string `json:"status,omitempty"`
	Progress int    `json:"progress,omitempty"`
}

type JavaManager struct {
	RuntimesDir string
	installing  map[int]*JavaRuntime
	mu          sync.Mutex
}

func NewJavaManager(runtimesDir string) *JavaManager {
	absDir, err := filepath.Abs(runtimesDir)
	if err == nil {
		runtimesDir = absDir
	}
	if _, err := os.Stat(runtimesDir); os.IsNotExist(err) {
		os.MkdirAll(runtimesDir, 0755)
	}
	return &JavaManager{
		RuntimesDir: runtimesDir,
		installing:  make(map[int]*JavaRuntime),
	}
}

func (m *JavaManager) ListInstalled() ([]JavaRuntime, error) {
	entries, err := os.ReadDir(m.RuntimesDir)
	if err != nil {
		return nil, err
	}

	runtimes := make([]JavaRuntime, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			// Basic parsing: java-<version>-<impl>
			// e.g., java-17-hotspot
			parts := strings.Split(entry.Name(), "-")
			version := 0
			if len(parts) >= 2 {
				fmt.Sscanf(parts[1], "%d", &version)
			}

			runtimes = append(runtimes, JavaRuntime{
				Name:    entry.Name(),
				Version: version,
				Path:    filepath.Join(m.RuntimesDir, entry.Name()),
			})
		}
	}
	return runtimes, nil
}

type AdoptiumRelease struct {
	Binaries []struct {
		Package struct {
			Link string `json:"link"`
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"package"`
	} `json:"binaries"`
	VersionData struct {
		Semver string `json:"semver"`
	} `json:"version_data"`
}

func (m *JavaManager) DownloadVersion(version int) error {
	fmt.Printf("[JavaManager] Starting download for version %d\n", version)
	m.mu.Lock()
	if _, exists := m.installing[version]; exists {
		m.mu.Unlock()
		return fmt.Errorf("java %d is already installing", version)
	}

	progressRuntime := &JavaRuntime{
		Name:     fmt.Sprintf("java-%d-hotspot", version), // Predicted name
		Version:  version,
		Status:   "Downloading",
		Progress: 0,
	}
	m.installing[version] = progressRuntime
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.installing, version)
		m.mu.Unlock()
		fmt.Printf("[JavaManager] Removed version %d from installing map\n", version)
	}()

	// 1. Fetch release info from Adoptium
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map GOARCH to Adoptium architecture
	if arch == "amd64" {
		arch = "x64"
	}
	// Map GOOS to Adoptium os
	if osName == "darwin" {
		osName = "mac"
	}

	url := fmt.Sprintf("https://api.adoptium.net/v3/assets/feature_releases/%d/ga?architecture=%s&heap_size=normal&image_type=jdk&jvm_impl=hotspot&os=%s&page=0&page_size=1&project=jdk&sort_method=DEFAULT&sort_order=DESC&vendor=eclipse", version, arch, osName)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[JavaManager] Failed to fetch releases: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("[JavaManager] Failed to fetch releases: status code %d\n", resp.StatusCode)
		return fmt.Errorf("failed to fetch releases: %s", resp.Status)
	}

	var releases []AdoptiumRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		fmt.Printf("[JavaManager] Failed to decode releases: %v\n", err)
		return err
	}

	if len(releases) == 0 {
		fmt.Printf("[JavaManager] No releases found for Java %d\n", version)
		return fmt.Errorf("no releases found for Java %d", version)
	}

	if len(releases[0].Binaries) == 0 {
		fmt.Printf("[JavaManager] No binaries found for Java %d\n", version)
		return fmt.Errorf("no binaries found for Java %d", version)
	}

	downloadUrl := releases[0].Binaries[0].Package.Link
	fileName := releases[0].Binaries[0].Package.Name
	fileSize := releases[0].Binaries[0].Package.Size

	fmt.Printf("[JavaManager] Downloading %s (Size: %d)\n", fileName, fileSize)

	localPath := filepath.Join(m.RuntimesDir, fileName)

	// Update status
	m.mu.Lock()
	progressRuntime.Status = "Downloading"
	m.mu.Unlock()

	// 2. Download file
	out, err := os.Create(localPath)
	if err != nil {
		fmt.Printf("[JavaManager] Failed to create file: %v\n", err)
		return err
	}
	defer out.Close()

	if err := m.downloadFile(downloadUrl, out, fileSize, func(p int) {
		m.mu.Lock()
		if progressRuntime.Progress != p {
			progressRuntime.Progress = p
		}
		m.mu.Unlock()
	}); err != nil {
		fmt.Printf("[JavaManager] Download failed: %v\n", err)
		return err
	}

	// Update status
	m.mu.Lock()
	progressRuntime.Status = "Extracting"
	progressRuntime.Progress = 100 // Download done
	m.mu.Unlock()

	fmt.Printf("[JavaManager] Extracting %s...\n", fileName)

	// 3. Extract
	extractDir := filepath.Join(m.RuntimesDir, fmt.Sprintf("java-%d-hotspot", version))
	if strings.HasSuffix(fileName, ".zip") {
		if err := m.unzip(localPath, extractDir); err != nil {
			fmt.Printf("[JavaManager] Unzip failed: %v\n", err)
			return err
		}
	} else if strings.HasSuffix(fileName, ".tar.gz") {
		if err := m.untar(localPath, extractDir); err != nil {
			fmt.Printf("[JavaManager] Untar failed: %v\n", err)
			return err
		}
	}

	// Cleanup archive
	os.Remove(localPath)

	fmt.Printf("[JavaManager] Install complete for Java %d\n", version)

	return nil
}

func (m *JavaManager) DeleteVersion(name string) error {
	return os.RemoveAll(filepath.Join(m.RuntimesDir, name))
}

func (m *JavaManager) downloadFile(url string, out *os.File, totalSize int64, progressCb func(int)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	counter := &WriteCounter{
		Total:    totalSize,
		Callback: progressCb,
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	return nil
}

func (m *JavaManager) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	// Optimization: Move the inner directory up if it's the only one
	// (Adoptium usually extracts to a subdirectory like jdk-17.0.1+12)
	return m.flattenDir(dest)
}

func (m *JavaManager) untar(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fpath := filepath.Join(dest, header.Name)
		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			return err
		}
		outFile.Close()
	}
	return m.flattenDir(dest)
}

func (m *JavaManager) flattenDir(dest string) error {
	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}

	if len(entries) == 1 && entries[0].IsDir() {
		innerDir := filepath.Join(dest, entries[0].Name())
		tmpDir := dest + "_tmp"
		if err := os.Rename(innerDir, tmpDir); err != nil {
			return err
		}
		os.RemoveAll(dest)
		if err := os.Rename(tmpDir, dest); err != nil {
			return err
		}
	}
	return nil
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Total    int64
	Current  int64
	Callback func(int)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Current += int64(n)
	if wc.Total > 0 {
		percentage := int(float64(wc.Current) * 100 / float64(wc.Total))
		if wc.Callback != nil {
			wc.Callback(percentage)
		}
	} else {
		if wc.Callback != nil {
			wc.Callback(0)
		}
	}
	return n, nil
}
