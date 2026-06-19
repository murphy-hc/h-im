package biz_test

import (
	"context"
	"errors"
	"testing"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
)

type mockRepo struct {
	start, end int64
	step       int32
	err        error
}

func (m *mockRepo) AllocateSegment(ctx context.Context, key string, size int32) (int64, int64, int32, error) {
	return m.start, m.end, m.step, m.err
}

func TestAllocateSegment_EmptyKey(t *testing.T) {
	uc := biz.NewSequenceUseCase(&mockRepo{})
	_, _, _, err := uc.AllocateSegment(context.Background(), "", 10)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestAllocateSegment_Success(t *testing.T) {
	mock := &mockRepo{start: 1, end: 100, step: 1}
	uc := biz.NewSequenceUseCase(mock)
	start, end, step, err := uc.AllocateSegment(context.Background(), "msg_id", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != 1 || end != 100 || step != 1 {
		t.Fatalf("unexpected result: start=%d end=%d step=%d", start, end, step)
	}
}

func TestAllocateSegment_RepoError(t *testing.T) {
	mock := &mockRepo{err: errors.New("db down")}
	uc := biz.NewSequenceUseCase(mock)
	_, _, _, err := uc.AllocateSegment(context.Background(), "msg_id", 10)
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
