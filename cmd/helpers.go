package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cactuuus/leet/internal/leetcode"
	"github.com/cactuuus/leet/internal/problem"
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

// fetchPreviewByIdentifier fetches a problem preview given either "daily" or a numeric string.
func fetchPreviewByIdentifier(c *leetcode.Client, id string) (problem.Preview, error) {
	if id == "daily" {
		p, err := c.GetDailyProblem()
		if err != nil {
			return problem.Preview{}, fmt.Errorf("failed to fetch daily problem: %w", err)
		}
		return p.Preview, nil
	}
	// else try to parse the argument as a number
	num, err := strconv.Atoi(id)
	if err != nil {
		return problem.Preview{}, fmt.Errorf("invalid problem number: %s", id)
	}
	// get the problem slug
	preview, err := c.GetProblemPreview(num)
	if err != nil {
		return problem.Preview{}, fmt.Errorf("could not find problem %d: %w", num, err)
	}
	return preview, nil
}

// fetchFullByIdentifier fetches a problem (full) given either "daily" or a numeric string.
func fetchFullByIdentifier(c *leetcode.Client, id string) (problem.Full, error) {
	if id == "daily" {
		return c.GetDailyProblem()
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return problem.Full{}, fmt.Errorf("invalid problem identifier: %q — use a problem number or 'daily'", id)
	}
	return c.GetProblemFull(num)
}
