package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateManager_LoadTemplates(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmplContent := `{"id": "test-template", "name": "Test Template"}`
	err = os.WriteFile(filepath.Join(tmpDir, "test.json"), []byte(tmplContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tm := NewTemplateManager(tmpDir)
	if err := tm.LoadTemplates(); err != nil {
		t.Fatalf("LoadTemplates failed: %v", err)
	}

	tmpl, ok := tm.GetTemplate("test-template")
	if !ok {
		t.Errorf("Expected template 'test-template' to be found")
	}
	if tmpl.Name != "Test Template" {
		t.Errorf("Expected template name 'Test Template', got '%s'", tmpl.Name)
	}
}

func TestTemplateManager_ListTemplates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"t1.json": `{"id": "t1", "name": "T1"}`,
		"t2.json": `{"id": "t2", "name": "T2"}`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", name, err)
		}
	}

	tm := NewTemplateManager(tmpDir)
	tm.LoadTemplates()

	list := tm.ListTemplates()
	if len(list) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(list))
	}
}

func TestTemplateManager_ThreadSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "t1.json"), []byte(`{"id": "t1", "name": "T1"}`), 0644)

	tm := NewTemplateManager(tmpDir)
	tm.LoadTemplates()

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			tm.ListTemplates()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {

			tm.LoadTemplates()
		}
		done <- true
	}()

	<-done
	<-done
}
