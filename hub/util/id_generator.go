package util

import (
	"github.com/google/uuid"
	"strconv"
)

type IdGenerator interface {
	NextId() string
}

type SimpleIdGenerator struct {
	CurrentId int
}

func (idg *SimpleIdGenerator) NextId() string {
	idg.CurrentId++

	return strconv.Itoa(idg.CurrentId)
}

// UUIDGenerator generates unique UUIDs for each ID
type UUIDGenerator struct {
}

func (idg *UUIDGenerator) NextId() string {
	return uuid.New().String()
}

// SequenceGenerator generates IDs from a preset sequence
type SequenceGenerator struct {
	idSequence   []string
	currentIndex int
}

func (idg *SequenceGenerator) NextId() string {
	if idg.currentIndex >= len(idg.idSequence) {
		// If we've exhausted the sequence, return empty string or panic
		// For now, returning empty string
		return ""
	}

	id := idg.idSequence[idg.currentIndex]
	idg.currentIndex++
	return id
}

// NewSequenceGenerator creates a new SequenceGenerator with the given ID sequence
func NewSequenceGenerator(idSequence []string) *SequenceGenerator {
	return &SequenceGenerator{
		idSequence:   idSequence,
		currentIndex: 0,
	}
}
