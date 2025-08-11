package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kodelint/wiper/pkg/logger"
	"github.com/kodelint/wiper/pkg/reclaimer"
	"github.com/kodelint/wiper/pkg/utils"
)

// ====================================================================================================
// SYSTEM CLEANUP FUNCTION
// ====================================================================================================

// CleanSystem performs a comprehensive system cleanup on macOS.
// It removes temporary files, caches, and other junk files based on predefined targets.
//
// Parameters:
//   - dryRun: A boolean flag for dry-run mode (no files are actually deleted).
//   - ignorePaths: A list of paths to explicitly exclude from deletion.
//   - summary: A pointer to a SummaryTable to record deleted items and their sizes.
//   - estimatedSummary: A pointer to a SummaryTable to record items found during a dry run.
//
// Returns:
//   - The total space reclaimed in bytes and an error, if any.
func CleanSystem(dryRun bool, ignorePaths []string, summary *reclaimer.SummaryTable, estimatedSummary *reclaimer.SummaryTable) (int64, error) {
	logger.Log.Debugf(utils.Cyan("Starting system cleanup..."))
	// getCleanupTargets() is assumed to be defined elsewhere and returns a slice of CleanupTarget structs.
	cleanupTargets := getCleanupTargets() // Get cleanup targets from the dedicated function

	// Pre-process ignorePaths to expand environment variables like ~ and $HOME once upfront.
	var expandedIgnorePaths []string
	for _, p := range ignorePaths {
		expandedIgnorePaths = append(expandedIgnorePaths, utils.ExpandPath(p))
	}

	showWarnings := os.Getenv("WIPER_SHOW_WARNINGS") == "true"
	var suppressedWarnings bool // To track if any warnings were suppressed

	// Collect all potential items to process as cleanupItems
	var itemsToProcess []cleanupItem

	for _, target := range cleanupTargets {
		logger.Log.Debugf("Scanning for %s using patterns: %v", target.Category, target.Paths)
		for _, pattern := range target.Paths {
			// filepath.Glob finds all file paths matching a pattern.
			matches, err := filepath.Glob(pattern)
			if err != nil {
				if showWarnings {
					logger.Log.Warnf("Error globbing pattern %s: %v", pattern, err)
				} else {
					suppressedWarnings = true
				}
				continue
			}

			for _, path := range matches {
				// Check if the path is in the list of paths to ignore.
				if utils.ContainsPath(path, expandedIgnorePaths) {
					logger.Log.Debugf(utils.Yellow("Skipping ignored path: %s"), path)
					continue
				}

				fileInfo, err := os.Stat(path)
				if err != nil {
					if showWarnings {
						logger.Log.Debugf("Error stating path %s: %v", path, err)
					} else {
						suppressedWarnings = true
					}
					continue
				}
				// Check if the file's modification time is recent, if a minimum age is specified.
				if target.MinAge > 0 && time.Since(fileInfo.ModTime()) < target.MinAge {
					logger.Log.Debugf("Skipping recent file/directory: %s (Modified: %s)", path, fileInfo.ModTime().Format("2006-01-02"))
					continue
				}
				// Get the size of the file to be able to calculate the total reclaimed space.
				size, err := utils.GetFileSizeInBytes(path)
				if err != nil {
					if showWarnings {
						logger.Log.Debugf("Could not get size of %s for aggregation: %v", path, err)
					} else {
						suppressedWarnings = true
					}
					continue
				}

				// The 'Path' field in cleanupItem is used for display. We aggregate files
				// by their cleanup target root for a cleaner-looking summary table.
				displayPath := path // Default to individual path
				for _, root := range target.LogAggregationRoots {
					if strings.HasPrefix(path, root) {
						displayPath = root
						break
					}
				}

				itemsToProcess = append(itemsToProcess, cleanupItem{
					Path:       displayPath,     // This is the aggregated path for display in the table
					Size:       size,            // Size of the file.
					Category:   target.Category, // This is the higher-level category for the summary table
					ActualPath: path,            // This is the actual path to delete
				})
			}
		}
	}

	if suppressedWarnings {
		logger.Log.Warn("Some warnings were suppressed. Set WIPER_SHOW_WARNINGS=true to see full warning details.")
	}

	// Call the generic processCleanupItems function to handle the deletion logic.
	// System cleanup is not interactive by default.
	reclaimed, err := processCleanupItems(itemsToProcess,
		dryRun,
		false,
		summary,
		estimatedSummary,
		"Folders that would be cleaned",
		false)
	if err != nil {
		return 0, fmt.Errorf("failed to process system cleanup: %w", err)
	}

	return reclaimed, nil
}
