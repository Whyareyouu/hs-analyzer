package main

import (
	"bufio"
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
	pgstore "hsanalyzer/apps/hs-analyzer-be/internal/storage/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	envLogsPath    = "HS_ANALYZER_LOGS_PATH"
	envCardsDBPath = "HS_ANALYZER_CARDS_DB_PATH"
	envDatabaseURL = "HS_ANALYZER_DATABASE_URL"
)

func main() {
	loadDotEnv(".env")

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
		log.Fatalf("%s is not set; configure it in .env or environment", envDatabaseURL)
	}
	log.Printf("using postgres %s", redactDatabaseURL(databaseURL))

	db, err := openPostgres(databaseURL)
	if err != nil {
		log.Fatalf("open postgres store: %v", err)
	}
	return pgstore.NewStore(db), func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("read .env skipped: %v", err)
		}
		return
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || strings.TrimSpace(os.Getenv(name)) != "" {
			continue
		}
		if len(value) >= 2 {
			first := value[0]
			last := value[len(value)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if err := os.Setenv(name, value); err != nil {
			log.Printf("set .env value %s skipped: %v", name, err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scan .env skipped: %v", err)
	}
}

func redactDatabaseURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.User == nil {
		return rawURL
	}
	if username := parsed.User.Username(); username != "" {
		parsed.User = url.UserPassword(username, "xxxxx")
	}
	return parsed.String()
}

func openPostgres(databaseURL string) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ensurePostgresDatabase(ctx, databaseURL); err != nil {
		return nil, err
	}

	db, err := gorm.Open(gormpostgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	if err := pgstore.ApplyMigrations(ctx, db); err != nil {
		_ = sqlDB.Close()
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
