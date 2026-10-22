package util

import "testing"

func TestNewIndexMapPool(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if pool == nil {
		t.Fatal("NewIndexMapPool returned nil")
	}

	if pool.count != 3 {
		t.Errorf("Expected count 3, got %d", pool.count)
	}

	if len(pool.keys) != 3 {
		t.Errorf("Expected keys length 3, got %d", len(pool.keys))
	}

	if len(pool.indexMap) != 3 {
		t.Errorf("Expected indexMap length 3, got %d", len(pool.indexMap))
	}
}

func TestNewIndexMapPoolEmpty(t *testing.T) {
	pool := NewIndexMapPool([]string{})

	if pool.count != 0 {
		t.Errorf("Expected count 0, got %d", pool.count)
	}
}

func TestGetKey(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if pool.GetKey(0) != "alpha" {
		t.Errorf("Expected 'alpha', got '%s'", pool.GetKey(0))
	}

	if pool.GetKey(1) != "beta" {
		t.Errorf("Expected 'beta', got '%s'", pool.GetKey(1))
	}

	if pool.GetKey(2) != "gamma" {
		t.Errorf("Expected 'gamma', got '%s'", pool.GetKey(2))
	}
}

func TestGetIndex(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if pool.GetIndex("alpha") != 0 {
		t.Errorf("Expected index 0, got %d", pool.GetIndex("alpha"))
	}

	if pool.GetIndex("beta") != 1 {
		t.Errorf("Expected index 1, got %d", pool.GetIndex("beta"))
	}

	if pool.GetIndex("gamma") != 2 {
		t.Errorf("Expected index 2, got %d", pool.GetIndex("gamma"))
	}
}

func TestGetIndexNonExistent(t *testing.T) {
	keys := []string{"alpha", "beta"}
	pool := NewIndexMapPool(keys)

	// Non-existent key should return -1 to avoid ambiguity with valid index 0
	result := pool.GetIndex("nonexistent")
	if result != -1 {
		t.Errorf("Expected index -1 for non-existent key, got %d", result)
	}
}

