package instances

import (
	"fmt"
)

// Source represents a provider of server software (e.g. Paper, Spigot, Vanilla, Modrinth)
type Source interface {
	// Search queries the source for available versions or projects
	Search(query string) ([]SourceResult, error)

	// GetVersion resolves a specific version ID to a downloadable artifact
	GetVersion(versionId string) (*SourceVersion, error)

	// GetLatestVersion returns the latest stable version
	GetLatestVersion() (string, error)
}

// SourceResult is a summary of a search result
type SourceResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // e.g. "mod", "modpack", "server"
}

// SourceVersion represents a concrete version ready for download
type SourceVersion struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`    // Display name (e.g. "1.20.4-Build-123")
	Version string       `json:"version"` // Minecraft version (e.g. "1.20.4")
	Files   []SourceFile `json:"files"`
}

type SourceFile struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
	HashAlgo string `json:"hash_algo"` // sha1, sha256
	Primary  bool   `json:"primary"`
}

// SourceRegistry holds available sources
type SourceRegistry struct {
	sources map[string]Source
}

var Registry = &SourceRegistry{
	sources: make(map[string]Source),
}

func (r *SourceRegistry) Register(id string, s Source) {
	r.sources[id] = s
}

func (r *SourceRegistry) Get(id string) (Source, error) {
	s, ok := r.sources[id]
	if !ok {
		return nil, fmt.Errorf("source not found: %s", id)
	}
	return s, nil
}
