package cmd

import (
	"fmt"                    // The fmt package is used for formatted I/O, specifically to print the version string.
	"github.com/spf13/cobra" // The cobra library is used to define and manage the command-line interface.
)

// ====================================================================================================
// VERSION COMMAND DEFINITION
// ====================================================================================================

// versionCmd represents the version command.
// It's a subcommand of the RootCmd and is responsible for showing the tool's version.
var versionCmd = &cobra.Command{
	// Use: defines the command's name, which is how it's invoked from the command line (e.g., `wiper version`).
	Use: "version",
	// Short: provides a brief description of the command, shown in the help output.
	Short: "Show the wiper tool version",
	// Run: is the main function executed when the 'version' command is called.
	Run: func(cmd *cobra.Command, args []string) {
		// This line prints the current version string.
		// NOTE: Currently, the version is hardcoded. In a real-world application,
		// this would typically be set dynamically at build time using a build-time variable.
		fmt.Println("Wiper Version: 0.1.0-alpha")
	},
}

// ====================================================================================================
// INITIALIZATION
// ====================================================================================================

// init registers the version command with the root command.
// This function is automatically called by the Go runtime before the main function.
func init() {
	// RootCmd.AddCommand() is used to add this specific subcommand to the main command-line tool.
	// This makes `wiper version` a valid command.
	RootCmd.AddCommand(versionCmd)
}
