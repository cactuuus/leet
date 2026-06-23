package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	// loadFrom should create the file with default values if it doesn't exist
	cfg, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom() error = %v", err)
	}

	// check that the default values are set correctly
	// these should match the defaults defined in config.template.toml
	home, _ := os.UserHomeDir()
	wantDir := filepath.Join(home, "leet-problems")
	if cfg.ProblemsDir != wantDir {
		t.Errorf("ProblemsDir = %q, want %q", cfg.ProblemsDir, wantDir)
	}
	if len(cfg.PreferredLanguages) != 0 {
		t.Errorf("PreferredLanguages = %v, want empty", cfg.PreferredLanguages)
	}
	if cfg.Editor != "" {
		t.Errorf("Editor = %q, want empty", cfg.Editor)
	}

	// file should now exist on disk
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config file to be created, got error: %v", err)
	}
}

func TestLoadFrom_ValidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	// create a valid config file with custom values
	problemsDir := "/tmp/problems"
	languages := []string{"golang", "python3"}
	editor := "nvim"
	content := fmt.Sprintf(`
problems_dir = "%s"
preferred_languages = ["%s", "%s"]
editor_cmd = "%s"
`, problemsDir, languages[0], languages[1], editor)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// loadFrom should read the values correctly
	cfg, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom() error = %v", err)
	}
	if cfg.ProblemsDir != problemsDir {
		t.Errorf("ProblemsDir = %q, want %q", cfg.ProblemsDir, problemsDir)
	}
	if len(cfg.PreferredLanguages) != 2 ||
			cfg.PreferredLanguages[0] != languages[0] ||
			cfg.PreferredLanguages[1] != languages[1] {
		t.Errorf("PreferredLanguages = %v, want %v", cfg.PreferredLanguages, languages)
	}
	if cfg.Editor != editor {
		t.Errorf("Editor = %q, want %q", cfg.Editor, editor)
	}
}

func TestLoadFrom_MalformedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("this is not valid toml ]["), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := loadFrom(path)
	if err == nil {
		t.Fatal("loadFrom() expected error for malformed file, got nil")
	}
}

func TestResetTo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	// write a bad config first
	if err := os.WriteFile(path, []byte("this is not valid toml ]["), 0644); err != nil {
		t.Fatalf("failed to write bad config: %v", err)
	}

	if err := resetAt(path); err != nil {
		t.Fatalf("resetAt() error = %v", err)
	}

	// should now load cleanly with defaults
	cfg, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom() after reset error = %v", err)
	}

	home, _ := os.UserHomeDir()
	wantDir := filepath.Join(home, "leet-problems")
	if cfg.ProblemsDir != wantDir {
		t.Errorf("ProblemsDir = %q, want %q", cfg.ProblemsDir, wantDir)
	}
}
