package main

import (
	"fmt"

	"hsanalyzer/internal/carddb"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newUpdateCardsCmd() *cobra.Command {
	locale := "ruRU"
	out := "data/cards.json"

	cmd := &cobra.Command{
		Use:   "update-cards",
		Short: "Download latest Hearthstone cards database",
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := loggerFromContext(cmd.Context())
			logger.Info("starting card database update",
				zap.String("locale", locale),
				zap.String("output", out),
			)
			fmt.Fprintf(cmd.OutOrStdout(), "Downloading latest cards for locale %s...\n", locale)
			if err := carddb.DownloadLatestCards(locale, out); err != nil {
				logger.Error("card database update failed", zap.Error(err))
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Cards database saved to %s\n", out)
			logger.Info("card database update completed",
				zap.String("locale", locale),
				zap.String("output", out),
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&locale, "locale", locale, "Hearthstone locale")
	cmd.Flags().StringVar(&out, "out", out, "Output file path")

	return cmd
}
