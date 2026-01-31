package downloader

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProgressCallback is a function called periodically during download
type ProgressCallback func(current, total int64, percent float64)

// Downloader handles file downloads
type Downloader struct {
	Client *http.Client
}

func New() *Downloader {
	return &Downloader{
		Client: &http.Client{
			Timeout: 0, // No timeout for large files? Or set reasonable one?
		},
	}
}

// DownloadOptions configures the download
type DownloadOptions struct {
	Url        string
	DestPath   string
	Hash       string // Expected hash
	HashAlgo   string // sha1, sha256, md5
	OnProgress ProgressCallback
	Force      bool // Overwrite even if exists and hash matches?
	UserAgent  string
}

// DownloadFile downloads a file with options
func (d *Downloader) DownloadFile(opts DownloadOptions) error {
	// 1. Check if file exists and verify hash if provided
	if !opts.Force && opts.Hash != "" && fileExists(opts.DestPath) {
		match, err := VerifyFile(opts.DestPath, opts.Hash, opts.HashAlgo)
		if err == nil && match {
			// File exists and is valid, skip download
			if opts.OnProgress != nil {
				// Report 100%
				info, _ := os.Stat(opts.DestPath)
				opts.OnProgress(info.Size(), info.Size(), 100.0)
			}
			return nil
		}
	}

	// 2. Prepare Request
	req, err := http.NewRequest("GET", opts.Url, nil)
	if err != nil {
		return err
	}

	ua := opts.UserAgent
	if ua == "" {
		ua = "JJMC/1.0"
	}
	req.Header.Set("User-Agent", ua)

	// 3. Execute Request
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, resp.Status)
	}

	// 4. Create Destination
	if err := os.MkdirAll(filepath.Dir(opts.DestPath), 0755); err != nil {
		return err
	}

	// Download to .tmp file first
	tmpPath := opts.DestPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 5. Copy with Progress
	var reader io.Reader = resp.Body
	if opts.OnProgress != nil {
		reader = &progressReader{
			Reader:     resp.Body,
			Total:      resp.ContentLength,
			OnProgress: opts.OnProgress,
		}
	}

	if _, err := io.Copy(file, reader); err != nil {
		return err
	}
	file.Close() // Close before rename/hash check

	// 6. Verify Hash of Downloaded File
	if opts.Hash != "" {
		match, err := VerifyFile(tmpPath, opts.Hash, opts.HashAlgo)
		if err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to verify hash: %v", err)
		}
		if !match {
			os.Remove(tmpPath)
			return fmt.Errorf("hash mismatch: expected %s (%s)", opts.Hash, opts.HashAlgo)
		}
	}

	// 7. Move to Final Destination
	if err := os.Rename(tmpPath, opts.DestPath); err != nil {
		// Try copy if rename fails (filesystems)
		// But usually tmp is in same dir, so rename should work.
		return err
	}

	return nil
}

// VerifyFile checks if a file's hash matches the expected value
func VerifyFile(path string, expectedHash string, algo string) (bool, error) {
	if expectedHash == "" {
		return true, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var h hash.Hash
	switch strings.ToLower(algo) {
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "md5":
		h = md5.New()
	default:
		// Default to sha1 if unknown? Or error?
		// Let's assume sha1 for legacy compatibility if unspecified, or error.
		return false, fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	return strings.EqualFold(actualHash, expectedHash), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type progressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress ProgressCallback
	LastUpdate time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	// Throttle updates to ~100ms
	if time.Since(pr.LastUpdate) > 100*time.Millisecond || err == io.EOF {
		percent := 0.0
		if pr.Total > 0 {
			percent = float64(pr.Current) / float64(pr.Total) * 100
		}
		pr.OnProgress(pr.Current, pr.Total, percent)
		pr.LastUpdate = time.Now()
	}

	return n, err
}
