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
		panic(fmt.Sprintf("failed connection to database %s: %v", err))
	}

	if err := dbPool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("failed ping database: %v", err))
	}

	log.Print("repository.NewPG: connection with database is successful")

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
	if cfg.Database.Schema != "" {
		connStr.WriteString(fmt.Sprintf("?schema=%s", cfg.Database.Schema))
	}
	return connStr.String()
}
