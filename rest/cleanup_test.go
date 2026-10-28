package rest

import (
	"errors"
	"testing"
)

func TestDeleteInBatchesSplitsInOrder(t *testing.T) {
	ids := make([]uint, 0, 1201)
	for i := uint(1); i <= 1201; i++ {
		ids = append(ids, i)
	}
	var seen []uint
	batches := 0
	if err := deleteInBatches(ids, func(batch []uint) error {
		batches++
		if len(batch) > 500 {
			t.Errorf("batch of %d exceeds the chunk size", len(batch))
		}
		seen = append(seen, batch...)
		return nil
	}); err != nil {
		t.Fatalf("deleteInBatches: %v", err)
	}
	if batches != 3 || len(seen) != len(ids) {
		t.Fatalf("batches = %d, seen = %d, want 3 and %d", batches, len(seen), len(ids))
	}
	for i := range ids {
		if seen[i] != ids[i] {
			t.Fatalf("id order changed at %d: got %d, want %d", i, seen[i], ids[i])
		}
	}
}

func TestDeleteInBatchesStopsOnError(t *testing.T) {
	stop := errors.New("boom")
	calls := 0
	err := deleteInBatches([]uint{1, 2, 3}, func([]uint) error {
		calls++
		return stop
	})
	if !errors.Is(err, stop) || calls != 1 {
		t.Errorf("error should stop after the first batch: calls = %d, err = %v", calls, err)
	}

	if err := deleteInBatches(nil, func([]uint) error {
		t.Fatal("deleteFn called for an empty id list")
		return nil
	}); err != nil {
		t.Errorf("empty input: %v", err)
	}
}
