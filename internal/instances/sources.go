package instances

import (
	"fmt"
)

type Source interface {
	Search(query string) ([]SourceResult, error)

	GetVersion(versionId string) (*SourceVersion, error)

	GetLatestVersion() (string, error)
}

type SourceResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type SourceVersion struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Files   []SourceFile `json:"files"`
}

type SourceFile struct {
	Url      string `json:"url"`
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
	HashAlgo string `json:"hash_algo"`
	Primary  bool   `json:"primary"`
}

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
