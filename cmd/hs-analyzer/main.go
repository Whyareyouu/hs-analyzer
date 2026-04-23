package main

import (
	"fmt"
	"os"

	"hs-analyzer/internal/classifier"
	"hs-analyzer/internal/exporter"
	"hs-analyzer/internal/parser"
	"hs-analyzer/internal/reader"
	"hs-analyzer/internal/state"
	"hs-analyzer/internal/timeline"
)

func main() {
	input := "logs/Power.log"
	outJSON := "output/timeline.json"
	outText := "output/timeline.txt"

	lines, err := reader.ReadLines(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read logs: %v\n", err)
		os.Exit(1)
	}

	rawEvents := parser.New().ParseLines(lines)
	gameEvents := classifier.New().Classify(rawEvents)
	matchState := state.New().Build(gameEvents)
	tl := timeline.Build(matchState)

	if err := exporter.ExportJSON(tl, outJSON); err != nil {
		fmt.Fprintf(os.Stderr, "export json: %v\n", err)
		os.Exit(1)
	}
	if err := exporter.ExportText(tl, outText); err != nil {
		fmt.Fprintf(os.Stderr, "export text: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("processed %d lines, %d raw events, %d turns\n", len(lines), len(rawEvents), len(tl.Turns))
}
