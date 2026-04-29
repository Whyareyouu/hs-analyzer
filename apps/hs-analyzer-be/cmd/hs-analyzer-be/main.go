package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
	"hsanalyzer/apps/hs-analyzer-be/internal/storage/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	envLogsPath    = "HS_ANALYZER_LOGS_PATH"
	envCardsDBPath = "HS_ANALYZER_CARDS_DB_PATH"
	envDatabaseURL = "HS_ANALYZER_DATABASE_URL"
)

func main() {
	store, closeStore := openMatchStore()
	defer closeStore()
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

func openMatchStore() (api.MatchStore, func()) {
	databaseURL := strings.TrimSpace(os.Getenv(envDatabaseURL))
	if databaseURL == "" {
		log.Printf("%s is not set; using in-memory match store", envDatabaseURL)
		return api.NewMemoryStore(), func() {}
	}

	db, err := openPostgres(databaseURL)
	if err != nil {
		log.Fatalf("open postgres store: %v", err)
	}
	return postgres.NewStore(db), func() {
		_ = db.Close()
	}
}

func openPostgres(databaseURL string) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ensurePostgresDatabase(ctx, databaseURL); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := postgres.ApplyMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	log.Printf("postgres match store enabled")
	return db, nil
}

func ensurePostgresDatabase(ctx context.Context, databaseURL string) error {
	target, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}
	databaseName := strings.Trim(strings.TrimSpace(target.Path), "/")
	if databaseName == "" {
		return nil
	}

	targetDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	if err := targetDB.PingContext(ctx); err == nil {
		_ = targetDB.Close()
		return nil
	}
	_ = targetDB.Close()

	maintenanceURL := *target
	maintenanceURL.Path = "/postgres"
	maintenanceDB, err := sql.Open("pgx", maintenanceURL.String())
	if err != nil {
		return err
	}
	defer func() {
		_ = maintenanceDB.Close()
	}()
	if err := maintenanceDB.PingContext(ctx); err != nil {
		return err
	}
	if _, err := maintenanceDB.ExecContext(ctx, "CREATE DATABASE "+quotePostgresIdentifier(databaseName)); err != nil {
		return err
	}
	return nil
}

var postgresIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func quotePostgresIdentifier(value string) string {
	if postgresIdentifierPattern.MatchString(value) {
		return `"` + value + `"`
	}
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
