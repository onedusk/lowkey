package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd   = &cobra.Command{Use: "lowkey", Short: "Filesystem monitor toolkit"}
	cfgFile   string
	appConfig = viper.New()
)

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

func execute(args []string) error {
	var remaining []string
	var err error
	cfgFile, remaining, err = parseConfigFlag(args)
	if err != nil {
		return err
	}

	rootCmd.SetArgs(remaining)
	cobra.ExecuteInitializers()
	return rootCmd.Execute()
}

func initConfig() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cfgFile = filepath.Join(home, ".lowkey.yaml")
		}
	}

	if cfgFile == "" {
		return
	}

	appConfig.SetConfigFile(cfgFile)
	_ = appConfig.ReadInConfig() // Optional; ignore errors when bootstrapping.
	appConfig.AutomaticEnv()
}

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

func loadWatchTargetsFromConfig() []string {
	dirs := appConfig.GetString("directories")
	if dirs == "" {
		return nil
	}
	return []string{dirs}
}
