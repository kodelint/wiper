package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kodelint/wiper/pkg/logger"
	"github.com/kodelint/wiper/pkg/reclaimer"
	"github.com/kodelint/wiper/pkg/utils"
)

// ====================================================================================================
// GLOBAL VARIABLES
// ====================================================================================================

// appInstallPaths defines the common directories where macOS applications can be installed.
// The search for application bundles will be limited to these locations.
var appInstallPaths = []string{
	"/Applications",
	filepath.Join(os.Getenv("HOME"), "Applications"),
}

// ====================================================================================================
// APPLICATION UNINSTALLATION FUNCTION
// ====================================================================================================

// UninstallApplication attempts to remove a specified macOS application and its leftover files.
// It returns the total space reclaimed in bytes and an error, if any.
//
// Parameters:
//   - appName: The name of the application to uninstall (e.g., "Google Chrome").
//   - dryRun: A boolean flag indicating whether to perform a dry run (simulate deletion without changes).
//   - ignorePaths: A slice of paths to be ignored during the cleanup process.
//   - summary: A pointer to a SummaryTable to record deleted items and their sizes.
//   - estimatedSummary: A pointer to a SummaryTable to record estimated items and their sizes (for dry runs).
func UninstallApplication(appName string, dryRun bool, ignorePaths []string, summary *reclaimer.SummaryTable, estimatedSummary *reclaimer.SummaryTable) (int64, error) {
	// Ensure the application name ends with ".app" for consistent searching.
	if !strings.HasSuffix(appName, ".app") {
		appName += ".app"
	}

	logger.Log.Infof(utils.Cyan("Searching for '%s' and its associated files..."), appName)

	var itemsToProcess []cleanupItem

	// =================================================================================================
	// Step 1: Find Application Bundles and Leftover Files
	// =================================================================================================

	// Find the main application bundle(s) in the defined common installation paths.
	appBundlePaths := utils.FindPaths(appInstallPaths, appName)
	if len(appBundlePaths) == 0 {
		logger.Log.Warnf(utils.Yellow("Application '%s' not found in common /Applications directories."), appName)
	} else {
		for _, bundlePath := range appBundlePaths {
			// Check if the path should be ignored.
			if !utils.IsPathIgnored(bundlePath, ignorePaths) {
				size, err := utils.GetFileSizeInBytes(bundlePath)
				if err == nil {
					itemsToProcess = append(itemsToProcess, cleanupItem{
						Path:       bundlePath,
						Size:       size,
						Category:   "Application Bundle",
						ActualPath: bundlePath,
					})
				}
			} else {
				logger.Log.Debugf(utils.Yellow("Skipping ignored application bundle: %s"), bundlePath)
			}
		}
	}

	// Search for related files and directories in common leftover locations.
	// We use `filepath.Glob` with patterns to find files that match a wildcard.
	logger.Log.Infof(utils.Cyan("Searching for leftover files for '%s'..."), strings.TrimSuffix(appName, ".app"))

	baseAppName := strings.TrimSuffix(appName, ".app")
	leftoverSearchPatterns := []string{
		// Common paths for application support, caches, preferences, and containers.
		filepath.Join(os.Getenv("HOME"), "Library", "Application Support", baseAppName),
		filepath.Join(os.Getenv("HOME"), "Library", "Caches", baseAppName),
		// Preferences files often follow a reverse-domain-name convention (e.g., com.google.chrome.plist).
		filepath.Join(os.Getenv("HOME"), "Library", "Preferences", "com."+strings.ToLower(strings.ReplaceAll(baseAppName, " ", ""))+".*"),
		filepath.Join(os.Getenv("HOME"), "Library", "Saved Application State", "com."+strings.ToLower(strings.ReplaceAll(baseAppName, " ", ""))+".*"),
		filepath.Join(os.Getenv("HOME"), "Library", "Containers", "*"+baseAppName+"*"),
		filepath.Join(os.Getenv("HOME"), "Library", "Group Containers", "*"+baseAppName+"*"),
		// System-wide library paths.
		filepath.Join("/Library", "Application Support", baseAppName),
		filepath.Join("/Library", "Caches", baseAppName),
		filepath.Join("/Library", "Preferences", "com."+strings.ToLower(strings.ReplaceAll(baseAppName, " ", ""))+".*"),
	}

	for _, pattern := range leftoverSearchPatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.Log.Debugf("Error globbing app data pattern %s: %v", pattern, err)
			continue
		}
		for _, match := range matches {
			if _, err := os.Stat(match); err == nil && !utils.IsPathIgnored(match, ignorePaths) {
				size, err := utils.GetFileSizeInBytes(match)
				if err == nil {
					itemsToProcess = append(itemsToProcess, cleanupItem{
						Path:       match,
						Size:       size,
						Category:   "Application Leftover",
						ActualPath: match,
					})
				}
			} else if err == nil && utils.IsPathIgnored(match, ignorePaths) {
				logger.Log.Debugf(utils.Yellow("Skipping ignored leftover path: %s"), match)
			}
		}
	}

	if len(itemsToProcess) == 0 {
		logger.Log.Info("No items found for cleanup.")
		return 0, nil
	}

	// =================================================================================================
	// Step 2: Process and Clean Up the Found Items
	// =================================================================================================

	// Call the generic processCleanupItems function to handle the deletion logic.
	// This function centralizes the logic for dry-run simulation, deletion, and summary updates.
	// Note: We pass `false` for the interactive flag as this feature is not supported for application uninstallation.
	reclaimed, err := processCleanupItems(
		itemsToProcess,
		dryRun,
		false, // interactiveMode is not enabled for app uninstall
		summary,
		estimatedSummary,
		fmt.Sprintf("Application Cleanup for '%s'", strings.TrimSuffix(appName, ".app")),
		true, // always show progress for this type of cleanup
	)

	return reclaimed, err
}
