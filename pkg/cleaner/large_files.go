package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kodelint/wiper/pkg/logger"
	"github.com/kodelint/wiper/pkg/reclaimer"
	"github.com/kodelint/wiper/pkg/utils"
)

// ====================================================================================================
// LARGE FILES CLEANUP FUNCTION
// ====================================================================================================

// CleanLargeFiles identifies and optionally removes large files based on a size threshold.
//
// Parameters:
//   - dryRun: A boolean flag for dry-run mode.
//   - ignorePaths: A slice of paths to be ignored during the scan.
//   - summary: A pointer to a SummaryTable to record deleted items.
//   - estimatedSummary: A pointer to a SummaryTable to record dry-run estimations.
//   - interactive: A boolean flag for interactive mode (prompts for each file).
//
// Returns:
//   - The total space reclaimed in bytes and an error, if any.
func CleanLargeFiles(dryRun bool, ignorePaths []string, summary *reclaimer.SummaryTable, estimatedSummary *reclaimer.SummaryTable, interactive bool) (int64, error) {
	logger.Log.Infof("Initiating large file scan (dryRun: %t, interactive: %t)", dryRun, interactive)

	// Define the threshold for a file to be considered "large" (100 MB).
	const largeFileThreshold = 100 * 1024 * 1024 // 100 MB

	// Directories to scan for large files.
	// We use utils.ExpandPath to handle environment variables like $HOME and user-friendly paths like `~`.
	dirsToScan := []string{
		utils.ExpandPath("/Users"),
		utils.ExpandPath("/private/var/folders"),
		utils.ExpandPath("/private/tmp"),
		utils.ExpandPath("$HOME/Downloads"),
		utils.ExpandPath("$HOME/Documents"),
	}

	// Prepare a cleaned list of absolute paths to ignore.
	var cleanedIgnorePaths = []string{
		// We automatically ignore the Applications folder to avoid scanning inside app bundles.
		utils.ExpandPath("$HOME/Applications/"),
	}
	for _, p := range ignorePaths {
		// Resolve user-provided ignore paths to absolute paths for reliable comparison.
		absPath, err := filepath.Abs(utils.ExpandPath(p))
		if err != nil {
			logger.Log.Warnf("Failed to resolve absolute path for ignore entry %s: %v", p, err)
			continue
		}
		cleanedIgnorePaths = append(cleanedIgnorePaths, absPath)
	}

	showWarnings := os.Getenv("WIPER_SHOW_WARNINGS") == "true"
	showDetails := os.Getenv("WIPER_SHOW_DETAILS") == "true"
	var suppressedWarnings bool // To track if any warnings were suppressed

	// Collect all large files as cleanupItems before processing.
	var itemsToProcess []cleanupItem

	for _, dir := range dirsToScan {
		// filepath.Walk traverses the file tree rooted at 'dir'.
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if showWarnings {
					logger.Log.Warnf("Error accessing path %s: %v", path, err)
				} else {
					suppressedWarnings = true
				}
				// Continue walking the rest of the tree despite the error on this path.
				return nil
			}

			// Check if the current path should be ignored.
			if utils.IsPathIgnored(path, cleanedIgnorePaths) {
				if info.IsDir() {
					// If the ignored path is a directory, skip the entire directory tree.
					return filepath.SkipDir
				}
				return nil
			}

			// If it's a directory, check for system paths that should be skipped.
			if info.IsDir() {
				if path == "/System" || path == "/Library" || path == "/usr" || path == "/Applications" || strings.HasPrefix(path, "/Developer") {
					return filepath.SkipDir
				}
				return nil
			}

			// Calculate the actual disk usage of the file using a system call.
			// This is more accurate for sparse files or files on HFS+ and APFS.
			var actualSize int64
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				actualSize = stat.Blocks * 512
			} else {
				actualSize = info.Size()
				logger.Log.Debugf("Could not get actual disk usage for %s, falling back to logical size.", path)
			}

			// Check if the file meets the large file size threshold.
			if actualSize >= largeFileThreshold {
				if showDetails {
					logger.Log.Infof("Found large file: %s (Actual Size: %s, Logical Size: %s)",
						path, reclaimer.FormatBytes(actualSize), reclaimer.FormatBytes(info.Size()))
				}

				// Assign a generic category to the file based on its path.
				category := categorizeLargeFilePath(path)
				itemsToProcess = append(itemsToProcess, cleanupItem{
					Path:       path, // For large files, Path is the actual file path for display in the table
					Size:       actualSize,
					Category:   category, // This is the aggregated category for the summary table
					ActualPath: path,     // Store the actual file path here
				})
			}
			return nil
		})

		if err != nil {
			if showWarnings {
				logger.Log.Errorf("Error walking directory %s: %v", dir, err)
			} else {
				suppressedWarnings = true
			}
		}
	}

	if suppressedWarnings {
		logger.Log.Warn("Some warnings were suppressed. Set WIPER_SHOW_WARNINGS=true to see full warning details.")
	}
	// Pass the collected items to the generic processing function.
	// The `isApp` flag is set to `false` as this is not an application uninstall.
	reclaimed, err := processCleanupItems(itemsToProcess,
		dryRun,
		interactive,
		summary,
		estimatedSummary,
		"Detected Large Files",
		false)
	if err != nil {
		return 0, fmt.Errorf("failed to process large files cleanup: %w", err)
	}

	return reclaimed, nil
}

