package util

import (
	"testing"
)

func TestNewIndexPool(t *testing.T) {
	t.Run("Creates pool with all indices", func(t *testing.T) {
		pool := NewIndexPool(5)

		if pool.Count() != 5 {
			t.Errorf("Expected 5 indices, got %d", pool.Count())
		}

		for i := 0; i < 5; i++ {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}
	})

	t.Run("Creates pool with zero indices", func(t *testing.T) {
		pool := NewIndexPool(0)

		if pool.Count() != 0 {
			t.Errorf("Expected 0 indices, got %d", pool.Count())
		}
	})

	t.Run("Creates pool with one index", func(t *testing.T) {
		pool := NewIndexPool(1)

		if pool.Count() != 1 {
			t.Errorf("Expected 1 index, got %d", pool.Count())
		}

		if !pool.Contains(0) {
			t.Error("Index 0 should be in pool")
		}
	})
}

func TestRemove(t *testing.T) {
	t.Run("Removes single index", func(t *testing.T) {
		pool := NewIndexPool(5)

		pool.Remove(2)

		if pool.Count() != 4 {
			t.Errorf("Expected 4 indices, got %d", pool.Count())
		}

		if pool.Contains(2) {
			t.Error("Index 2 should be removed")
		}

		for _, i := range []int{0, 1, 3, 4} {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}
	})

	t.Run("Removes multiple indices", func(t *testing.T) {
		pool := NewIndexPool(5)

		pool.Remove(0)
		pool.Remove(2)
		pool.Remove(4)

		if pool.Count() != 2 {
			t.Errorf("Expected 2 indices, got %d", pool.Count())
		}

		for _, i := range []int{0, 2, 4} {
			if pool.Contains(i) {
				t.Errorf("Index %d should be removed", i)
			}
		}

		for _, i := range []int{1, 3} {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}
	})

	t.Run("Removes all indices", func(t *testing.T) {
		pool := NewIndexPool(3)

		pool.Remove(0)
		pool.Remove(1)
		pool.Remove(2)

		if pool.Count() != 0 {
			t.Errorf("Expected 0 indices, got %d", pool.Count())
		}

		for i := 0; i < 3; i++ {
			if pool.Contains(i) {
				t.Errorf("Index %d should be removed", i)
			}
		}
	})

	t.Run("Removes same index twice (idempotent)", func(t *testing.T) {
		pool := NewIndexPool(5)

		pool.Remove(2)
		pool.Remove(2)

		if pool.Count() != 4 {
			t.Errorf("Expected 4 indices after double remove, got %d", pool.Count())
		}

		if pool.Contains(2) {
			t.Error("Index 2 should be removed")
		}
	})

	t.Run("Removes indices in reverse order", func(t *testing.T) {
		pool := NewIndexPool(5)

		pool.Remove(4)
		pool.Remove(3)
		pool.Remove(2)

		if pool.Count() != 2 {
			t.Errorf("Expected 2 indices, got %d", pool.Count())
		}

		for _, i := range []int{0, 1} {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}
	})

	t.Run("Removes last index when only one remains", func(t *testing.T) {
		pool := NewIndexPool(3)

		pool.Remove(0)
		pool.Remove(1)

		if pool.Count() != 1 {
			t.Errorf("Expected 1 index, got %d", pool.Count())
		}

		pool.Remove(2)

		if pool.Count() != 0 {
			t.Errorf("Expected 0 indices, got %d", pool.Count())
		}
	})
}

