package reclaimer

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kodelint/wiper/pkg/logger"
	"github.com/kodelint/wiper/pkg/utils"
	"os"
	"sort"
)

// ====================================================================================================
// DATA STRUCTURES
// ====================================================================================================

// ReclaimedEntry represents a single entry in the cleanup summary before aggregation.
// It holds all the details of one file or directory that was processed.
type ReclaimedEntry struct {
	Path          string // The path of the file or directory that was processed.
	SizeReclaimed int64  // The size of the file/directory.
	WasRemoved    bool   // A boolean flag indicating if the item was actually deleted.
	Category      string // The high-level category of the item (e.g., "User Cache", "Application Bundle").
}

// SummaryTable holds all the ReclaimedEntry items for a single cleanup operation.
// This is used to generate the final summary report.
type SummaryTable struct {
	Entries []ReclaimedEntry
}

// ====================================================================================================
// CONSTRUCTOR AND METHODS
// ====================================================================================================

// NewSummaryTable creates a new empty SummaryTable instance.
// This is the standard way to initialize a summary report.
func NewSummaryTable() *SummaryTable {
	return &SummaryTable{}
}

// AddEntry adds a new entry to the summary table.
// path: The file or directory path.
// sizeReclaimed: The size in bytes.
// wasRemoved: True if the item was deleted, false otherwise (e.g., in a dry run).
// category: The category of the item for aggregation and display.
func (st *SummaryTable) AddEntry(path string, sizeReclaimed int64, wasRemoved bool, category string) {
	st.Entries = append(st.Entries, ReclaimedEntry{
		Path:          path,
		SizeReclaimed: sizeReclaimed,
		WasRemoved:    wasRemoved,
		Category:      category,
	})
}

// TotalReclaimedBytes calculates the total bytes reclaimed from all entries in the summary table.
func (st *SummaryTable) TotalReclaimedBytes() int64 {
	var total int64
	for _, entry := range st.Entries {
		total += entry.SizeReclaimed
	}
	return total
}

// PrintTable renders and prints a formatted summary table to standard output.
// It groups entries by category for a clean, readable report.
//
// Parameters:
//   - title: The title of the summary table.
func (st *SummaryTable) PrintTable(dryRun bool, title string) {
	// If there are no entries, there's nothing to display.
	if len(st.Entries) == 0 {
		logger.Log.Debugf("No files or directories were processed for cleanup.")
		return
	}

	//	// Older Logic
	//Step 1: Group entries by category to aggregate totals.
	//groupedTotals := make(map[string]int64)
	//for _, entry := range st.Entries {
	//	groupedTotals[entry.Category] += entry.SizeReclaimed
	//}

	// Older Logic
	//Step 2: Sort categories for a consistent and predictable table order.
	//categories := make([]string, 0, len(groupedTotals))
	//for category := range groupedTotals {
	//	categories = append(categories, category)
	//}

	// Step 1: Group entries by category to aggregate totals. {New}
	groupedTotals := make(map[string]int64)
	if dryRun {
		for _, entry := range st.Entries {
			groupedTotals[entry.Category] += entry.SizeReclaimed
		}
	} else {
		for _, entry := range st.Entries {
			// Only aggregate space from items that were actually removed.
			if entry.WasRemoved {
				groupedTotals[entry.Category] += entry.SizeReclaimed
			}
		}
	}

	// Step 2: Sort categories for a consistent and predictable table order. {New}
	categories := make([]string, 0, len(groupedTotals))
	for category := range groupedTotals {
		categories = append(categories, category)
	}

	sort.Strings(categories)
	// Step 3: Configure and render the table using the `go-pretty/v6/table` library.
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	// Add a newline for better visual separation.
	println("")
	tw.SetTitle(title)
	tw.AppendHeader(table.Row{utils.Blue("CATEGORY"), utils.Blue("RECLAIMED")})
	// Use a dark table style that works well with colored text.
	tw.SetStyle(table.StyleColoredDark)

	for _, category := range categories {
		totalSize := groupedTotals[category]
		tw.AppendRow(table.Row{category, utils.Green(utils.FormatBytes(totalSize))})
	}
	// Step 4: Add a footer row with the total reclaimed size.
	tw.AppendFooter(table.Row{utils.Blue("TOTAL RECLAIMED:"), utils.Blue(utils.FormatBytes(st.TotalReclaimedBytes()))})

	tw.Render()
}

// ====================================================================================================
// HELPER FUNCTIONS
// ====================================================================================================

// FormatBytes is a convenience function that wraps the `utils.FormatBytes` function.
// It is exposed here to be used directly by other packages that import `reclaimer`.
func FormatBytes(b int64) string {
	return utils.FormatBytes(b)
}
