package instances

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"jjmc/manager"
)

type VersionsManager struct {
	manager *manager.Manager
}

func NewVersionsManager(m *manager.Manager) *VersionsManager {
	return &VersionsManager{manager: m}
}

type WriteCounter struct {
	Total   uint64
	Current uint64
	Manager *manager.Manager
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Current += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	percent := float64(wc.Current) / float64(wc.Total) * 100
	msg := fmt.Sprintf("Downloading... %.2f%%", percent)
	wc.Manager.Broadcast(msg)
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