func TestGetRandom(t *testing.T) {
	t.Run("Returns valid index from all indices", func(t *testing.T) {
		pool := NewIndexPool(5)

		randomValues := []float64{0.0, 0.2, 0.4, 0.6, 0.8}
		indices := make(map[int]bool)

		for _, rv := range randomValues {
			idx := pool.GetRandom(rv)
			if idx < 0 || idx >= 5 {
				t.Errorf("Invalid index %d", idx)
			}
			indices[idx] = true
		}

		if len(indices) < 2 {
			t.Error("Expected to get multiple different indices")
		}
	})

	t.Run("Returns only active indices", func(t *testing.T) {
		pool := NewIndexPool(5)
		pool.Remove(0)
		pool.Remove(2)
		pool.Remove(4)

		// Active indices: 1, 3
		for i := 0; i < 20; i++ {
			randomValue := float64(i) / 20.0
			idx := pool.GetRandom(randomValue)

			if idx != 1 && idx != 3 {
				t.Errorf("Got index %d, but only 1 and 3 should be active", idx)
			}

			if !pool.Contains(idx) {
				t.Errorf("GetRandom returned removed index %d", idx)
			}
		}
	})

	t.Run("Returns -1 when pool is empty", func(t *testing.T) {
		pool := NewIndexPool(3)
		pool.Remove(0)
		pool.Remove(1)
		pool.Remove(2)

		idx := pool.GetRandom(0.5)

		if idx != -1 {
			t.Errorf("Expected -1 when pool empty, got %d", idx)
		}
	})

	t.Run("Returns only index when one remains", func(t *testing.T) {
		pool := NewIndexPool(5)
		pool.Remove(0)
		pool.Remove(1)
		pool.Remove(2)
		pool.Remove(4)
		// Only index 3 remains

		for i := 0; i < 10; i++ {
			randomValue := float64(i) / 10.0
			idx := pool.GetRandom(randomValue)
			if idx != 3 {
				t.Errorf("Expected index 3, got %d", idx)
			}
		}
	})

	t.Run("Handles edge case randomValue == 1.0", func(t *testing.T) {
		pool := NewIndexPool(5)

		// randomValue of 1.0 (or very close) should not cause index out of bounds
		idx := pool.GetRandom(0.9999999)

		if idx < 0 || idx >= 5 {
			t.Errorf("Edge case: got invalid index %d", idx)
		}
	})

	t.Run("Maps random values correctly", func(t *testing.T) {
		pool := NewIndexPool(5)

		tests := []struct {
			randomValue    float64
			expectedInPool bool
		}{
			{0.0, true},
			{0.1, true},
			{0.19, true},
			{0.2, true},
			{0.4, true},
			{0.6, true},
			{0.8, true},
			{0.99, true},
		}

		for _, tt := range tests {
			idx := pool.GetRandom(tt.randomValue)
			if tt.expectedInPool && !pool.Contains(idx) {
				t.Errorf("Random %v: index %d should be in pool", tt.randomValue, idx)
			}
		}
	})
}

func TestIndexPoolSequence(t *testing.T) {
	t.Run("Complex removal sequence", func(t *testing.T) {
		pool := NewIndexPool(10)

		pool.Remove(3)
		if pool.Count() != 9 {
			t.Errorf("After 1 removal: expected 9 active, got %d", pool.Count())
		}

		pool.Remove(7)
		pool.Remove(1)
		if pool.Count() != 7 {
			t.Errorf("After 3 removals: expected 7 active, got %d", pool.Count())
		}

		for _, i := range []int{1, 3, 7} {
			if pool.Contains(i) {
				t.Errorf("Index %d should be removed", i)
			}
		}

		idx := pool.GetRandom(0.5)
		if !pool.Contains(idx) {
			t.Errorf("Random index %d should be in pool", idx)
		}

		pool.Remove(0)
		pool.Remove(9)
		pool.Remove(5)

		if pool.Count() != 4 {
			t.Errorf("After 6 removals: expected 4 active, got %d", pool.Count())
		}

		// Active should be: 2, 4, 6, 8
		active := []int{2, 4, 6, 8}
		for _, i := range active {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}
	})
}

func TestIndexPoolLargeScale(t *testing.T) {
	t.Run("Works with 1000+ indices", func(t *testing.T) {
		count := 1000
		pool := NewIndexPool(count)

		if pool.Count() != count {
			t.Errorf("Expected %d indices, got %d", count, pool.Count())
		}

		// Remove every other index
		for i := 0; i < count; i += 2 {
			pool.Remove(i)
		}

		expectedActive := count / 2
		if pool.Count() != expectedActive {
			t.Errorf("Expected %d active indices, got %d", expectedActive, pool.Count())
		}

		// Verify removed
		for i := 0; i < count; i += 2 {
			if pool.Contains(i) {
				t.Errorf("Index %d should be removed", i)
			}
		}

		// Verify active
		for i := 1; i < count; i += 2 {
			if !pool.Contains(i) {
				t.Errorf("Index %d should be in pool", i)
			}
		}

		// Get random indices
		for i := 0; i < 100; i++ {
			randomValue := float64(i) / 100.0
			idx := pool.GetRandom(randomValue)

			if !pool.Contains(idx) {
				t.Errorf("Random index %d should be in pool", idx)
			}

			if idx%2 == 0 {
				t.Errorf("Random index %d should be odd (even indices removed)", idx)
			}
		}
	})
}
