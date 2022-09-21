package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"github.com/JenswBE/dead-link-checker/cmd/config"
	"github.com/JenswBE/dead-link-checker/internal"
)

func main() {
	// Parse flags
	verbose := flag.BoolP("verbose", "v", false,
		"Enables verbose output. Will be enabled if either this flag, config option or env var is provided.")
	configPath := flag.StringP("config", "c", "./config.yml", "Path to the config file")
	printJSON := flag.Bool("json", false, "Print all site reports as JSON to stdout")
	runNow := flag.Bool("now", false, "Overrides cron and forces an immediate check")
	flag.Parse()

	// Setup logging
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Parse config
	delicConfig, err := config.ParseConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	// Setup Debug logging if enabled
	if delicConfig.Verbose || *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	// Create manager
	manager := internal.NewManager()

	// Run DeLiC
	if delicConfig.Cron == "" || *runNow {
		// Run once
		if err = runDeLiC(context.Background(), manager, delicConfig, *printJSON); err != nil {
			log.Fatal().Err(err).Msg("Error while running DeLiC")
		}
	} else {
		// Run at cron interval
		log.Info().Str("spec", delicConfig.Cron).Msg("DeLiC started with cron")
		tasker.
			New(tasker.Option{Verbose: delicConfig.Verbose}).
			Task(delicConfig.Cron, newDeLiCTask(manager, delicConfig, *printJSON)).
			Run()
	}
}

func runDeLiC(ctx context.Context, manager *internal.Manager, delicConfig *config.Config, printJSON bool) error {
	// Run manager
	reports := manager.Run(ctx, delicConfig)

	// Print JSON results if enabled
	if printJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(reports); err != nil {
			// Both log and return error to have correct severity in logs
			log.Error().Err(err).Msg("Failed to print reports as JSON to stdout")
			return fmt.Errorf("failed to print reports as JSON to stdout: %w", err)
		}
	}
	return nil
}

func newDeLiCTask(manager *internal.Manager, delicConfig *config.Config, printJSON bool) tasker.TaskFunc {
	return func(ctx context.Context) (int, error) {
		if err := runDeLiC(ctx, manager, delicConfig, printJSON); err != nil {
			return 1, err
		}
		return 0, nil
	}
}