// ====================================================================================================
// PATH CATEGORIZATION FUNCTION
// ====================================================================================================

// categorizeLargeFilePath determines a higher-level, generic category for a given large file path.
// This helps in creating a clean summary table for the user.
func categorizeLargeFilePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if strings.HasPrefix(path, "/private/var") || strings.HasPrefix(path, "/tmp") {
			return "System Temporary Files"
		}
		return "Other Large Files"
	}

	normalizedPath := path
	if strings.HasPrefix(path, homeDir) {
		normalizedPath = strings.Replace(path, homeDir, "~", 1)
	}

	if strings.Contains(normalizedPath, "~/Library/Application Support/Google/Chrome") ||
		strings.Contains(normalizedPath, "~/Library/Caches/Google/Chrome") ||
		strings.Contains(normalizedPath, "~/Library/Application Support/BraveSoftware/Brave-Browser") ||
		strings.Contains(normalizedPath, "~/Library/Caches/com.apple.Safari") ||
		strings.Contains(normalizedPath, "~/Library/Application Support/Firefox") {
		return "Browser Caches"
	}

	if strings.HasPrefix(normalizedPath, "~/Downloads") {
		return "User Downloads"
	}
	if strings.HasPrefix(normalizedPath, "~/Documents") {
		return "User Documents"
	}

	if strings.HasPrefix(normalizedPath, "~/Library/Application Support") {
		return "Application Support Files"
	}
	if strings.HasPrefix(normalizedPath, "~/Library/Caches") {
		return "User Caches"
	}
	if strings.HasPrefix(normalizedPath, "~/Library/Containers") {
		return "User Container Data"
	}
	if strings.HasPrefix(normalizedPath, "~/Library/Group Containers") {
		return "User Group Container Data"
	}
	if strings.HasPrefix(normalizedPath, "~/Library/Messages/Attachments") {
		return "Messages Attachments"
	}
	if strings.HasPrefix(normalizedPath, "~/Library/Metadata/CoreSpotlight") {
		return "Spotlight Metadata"
	}
	if strings.HasPrefix(normalizedPath, "~/Library") {
		return "Other User Library Files"
	}

	if strings.HasPrefix(path, "/private/var/folders") ||
		strings.HasPrefix(path, "/private/var/tmp") ||
		strings.HasPrefix(path, "/tmp") {
		return "System Temporary Files"
	}

	if strings.Contains(normalizedPath, ".rustup") ||
		strings.Contains(normalizedPath, ".npm") ||
		strings.Contains(normalizedPath, ".gradle") ||
		strings.Contains(normalizedPath, "Xcode/iOS DeviceSupport") ||
		strings.Contains(normalizedPath, "Android/Sdk") ||
		strings.Contains(normalizedPath, "JetBrains") {
		return "Developer Tool Caches/Data"
	}

	if strings.HasPrefix(normalizedPath, "~/") {
		return "User Home Files"
	}

	if strings.HasPrefix(path, "/var") || strings.HasPrefix(path, "/usr") || strings.HasPrefix(path, "/opt") {
		return "System Files"
	}

	return "Other Large Files"
}

