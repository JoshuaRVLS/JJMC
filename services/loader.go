package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jjmc/models"
)

type TemplateManager struct {
	TemplatesDir string
	Templates    map[string]models.Template
}

func NewTemplateManager(dir string) *TemplateManager {
	return &TemplateManager{
		TemplatesDir: dir,
		Templates:    make(map[string]models.Template),
	}
}

func (tm *TemplateManager) LoadTemplates() error {
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
			fmt.Printf("Failed to read template %s: %v\n", entry.Name(), err)
			continue
		}

		var tmpl models.Template
		if err := json.Unmarshal(data, &tmpl); err != nil {
			fmt.Printf("Failed to parse template %s: %v\n", entry.Name(), err)
			continue
		}

		// ID defaults to filename info if missing?
		if tmpl.ID == "" {
			tmpl.ID = strings.TrimSuffix(entry.Name(), ".json")
		}

		tm.Templates[tmpl.ID] = tmpl
	}
	return nil
}

func (tm *TemplateManager) GetTemplate(id string) (models.Template, bool) {
	t, ok := tm.Templates[id]
	return t, ok
}

func (tm *TemplateManager) ListTemplates() []models.Template {
	list := make([]models.Template, 0, len(tm.Templates))
	for _, t := range tm.Templates {
		list = append(list, t)
	}
	return list
}
