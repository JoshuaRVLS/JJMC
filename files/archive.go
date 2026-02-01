package files

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(rootDir string, relPaths []string, destRelPath string) error {

	cleanDest := filepath.Clean(destRelPath)
	if strings.Contains(cleanDest, "..") {
		return fmt.Errorf("invalid destination path")
	}
	destPath := filepath.Join(rootDir, cleanDest)

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	for _, relPath := range relPaths {
		cleanRel := filepath.Clean(relPath)
		if strings.Contains(cleanRel, "..") {
			continue
		}
		fullPath := filepath.Join(rootDir, cleanRel)

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		walker := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relInInstance, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Name = relInInstance
			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := w.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		}

		if info.IsDir() {
			filepath.Walk(fullPath, walker)
		} else {
			walker(fullPath, info, nil)
		}
	}

	return nil
}

func Decompress(rootDir string, zipRelPath string, destRelPath string) error {
	cleanZip := filepath.Clean(zipRelPath)
	if strings.Contains(cleanZip, "..") {
		return fmt.Errorf("invalid zip path")
	}
	zipPath := filepath.Join(rootDir, cleanZip)

	cleanDest := filepath.Clean(destRelPath)
	if strings.Contains(cleanDest, "..") {
		return fmt.Errorf("invalid destination path")
	}
	destDir := filepath.Join(rootDir, cleanDest)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {

			continue
		}

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
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
