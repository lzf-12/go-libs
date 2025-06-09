package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	//initialize ephemeral sqlite in memory
	log.Println("starting migration test with in-memory SQLite database")
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to open SQLite in-memory database: %v", err)
	}
	defer db.Close()

	// test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("successfully connected to in-memory SQLite database")

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		log.Fatalf("failed to get migrations path: %v", err)
	}
	log.Printf("using migrations path: %s", migrationsPath)

	// init migration instance
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "sqlite3", driver)
	if err != nil {
		log.Fatalf("failed to initialize migrate instance: %v", err)
	}
	defer m.Close()

	// get state
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("failed to get current migration version: %v", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("current state: No migrations applied yet")
	} else {
		log.Printf("current migration version: %d, dirty: %t", version, dirty)
	}

	log.Println("\nTesting UP migrations...")
	if err := testMigrationsUp(m); err != nil {
		log.Printf("UP migrations failed: %v", err)
		forceRollback(m)
		os.Exit(1)
	}

	log.Println("\nTesting DOWN migrations...")
	if err := testMigrationsDown(m); err != nil {
		log.Printf("DOWN migrations failed: %v", err)
		forceRollback(m)
		os.Exit(1)
	}

	// log any failure and force rollback version
	log.Println("\nAll migration tests completed successfully!")

}

func testMigrationsUp(m *migrate.Migrate) error {

	// up all to latest
	err := m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run UP migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("INFO: No new migrations to apply")
	} else {
		log.Println("SUCCESS: all UP migrations applied successfully")
	}

	// log current version after up
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get version after UP: %w", err)
	}
	log.Printf("Current version after UP: %d, dirty: %t", version, dirty)

	return nil
}

func testMigrationsDown(m *migrate.Migrate) error {

	// start version
	startVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// apply down to earliest
	err = m.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run DOWN migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("INFO: No migrations to roll back")
	} else {
		log.Printf("SUCCESS: All DOWN migrations applied successfully (rolled back from version %d)", startVersion)
	}

	// log current version after down
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get version after DOWN: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("current state after DOWN: No migrations applied (clean slate)")
	} else {
		log.Printf("current version after DOWN: %d, dirty: %t", version, dirty)
	}

	return nil
}

func forceRollback(m *migrate.Migrate) {
	log.Println("attempting to force rollback due to failure...")

	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("WARNING: Could not get current version for rollback: %v", err)
		return
	}

	if dirty {
		log.Printf("Database is in dirty state at version %d, forcing version...", version)
		if err := m.Force(int(version)); err != nil {
			log.Printf("ERROR: Failed to force version %d: %v", version, err)
			return
		}
		log.Printf("SUCCESS: Forced database to clean state at version %d", version)
	}

	// try to rollback to previous version or to 0
	targetVersion := int(version) - 1
	if targetVersion < 0 {
		targetVersion = 0
	}

	log.Printf("Rolling back to version %d...", targetVersion)
	if err := m.Migrate(uint(targetVersion)); err != nil && err != migrate.ErrNoChange {
		log.Printf("ERROR: Failed to rollback to version %d: %v", targetVersion, err)
		return
	}

	log.Printf("SUCCESS: Successfully rolled back to version %d", targetVersion)
}

func getMigrationsPath() (string, error) {
	// Try common migration directory locations
	possiblePaths := []string{
		"./migrations",
		"./db/migrations",
		"../migrations",
		".",
	}

	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		// Check if directory exists and contains .sql files
		if dirExists(absPath) && hasSQLFiles(absPath) {
			return "file://" + filepath.ToSlash(absPath), nil
		}
	}

	// If no migrations directory found, create a default one
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	migrationsDir := filepath.Join(currentDir, "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	log.Printf("Created migrations directory at: %s", migrationsDir)
	log.Println("INFO: please place your .sql migration files in this directory")

	return "file://" + filepath.ToSlash(migrationsDir), nil
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// hasSQLFiles checks if directory contains any .sql files
func hasSQLFiles(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			return true
		}
	}
	return false
}
