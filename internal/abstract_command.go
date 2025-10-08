package internal

import (
	"flag"
	"fmt"
	"os"
)

// AbstractCommand provides common functionality for all git-tree commands.
type AbstractCommand struct {
	Config         *Config
	Args           []string
	Serial         bool
	AllowEmptyArgs bool
}

// NewAbstractCommand creates a new AbstractCommand instance.
func NewAbstractCommand(args []string, allowEmptyArgs bool) *AbstractCommand {
	cmd := &AbstractCommand{
		Config:         NewConfig(),
		Args:           args,
		AllowEmptyArgs: allowEmptyArgs,
	}

	// Set initial verbosity from config
	SetVerbosity(cmd.Config.Verbosity)

	return cmd
}

// ParseCommonFlags parses common flags like -h, -q, -v, -s.
// Returns the remaining non-flag arguments.
func (cmd *AbstractCommand) ParseCommonFlags(helpFunc func()) []string {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help := fs.Bool("h", false, "Show this help message and exit")
	fs.BoolVar(help, "help", false, "Show this help message and exit")

	quiet := fs.Bool("q", false, "Suppress normal output, only show errors")
	fs.BoolVar(quiet, "quiet", false, "Suppress normal output, only show errors")

	serial := fs.Bool("s", false, "Run tasks serially in a single thread")
	fs.BoolVar(serial, "serial", false, "Run tasks serially in a single thread")

	verboseCount := 0
	fs.Func("v", "Increase verbosity (can be used multiple times)", func(s string) error {
		verboseCount++
		return nil
	})
	fs.Func("verbose", "Increase verbosity (can be used multiple times)", func(s string) error {
		verboseCount++
		return nil
	})

	// Parse the flags
	if err := fs.Parse(cmd.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Handle help
	if *help {
		helpFunc()
		os.Exit(0)
	}

	// Handle quiet
	if *quiet {
		SetVerbosity(LogQuiet)
	}

	// Handle verbose
	if verboseCount > 0 {
		SetVerbosity(GetVerbosity() + verboseCount)
	}

	// Handle serial
	cmd.Serial = *serial

	// Get remaining args
	remainingArgs := fs.Args()

	// Check if empty args are allowed
	if !cmd.AllowEmptyArgs && len(remainingArgs) == 0 {
		Log(LogQuiet, "Error: No arguments provided", ColorRed)
		helpFunc()
		os.Exit(1)
	}

	return remainingArgs
}

// ParseFlagsWithCallback parses flags with a custom callback for additional flags.
// The callback receives a FlagSet to add custom flags.
func (cmd *AbstractCommand) ParseFlagsWithCallback(helpFunc func(), callback func(*flag.FlagSet)) []string {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help := fs.Bool("h", false, "Show this help message and exit")
	fs.BoolVar(help, "help", false, "Show this help message and exit")

	quiet := fs.Bool("q", false, "Suppress normal output, only show errors")
	fs.BoolVar(quiet, "quiet", false, "Suppress normal output, only show errors")

	serial := fs.Bool("s", false, "Run tasks serially in a single thread")
	fs.BoolVar(serial, "serial", false, "Run tasks serially in a single thread")

	verboseCount := 0
	fs.Func("v", "Increase verbosity (can be used multiple times)", func(s string) error {
		verboseCount++
		return nil
	})
	fs.Func("verbose", "Increase verbosity (can be used multiple times)", func(s string) error {
		verboseCount++
		return nil
	})

	// Allow custom flags
	if callback != nil {
		callback(fs)
	}

	// Parse the flags
	if err := fs.Parse(cmd.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Handle help
	if *help {
		helpFunc()
		os.Exit(0)
	}

	// Handle quiet
	if *quiet {
		SetVerbosity(LogQuiet)
	}

	// Handle verbose
	if verboseCount > 0 {
		SetVerbosity(GetVerbosity() + verboseCount)
	}

	// Handle serial
	cmd.Serial = *serial

	// Get remaining args
	remainingArgs := fs.Args()

	// Check if empty args are allowed
	if !cmd.AllowEmptyArgs && len(remainingArgs) == 0 {
		Log(LogQuiet, "Error: No arguments provided", ColorRed)
		helpFunc()
		os.Exit(1)
	}

	return remainingArgs
}
