//package cmd
//
//import (
//	"fmt"                    // The fmt package is used for formatted I/O, specifically to print the version string.
//	"github.com/spf13/cobra" // The cobra library is used to define and manage the command-line interface.
//)
//
//// ====================================================================================================
//// VERSION COMMAND DEFINITION
//// ====================================================================================================
//
//// versionCmd represents the version command.
//// It's a subcommand of the RootCmd and is responsible for showing the tool's version.
//var versionCmd = &cobra.Command{
//	// Use: defines the command's name, which is how it's invoked from the command line (e.g., `wiper version`).
//	Use: "version",
//	// Short: provides a brief description of the command, shown in the help output.
//	Short: "Show the wiper tool version",
//	// Run: is the main function executed when the 'version' command is called.
//	Run: func(cmd *cobra.Command, args []string) {
//		// This line prints the current version string.
//		// NOTE: Currently, the version is hardcoded. In a real-world application,
//		// this would typically be set dynamically at build time using a build-time variable.
//		fmt.Println("Wiper Version: 0.1.0-alpha")
//	},
//}
//
//// ====================================================================================================
//// INITIALIZATION
//// ====================================================================================================
//
//// init registers the version command with the root command.
//// This function is automatically called by the Go runtime before the main function.
//func init() {
//	// RootCmd.AddCommand() is used to add this specific subcommand to the main command-line tool.
//	// This makes `wiper version` a valid command.
//	RootCmd.AddCommand(versionCmd)
//}

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/kodelint/wiper/pkg/utils"
	"net/http"
	"strings"
	"time"

	"github.com/kodelint/wiper/pkg/logger"
	"github.com/spf13/cobra"
)

// ====================================================================================================
// VERSION COMMAND DEFINITION
// ====================================================================================================

// These variables are populated at build time using -ldflags.
// A typical build command would look like:
// go build -ldflags "-X 'github.com/kodelint/wiper/cmd.version=$(git describe --tags --always)'"
var version string = "development"

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the wiper tool version and check for updates",
	Run: func(cmd *cobra.Command, args []string) {
		// Use the version variable that is populated at build time
		fmt.Printf("Wiper Version: %s\n", version)

		// Check for a new version
		checkForNewVersion(version)
	},
}

// ====================================================================================================
// NEW VERSION CHECK LOGIC
// ====================================================================================================

const (
	repoOwner    = "kodelint"
	repoName     = "wiper"
	githubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

// githubRelease represents the relevant fields from the GitHub API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// checkForNewVersion queries the GitHub API for the latest release and compares it to the current version.
func checkForNewVersion(currentVersion string) {
	logger.Log.Debug("Checking for new version...")

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Construct the API URL
	url := fmt.Sprintf(githubAPIURL, repoOwner, repoName)

	resp, err := client.Get(url)
	if err != nil {
		logger.Log.Debugf("Failed to check for updates: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Debugf("Failed to check for updates, received status code %d", resp.StatusCode)
		return
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		logger.Log.Debugf("Failed to decode GitHub API response: %v", err)
		return
	}

	latestVersion := strings.TrimSpace(release.TagName)
	latestVersion = strings.TrimPrefix(latestVersion, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	if latestVersion != "" && latestVersion != currentVersion {
		// A more robust version comparison would be needed for complex schemes
		// but a simple string compare is often sufficient for basic use cases.
		if latestVersion > currentVersion {
			fmt.Printf("A new version is available: %s. You are using %s.\n", utils.GreenBold(release.TagName), utils.Cyan(currentVersion))
			fmt.Printf("Please download the new version from: https://github.com/%s/%s/releases\n", repoOwner, repoName)
		}
	} else {
		fmt.Println(utils.GreenBold("You are running the latest version."))
	}
}

// ====================================================================================================
// INITIALIZATION
// ====================================================================================================

// init registers the version command with the root command.
func init() {
	RootCmd.AddCommand(versionCmd)
}
