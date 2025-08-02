package main

import (
	// The "cmd" package contains the root command and all subcommands for the CLI application.
	"github.com/kodelint/wiper/cmd"
)

// main is the entry point for the wiper CLI tool.
func main() {
	// Execute is the primary function of the CLI tool's command package.
	// It parses command-line arguments, flags, and executes the appropriate command.
	cmd.Execute()
}
