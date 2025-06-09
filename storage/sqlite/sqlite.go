package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

type SQLiteConfig struct {
	Filepath   string
	AutoCreate bool
}

func New(sc SQLiteConfig) (*SQLite, error) {
	if err := validateFilepath(sc.Filepath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", sc.Filepath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	if sc.AutoCreate {
		_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value TEXT
		)
	`)
		if err != nil {
			return nil, fmt.Errorf("failed to autocreate table: %w", err)
		}
	}

	return &SQLite{db: db}, nil
}

func validateFilepath(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	return nil
}

func (s *SQLite) Ping() error {
	return s.db.Ping()
}

func (s *SQLite) IsReady() error {
	row := s.db.QueryRow("SELECT 1")
	var result int
	return row.Scan(&result)
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

// ////////////////////
// SQLite as KV storage
func (s *SQLite) SetValueWithKey(key, value string) error {
	_, err := s.db.Exec("REPLACE INTO kv (key, value) VALUES (?, ?)", key, value)
	return err
}

func (s *SQLite) GetValueWithKey(key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM kv WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}