//package cleaner
//
//import (
//	"fmt"
//	"os"
//	"path/filepath"
//	"strings"
//	"syscall"
//
//	"github.com/kodelint/wiper/pkg/logger"
//	"github.com/kodelint/wiper/pkg/reclaimer"
//	"github.com/kodelint/wiper/pkg/utils"
//)
//
//// ====================================================================================================
//// LARGE FILES CLEANUP FUNCTION
//// ====================================================================================================
//
//// CleanLargeFiles identifies and optionally removes large files based on a size threshold.
////
//// Parameters:
////   - dryRun: A boolean flag for dry-run mode.
////   - ignorePaths: A slice of paths to be ignored during the scan.
////   - summary: A pointer to a SummaryTable to record deleted items.
////   - estimatedSummary: A pointer to a SummaryTable to record dry-run estimations.
////   - interactive: A boolean flag for interactive mode (prompts for each file).
////
//// Returns:
////   - The total space reclaimed in bytes and an error, if any.
//func CleanLargeFiles(dryRun bool, ignorePaths []string, summary *reclaimer.SummaryTable, estimatedSummary *reclaimer.SummaryTable, interactive bool) (int64, error) {
//	logger.Log.Infof("Initiating large file scan (dryRun: %t, interactive: %t)", dryRun, interactive)
//
//	// Define the threshold for a file to be considered "large" (100 MB).
//	const largeFileThreshold = 100 * 1024 * 1024
//
//	// Directories to scan for large files.
//	// We use utils.ExpandPath to handle environment variables like $HOME and user-friendly paths like `~`.
//	dirsToScan := []string{
//		utils.ExpandPath("/Users"),
//		utils.ExpandPath("/private/var/folders"),
//		utils.ExpandPath("/private/tmp"),
//		utils.ExpandPath("$HOME/Downloads"),
//		utils.ExpandPath("$HOME/Documents"),
//	}
//
//	// Prepare a cleaned list of absolute paths to ignore.
//	var cleanedIgnorePaths = []string{
//		// We automatically ignore the Applications folder to avoid scanning inside app bundles.
//		utils.ExpandPath("$HOME/Applications/"),
//	}
//	for _, p := range ignorePaths {
//		// Resolve user-provided ignore paths to absolute paths for reliable comparison.
//		absPath, err := filepath.Abs(utils.ExpandPath(p))
//		if err != nil {
//			logger.Log.Warnf("Failed to resolve absolute path for ignore entry %s: %v", p, err)
//			continue
//		}
//		cleanedIgnorePaths = append(cleanedIgnorePaths, absPath)
//	}
//
//	showWarnings := os.Getenv("WIPER_SHOW_WARNINGS") == "true"
//	showDetails := os.Getenv("WIPER_SHOW_DETAILS") == "true"
//	var suppressedWarnings bool
//
//	// Collect all large files as cleanupItems before processing.
//	var itemsToProcess []cleanupItem
//
//	for _, dir := range dirsToScan {
//		// filepath.Walk traverses the file tree rooted at 'dir'.
//		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
//			if err != nil {
//				if showWarnings {
//					logger.Log.Warnf("Error accessing path %s: %v", path, err)
//				} else {
//					suppressedWarnings = true
//				}
//				// Continue walking the rest of the tree despite the error on this path.
//				return nil
//			}
//
//			// Check if the current path should be ignored.
//			if utils.IsPathIgnored(path, cleanedIgnorePaths) {
//				if info.IsDir() {
//					// If the ignored path is a directory, skip the entire directory tree.
//					return filepath.SkipDir
//				}
//				return nil
//			}
//
//			// If it's a directory, check for system paths that should be skipped.
//			if info.IsDir() {
//				if path == "/System" || path == "/Library" || path == "/usr" || path == "/Applications" || strings.HasPrefix(path, "/Developer") {
//					return filepath.SkipDir
//				}
//				return nil
//			}
//
//			// Calculate the actual disk usage of the file using a system call.
//			// This is more accurate for sparse files or files on HFS+ and APFS.
//			var actualSize int64
//			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
//				actualSize = stat.Blocks * 512
//			} else {
//				actualSize = info.Size()
//				logger.Log.Debugf("Could not get actual disk usage for %s, falling back to logical size.", path)
//			}
//
//			// Check if the file meets the large file size threshold.
//			if actualSize >= largeFileThreshold {
//				if showDetails {
//					logger.Log.Infof("Found large file: %s (Actual Size: %s, Logical Size: %s)",
//						path, reclaimer.FormatBytes(actualSize), reclaimer.FormatBytes(info.Size()))
//				}
//				// Assign a generic category to the file based on its path.
//				category := categorizeLargeFilePath(path)
//				itemsToProcess = append(itemsToProcess, cleanupItem{
//					Path:       path, // For large files, Path is the actual file path for display.
//					Size:       actualSize,
//					Category:   category,
//					ActualPath: path,
//				})
//			}
//			return nil
//		})
//
//		if err != nil {
//			if showWarnings {
//				logger.Log.Errorf("Error walking directory %s: %v", dir, err)
//			} else {
//				suppressedWarnings = true
//			}
//		}
//	}
//
//	if suppressedWarnings {
//		logger.Log.Warn("Some warnings were suppressed. Set WIPER_SHOW_WARNINGS=true to see full warning details.")
//	}
//
//	// Pass the collected items to the generic processing function.
//	// The `isApp` flag is set to `false` as this is not an application uninstall.
//	reclaimed, err := processCleanupItems(itemsToProcess,
//		dryRun,
//		interactive,
//		summary,
//		estimatedSummary,
//		"Detected Large Files",
//		false)
//	if err != nil {
//		return 0, fmt.Errorf("failed to process large files cleanup: %w", err)
//	}
//
//	return reclaimed, nil
//}
//
//// ====================================================================================================
//// PATH CATEGORIZATION FUNCTION
//// ====================================================================================================
//
//// categorizeLargeFilePath determines a higher-level, generic category for a given large file path.
//// This helps in creating a clean summary table for the user.
//func categorizeLargeFilePath(path string) string {
//	homeDir, err := os.UserHomeDir()
//	if err != nil {
//		// Fallback for paths that don't belong to a user.
//		if strings.HasPrefix(path, "/private/var") || strings.HasPrefix(path, "/tmp") {
//			return "System Temporary Files"
//		}
//		return "Other Large Files"
//	}
//
//	normalizedPath := path
//	// Normalize paths by replacing the user's home directory with a tilde (~).
//	if strings.HasPrefix(path, homeDir) {
//		normalizedPath = strings.Replace(path, homeDir, "~", 1)
//	}
//
//	// Check for common browser caches.
//	if strings.Contains(normalizedPath, "~/Library/Application Support/Google/Chrome") ||
//		strings.Contains(normalizedPath, "~/Library/Caches/Google/Chrome") ||
//		strings.Contains(normalizedPath, "~/Library/Application Support/BraveSoftware/Brave-Browser") ||
//		strings.Contains(normalizedPath, "~/Library/Caches/com.apple.Safari") ||
//		strings.Contains(normalizedPath, "~/Library/Application Support/Firefox") {
//		return "Browser Caches"
//	}
//
//	// Check for common user directories.
//	if strings.HasPrefix(normalizedPath, "~/Downloads") {
//		return "User Downloads"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Documents") {
//		return "User Documents"
//	}
//
//	// Check for various user library folders.
//	if strings.HasPrefix(normalizedPath, "~/Library/Application Support") {
//		return "Application Support Files"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library/Caches") {
//		return "User Caches"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library/Containers") {
//		return "User Container Data"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library/Group Containers") {
//		return "User Group Container Data"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library/Messages/Attachments") {
//		return "Messages Attachments"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library/Metadata/CoreSpotlight") {
//		return "Spotlight Metadata"
//	}
//	if strings.HasPrefix(normalizedPath, "~/Library") {
//		return "Other User Library Files"
//	}
//
//	// Check for system temporary files.
//	if strings.HasPrefix(path, "/private/var/folders") ||
//		strings.HasPrefix(path, "/private/var/tmp") ||
//		strings.HasPrefix(path, "/tmp") {
//		return "System Temporary Files"
//	}
//
//	// Check for developer tool caches and data.
//	if strings.Contains(normalizedPath, ".rustup") ||
//		strings.Contains(normalizedPath, ".npm") ||
//		strings.Contains(normalizedPath, ".gradle") ||
//		strings.Contains(normalizedPath, "Xcode/iOS DeviceSupport") ||
//		strings.Contains(normalizedPath, "Android/Sdk") ||
//		strings.Contains(normalizedPath, "JetBrains") {
//		return "Developer Tool Caches/Data"
//	}
//
//	// A final check for generic user home files.
//	if strings.HasPrefix(normalizedPath, "~/") {
//		return "User Home Files"
//	}
//
//	// Check for system-level files.
//	if strings.HasPrefix(path, "/var") || strings.HasPrefix(path, "/usr") || strings.HasPrefix(path, "/opt") {
//		return "System Files"
//	}
//
//	// If no specific category is found, return a generic one.
//	return "Other Large Files"
//}
