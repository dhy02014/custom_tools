package main

import (
	"fmt"
	"os"
	"time"
)

// notifyCommands are terraform subcommands that trigger Telegram notifications.
var notifyCommands = map[string]bool{
	"plan":     true,
	"apply":    true,
	"init":     true,
	"validate": true,
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// No args: just run terraform with no args (shows help)
		result := runTerraform(args)
		os.Exit(result.ExitCode)
	}

	// Detect the terraform subcommand (first arg that doesn't start with '-')
	subcommand := ""
	for _, a := range args {
		if a[0] != '-' {
			subcommand = a
			break
		}
	}

	// Run terraform
	start := time.Now()
	result := runTerraform(args)
	elapsed := time.Since(start)

	// Send notification if this is a notify-worthy command
	if notifyCommands[subcommand] {
		cfg := loadConfig()
		if cfg.isValid() {
			workDir := extractWorkDir(args)
			err := sendNotification(cfg, notifyPayload{
				Subcommand: subcommand,
				Args:       args,
				WorkDir:    workDir,
				ExitCode:   result.ExitCode,
				Duration:   elapsed,
				Stderr:     result.Stderr,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "[tfn] notification failed: %v\n", err)
			}
		}
	}

	os.Exit(result.ExitCode)
}

// extractWorkDir finds the directory name from -chdir=<path> argument.
func extractWorkDir(args []string) string {
	for _, a := range args {
		if len(a) > 7 && a[:7] == "-chdir=" {
			path := a[7:]
			// Return just the last component of the path
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' {
					return path[i+1:]
				}
			}
			return path
		}
	}

	// Fallback: use current directory name
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	for i := len(dir) - 1; i >= 0; i-- {
		if dir[i] == '/' {
			return dir[i+1:]
		}
	}
	return dir
}
