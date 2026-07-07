package main

import (
	"arena/migrations"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	cmd := os.Args[1]

	switch cmd {
	case "up":
		runUp(dbURL)
	case "down":
		runDown(dbURL)
	case "version":
		runVersion(dbURL)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force VERSION")
		}
		runForce(dbURL, os.Args[2])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up       Apply all pending migrations")
	fmt.Println("  down     Rollback the last migration")
	fmt.Println("  version  Show current migration version")
	fmt.Println("  force    Force a specific version (for recovery)")
}

func newMigrate(dbURL string) *migrate.Migrate {
	d, err := iofs.New(migrations.Files, ".")
	if err != nil {
		log.Fatalf("failed to create migration source: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	return m
}

func runUp(dbURL string) {
	m := newMigrate(dbURL)
	defer m.Close()

	err := m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("database is already up to date")
			return
		}
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("migrations applied successfully")
}

func runDown(dbURL string) {
	m := newMigrate(dbURL)
	defer m.Close()

	err := m.Steps(-1)
	if err != nil {
		log.Fatalf("rollback failed: %v", err)
	}
	fmt.Println("migration rolled back successfully")
}

func runVersion(dbURL string) {
	m := newMigrate(dbURL)
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("no migrations applied yet (version: 0)")
			return
		}
		log.Fatalf("failed to get version: %v", err)
	}

	status := "clean"
	if dirty {
		status = "DIRTY"
	}
	fmt.Printf("current version: %d (%s)\n", version, status)
}

func runForce(dbURL string, versionStr string) {
	var version uint
	_, err := fmt.Sscanf(versionStr, "%d", &version)
	if err != nil {
		log.Fatalf("invalid version number: %s", versionStr)
	}

	m := newMigrate(dbURL)
	defer m.Close()

	err = m.Force(int(version))
	if err != nil {
		log.Fatalf("failed to force version: %v", err)
	}
	fmt.Printf("version forced to %d\n", version)
}
