package main

import (
	"flag"
	"fmt"
	"os"

	"hsanalyzer/internal/carddb"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "update-cards":
		if err := runUpdateCards(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "update-cards failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runUpdateCards(args []string) error {
	cmd := flag.NewFlagSet("update-cards", flag.ContinueOnError)
	cmd.SetOutput(os.Stderr)

	locale := cmd.String("locale", "ruRU", "Hearthstone locale (e.g. ruRU, enUS)")
	out := cmd.String("out", "data/cards.json", "Output file path")

	if err := cmd.Parse(args); err != nil {
		return err
	}

	fmt.Printf("Downloading latest cards for locale %s...\n", *locale)
	if err := carddb.DownloadLatestCards(*locale, *out); err != nil {
		return err
	}
	fmt.Printf("Cards database saved to %s\n", *out)
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `hsanalyzer CLI

Usage:
  hsanalyzer <command> [options]

Commands:
  update-cards   Download latest Hearthstone cards database

Examples:
  hsanalyzer update-cards
  hsanalyzer update-cards --locale enUS
  hsanalyzer update-cards --locale ruRU --out data/cards.json
`)
}
