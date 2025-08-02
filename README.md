# Wiper

A powerful command-line utility for macOS to completely uninstall applications and clean up system junk.

---

![Go Version](https://img.shields.io/badge/go-1.24.5-blue.svg)
[![Go CI](https://github.com/kodelint/wiper/actions/workflows/release.yml/badge.svg)](https://github.com/kodelint/wiper/actions/workflows/go.yml)
[![GitHub release](https://img.shields.io/github/v/release/kodelint/wiper)](https://github.com/kodelint/wiper/releases)
[![GitHub stars](https://img.shields.io/github/stars/kodelint/wiper.svg)](https://github.com/kodelint/wiper/stargazers)
[![Last commit](https://img.shields.io/github/last-commit/kodelint/wiper.svg)](https://github.com/kodelint/wiper/commits/main)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/kodelint/wiper/pulls)

<p align="center">
  <img src="https://raw.githubusercontent.com/kodelint/blog-images/main/common/01-Wiper.png" width="600" height="600" alt="Wiper Command-Line Output" />
</p>

### Key Features

* **Complete Application Uninstallation**: Wiper not only removes the main `.app` bundle but also intelligently finds and deletes associated caches, temporary files, and configuration data scattered across your system.
* **Comprehensive System Cleanup**: Optimize your macOS performance by removing old and unnecessary files from common locations like `/tmp`, user and system caches, logs, and more.
* **Large File Cleanup**: Quickly identify and remove unusually large files (over 100MB) from directories like `~/Downloads` and `~/Documents`.
* **Dry-Run Mode**: Safely preview all files and directories that would be removed using the `--dry-run` flag before committing to any changes.
* **Interactive Control**: Gain granular control over the cleanup process with the `--interactive` flag, which prompts you for confirmation before deleting each individual file or directory.
* **Path Exclusion**: Use the `--ignore` flag to specify a comma-separated list of paths that you want to exclude from the cleanup process.
* **Clear Reporting**: All cleanup operations conclude with a summary table that clearly shows the total disk space reclaimed.

## Installation

### Prerequisites

You need to have Go version `1.24.5` or higher installed to build and run Wiper.

### From Source

You can install Wiper directly using the Go command-line tool.

```sh
go install github.com/kodelint/wiper@latest
```
This command will download the source code, build the binary, and place it in your `$GOPATH/bin` directory. Ensure this directory is in your system's `PATH`.

## Usage

### General Syntax
```bash
wiper [command] [arguments] [flags]
```

```Bash
wiper is a comprehensive macOS utility designed to help you:

1.  Uninstall Applications: Completely remove applications and all their leftover files, reclaiming disk space.
2.  Clean System Junk: Remove temporary files, caches (system, user, browser), and other unnecessary data to optimize your macOS performance.

It provides detailed output and supports dry-run modes to show you what will be removed before any changes are made.

Usage:
  wiper [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Show the wiper tool version
  wipe        Uninstall an application or clean up the system.

Flags:
  -d, --debug           Enable debug logging.
  -n, --dry-run         Perform a dry run without making any changes.
  -h, --help            help for wiper
  -i, --ignore string   Comma-separated list of paths to ignore during cleanup.

Use "wiper [command] --help" for more information about a command.
```

### Examples

#### Uninstall a application

```Bash
# Uninstall Google Chrome and all its associated files
wiper wipe "Google Chrome"

# You can also specify the .app extension
wiper wipe "Google Chrome.app"
```

#### Perform System-wide Cleanup
To run a general system cleanup of caches, temporary files, and logs, run `wipe` without any arguments.
```Bash
# Clean up system junk
wiper wipe
```

#### Find and Clean Large Files
To perform a cleanup specifically targeting large files (over **100MB**), use the `--large-files` flag.
```Bash
# Find and clean large files with a dry-run
wiper wipe --large-files --dry-run
```

#### Using Global Flags
`Wiper` includes several global flags that can be used with any command.

```bash
# Perform a dry-run of an application uninstall
wiper wipe "Spotify" --dry-run

# Run a system cleanup while ignoring a specific path
wiper wipe --ignore "~/Library/Caches/important-data, /private/var/folders/other-stuff"

# Enable debug logging for a detailed look at the process
wiper wipe --debug
```

### Commands

#### `wipe`
The primary command for all cleanup operations.

| Argument             | Description                                                                               |
|----------------------|-------------------------------------------------------------------------------------------|
| `[application-name]` | The name of the application to uninstall. If omitted, a system-wide cleanup is performed. |

| Flag            | Shortcut | Description                                                                                          |
|-----------------|----------|------------------------------------------------------------------------------------------------------|
| `--large-files` | None     | Perform a cleanup of large files instead of a standard system cleanup.                               |
| `--interactive` | `-i`     | Use interactive mode for large file cleanup, prompting for confirmation before each file is deleted. |

#### `version`
Displays the current version of the **Wiper** tool. Also check if there is new release

```bash
wiper version
```

### Global Flags
| Flag        | Shortcut | Description                                                                                                |
|-------------|----------|------------------------------------------------------------------------------------------------------------|
| `--debug`   | `-d`     | Enables debug logging, providing verbose output about the tool's actions.                                  |
| `--dry-run` | `-n`     | Simulates the cleanup process without deleting any files. A summary of what would be removed is displayed. |
| `--ignore`  | `-e`     | A comma-separated list of paths to exclude from cleanup. Supports `~` and environment variable `$HOME.`    |

---

### Contributing
If you encounter a bug, have a feature request, or want to submit a code change, please open an issue or a pull request on the GitHub repository.