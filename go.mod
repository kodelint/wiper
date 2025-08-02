module github.com/kodelint/wiper

go 1.24.5

require (
	// fatih/color is used for adding color-coded output to the terminal,
	// making log messages and tables more readable.
	github.com/fatih/color v1.18.0

	// jedib0t/go-pretty is a powerful library for generating formatted tables
	// in the terminal, used for displaying cleanup summaries.
	github.com/jedib0t/go-pretty/v6 v6.6.7

	// spf13/cobra is the popular library for creating powerful and modern CLI
	// applications. It handles commands, flags, and arguments.
	github.com/spf13/cobra v1.9.1

	// Indirect dependencies required by the direct dependencies above.
	// They are automatically managed by the Go toolchain.
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13       // indirect
	github.com/mattn/go-isatty v0.0.20          // indirect
	github.com/mattn/go-runewidth v0.0.16       // indirect
	github.com/rivo/uniseg v0.4.7               // indirect
	github.com/sirupsen/logrus v1.9.3           // indirect
	github.com/spf13/pflag v1.0.6               // indirect
	golang.org/x/sys v0.30.0                      // indirect
	golang.org/x/text v0.22.0                     // indirect
)