package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
)

//go:embed config.template.toml
var configTemplate string

// Config represents the configuration for the leet tool.
type Config struct {
	ProblemsDir        string   `toml:"problems_dir"`
	PreferredLanguages []string `toml:"preferred_languages"`
	Editor             string   `toml:"editor_cmd"`
}

// Load loads the configuration from the standard config file location.
// If the config file does not exist, it will be created with default values.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	// If the config file does not exist, create it with default values.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := writeDefault(path); err != nil {
			return Config{}, err
		}
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config file: %w", err)
	}
	return cfg, nil
}

// Reset resets the configuration file to default values. This simply deletes the config file and creates a new one with default values.
func Reset() error {
	path, err := Path()
	if err != nil {
		return err
	}
	return writeDefault(path)
}

// Path returns the standard location of the config file: ~/.config/leet/config.toml
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "leet", "config.toml"), nil
}


// String returns a human-readable summary of the configuration.
func (c Config) String() string {
	return fmt.Sprintf(
		"# Leet Configuration #\n" +
		"Problems directory.: %s\n" +
		"Preferred languages: %v\n" +
		"Editor command.....: %s",
		c.ProblemsDir, c.PreferredLanguages, c.Editor,
	)
}


// writeDefault writes the default configuration to the given path, creating any necessary directories.
func writeDefault(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"Home": home}); err != nil {
		return fmt.Errorf("failed to render config template: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// LoadDefault returns the default configuration without writing it to disk.
func LoadDefault() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"Home": home}); err != nil {
		return Config{}, fmt.Errorf("failed to render config template: %w", err)
	}
	var cfg Config
	if _, err := toml.Decode(buf.String(), &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode default config: %w", err)
	}
	return cfg, nil
}
