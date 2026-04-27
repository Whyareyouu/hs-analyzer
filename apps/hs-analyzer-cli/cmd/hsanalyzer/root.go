package main

import (
	"context"

	"github.com/spf13/cobra"
)

const appVersion = "0.0.2"

func newRootCmd() *cobra.Command {
	options := rootOptions{logLevel: "info"}

	cmd := &cobra.Command{
		Use:           "hsanalyzer",
		Short:         "Hearthstone log analyzer",
		Version:       appVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			logger, err := buildLogger(options.logLevel)
			if err != nil {
				return err
			}

			ctx := context.WithValue(cmd.Context(), loggerContextKey{}, logger)
			cmd.SetContext(ctx)
			cmd.Root().SetContext(ctx)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			_ = syncLogger(loggerFromContext(cmd.Context()))
		},
	}

	cmd.PersistentFlags().StringVar(&options.logLevel, "log-level", options.logLevel, "Log level: debug, info, warn, error")
	cmd.AddCommand(newAnalyzeCmd())
	cmd.AddCommand(newUpdateCardsCmd())
	cmd.AddCommand(newMonitorCmd())

	return cmd
}

type rootOptions struct {
	logLevel string
}
