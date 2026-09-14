package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/migrate [up|down|version|force]")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("failed to initialize migrations: %v", err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()

		if sourceErr != nil {
			log.Printf("failed to close migration source: %v", sourceErr)
		}

		if databaseErr != nil {
			log.Printf("failed to close migration database: %v", databaseErr)
		}
	}()

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migration up failed: %v", err)
		}

		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Steps(-1); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}

		fmt.Println("Last migration rolled back successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("No migrations applied")
				return
			}

			log.Fatalf("failed to get migration version: %v", err)
		}

		fmt.Printf("Version: %d, Dirty: %t\n", version, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: go run ./cmd/migrate force <version>")
		}

		var version int
		if _, err := fmt.Sscanf(os.Args[2], "%d", &version); err != nil {
			log.Fatalf("invalid migration version: %v", err)
		}

		if err := m.Force(version); err != nil {
			log.Fatalf("failed to force migration version: %v", err)
		}

		fmt.Printf("Migration version forced to %d\n", version)
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}