func TestRemoveKeyMiddle(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("beta"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// After removing "beta", "gamma" should be moved to index 1
	if pool.GetKey(1) != "gamma" {
		t.Errorf("Expected 'gamma' at index 1, got '%s'", pool.GetKey(1))
	}

	// The index of "gamma" should now be 1
	if pool.GetIndex("gamma") != 1 {
		t.Errorf("Expected 'gamma' to be at index 1, got %d", pool.GetIndex("gamma"))
	}

	// "alpha" should still be at index 0
	if pool.GetKey(0) != "alpha" {
		t.Errorf("Expected 'alpha' at index 0, got '%s'", pool.GetKey(0))
	}

	// Count should be decremented
	if pool.count != 2 {
		t.Errorf("Expected count 2 after removal, got %d", pool.count)
	}

	// The removed key should no longer be in the index map or should have invalid index
	idx := pool.GetIndex("beta")
	if idx >= 0 && idx < pool.count {
		t.Errorf("Removed key 'beta' should not have valid index, got %d", idx)
	}
}

func TestRemoveKeyFirst(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("alpha"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// After removing "alpha", "gamma" should be moved to index 0
	if pool.GetKey(0) != "gamma" {
		t.Errorf("Expected 'gamma' at index 0, got '%s'", pool.GetKey(0))
	}

	// The index of "gamma" should now be 0
	if pool.GetIndex("gamma") != 0 {
		t.Errorf("Expected 'gamma' to be at index 0, got %d", pool.GetIndex("gamma"))
	}

	// Count should be decremented
	if pool.count != 2 {
		t.Errorf("Expected count 2 after removal, got %d", pool.count)
	}
}

func TestRemoveKeyLast(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("gamma"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// After removing "gamma", elements at 0 and 1 should remain unchanged
	if pool.GetKey(0) != "alpha" {
		t.Errorf("Expected 'alpha' at index 0, got '%s'", pool.GetKey(0))
	}

	if pool.GetKey(1) != "beta" {
		t.Errorf("Expected 'beta' at index 1, got '%s'", pool.GetKey(1))
	}

	// Count should be decremented
	if pool.count != 2 {
		t.Errorf("Expected count 2 after removal, got %d", pool.count)
	}
}

func TestRemoveMultipleKeys(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	pool := NewIndexMapPool(keys)

	// Remove "beta" (index 1)
	if err := pool.RemoveKey("beta"); err != nil {
		t.Fatalf("Unexpected error removing 'beta': %v", err)
	}
	if pool.count != 4 {
		t.Errorf("Expected count 4 after first removal, got %d", pool.count)
	}

	// Remove "delta" (which might have moved)
	if err := pool.RemoveKey("delta"); err != nil {
		t.Fatalf("Unexpected error removing 'delta': %v", err)
	}
	if pool.count != 3 {
		t.Errorf("Expected count 3 after second removal, got %d", pool.count)
	}

	// Remove "alpha"
	if err := pool.RemoveKey("alpha"); err != nil {
		t.Fatalf("Unexpected error removing 'alpha': %v", err)
	}
	if pool.count != 2 {
		t.Errorf("Expected count 2 after third removal, got %d", pool.count)
	}

	// Verify remaining keys can still be accessed correctly
	// We should still be able to get valid keys at indices 0 and 1
	key0 := pool.GetKey(0)
	key1 := pool.GetKey(1)

	if key0 == "" || key1 == "" {
		t.Errorf("Keys at valid indices should not be empty: key0='%s', key1='%s'", key0, key1)
	}

	// Verify indices are correct for remaining keys
	if pool.GetIndex(key0) != 0 {
		t.Errorf("Key '%s' should be at index 0, got %d", key0, pool.GetIndex(key0))
	}

	if pool.GetIndex(key1) != 1 {
		t.Errorf("Key '%s' should be at index 1, got %d", key1, pool.GetIndex(key1))
	}
}

func TestRemoveKeySingleElement(t *testing.T) {
	keys := []string{"alpha"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("alpha"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if pool.count != 0 {
		t.Errorf("Expected count 0 after removing single element, got %d", pool.count)
	}
}

func TestRemoveKeyLastElementIndexMapCleanup(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	// Remove the last key "gamma"
	if err := pool.RemoveKey("gamma"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// The removed key should either be deleted from the map or marked as invalid (-1)
	idx, exists := pool.indexMap["gamma"]

	// The key should either not exist, or if it exists, should be -1 or >= count
	if exists && idx >= 0 && idx < pool.count {
		t.Errorf("Removed key 'gamma' should not have a valid index, got %d (count=%d)", idx, pool.count)
	}

	// The remaining keys should still be accessible
	if pool.GetIndex("alpha") != 0 {
		t.Errorf("Expected 'alpha' at index 0, got %d", pool.GetIndex("alpha"))
	}
	if pool.GetIndex("beta") != 1 {
		t.Errorf("Expected 'beta' at index 1, got %d", pool.GetIndex("beta"))
	}
}

func TestRemoveKeyIndexMapConsistency(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma", "delta"}
	pool := NewIndexMapPool(keys)

	// Remove "delta" (last element)
	if err := pool.RemoveKey("delta"); err != nil {
		t.Fatalf("Unexpected error removing 'delta': %v", err)
	}

	// Verify "delta" is not in the valid range
	idx := pool.GetIndex("delta")
	if idx >= 0 && idx < pool.count {
		t.Errorf("After removing 'delta', its index should be invalid, got %d (count=%d)", idx, pool.count)
	}

	// Remove "alpha" (will be swapped with "gamma")
	if err := pool.RemoveKey("alpha"); err != nil {
		t.Fatalf("Unexpected error removing 'alpha': %v", err)
	}

	// Verify "alpha" is not in the valid range
	idx = pool.GetIndex("alpha")
	if idx >= 0 && idx < pool.count {
		t.Errorf("After removing 'alpha', its index should be invalid, got %d (count=%d)", idx, pool.count)
	}

	// Verify remaining keys have valid indices
	for _, key := range []string{"beta", "gamma"} {
		idx := pool.GetIndex(key)
		if idx < 0 || idx >= pool.count {
			t.Errorf("Key '%s' should have valid index, got %d (count=%d)", key, idx, pool.count)
		}

		// Verify bidirectional consistency
		if pool.GetKey(idx) != key {
			t.Errorf("Index %d should map to '%s', got '%s'", idx, key, pool.GetKey(idx))
		}
	}
}

func TestGetIndexAmbiguityWithZero(t *testing.T) {
	keys := []string{"zero", "one", "two"}
	pool := NewIndexMapPool(keys)

	// "zero" is at index 0
	zeroIdx := pool.GetIndex("zero")
	if zeroIdx != 0 {
		t.Errorf("Expected 'zero' at index 0, got %d", zeroIdx)
	}

	// Non-existent key should return -1, not 0 (which would be ambiguous)
	nonExistentIdx := pool.GetIndex("nonexistent")
	if nonExistentIdx == zeroIdx {
		t.Errorf("Non-existent key returns %d, same as valid key 'zero' - AMBIGUITY BUG", nonExistentIdx)
	}

	if nonExistentIdx != -1 {
		t.Errorf("Expected -1 for non-existent key, got %d", nonExistentIdx)
	}
}

func TestRemoveKeyNonExistent(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	initialCount := pool.count

	// Removing non-existent key should return an error
	err := pool.RemoveKey("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent key, got nil")
	}

	// Verify the pool state is unchanged
	if pool.count != initialCount {
		t.Errorf("Count changed after removing non-existent key: was %d, now %d", initialCount, pool.count)
	}

	// Verify existing keys are still accessible
	if pool.GetIndex("alpha") != 0 {
		t.Errorf("Existing key 'alpha' affected by invalid removal")
	}
	if pool.GetIndex("beta") != 1 {
		t.Errorf("Existing key 'beta' affected by invalid removal")
	}
	if pool.GetIndex("gamma") != 2 {
		t.Errorf("Existing key 'gamma' affected by invalid removal")
	}
}

func TestRemoveKeyAlreadyRemoved(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	// Remove a key
	if err := pool.RemoveKey("beta"); err != nil {
		t.Fatalf("Unexpected error on first removal: %v", err)
	}

	if pool.count != 2 {
		t.Errorf("Expected count 2 after first removal, got %d", pool.count)
	}

	// Remove the same key again - should return error
	err := pool.RemoveKey("beta")
	if err == nil {
		t.Error("Expected error when removing already-removed key, got nil")
	}

	// Check if count is still correct
	if pool.count != 2 {
		t.Errorf("Count should still be 2 after removing already-removed key, got %d", pool.count)
	}
}

func TestGetKeyWithNegativeIndex(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	// This will cause a panic from array access with negative index
	// The implementation should validate and return empty string or panic gracefully
	result := pool.GetKey(-1)

	// If we get here without panic, check the result
	if result != "" {
		t.Errorf("Expected empty string for negative index, got '%s'", result)
	}
}

func TestGetKeyWithIndexEqualToCount(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	// Accessing index == count is out of bounds
	result := pool.GetKey(pool.count)

	// Should return empty string (valid keys were cleared from removed positions)
	if result != "" {
		t.Errorf("Expected empty string for index == count, got '%s'", result)
	}
}

func TestGetKeyWithIndexGreaterThanCount(t *testing.T) {
	keys := []string{"alpha", "beta"}
	pool := NewIndexMapPool(keys)

	// This will panic - array out of bounds
	result := pool.GetKey(100)

	if result != "" {
		t.Errorf("Expected empty string for index > count, got '%s'", result)
	}
}

func TestGetKeyAfterRemoval(t *testing.T) {
	keys := []string{"alpha", "beta", "gamma"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("beta"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// After removing "beta", position 2 should be empty string
	result := pool.GetKey(2)
	if result != "" {
		t.Errorf("Expected empty string at invalidated position 2, got '%s'", result)
	}

	// Positions 0 and 1 should still have valid keys
	key0 := pool.GetKey(0)
	key1 := pool.GetKey(1)

	if key0 == "" || key1 == "" {
		t.Errorf("Valid positions returned empty: key0='%s', key1='%s'", key0, key1)
	}
}

func TestGetKeyAfterAllRemoved(t *testing.T) {
	keys := []string{"alpha"}
	pool := NewIndexMapPool(keys)

	if err := pool.RemoveKey("alpha"); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if pool.count != 0 {
		t.Errorf("Expected count 0, got %d", pool.count)
	}

	// Accessing index 0 when count is 0 should still be safe (returns empty string)
	result := pool.GetKey(0)

	if result != "" {
		t.Errorf("Expected empty string after all keys removed, got '%s'", result)
	}
}

func TestRemoveKeyEmptyString(t *testing.T) {
	keys := []string{"alpha", "beta", ""}
	pool := NewIndexMapPool(keys)

	// Remove the empty string key
	if err := pool.RemoveKey(""); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if pool.count != 2 {
		t.Errorf("Expected count 2 after removing empty string key, got %d", pool.count)
	}

	// Verify empty string is no longer at a valid index
	idx := pool.GetIndex("")
	if idx >= 0 && idx < pool.count {
		t.Errorf("Empty string key should not have valid index after removal, got %d", idx)
	}
}
