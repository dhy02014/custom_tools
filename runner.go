package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

// runResult holds the outcome of a terraform execution.
type runResult struct {
	ExitCode int
	Stderr   string
}

// runTerraform executes terraform with the given arguments.
// It streams stdout/stderr to the terminal in real-time while capturing stderr.
// stdin is connected directly for interactive prompts (e.g., apply yes/no).
func runTerraform(args []string) runResult {
	cmd := exec.Command("terraform", args...)

	// stdin passthrough for interactive prompts
	cmd.Stdin = os.Stdin

	// stdout: stream directly to terminal
	cmd.Stdout = os.Stdout

	// stderr: stream to terminal + capture in buffer
	var stderrBuf bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Command not found or other OS-level error
			exitCode = 1
		}
	}

	return runResult{
		ExitCode: exitCode,
		Stderr:   lastNLines(stderrBuf.String(), 10),
	}
}

// lastNLines returns the last n non-empty lines of s.
func lastNLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
