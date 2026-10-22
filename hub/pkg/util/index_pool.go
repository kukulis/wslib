package util

// IndexPool efficiently tracks a pool of indices with O(1) removal and selection.
// Uses swap-and-pop algorithm for constant-time operations.
//
// IMPORTANT: The order of indices in the pool is NOT maintained. After removals,
// indices may be in any order. Use only when order doesn't matter (e.g., random selection).
type IndexPool struct {
	activeIndices []int // indices currently in the pool (order not maintained)
	indexPosition []int // maps index -> position in activeIndices (-1 if removed)
}

// NewIndexPool creates a new index pool for the given number of indices.
// Initially all indices [0, count-1] are in the pool.
func NewIndexPool(count int) *IndexPool {
	activeIndices := make([]int, count)
	indexPosition := make([]int, count)

	for i := 0; i < count; i++ {
		activeIndices[i] = i
		indexPosition[i] = i
	}

	return &IndexPool{
		activeIndices: activeIndices,
		indexPosition: indexPosition,
	}
}

// Remove removes the given index from the pool.
// Uses swap-and-pop for O(1) complexity.
// If the index is already removed, this is a no-op.
func (ip *IndexPool) Remove(index int) {
	position := ip.indexPosition[index]

	// Already removed
	if position == -1 {
		return
	}

	lastPos := len(ip.activeIndices) - 1

	// Swap with last element
	lastIndex := ip.activeIndices[lastPos]
	ip.activeIndices[position] = lastIndex
	ip.indexPosition[lastIndex] = position

	// Pop last element
	ip.activeIndices = ip.activeIndices[:lastPos]
	ip.indexPosition[index] = -1
}

// GetRandom returns a random index from the pool using the provided random value [0.0, 1.0).
// Returns -1 if the pool is empty.
func (ip *IndexPool) GetRandom(randomValue float64) int {
	if len(ip.activeIndices) == 0 {
		return -1
	}

	// Map random [0, 1) to [0, count)
	position := int(randomValue * float64(len(ip.activeIndices)))

	// Clamp to valid range (handles edge case where randomValue == 1.0)
	if position >= len(ip.activeIndices) {
		position = len(ip.activeIndices) - 1
	}

	return ip.activeIndices[position]
}

// Count returns the number of indices currently in the pool.
func (ip *IndexPool) Count() int {
	return len(ip.activeIndices)
}

// Contains returns true if the given index is in the pool.
func (ip *IndexPool) Contains(index int) bool {
	return ip.indexPosition[index] != -1
}
