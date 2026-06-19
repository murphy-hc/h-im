package data_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/murphy-hc/h-im/services/sequence/internal/data"
)

// Requires a running PostgreSQL with the sequences table created.
// Skip if no PG available.
func TestAllocateSegment_Integration(t *testing.T) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://him:him_secret@localhost:5432/him?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skipf("PG not available: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("PG ping failed: %v", err)
	}

	// Ensure the sequences table exists.
	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS sequences (
			key TEXT PRIMARY KEY,
			next_val BIGINT NOT NULL DEFAULT 1,
			step INT NOT NULL DEFAULT 1,
			segment_size INT NOT NULL DEFAULT 100
		)`,
	)
	if err != nil {
		t.Fatalf("failed to create sequences table: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM sequences WHERE key LIKE 'test_%'")
	})

	d := &data.Data{PG: pool}
	repo := data.NewSequenceRepo(d)

	key := "test_segment_id"
	size := int32(50)

	// First allocation.
	start1, end1, step1, err := repo.AllocateSegment(context.Background(), key, size)
	if err != nil {
		t.Fatalf("first allocation failed: %v", err)
	}
	if start1 < 1 || end1 <= start1 || step1 <= 0 {
		t.Fatalf("invalid first range: start=%d end=%d step=%d", start1, end1, step1)
	}

	// Second allocation — must be non-overlapping.
	start2, end2, step2, err := repo.AllocateSegment(context.Background(), key, size)
	if err != nil {
		t.Fatalf("second allocation failed: %v", err)
	}
	if start2 <= end1 {
		t.Fatalf("ranges overlap: first=[%d,%d] second=[%d,%d]", start1, end1, start2, end2)
	}
	if end2 <= start2 || step2 <= 0 {
		t.Fatalf("invalid second range: start=%d end=%d step=%d", start2, end2, step2)
	}

	// Step should be consistent across allocations.
	if step1 != step2 {
		t.Fatalf("step mismatch: first=%d second=%d", step1, step2)
	}
}
