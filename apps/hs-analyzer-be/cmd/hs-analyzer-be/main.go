package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
)

const (
	envLogsPath    = "HS_ANALYZER_LOGS_PATH"
	envCardsDBPath = "HS_ANALYZER_CARDS_DB_PATH"
)

func main() {
	store := api.NewMemoryStore()
	dashboardStore := api.NewMatchDashboardStore(store)

	logsPath := strings.TrimSpace(os.Getenv(envLogsPath))
	cardsDBPath := strings.TrimSpace(os.Getenv(envCardsDBPath))
	if logsPath != "" {
		if cardsDBPath == "" {
			cardsDBPath = "data/cards.json"
		}
		source, err := api.NewFileDeckCatalogSource(logsPath, cardsDBPath)
		if err != nil {
			log.Printf("deck catalog disabled: %v", err)
		} else {
			dashboardStore = api.NewMatchDashboardStore(store, source)
		}
	}

	server := api.NewServer(store, dashboardStore)

	log.Println("hs-analyzer-be listening on 127.0.0.1:8080")
	if err := http.ListenAndServe("127.0.0.1:8080", server.Routes()); err != nil {
		log.Fatal(err)
	}
}
