package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/kodelint/wiper/pkg/logger"
)

// ====================================================================================================
// COLOR UTILITY FUNCTIONS
// ====================================================================================================

// These global variables are pre-configured color functions from the `fatih/color` library.
// They are used to apply consistent coloring to text output in the terminal.
var (
	Green     = color.New(color.FgGreen).SprintFunc()
	GreenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
	White     = color.New(color.FgHiWhite).SprintFunc()
	WhiteBold = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	Yellow    = color.New(color.FgYellow).SprintFunc()
	Red       = color.New(color.FgRed).SprintFunc()
	Cyan      = color.New(color.FgCyan).SprintFunc()
	CyanBold  = color.New(color.FgCyan, color.Bold).SprintFunc()
	Blue      = color.New(color.FgBlue).SprintFunc()
)

// ====================================================================================================
// FILE SYSTEM UTILITY FUNCTIONS
// ====================================================================================================

// GetFileSizeInBytes calculates the total size of a file or directory recursively.
// It uses `os.Lstat` to correctly handle symbolic links and `syscall.Stat_t` to get
// the more accurate "actual disk usage" rather than the logical file size.
//
// Parameters:
//   - path: The file or directory path to check.
//
// Returns:
//   - The total size in bytes and an error, if any.
func GetFileSizeInBytes(path string) (int64, error) {
	var totalSize int64

	// First, check if the path exists
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // Path doesn't exist, size is 0
		}
		return 0, fmt.Errorf("failed to get info for %s: %w", path, err)
	}

	// Use stat.Blocks * 512 for a more accurate size on disk
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		// Use the number of 512-byte blocks, which is more accurate for sparse files
		if !info.IsDir() {
			return stat.Blocks * 512, nil
		}
	} else {
		// Fallback to logical size if not a syscall.Stat_t
		if !info.IsDir() {
			logger.Log.Debugf("Could not get actual disk usage for %s, falling back to logical size.", path)
			return info.Size(), nil
		}
	}

	// For a directory, we need to walk it to get the total size of all its contents
	err = filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
		if err != nil {
			logger.Log.Debugf("Error walking path %s for size calculation: %v", subPath, err)
			return filepath.SkipDir
		}

		if subInfo.IsDir() {
			// Get the size of the directory itself
			if stat, ok := subInfo.Sys().(*syscall.Stat_t); ok {
				totalSize += stat.Blocks * 512
			} else {
				totalSize += subInfo.Size()
			}
			return nil
		}

		// Get the size of the file
		if stat, ok := subInfo.Sys().(*syscall.Stat_t); ok {
			totalSize += stat.Blocks * 512
		} else {
			totalSize += subInfo.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to walk path %s: %w", path, err)
	}
	return totalSize, nil
}

// RemovePath removes a file or directory.
// It handles symbolic links and includes a dry-run option.
//
// Parameters:
//   - path: The path of the file or directory to remove.
//   - dryRun: If true, the function will only log what it would do, without making changes.
//
// Returns:
//   - The size of the removed item in bytes and an error, if any.
func RemovePath(path string, dryRun bool) (int64, error) {
	size, err := GetFileSizeInBytes(path)
	if err != nil {
		return 0, fmt.Errorf("could not get size of %s before removal: %w", path, err)
	}

	if dryRun {
		logger.Log.Debugf(Yellow("DRY RUN: Would remove granular item: %s (Size: %s)"), path, FormatBytes(size))
		return size, nil
	}

	logger.Log.Debugf(Red("Removing granular item: %s (Size: %s)"), path, FormatBytes(size))
	//Enable it if we really need to remove it
	//logger.Log.Infof("Removing granular item: %s (Size: %s)", path, FormatBytes(size))
	//if err := os.RemoveAll(path); err != nil {
	//	return 0, fmt.Errorf("failed to remove %s: %w", path, err)
	//}
	return size, nil
}

// ====================================================================================================
// PATH AND STRING UTILITY FUNCTIONS
// ====================================================================================================

// ExpandPath replaces `~` with the user's home directory and expands
// environment variables like `$HOME` or `${HOME}`.
// This is crucial for handling user-provided paths reliably.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	} else if strings.Contains(path, "$HOME") {
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			path = strings.ReplaceAll(path, "$HOME", homeDir)
		}
	} else if strings.Contains(path, "${HOME}") {
		homeDir := os.Getenv("HOME")
		if homeDir != "" {
			path = strings.ReplaceAll(path, "${HOME}", homeDir)
		}
	}
	return path
}

