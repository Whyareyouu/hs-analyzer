package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"hs-analyzer/internal/timeline"
)

func ExportJSON(tl timeline.Timeline, path string) error {
	content, err := json.MarshalIndent(tl, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func ExportText(tl timeline.Timeline, path string) error {
	var b strings.Builder
	b.WriteString("Hearthstone Match Timeline\n")
	b.WriteString("==========================\n\n")

	for _, turn := range tl.Turns {
		b.WriteString(fmt.Sprintf("Turn %d (%s)\n", turn.TurnNumber, turn.Player))
		if len(turn.Actions) == 0 {
			b.WriteString("- no actions\n")
		}
		for _, action := range turn.Actions {
			b.WriteString("- " + action + "\n")
		}
		b.WriteString("\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}
