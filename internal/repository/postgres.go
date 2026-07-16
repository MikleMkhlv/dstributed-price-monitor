package repository

import (
	"context"
	"dstributed-price-monitor/config"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
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

func (p *Postgres) Create(ctx context.Context, id string, data []byte) error {
	query := "INSERT INTO data (object_id, object_data) VALUES ($1, $2)"
	_, err := p.Pool.Exec(ctx, query, id, data)
	if err != nil {
		return fmt.Errorf("repository.Postgres.Create:{%v} error get data: %v", ctx.Value("operId"), err)
	}

	log.Printf("repository.Postgres.Create:{%v} create data in database is successful. id = %s", ctx.Value("operId"), id)

	return nil
}

func (p *Postgres) GetByID(ctx context.Context, id string) ([]byte, error) {
	query := "SELECT object_data FROM data WHERE object_id = $1"
	var data []byte
	if err := p.Pool.QueryRow(ctx, query, id).Scan(&data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.Postgres.GetByID:{%v} data is not found. id=%s", ctx.Value("operId"), id)
		}
		return nil, fmt.Errorf("repository.Postgres.GetByID:{%v}: %v", ctx.Value("operId"), err)
	}
	log.Printf("repository.Postgres.GetByID:{%v} data is found. id=%s", ctx.Value("operId"), id)
	return data, nil
}

func (p *Postgres) Update(ctx context.Context, id string, data []byte) ([]byte, error) {
	trz, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.Postgres.Update:{%v} error create transsaction for id=%s. %v", ctx.Value("operId"), id, err)
	}

	defer func() {
		if err != nil {
			trz.Rollback(ctx)
		}
	}()
	tempData := struct {
		Id     int
		Obj_id int
		Data   []byte
	}{}

	findQuery := "SELECT id, object_id, object_data FROM data WHERE object_id = $1"
	if err := trz.QueryRow(ctx, findQuery, id).Scan(
		&tempData.Id,
		&tempData.Obj_id,
		&tempData.Data,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.Postgres.Update:{%v} data not found by id = %s. %v", ctx.Value("operId"), id, err)
		}
	}
	if string(tempData.Data) != string(data) {
		return nil, fmt.Errorf("repository.Postgres.Update:{%v} The data has not changed by id = %s", ctx.Value("operId"), id)
	}

	updateQuery := "UPDATE data SET object_data = $1 WHERE object_id = $2"

	_, err = trz.Exec(ctx, updateQuery, data, id)
	if err != nil {
		return nil, fmt.Errorf("repository.Postgres.Update:{%v} error update data by id = %s. %v", ctx.Value("operId"), id, err)
	}

	err = trz.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository.Postgres.Update:{%v} %v", ctx.Value("operId"), err)
	}

	log.Printf("repository.Postgres.Update:{%v} update data by id = %s is successful", ctx.Value("operId"), id)
	return tempData.Data, nil
}

func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM data WHERE object_id = $1"
	_, err := p.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repository.Postgres.Delete:{%v} error delete data by id = %s. %v", ctx.Value("operId"), id, err)
	}
	log.Printf("repository.Postgres.Delete:{%v} update data by id = %s is successful", ctx.Value("operId"), id)
	return nil
}
