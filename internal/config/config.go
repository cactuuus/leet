package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Paths		PathsConfig		`toml:"paths"`
	Languages	LanguagesConfig	`toml:"languages"`
	Editor		EditorConfig	`toml:"editor"`
}

type PathsConfig struct {
	Cache		string	`toml:"cache"`
	Problems 	string	`toml:"problems"`
}

type LanguagesConfig struct {
	Preferred []string `toml:"preferred"`
}

type EditorConfig struct {
	Command string `toml:"command"`
}

func LoadConfig() (Config, error) {
	// check if the config file exists, if not create it with default values
	configFilePath, err := configPath()
	if err != nil {
		return Config{}, err
	}
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		defaultConfig, err := DefaultConfig()
		if err != nil {
			return Config{}, err
		}
		err = UpdateConfig(defaultConfig)
		if err != nil {
			return Config{}, err
		}
	}
	var cfg Config
	if _, err := toml.DecodeFile(configFilePath, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func UpdateConfig(cfg Config) error {
	configFilePath, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
		return err
	}
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(cfg)
}

func (c Config) CachePath() string {
	return c.Paths.Cache
}

func (c Config) ProblemsPath() string {
	return c.Paths.Problems
}

func (c Config) OpenInEditor(path string) error {
	if c.Editor.Command == "" {
		return fmt.Errorf("no editor configured")
	}
	cmd := exec.Command(c.Editor.Command, path)
	return cmd.Start()
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "leet", "config.toml"), nil
}

func DefaultConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Paths: PathsConfig{
			Cache:    filepath.Join(home, ".cache", "leet", "problems.json"),
			Problems: filepath.Join(home, "leet-problems"),
		},
		Languages: LanguagesConfig{
			Preferred: []string{},
		},
		Editor: EditorConfig{
			Command: "code",
		},
	}, nil
}
