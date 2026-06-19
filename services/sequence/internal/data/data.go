package data

import (
	"context"
	"fmt"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewSequenceRepo)

// Data holds data source clients.
type Data struct {
	PG *pgxpool.Pool
}

// NewData creates a Data instance with a PG connection pool.
func NewData() (*Data, func(), error) {
	// TODO: read DSN from config
	dsn := "postgres://him:him_secret@localhost:5432/him?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("data: connect pg: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("data: ping pg: %w", err)
	}

	d := &Data{PG: pool}
	cleanup := func() {
		pool.Close()
	}
	return d, cleanup, nil
}
