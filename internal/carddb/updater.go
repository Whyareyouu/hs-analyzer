package carddb

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	defaultCardsHost = "https://api.hearthstonejson.com"
)

// DownloadLatestCards скачивает актуальную базу карт HearthstoneJSON.
func DownloadLatestCards(locale, outFile string) error {
	if locale == "" {
		locale = "ruRU"
	}
	if outFile == "" {
		outFile = filepath.Join("data", "cards.json")
	}

	url := fmt.Sprintf("%s/v1/latest/%s/cards.json", defaultCardsHost, locale)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download cards from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download cards from %s: unexpected status %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("save cards file: %w", err)
	}

	return nil
}
