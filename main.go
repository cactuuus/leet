package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cactuuus/leet/cmd"
	"github.com/cactuuus/leet/internal/auth"
	"github.com/cactuuus/leet/internal/config"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/scaffold"
)

// App implements the AppContext interface defined in cmd package.
// It provides access to the application's core components, lazily initialized.
type App struct {
	homeDir	 	string
	configPath 	string
	config      *config.Manager
	client      *leetcode.Client
	scaffolder  *scaffold.Scaffolder
}

func main() {
	// generate the default config path
	home, err := getHomeDir()
	if err != nil {
		panic(err)
	}
	// execute with the given the app context
	cmd.Execute(&App{
		homeDir: home,
		configPath: filepath.Join(home, ".config", "leet", "config.toml"),
	})
}

// Config initializes a configuration manager and returns it, caching it for future calls.
// If the configuration manager fails to initialize, it returns its default values and prints a
// warning message. This, instead of exiting the program, allows the application to continue with
// defaults, allowing to at least attempt to fix the issue via other commands.
func (a *App) Config() *config.Manager {
	// if the config is already loaded, return it
	if a.config != nil {
		return a.config
	}

	// get 'best guess' at editor, based on OS
	var fileCmd, folderCmd string
	if editor := os.Getenv("EDITOR"); editor != "" {
		fileCmd = editor
		folderCmd = editor
	} else {
		switch runtime.GOOS {
		case "windows":
			fileCmd = "notepad"
			folderCmd = "explorer"
		case "darwin":	// my best guesses, I have no idea about macs
			fileCmd = "open -a TextEdit"
			folderCmd = "open"
		default:
			fileCmd = "xdg-open"
			folderCmd = "xdg-open"
		}
	}

	// else load it then return it
	defaultCfg := config.ConfigData{
		Version:            -1, // this will be overwritten by the config manager
		ProblemsDir:        filepath.Join(a.homeDir, "leet-problems"),
		PreferredLanguages: []string{},
		OpenFilesWith:      fileCmd,
		OpenFoldersWith:    folderCmd,
		Credentials:        auth.Credentials{},
		BaseURL:	        "https://leetcode.com/",
		TemplatesDir:       filepath.Join(a.homeDir, ".config", "leet", "templates"),
		CachePath:          filepath.Join(a.homeDir, ".cache", "leet", "problems.json"),
	}
	cfg := config.NewManager(a.configPath, defaultCfg)
	if err := cfg.LoadFromFile(); err != nil {
		fmt.Printf(
			"Warning: failed to load config file, using defaults instead." +
			"If the error persists, check your config file or run `leet config reset` to reset it. " +
			"(reason: %v)\n", err)
	}
	a.config = cfg
	return cfg
}

// Client initializes a LeetCode client and returns it, caching it for future calls.
// If the client fails to initialize, it prints an error message and exits the program.
func (a *App) Client() (*leetcode.Client, error) {
	// if the client is already initialized, return it
	if a.client != nil {
		return a.client, nil
	}
	// else initialize it then return it
	cfg := a.Config()
	client, err := leetcode.NewClient(cfg.CachePath, cfg.BaseURL, http.DefaultClient, cfg.Credentials)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize LeetCode client:\n%w", err)
	}
	a.client = client
	return client, nil
}

// Scaffolder initializes a scaffolder instance and returns it, caching it for future calls.
// If the scaffolder fails to initialize, it prints an error message and exits the program.
func (a *App) Scaffolder() (*scaffold.Scaffolder, error) {
	// if the scaffolder is already initialized, return it
	if a.scaffolder != nil {
		return a.scaffolder, nil
	}
	// else initialize it then return it
	s, err := scaffold.NewScaffolder(a.Config().ProblemsDir, a.Config().TemplatesDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize file scaffolder:\n%w", err)
	}
	a.scaffolder = s
	return s, nil
}

// getHomeDir returns the user's home directory, or an error if it cannot be determined.
func getHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}
