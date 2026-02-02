package instances

import (
	"jjmc/files"
	"mime/multipart"
	"os"
)

func (i *Instance) ListFiles(relPath string) ([]files.FileInfo, error) {
	return files.List(i.Directory, relPath)
}

func (i *Instance) ReadFile(relPath string) ([]byte, error) {
	return files.Read(i.Directory, relPath)
}

func (i *Instance) ReadFileStream(relPath string) (*os.File, error) {
	return files.GetStream(i.Directory, relPath)
}

func (i *Instance) WriteFile(relPath string, data []byte) error {
	return files.Write(i.Directory, relPath, data)
}

func (i *Instance) DeleteFile(relPath string) error {
	return files.Delete(i.Directory, relPath)
}

func (i *Instance) Mkdir(relPath string) error {
	return files.Mkdir(i.Directory, relPath)
}

func (i *Instance) HandleUpload(relPath string, file *multipart.FileHeader) error {
	return files.HandleUpload(i.Directory, relPath, file)
}

func (i *Instance) CompressFiles(relPaths []string, destRelPath string) error {
	return files.Compress(i.Directory, relPaths, destRelPath)
}

func (i *Instance) DecompressFile(zipRelPath string, destRelPath string) error {
	return files.Decompress(i.Directory, zipRelPath, destRelPath)
}
