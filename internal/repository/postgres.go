package repository

import (
	"context"
	"dstributed-price-monitor/config"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPG(ctx context.Context, cfg *config.Config) *Postgres {
	connStr := createConnectionStr(cfg)
	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("failed connection to database %s: %v", connStr, err))
	}

	if err := dbPool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("failed ping database: %v", err))
	}

	log.Print("repository.NewPG: connection with database is successful")

	if err := runMigrations(connStr); err != nil {
		dbPool.Close()
		panic(fmt.Sprintf("repository.Postgres.NewPG: %v", err))

	}

	return &Postgres{
		Pool: dbPool,
	}
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func createConnectionStr(cfg *config.Config) string {
	var connStr strings.Builder
	connStr.WriteString("postgres://")
	connStr.WriteString(cfg.Database.Login)
	connStr.WriteString(":")
	connStr.WriteString(cfg.Database.Password)
	connStr.WriteString("@")
	connStr.WriteString(cfg.Database.Address)
	connStr.WriteString("/")
	connStr.WriteString(cfg.Database.Table)
	if cfg.Database.SslMode != "" {
		connStr.WriteString(fmt.Sprintf("?sslmode=%s", cfg.Database.SslMode))
	}
	return connStr.String()
}

// func (p *Postgres) Create(ctx context.Context, id string, data []byte) error {
// }

// func (p *Postgres) GetByID(ctx context.Context, id string) ([]byte, error) {
// }

// func (p *Postgres) Update(ctx context.Context, id string, data []byte) ([]byte, error) {
// }

// func (p *Postgres) Delete(ctx context.Context, id string) error {
// }
