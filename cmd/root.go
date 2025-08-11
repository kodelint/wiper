package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kodelint/wiper/pkg/logger"
	"github.com/spf13/cobra"
)

// ====================================================================================================
// GLOBAL VARIABLES AND FLAGS
// ====================================================================================================

// These global variables are used to hold the values of persistent flags.
// They are accessible to all commands and subcommands.
var (
	// debugFlag enables verbose debug logging for more detailed output.
	debugFlag bool
	// dryRunFlag indicates that the command should not make any changes to the system.
	// It's used for previewing what would be done.
	dryRunFlag bool
	// ignorePathsStr holds the raw comma-separated string of paths from the --ignore flag.
	ignorePathsStr string
	// IgnorePaths will hold the parsed slice of paths, used by subcommands
	// after being processed in PersistentPreRunE.
	IgnorePaths []string
)

// ====================================================================================================
// ROOT COMMAND DEFINITION
// ====================================================================================================

// RootCmd represents the base command when called without any subcommands.
// It is the entry point for the wiper CLI application.
var RootCmd = &cobra.Command{
	Use:   "wiper",
	Short: "A powerful tool to uninstall applications and clean up macOS.",
	Long: `wiper is a comprehensive macOS utility designed to help you:

1.  Uninstall Applications: Completely remove applications and all their leftover files, reclaiming disk space.
2.  Clean System Junk: Remove temporary files, caches (system, user, browser), and other unnecessary data to optimize your macOS performance.

It provides detailed output and supports dry-run modes to show you what will be removed before any changes are made.`,

	// PersistentPreRunE is a function that is executed before any command (including subcommands).
	// It is used to initialize common settings or pre-process flags that apply to all commands.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the logger based on the debug flag.
		// If the debug flag is set, we enable a more verbose logging level.
		if debugFlag {
			logger.SetDebug(true)
		}

		// Parse the ignorePathsStr into the IgnorePaths slice.
		// This logic ensures that the --ignore flag is processed once and the result
		// is available as a slice of strings for all subcommands.
		if ignorePathsStr != "" {
			// Split the comma-separated string into a slice of strings.
			paths := strings.Split(ignorePathsStr, ",")
			for _, p := range paths {
				// Trim leading/trailing whitespace from each path.
				trimmedPath := strings.TrimSpace(p)
				if trimmedPath != "" {
					// Append the cleaned path to the global slice.
					IgnorePaths = append(IgnorePaths, trimmedPath)
				}
			}
			// Log the ignored paths at a debug level for verification.
			logger.Log.Debugf("Ignoring paths: %v", IgnorePaths)
		}

		// Return nil to indicate that the setup was successful.
		return nil
	},
}

// ====================================================================================================
// APPLICATION ENTRY POINT
// ====================================================================================================

// Execute adds all child commands to the root command and sets flags appropriately.
// It is the main entry point for the cobra application and is called by the main() function.
// It only needs to be called once to execute the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// If an error occurs during execution, print the error to standard error
		// and exit the program with a non-zero status code.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ====================================================================================================
// INITIALIZATION
// ====================================================================================================

// init initializes global flags for the root command.
// This function is automatically called by the Go runtime before the main function.
func init() {
	// Persistent flags are available to the command they are declared on, as well as all of its subcommands.
	// This is where we define the global flags like --debug, --dry-run, and --ignore.

	// BoolVarP binds a boolean flag to a variable.
	// &debugFlag: The variable to store the flag's value.
	// "debug": The long name of the flag (--debug).
	// "d": The short name of the flag (-d).
	// false: The default value.
	// "Enable debug logging.": The usage description.
	RootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug logging.")

	// BoolVarP for the dry-run flag.
	RootCmd.PersistentFlags().BoolVarP(&dryRunFlag, "dry-run", "n", false, "Perform a dry run without making any changes.")

	// StringVarP binds a string flag to a variable.
	// &ignorePathsStr: The variable to store the flag's string value.
	// "ignore": The long name of the flag (--ignore).
	// "i": The short name of the flag (-i).
	// "": The default value (an empty string).
	// "Comma-separated list of paths to ignore during cleanup.": The usage description.
	RootCmd.PersistentFlags().StringVarP(&ignorePathsStr, "ignore", "i", "", "Comma-separated list of paths to ignore during cleanup.")
}
