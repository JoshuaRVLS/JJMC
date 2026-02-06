package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"jjmc/internal/models"
	"jjmc/pkg/logger"
)

type TemplateManager struct {
	TemplatesDir string
	Templates    map[string]models.Template
	mu           sync.RWMutex
}

func NewTemplateManager(dir string) *TemplateManager {
	return &TemplateManager{
		TemplatesDir: dir,
		Templates:    make(map[string]models.Template),
	}
}

func (tm *TemplateManager) LoadTemplates() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	entries, err := os.ReadDir(tm.TemplatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(tm.TemplatesDir, 0755)
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(tm.TemplatesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Error("Failed to read template", "file", entry.Name(), "error", err)
			continue
		}

		var tmpl models.Template
		if err := json.Unmarshal(data, &tmpl); err != nil {
			logger.Error("Failed to parse template", "file", entry.Name(), "error", err)
			continue
		}

		if tmpl.ID == "" {
			tmpl.ID = strings.TrimSuffix(entry.Name(), ".json")
		}

		tm.Templates[tmpl.ID] = tmpl
	}
	return nil
}

func (tm *TemplateManager) GetTemplate(id string) (models.Template, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.Templates[id]
	return t, ok
}

func (tm *TemplateManager) ListTemplates() []models.Template {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	list := make([]models.Template, 0, len(tm.Templates))
	for _, t := range tm.Templates {
		list = append(list, t)
	}
	return list
}
