package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"lowkey/internal/daemon"
	"lowkey/internal/state"
	"lowkey/pkg/config"
	"lowkey/pkg/output"
)

var (
	// rootCmd is the root command for the lowkey CLI. All other commands are
	// added as subcommands to it.
	rootCmd = &cobra.Command{Use: "lowkey", Short: "Filesystem monitor toolkit"}
	// cfgFile holds the path to the configuration file.
	cfgFile string
	// appConfig is the Viper instance used for managing application configuration.
	appConfig = viper.New()
	// manifestFromConfig stores the daemon manifest loaded from the config file.
	manifestFromConfig *config.Manifest
	// outputFormat determines the format for command output (e.g., "plain", "json").
	outputFormat = "plain"
	// outputRenderer is the renderer instance used for printing command output.
	outputRenderer output.Renderer
)

// init initializes the command-line interface, setting up commands and
// configuration handling.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(
		newWatchCmd(),
		newStartCmd(),
		newStopCmd(),
		newStatusCmd(),
		newLogCmd(),
		newTailCmd(),
		newSummaryCmd(),
		newClearCmd(),
	)
}

// execute is the main entry point for the CLI client. It parses global flags,
// sets up the output renderer, and executes the appropriate command.
func execute(args []string) error {
	var remaining []string
	var err error
	cfgFile, remaining, err = parseConfigFlag(args)
	if err != nil {
		return err
	}

	format, remaining := extractOption(remaining, "--output", "-o")
	if format != "" {
		outputFormat = format
	}

	rootCmd.SetArgs(remaining)
	cobra.ExecuteInitializers()
	if err := ensureRenderer(); err != nil {
		return err
	}
	return rootCmd.Execute()
}

// initConfig loads the application configuration from a file and sets up Viper
// to read from environment variables. It searches for a configuration file in
// standard locations if one is not specified explicitly.
func initConfig() {
	// This function initializes the application configuration. It searches for a
	// configuration file in standard locations (e.g., user's home directory)
	// and loads it if found. It also sets up Viper for environment variable
	// support.
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			candidate := filepath.Join(home, ".lowkey.json")
			if _, err := os.Stat(candidate); err == nil {
				cfgFile = candidate
			}
		}
	}

	tryPaths := []string{}
	if cfgFile != "" {
		tryPaths = append(tryPaths, cfgFile)
	} else {
		if stateDir, err := state.DefaultStateDir(); err == nil {
			tryPaths = append(tryPaths, filepath.Join(stateDir, "daemon.json"))
		}
	}

	for _, candidate := range tryPaths {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		manifest, err := config.LoadManifest(candidate)
		if err != nil {
			continue
		}
		manifestFromConfig = manifest
		cfgFile = candidate
		break
	}

	if cfgFile != "" {
		appConfig.SetConfigFile(cfgFile)
		_ = appConfig.ReadInConfig()
	}
	appConfig.AutomaticEnv()
}

// parseConfigFlag manually parses the --config flag from the arguments list.
// This is necessary to ensure the config file is loaded by initConfig before
// Cobra parses the rest of the flags.
func parseConfigFlag(args []string) (string, []string, error) {
	var cfg string
	remaining := make([]string, 0, len(args))

	skip := false
	for i, arg := range args {
		if skip {
			skip = false
			continue
		}

		switch {
		case arg == "--config" || arg == "-c":
			if i+1 >= len(args) {
				return "", nil, errors.New("--config flag requires a value")
			}
			cfg = args[i+1]
			skip = true
		case len(arg) > 9 && arg[:9] == "--config=":
			cfg = arg[9:]
		default:
			remaining = append(remaining, arg)
		}
	}

	return cfg, remaining, nil
}

// loadWatchTargetsFromConfig retrieves the list of directories to watch from the
// manifest that was loaded from the configuration file.
func loadWatchTargetsFromConfig() []string {
	if manifestFromConfig != nil {
		return append([]string(nil), manifestFromConfig.Directories...)
	}
	return nil
}

// ensureRenderer initializes the output renderer if it hasn't been already.
// This ensures that commands can safely use the renderer to print output.
func ensureRenderer() error {
	if outputRenderer != nil {
		return nil
	}
	renderer, err := output.NewRenderer(outputFormat)
	if err != nil {
		return err
	}
	outputRenderer = renderer
	return nil
}

// renderStatus uses the configured output renderer to display the daemon's status.
func renderStatus(status daemon.ManagerStatus) error {
	if err := ensureRenderer(); err != nil {
		return err
	}
	return outputRenderer.Status(status)
}

// extractOption manually parses a key-value option from the arguments list.
// This is used for options that need to be processed before Cobra's parsing,
// such as the --output format.
func extractOption(args []string, keys ...string) (string, []string) {
	if len(keys) == 0 {
		return "", args
	}

	remaining := make([]string, 0, len(args))
	var value string
	skip := false
	for i, arg := range args {
		if skip {
			skip = false
			continue
		}

		matched := ""
		for _, key := range keys {
			if arg == key {
				matched = key
				break
			}
			if strings.HasPrefix(arg, key+"=") {
				value = arg[len(key)+1:]
				matched = key
				break
			}
		}

		if matched != "" {
			if value == "" {
				if i+1 >= len(args) {
					value = ""
				} else if strings.HasPrefix(args[i+1], "-") {
					value = ""
				} else {
					value = args[i+1]
					skip = true
				}
			}
			continue
		}

		remaining = append(remaining, arg)
	}

	return value, remaining
}
