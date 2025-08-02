package cmd

import (
	"fmt" // Used for formatted I/O, primarily for printing messages and errors.

	"github.com/kodelint/wiper/pkg/cleaner"   // Contains the core cleanup logic, such as uninstalling and cleaning files.
	"github.com/kodelint/wiper/pkg/logger"    // Provides a structured logging interface for debug and info messages.
	"github.com/kodelint/wiper/pkg/reclaimer" // Manages and formats disk space reclaimed during cleanup.
	"github.com/kodelint/wiper/pkg/utils"     // A collection of utility functions, such as for colored output.
	"github.com/spf13/cobra"                  // The primary library for building the command-line interface.
)

// ====================================================================================================
// COMMAND-SPECIFIC FLAGS
// ====================================================================================================

// largeFilesFlag controls whether to perform a large files cleanup.
// It is a local flag, specific to the `wipe` command.
var largeFilesFlag bool

// interactiveFlag enables interactive mode, prompting for confirmation before each deletion.
// It is a local flag for the `wipe` command.
var interactiveFlag bool

// ====================================================================================================
// WIPE COMMAND DEFINITION
// ====================================================================================================

// wipeCmd represents the wipe command.
// It is a powerful subcommand that handles both application uninstallation and system-wide cleanup,
// depending on the arguments and flags provided.
var wipeCmd = &cobra.Command{
	Use:   "wipe [application-name]",
	Short: "Uninstall an application or clean up the system.",
	Long: `The 'wipe' command performs two primary functions:

1.  Application Uninstallation: If an application name is provided (e.g., 'wiper wipe "Google Chrome"'),
   it will attempt to uninstall the specified application and remove its associated files.

2.  System Cleanup: If no application name is provided (e.g., 'wiper wipe'),
   it will perform a comprehensive system cleanup, removing junk files, temporary files,
   and caches from various locations on macOS.

3.  Large Files Cleanup: If the '--large-files' flag is used (e.g., 'wiper wipe --large-files'),
   it will identify and offer to clean up large files that are not typically part of
   standard system cleanup.

Use the '--dry-run' flag to see what will be removed without making actual changes.
Use the '--ignore' flag to specify paths to exclude from system cleanup.
Use the '--interactive' flag to confirm each deletion individually.`,
	Example: `
 # Uninstall an application
 wiper wipe "Google Chrome"
 wiper wipe "VS Code" --dry-run

 # Perform a full system cleanup
 wiper wipe
 wiper wipe --dry-run

 # Perform a large files cleanup
 wiper wipe --large-files
 wiper wipe --dry-run --large-files
 wiper wipe --large-files --interactive

 # Perform system cleanup, ignoring specific paths
 wiper wipe --ignore "/Users/john/Downloads,/System/Library/Caches"`,

	// Args specifies the number of arguments the command expects.
	// cobra.MaximumNArgs(1) means the command can have 0 or 1 argument.
	Args: cobra.MaximumNArgs(1),

	// RunE is the function that contains the core logic for the command.
	// It returns an error, which cobra will handle automatically.
	RunE: func(cmd *cobra.Command, args []string) error {
		// Log the status of persistent flags from the root command.
		// These variables (dryRunFlag, IgnorePaths) are populated by RootCmd.PersistentPreRunE.

		logger.Log.Debugf("Dry Run: %t", dryRunFlag)
		if len(IgnorePaths) > 0 {
			logger.Log.Debugf("Ignore Paths: %v", IgnorePaths)
		}

		// Log the status of local flags for the wipe command.
		if largeFilesFlag {
			logger.Log.Debugf("Large Files Cleanup: %t", largeFilesFlag)
		}
		if interactiveFlag {
			logger.Log.Debugf("Interactive Mode: %t", interactiveFlag)
		}

		var reclaimed int64
		summary := reclaimer.NewSummaryTable()
		estimatedSummary := reclaimer.NewSummaryTable()
		var err error

		// =================================================================
		// Logic Branching: Large Files, Application, or System Cleanup
		// =================================================================

		// Case 1: Large Files Cleanup
		if largeFilesFlag {
			// Ensure that an application name is not provided with the --large-files flag.
			if len(args) > 0 {
				return fmt.Errorf("the --large-files flag cannot be used with an application name")
			}
			logger.Log.Info("Performing large files cleanup...")

			// Call the CleanLargeFiles function from the cleaner package.
			// The dryRunFlag and IgnorePaths are passed to control the cleanup process.
			// The interactiveFlag is used to prompt for each deletion.
			reclaimed, err = cleaner.CleanLargeFiles(dryRunFlag, IgnorePaths, summary, estimatedSummary, interactiveFlag)
			if err != nil {
				return fmt.Errorf("failed to clean large files: %w", err)
			}

			// Case 2: Application Uninstallation
		} else if len(args) == 1 {
			appName := args[0]
			// Warn the user that interactive mode is not supported for this action.
			if interactiveFlag {
				logger.Log.Warn("Interactive mode is not supported for application uninstallation and will be ignored.")
			}
			logger.Log.Infof("Attempting to uninstall application: %s", appName)

			// Confirm with the user before proceeding with the uninstallation.
			prompt := fmt.Sprintf("Do you really want to uninstall application: %s?", appName)
			if cleaner.ConfirmAction(prompt) {
				// Call the UninstallApplication function from the cleaner package.
				reclaimed, err = cleaner.UninstallApplication(appName, dryRunFlag, IgnorePaths, summary, estimatedSummary)
				if err != nil {
					return fmt.Errorf("failed to uninstall %s: %w", appName, err)
				}
				logger.Log.Infof("Application uninstallation completed. Space reclaimed: %s", reclaimer.FormatBytes(reclaimed))
			} else {
				return fmt.Errorf("aborting uninstallation of %s", appName)
			}

			// Case 3: System Cleanup (Default)
		} else {
			logger.Log.Info("Performing system-wide cleanup...")
			// Warn the user that interactive mode is not supported for this action.
			if interactiveFlag {
				logger.Log.Warn("Interactive mode is not supported for system-wide cleanup and will be ignored.")
			}

			// Call the CleanSystem function from the cleaner package.
			space, err := cleaner.CleanSystem(dryRunFlag, IgnorePaths, summary, estimatedSummary)
			if err != nil {
				return fmt.Errorf("failed to clean system: %w", err)
			}
			reclaimed = space
		}

		// =================================================================
		// Final Output and Summary
		// =================================================================

		// Print a summary table of the disk space reclaimed.
		summary.PrintTable(false, "Reclaimed Disk Summary")
		println("\n")

		// Print the final message based on whether it was a dry run or an actual cleanup.
		if dryRunFlag {
			logger.Log.Infof(utils.CyanBold("Cleanup estimation finished. Estimated space reclaimed: %s"), utils.GreenBold(reclaimer.FormatBytes(reclaimed)))
		} else {
			logger.Log.Infof("Cleanup completed. Space reclaimed: %s", utils.GreenBold(reclaimer.FormatBytes(reclaimed)))
		}

		return nil
	},
}

// ====================================================================================================
// INITIALIZATION
// ====================================================================================================

// init registers the wipe command with the root command.
// This function is automatically called by the Go runtime before the main function.
func init() {
	// Add the wipeCmd as a subcommand to the RootCmd.
	RootCmd.AddCommand(wipeCmd)

	// Define local flags for the wipe command.
	// These flags are only available when the `wiper wipe` command is used.

	// BoolVar defines a boolean flag.
	// It binds the --large-files flag to the largeFilesFlag variable.
	wipeCmd.Flags().BoolVar(&largeFilesFlag, "large-files", false, "Perform a cleanup of large files")

	// BoolVarP defines a boolean flag with both a long name and a short name.
	// It binds the --interactive or -I flag to the interactiveFlag variable.
	wipeCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "I", false, "Prompt for confirmation before each deletion (only for --large-files)")
}
