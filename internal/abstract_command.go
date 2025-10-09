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

// handleVerboseFlag manually counts and removes verbose flags.
func (cmd *AbstractCommand) handleVerboseFlag() {
	verboseCount := 0
	remainingArgs := []string{}
	for _, arg := range cmd.Args {
		if arg == "-v" || arg == "--verbose" {
			verboseCount++
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	if verboseCount > 0 {
		SetVerbosity(GetVerbosity() + verboseCount)
	}
	cmd.Args = remainingArgs
}

// ParseCommonFlags parses common flags like -h, -q, -s.
// Returns the remaining non-flag arguments.
func (cmd *AbstractCommand) ParseCommonFlags(helpFunc func()) []string {
	// Manually handle verbose flags before parsing other flags
	cmd.handleVerboseFlag()

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help := fs.Bool("h", false, "Show this help message and exit")
	fs.BoolVar(help, "help", false, "Show this help message and exit")

	quiet := fs.Bool("q", false, "Suppress normal output, only show errors")
	fs.BoolVar(quiet, "quiet", false, "Suppress normal output, only show errors")

	serial := fs.Bool("s", false, "Run tasks serially in a single thread")
	fs.BoolVar(serial, "serial", false, "Run tasks serially in a single thread")

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
	// Manually handle verbose flags before parsing other flags
	cmd.handleVerboseFlag()

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	help := fs.Bool("h", false, "Show this help message and exit")
	fs.BoolVar(help, "help", false, "Show this help message and exit")

	quiet := fs.Bool("q", false, "Suppress normal output, only show errors")
	fs.BoolVar(quiet, "quiet", false, "Suppress normal output, only show errors")

	serial := fs.Bool("s", false, "Run tasks serially in a single thread")
	fs.BoolVar(serial, "serial", false, "Run tasks serially in a single thread")

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
