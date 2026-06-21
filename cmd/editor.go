package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cactuuus/leet/internal/config"
)

// resolveEditor returns the editor command to use: the one configured in config.toml, falling
// back to $EDITOR if unset.
func resolveEditor(cfg config.Config) (string, error) {
	if cfg.Editor != "" {
		return cfg.Editor, nil
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, nil
	}
	return "", fmt.Errorf("no editor configured — set 'editor' in your config file, or set $EDITOR")
}

// openInEditor opens path in the configured editor, waiting for it to close with stdio wired to
// the terminal.
func openInEditor(cfg config.Config, path string) error {
	editor, err := resolveEditor(cfg)
	if err != nil {
		return err
	}
	c := exec.Command(editor, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}
