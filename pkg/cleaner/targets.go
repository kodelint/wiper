package cleaner

import (
	"path/filepath" // Imported for filepath.Join and other path manipulations
	"time"          // Imported for time.Duration

	"github.com/kodelint/wiper/pkg/utils" // Imported for utils.ExpandPath
)

// ====================================================================================================
// DATA STRUCTURES
// ====================================================================================================

// cleanupTarget defines a category of files to be cleaned. It provides all the necessary
// information for the system cleanup function to know what to look for and how to handle it.
type cleanupTarget struct {
	// Paths is a slice of glob patterns to find files and directories for this target.
	Paths []string
	// Category is a user-friendly name for the type of files being cleaned (e.g., "User Caches").
	Category string
	// MinAge is the minimum age a file must have to be considered for deletion.
	// A value of 0 means all files matching the pattern will be considered.
	MinAge time.Duration
	// LogAggregationRoots is a list of root paths used to group found items
	// in the log output for a cleaner, more readable summary table.
	LogAggregationRoots []string
}

// ====================================================================================================
// CLEANUP TARGETS CONFIGURATION
// ====================================================================================================

// getCleanupTargets initializes and returns the slice of cleanup targets.
// This function acts as the central configuration for the system cleanup feature, defining
// the specific files and directories that the tool will target for removal.
func getCleanupTargets() []cleanupTarget {
	homeDir := utils.ExpandPath("~") // Ensure homeDir is expanded once
	return []cleanupTarget{
		{
			Paths:               []string{filepath.Join(homeDir, "Library", "Caches", "TemporaryItems", "*"), "/private/var/folders/*/*/T/*"},
			Category:            "User Temporary Files",
			MinAge:              24 * time.Hour,
			LogAggregationRoots: []string{filepath.Join(homeDir, "Library", "Caches", "TemporaryItems"), "/private/var/folders"},
		},
		{
			Paths:               []string{"/private/var/tmp/*", "/tmp/*"},
			Category:            "System Temporary Files",
			MinAge:              24 * time.Hour,
			LogAggregationRoots: []string{"/private/var/tmp", "/tmp"},
		},
		{
			Paths:               []string{filepath.Join(homeDir, "Library", "Caches", "*")},
			Category:            "User Caches",
			MinAge:              0,
			LogAggregationRoots: []string{filepath.Join(homeDir, "Library", "Caches")},
		},
		{
			Paths:               []string{"/Library/Caches/*"},
			Category:            "System Caches",
			MinAge:              0,
			LogAggregationRoots: []string{"/Library/Caches"},
		},
		{
			Paths:               []string{filepath.Join(homeDir, "Library", "Logs", "*")},
			Category:            "User Logs",
			MinAge:              30 * 24 * time.Hour,
			LogAggregationRoots: []string{filepath.Join(homeDir, "Library", "Logs")},
		},
		{
			Paths: []string{
				filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "Cache", "*"),
				filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "Service Worker", "CacheStorage", "*"),
				filepath.Join(homeDir, "Library", "Caches", "Google", "Chrome", "*"),
				filepath.Join(homeDir, "Library", "Caches", "com.apple.Safari", "*"),
				filepath.Join(homeDir, "Library", "Application Support", "Firefox", "Profiles", "*", "cache2", "entries", "*"),
				filepath.Join(homeDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Cache", "*"),
				filepath.Join(homeDir, "Library", "Caches", "BraveSoftware", "Brave-Browser", "*"),
			},
			Category: "Browser Caches",
			MinAge:   0,
			LogAggregationRoots: []string{
				filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome"),
				filepath.Join(homeDir, "Library", "Caches", "Google", "Chrome"),
				filepath.Join(homeDir, "Library", "Caches", "com.apple.Safari"),
				filepath.Join(homeDir, "Library", "Application Support", "Firefox"),
				filepath.Join(homeDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser"),
				filepath.Join(homeDir, "Library", "Caches", "BraveSoftware", "Brave-Browser"),
			},
		},
		{
			Paths:               []string{filepath.Join(homeDir, ".Trash", "*")},
			Category:            "Trash Bin",
			MinAge:              0,
			LogAggregationRoots: []string{filepath.Join(homeDir, ".Trash")},
		},
		{
			Paths:               []string{filepath.Join(homeDir, "Downloads", "*")},
			Category:            "Downloads (old)",
			MinAge:              90 * 24 * time.Hour,
			LogAggregationRoots: []string{filepath.Join(homeDir, "Downloads")},
		},
	}
}
