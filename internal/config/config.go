package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/cactuuus/leet/internal/auth"
)

//go:embed config.template.toml
var configTemplate string

// Increment this when breaking changes are introduced to force an automatic file layout migration.
const configVersion = 1

// ConfigData represents the runtime application configuration.
type ConfigData struct {
	Version            int              `toml:"version"`
	ProblemsDir        string           `toml:"problems_dir"`
	PreferredLanguages []string         `toml:"preferred_languages"`
	Editor             string           `toml:"editor_cmd"`
	Credentials        auth.Credentials `toml:"credentials"`

	// Internal configurations not exported directly via template placeholders
	CachePath string `toml:"-"`
	BaseURL   string `toml:"-"`
}

// Manager orchestrates the lifecycle, state mutations, and disk storage of ConfigData.
type Manager struct {
	Path        string
	defaultData ConfigData
	ConfigData
}

// NewManager builds a new configuration manager.
// It is initialized with the provided default configuration values. To properly load existing
// configuration from disk, call LoadFromFile() after creating the manager.
func NewManager(path string, defaultData ConfigData) *Manager {
	// Version is hardcoded on the baseline defaults
	defaultData.Version = configVersion
	m := &Manager{Path: path, defaultData: defaultData, ConfigData: defaultData}
	return m
}

// LoadFromFile initializes the configuration manager by loading existing data from disk.
// If the file doesn't exist, it creates a new one with default values.
// If the version is outdated, it updates the file to the latest version, keeping existing values
// where possible.
func (m *Manager) LoadFromFile() error {
	// File doesn't exist yet -> write defaults using template
	if _, err := os.Stat(m.Path); os.IsNotExist(err) {
		m.ConfigData = m.defaultData
		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to save initial default config: %w", err)
		}
		return nil
	}
	// File exists -> parse it
	if _, err := toml.DecodeFile(m.Path, &m.ConfigData); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}
	// Self-healing schema check -> Outdated version found
	if m.ConfigData.Version != configVersion {
		m.ConfigData.Version = configVersion
		if err := m.Save(); err != nil {
			return fmt.Errorf("failed to automatically update config layout version: %w", err)
		}
	}
	return nil
}

// Save executes your text/template writer to persist changes while retaining all comments.
func (m *Manager) Save() error {
	// Ensure the directory exists before writing the file.
	if err := os.MkdirAll(filepath.Dir(m.Path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	// Convert a slice of strings into a TOML array representation.
	funcMap := template.FuncMap{
		"tomlArray": func(s []string) string {
			if len(s) == 0 {
				return "[]"
			}
			quoted := make([]string, len(s))
			for i, v := range s {
				quoted[i] = fmt.Sprintf("%q", v)
			}
			return "[" + strings.Join(quoted, ", ") + "]"
		},
	}
	tmpl, err := template.New("config").Funcs(funcMap).Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}
	// Write it to disk.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m.ConfigData); err != nil {
		return fmt.Errorf("failed to render config template: %w", err)
	}
	return os.WriteFile(m.Path, buf.Bytes(), 0600)
}

// Reset clears out custom data and forces a fallback to default values.
func (m *Manager) Reset() error {
	m.ConfigData = m.defaultData
	return m.Save()
}

// String returns a safe summary of the config, redacting credentials.
func (m *Manager) String() string {
	credStatus := "not set"
	if m.ConfigData.Credentials.IsSet() {
		credStatus = "set"
	}
	return fmt.Sprintf(
		"# Leet Configuration #\n"+
		"\tProblems directory.: %s\n"+
		"\tPreferred languages: %v\n"+
		"\tEditor command.....: %s\n"+
		"\tCredentials........: %s",
		m.ConfigData.ProblemsDir, m.ConfigData.PreferredLanguages, m.ConfigData.Editor, credStatus,
	)
}
