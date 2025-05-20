package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

const (
	defaultmaxopen, defaultmaxidle int = 20, 10
	defaultmaxlifetime                 = 1 * time.Hour
)

// maxopen, maxidle, maxlifetime configuration is optional will given default value if nil.
// maxopen default is 20
// maxidle default is 10
// maxlifetime default is 1 hour
func NewPostgres(dsn string, maxopen, maxidle *int, maxlifetime *time.Duration) (*Postgres, error) {

	err := validateDSN(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	openConns := defaultmaxopen
	if maxopen != nil {
		openConns = *maxopen
	}

	iddleConns := defaultmaxidle
	if maxidle != nil {
		iddleConns = *maxidle
	}

	lifetimeConns := defaultmaxlifetime
	if maxlifetime != nil {
		lifetimeConns = *maxlifetime
	}

	db.SetMaxOpenConns(openConns)
	db.SetMaxIdleConns(iddleConns)
	db.SetConnMaxLifetime(lifetimeConns)

	return &Postgres{db: db}, nil
}

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.db.PingContext(ctx)
}

func validateDSN(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid DSN format: %w", err)
	}
	if parsed.Scheme != "postgres" {
		return errors.New("DSN must use postgres scheme")
	}
	if parsed.Host == "" {
		return errors.New("DSN must include host")
	}
	return nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) IsReady() error {
	row := p.db.QueryRow("SELECT 1")
	var dummy int
	return row.Scan(&dummy)
}

func (p *Postgres) DB() *sql.DB {
	return p.db
}