// FormatBytes converts an integer size in bytes into a human-readable string.
// For example, 1024 becomes "1.0 KB", and 1234567 becomes "1.2 MB".
func FormatBytes(b int64) string {
	const (
		_        = iota // ignore first value by assigning to blank identifier
		KB int64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f TB", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d Bytes", b)
	}
}

// ContainsPath checks if a given path is a sub-path of any path in a list.
// This is used to implement the `--ignore` functionality.
// It handles cases where an item to be checked is a child of an ignored directory.
func ContainsPath(targetPath string, ignorePaths []string) bool {
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		logger.Log.Warnf("Could not get absolute path for target %s: %v", targetPath, err)
		return false
	}

	for _, ignored := range ignorePaths {
		// IMPORTANT: Expand the ignored path first, then absolutize it
		expandedIgnored := ExpandPath(ignored)
		absIgnoredPath, err := filepath.Abs(expandedIgnored)
		if err != nil {
			logger.Log.Warnf("Could not get absolute path for ignore path %s (expanded from %s): %v", expandedIgnored, ignored, err)
			continue
		}

		// Clean the paths to handle cases like /tmp/../var or double slashes
		cleanTargetPath := filepath.Clean(absTargetPath)
		cleanIgnoredPath := filepath.Clean(absIgnoredPath)

		// Check if the targetPath is the ignoredPath itself or a sub-path of the ignoredPath
		if cleanTargetPath == cleanIgnoredPath || strings.HasPrefix(cleanTargetPath, cleanIgnoredPath+string(os.PathSeparator)) {
			logger.Log.Debugf("Path %s is ignored because it's under %s", targetPath, ignored)
			return true
		}
	}
	return false
}

// IsPathIgnored is a simple wrapper function that calls ContainsPath.
// It provides a more descriptive name for the common use case.
func IsPathIgnored(path string, ignorePaths []string) bool {
	return ContainsPath(path, ignorePaths)
}

// FindPaths searches for application bundles and associated data in a list of root directories.
// This function is specifically designed to support the application uninstallation logic.
//
// Parameters:
//   - rootDirs: A slice of root directories to begin the search (e.g., "/Applications").
//   - appName: The name of the application, including the ".app" suffix.
//
// Returns:
//   - A sorted slice of unique paths found.
func FindPaths(basePaths []string, name string) []string {
	var foundPaths []string
	for _, basePath := range basePaths {
		// Use filepath.Join to ensure correct path separators
		pattern := filepath.Join(basePath, name)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.Log.Debugf("Error globbing pattern %s: %v", pattern, err)
			continue
		}
		for _, match := range matches {
			foundPaths = append(foundPaths, match)
		}
	}
	// Also check common Application Support/Caches/Preferences directories within user Library
	if strings.HasSuffix(name, ".app") { // Only for .app bundles, look for associated files
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Log.Debugf("Could not get user home directory: %v", err)
		} else {
			// Remove the .app suffix for finding related directories
			appName := strings.TrimSuffix(name, ".app")

			// Common locations for leftover files
			appDataPaths := []string{
				filepath.Join(homeDir, "Library", "Application Support", appName),
				filepath.Join(homeDir, "Library", "Caches", appName),
				filepath.Join(homeDir, "Library", "Preferences", "com."+strings.ToLower(strings.ReplaceAll(appName, " ", ""))+".*"), // Common preference format
				filepath.Join(homeDir, "Library", "Saved Application State", "com."+strings.ToLower(strings.ReplaceAll(appName, " ", ""))+".*"),
				filepath.Join(homeDir, "Library", "Containers", "*"+appName+"*"),       // Sandboxed app containers
				filepath.Join(homeDir, "Library", "Group Containers", "*"+appName+"*"), // Group containers
				filepath.Join("/Library", "Application Support", appName),              // System-wide
				filepath.Join("/Library", "Caches", appName),                           // System-wide
				filepath.Join("/Library", "Preferences", "com."+strings.ToLower(strings.ReplaceAll(appName, " ", ""))+".*"),
			}
			for _, p := range appDataPaths {
				matches, err := filepath.Glob(p)
				if err != nil {
					logger.Log.Debugf("Error globbing app data pattern %s: %v", p, err)
					continue
				}
				for _, match := range matches {
					// Check if the path actually exists to avoid adding non-existent paths from glob patterns
					if _, err := os.Stat(match); err == nil {
						foundPaths = append(foundPaths, match)
					}
				}
			}
		}
	}

	// Remove duplicates and sort for consistent output
	uniquePaths := make(map[string]bool)
	var result []string
	for _, p := range foundPaths {
		if _, ok := uniquePaths[p]; !ok {
			uniquePaths[p] = true
			result = append(result, p)
		}
	}
	sort.Strings(result)
	return result
}
