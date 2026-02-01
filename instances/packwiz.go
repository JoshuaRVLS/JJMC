package instances

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"jjmc/pkg/downloader"

	"github.com/pelletier/go-toml/v2"
)

type PackwizPack struct {
	Index struct {
		File       string `toml:"file"`
		Hash       string `toml:"hash"`
		HashFormat string `toml:"hash-format"`
	} `toml:"index"`
	Name string `toml:"name"`
}

type PackwizIndex struct {
	HashFormat string        `toml:"hash-format"`
	Files      []PackwizFile `toml:"files"`
}

type PackwizFile struct {
	File     string `toml:"file"`
	Hash     string `toml:"hash"`
	MetaFile bool   `toml:"metafile"`
	Download struct {
		Url        string `toml:"url"`
		Hash       string `toml:"hash"`
		HashFormat string `toml:"hash-format"`
	} `toml:"download"`
}

func (inst *Instance) InstallPackwiz(packUrl string) error {
	dl := downloader.New()
	workDir := inst.Directory

	inst.Manager.Broadcast("Downloading pack.toml...")

	packTomlPath := filepath.Join(workDir, "pack.toml")
	err := dl.DownloadFile(downloader.DownloadOptions{
		Url:      packUrl,
		DestPath: packTomlPath,
		Force:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to download pack.toml: %v", err)
	}

	var pack PackwizPack
	if err := parseToml(packTomlPath, &pack); err != nil {
		return err
	}

	inst.Manager.Broadcast(fmt.Sprintf("Installing Packwiz Pack: %s", pack.Name))

	indexUrl, err := resolveRelativeUrl(packUrl, pack.Index.File)
	if err != nil {
		return err
	}

	indexTomlPath := filepath.Join(workDir, pack.Index.File)
	inst.Manager.Broadcast("Downloading index...")
	err = dl.DownloadFile(downloader.DownloadOptions{
		Url:      indexUrl,
		DestPath: indexTomlPath,
		Hash:     pack.Index.Hash,
		HashAlgo: pack.Index.HashFormat,
		Force:    true,
	})
	if err != nil {
		return fmt.Errorf("failed to download index: %v", err)
	}

	var index PackwizIndex
	if err := parseToml(indexTomlPath, &index); err != nil {
		return err
	}

	total := len(index.Files)
	for i, f := range index.Files {

		destPath := filepath.Join(workDir, f.File)

		inst.Manager.Broadcast(fmt.Sprintf("Downloading %d/%d: %s", i+1, total, f.File))

		if f.MetaFile || strings.HasSuffix(f.File, ".toml") {

			metaUrl, _ := resolveRelativeUrl(packUrl, f.File)
			err := dl.DownloadFile(downloader.DownloadOptions{
				Url:      metaUrl,
				DestPath: destPath,
				Hash:     f.Hash,
				HashAlgo: index.HashFormat,
				Force:    true,
			})
			if err != nil {
				return err
			}

			var modFile struct {
				Filename string `toml:"filename"`
				Side     string `toml:"side"`
				Download struct {
					Url        string `toml:"url"`
					Hash       string `toml:"hash"`
					HashFormat string `toml:"hash-format"`
				} `toml:"download"`
			}
			if err := parseToml(destPath, &modFile); err != nil {
				return err
			}

			jarPath := filepath.Join(filepath.Dir(destPath), modFile.Filename)

			err = dl.DownloadFile(downloader.DownloadOptions{
				Url:      modFile.Download.Url,
				DestPath: jarPath,
				Hash:     modFile.Download.Hash,
				HashAlgo: modFile.Download.HashFormat,
			})
			if err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Failed to download mod %s: %v", modFile.Filename, err))
			}

		} else {

			fileUrl := f.Download.Url
			if fileUrl == "" {
				fileUrl, _ = resolveRelativeUrl(packUrl, f.File)
			}

			err := dl.DownloadFile(downloader.DownloadOptions{
				Url:      fileUrl,
				DestPath: destPath,
				Hash:     f.Hash,
				HashAlgo: index.HashFormat,
			})
			if err != nil {
				inst.Manager.Broadcast(fmt.Sprintf("Failed to download config %s: %v", f.File, err))
			}
		}
	}

	inst.Manager.Broadcast("Packwiz installation complete.")
	return nil
}

func parseToml(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, v)
}

func resolveRelativeUrl(base, rel string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(rel)
	if err != nil {
		return "", err
	}
	return u.ResolveReference(r).String(), nil
}
