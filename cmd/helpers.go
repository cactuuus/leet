package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// promptYesNo prompts the user with a yes/no question, returning true for yes.
// Users can press Ctrl+C to abort entirely.
func promptYesNo(msg string) (bool, error) {
	fmt.Printf("%s [y/n] ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, fmt.Errorf("failed to read input: %w", err)
			}
			return false, fmt.Errorf("no input received")
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Print("Please enter 'y' or 'n'")
		}
	}
}

// openInEditor opens path with the given editor command.
// It returns an error if the editor command is empty or if the command fails to run.
func openInEditor(editorCmd string, path string) error {
	// if still empty, return an error
	if editorCmd == "" {
		return fmt.Errorf("no editor configured — set 'editor' in your config file")
	}
	c := exec.Command(editorCmd, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}
