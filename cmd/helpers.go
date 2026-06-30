package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cactuuus/leet/internal/language"
	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/problem"
)

// cmdBoolFlags defines a struct to hold information about boolean flags for commands.
// Simple helper for commands with lots of boolean flags.
type cmdBoolFlags struct {
	Long 	string
	Short 	string
	Value 	bool
	Desc  	string
}

// promptYesNo prompts the user with a yes/no question, returning true for yes.
// Users can press Ctrl+C to abort entirely.
func promptYesNo(msg string) (bool, error) {
	fmt.Printf("%s [y/n] ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, fmt.Errorf("Failed to read input:\n%w", err)
			}
			return false, fmt.Errorf("No input received.")
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
	printActionStart("Opening in editor")
	// if still empty, return an error
	if editorCmd == "" {
		return fmt.Errorf("No editor configured, set 'editor' in your config file.")
	}
	c := exec.Command(editorCmd, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("Failed to open %s in editor:\n%w", path, err)
	}
	printActionSuccess()
	return nil
}

// fetchPreviewByIdentifier fetches a problem preview given either "daily" or a numeric string.
func fetchPreviewByIdentifier(c *leetcode.Client, id string) (problem.Preview, error) {
	if id == "daily" {
		printActionStart("Fetching daily problem")
		full, err := c.GetDailyProblem()
		if err != nil {
			return problem.Preview{}, fmt.Errorf("Failed to fetch daily problem:\n%w", err)
		}
		printActionSuccess()
		return full.Preview, nil
	} else {
		// else try to parse the argument as a number
		num, err := strconv.Atoi(id)
		if err != nil {
			return problem.Preview{}, fmt.Errorf("Invalid problem identifier %q", id)
		}
		printActionStart(fmt.Sprintf("Fetching problem #%d", num))
		// get the problem slug
		preview, err := c.GetProblemPreview(num)
		if err != nil {
			return problem.Preview{}, fmt.Errorf("Could not find problem %d:\n%w", num, err)
		}
		printActionSuccess()
		return preview, nil
	}
}

// fetchFullByIdentifier fetches a problem (full) given either "daily" or a numeric string.
func fetchFullByIdentifier(c *leetcode.Client, id string) (problem.Full, error) {
	if id == "daily" {
		printActionStart("Fetching daily problem")
		p, err := c.GetDailyProblem()
		if err != nil {
			return problem.Full{}, fmt.Errorf("Failed to fetch daily problem:\n%w", err)
		}
		printActionSuccess()
		return p, nil
	} else {
		// else try to parse the argument as a number
		num, err := strconv.Atoi(id)
		if err != nil {
			return problem.Full{}, fmt.Errorf("Invalid problem identifier %q", id)
		}
		printActionStart(fmt.Sprintf("Fetching problem #%d", num))
		p, err := c.GetProblemFull(num)
		if err != nil {
			return problem.Full{}, fmt.Errorf("Could not find problem %d:\n%w", num, err)
		}
		printActionSuccess()
		return p, nil
	}
}

// printActionStart prints a message indicating the start of an action.
func printActionStart(action string) {
	fmt.Printf("%s...", action)
}

// printActionSuccess prints a checkmark indicating the successful completion of an action.
func printActionSuccess() {
	fmt.Println(" ✓")
}

// parseLanguage parses a language identifier and returns the corresponding Language object, else
// returns an error.
func parseLanguage(id string) (language.Language, error) {
	lang, ok := language.Get(id)
	if !ok {
		return language.Language{},
			fmt.Errorf(
				"Unknown language: %q. Run 'leet languages' to see supported languages",
				id)
	}
	return lang, nil
}
