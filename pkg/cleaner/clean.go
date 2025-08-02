package cleaner

import (
	"bufio"
	"fmt"
	"github.com/kodelint/wiper/pkg/logger"
	"github.com/kodelint/wiper/pkg/reclaimer"
	"github.com/kodelint/wiper/pkg/utils"
	"os"
	"strings"
)

// ====================================================================================================
// DATA STRUCTURES
// ====================================================================================================

// cleanupItem represents an item (file or directory) that can be cleaned.
// This struct holds all the necessary information for the cleanup process.
type cleanupItem struct {
	Path       string // The aggregated category or display path for the dry run table
	Size       int64
	Category   string // The actual category for the summary table
	ActualPath string // The actual file/directory path to delete
}

// dryRunItem represents a folder and its size that would be removed in a dry run.
// This is specifically for the aggregated table display to show total sizes by category.
type dryRunItem struct {
	Path string
	Size int64
}

// ====================================================================================================
// UTILITY FUNCTIONS
// ====================================================================================================

// ConfirmAction asks the user for a yes/no confirmation.
// This function is now shared by all cleanup processes that require user interaction.
func ConfirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s (y/N): ", prompt)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" || input == "yes" {
			println("")
			return true
		}
		if input == "n" || input == "no" || input == "" { // Default to No on empty input
			println("")
			return false
		}
		fmt.Println("Invalid input. Please enter 'y' or 'n'.")
	}
}

// ====================================================================================================
// CORE CLEANUP LOGIC
// ====================================================================================================

// processCleanupItems handles the confirmation and removal logic for a list of items.
// This is a central function that manages different cleanup modes (dry run, interactive, etc.).
//
// Parameters:
//   - items: The slice of cleanupItem structs to process.
//   - dryRun: A boolean flag for dry-run mode.
//   - interactive: A boolean flag for interactive mode (per-file confirmation).
//   - summary: A pointer to a SummaryTable to record actual deletions.
//   - estimatedSummary: A pointer to a SummaryTable to record dry-run estimations.
//   - tableTitle: The title for the summary table.
//   - isApp: A boolean flag indicating if the cleanup is for an application uninstallation.
//
// Returns:
//   - The total space reclaimed in bytes and an error, if any.
func processCleanupItems(
	items []cleanupItem,
	dryRun bool,
	interactive bool,
	summary *reclaimer.SummaryTable,
	estimatedSummary *reclaimer.SummaryTable,
	tableTitle string,
	isApp bool,
) (int64, error) {
	var totalReclaimed int64

	if len(items) == 0 {
		logger.Log.Info("No items found for cleanup.")
		return 0, nil
	}

	// Step 1: Aggregate and Display Items for Dry Run or Confirmation
	// This logic groups similar items together for a cleaner table display.
	aggregatedForTable := make(map[string]int64)
	for _, item := range items {
		displayKey := item.Path
		if item.Path == item.ActualPath && item.Category != "" {
			displayKey = item.Category
		}
		aggregatedForTable[displayKey] += item.Size
		estimatedSummary.AddEntry(item.ActualPath, item.Size, false, item.Category)
	}

	var tableItems []dryRunItem
	for category, size := range aggregatedForTable {
		tableItems = append(tableItems, dryRunItem{Path: category, Size: size})
	}

	// Print the table of detected items by category [Estimated]
	estimatedSummary.PrintTable(true, "Estimated Reclaimed Summary")

	// If dry run mode is enabled, we stop here and just return the estimated total.
	if dryRun {
		for _, item := range tableItems { // Sum from tableItems for dry run estimate
			totalReclaimed += item.Size
		}
		return totalReclaimed, nil
	}

	// Step 2: Actual Deletion Logic (Non-Dry Run)
	var actualRemovedSize int64

	// Case 1: Interactive Mode
	// The user is prompted to confirm each deletion individually.
	if interactive {
		logger.Log.Info("Starting interactive cleanup. You will be prompted for each item.")
		for _, item := range items { // Loop through actual files for deletion (original `items` list)
			prompt := fmt.Sprintf("Delete %s (%s, Category: %s)?", item.ActualPath, utils.FormatBytes(item.Size), item.Category)
			if ConfirmAction(prompt) {
				reclaimed, err := utils.RemovePath(item.ActualPath, false) // false for not dry run
				if err != nil {
					logger.Log.Errorf("Failed to remove %s: %v", item.ActualPath, err)
					summary.AddEntry(item.ActualPath, item.Size, false, item.Category) // Mark as not removed on error
				} else {
					actualRemovedSize += reclaimed
					summary.AddEntry(item.ActualPath, reclaimed, true, item.Category) // Mark as removed
					if os.Getenv("WIPER_SHOW_DETAILS") == "true" {                    // Use the same detail env var
						logger.Log.Infof("Removed %s", item.ActualPath)
					}
				}
			} else {
				logger.Log.Infof("Skipped %s", item.ActualPath)
				summary.AddEntry(item.ActualPath, item.Size, false, item.Category) // Add to summary but mark as not removed
			}
		}
		// Case 2: Application Uninstallation Mode
		// This mode assumes a single confirmation was already given for the entire application.
		// It proceeds to delete all files found without further prompts.
	} else if isApp {
		for _, item := range items { // Loop through actual files for deletion (original `items` list)
			reclaimed, err := utils.RemovePath(item.ActualPath, false) // false for not dry run
			if err != nil {
				logger.Log.Errorf("Failed to remove %s: %v", item.ActualPath, err)
				summary.AddEntry(item.ActualPath, item.Size, false, item.Category) // Mark as not removed on error
			} else {
				actualRemovedSize += reclaimed
				summary.AddEntry(item.ActualPath, reclaimed, true, item.Category) // Mark as removed
				if os.Getenv("WIPER_SHOW_DETAILS") == "true" {                    // Use the same detail env var
					logger.Log.Infof("Removed %s", item.ActualPath)
				}
			}
		}
		// Case 3: Single Confirmation Mode (Default for System Cleanup)
		// This mode prompts the user once to confirm the deletion of all items.
	} else {
		// Single confirmation mode: ask once for all detected files
		totalPotentialReclaimed := int64(0)
		for _, item := range tableItems {
			totalPotentialReclaimed += item.Size
		}
		println()
		prompt := fmt.Sprintf("Do you want to clean up these items (Total: %s)?", reclaimer.FormatBytes(totalPotentialReclaimed))
		if ConfirmAction(prompt) {
			println(utils.Yellow("  Proceeding with cleanup...🚀"))
			println(utils.CyanBold("================================"))
			for _, item := range items { // Loop through actual files for deletion (original `items` list)
				reclaimed, err := utils.RemovePath(item.ActualPath, false) // false for not dry run
				if err != nil {
					logger.Log.Errorf("Failed to remove %s: %v", item.ActualPath, err)
					summary.AddEntry(item.ActualPath, item.Size, false, item.Category) // Mark as not removed on error
				} else {
					actualRemovedSize += reclaimed
					summary.AddEntry(item.ActualPath, reclaimed, true, item.Category) // Mark as removed
					if os.Getenv("WIPER_SHOW_DETAILS") == "true" {                    // Use the same detail env var
						logger.Log.Infof("Removed %s", item.ActualPath)
					}
				}
			}
		} else {
			logger.Log.Info("Cleanup cancelled by user.")
			return 0, nil // Return 0 reclaimed and no error if cancelled
		}
	}

	totalReclaimed = actualRemovedSize
	return totalReclaimed, nil
}
