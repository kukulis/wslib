package util

import "fmt"

type IndexMapPool struct {
	indexMap map[string]int
	keys     []string
	count    int
}

func NewIndexMapPool(keys []string) *IndexMapPool {
	indexMap := make(map[string]int)
	for i, key := range keys {
		indexMap[key] = i
	}
	return &IndexMapPool{
		keys:     keys,
		indexMap: indexMap,
		count:    len(keys),
	}
}

func (p *IndexMapPool) GetKey(i int) string {
	if i < 0 || i >= len(p.keys) {
		return ""
	}
	return p.keys[i]
}

func (p *IndexMapPool) GetIndex(key string) int {
	if idx, exists := p.indexMap[key]; exists {
		return idx
	}
	return -1
}

func (p *IndexMapPool) RemoveKey(keyToRemove string) error {
	indexToRemove, exists := p.indexMap[keyToRemove]
	if !exists || indexToRemove < 0 || indexToRemove >= p.count {
		return fmt.Errorf("key not found: %s", keyToRemove)
	}

	lastIndex := p.count - 1
	lastKey := p.keys[lastIndex]

	p.keys[indexToRemove] = lastKey
	p.keys[lastIndex] = ""

	p.indexMap[keyToRemove] = -1
	p.indexMap[lastKey] = indexToRemove

	p.count--

	return nil
}

func (p *IndexMapPool) Count() int {
	return p.count
}
