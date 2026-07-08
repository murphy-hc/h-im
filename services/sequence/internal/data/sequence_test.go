package data_test

import (
	"context"
	"os"
	"testing"

	"github.com/murphy-hc/h-im/services/sequence/internal/data"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Requires a running MySQL.
func TestAllocateSegment_Integration(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "him:him_secret@tcp(localhost:3306)/him?charset=utf8mb4&parseTime=True"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("MySQL not available: %v", err)
	}

	// Ensure the sequences table exists.
	if err := db.AutoMigrate(&data.SequenceModel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() {
		db.Exec("DELETE FROM sequences WHERE `key` LIKE 'test_%'")
	})

	d := &data.Data{DB: db}
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
