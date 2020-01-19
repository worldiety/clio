package clio

import (
	"fmt"
	"hash"
	"io"
)

// HashingReader calculates for every transferred byte the hash until Sum() is called.
type HashingReader struct {
	hasher hash.Hash
	reader io.Reader
	count  uint64
}

// NewHashingReader creates a new instance
func NewHashingReader(hasher hash.Hash, reader io.Reader) *HashingReader {
	return &HashingReader{hasher: hasher, reader: reader}
}

func (h *HashingReader) Read(p []byte) (n int, err error) {
	n, err = h.reader.Read(p)
	n2, err2 := h.hasher.Write(p[0:n])
	if err != nil && err2 != nil {
		return n, fmt.Errorf("failed to hash: %w", fmt.Errorf("failed to read: %w", err))
	}

	if err != nil {
		return n, err
	}
	if err2 != nil {
		return n2, err2
	}
	if n != n2 {
		return n, fmt.Errorf("unable to hash the buffer properly")
	}
	return n, nil
}

// Sum returns the resulting slice.
// It does not change the underlying hash state.
func (h *HashingReader) Sum() []byte {
	return h.hasher.Sum(nil)
}

// Hash returns the wrapped hasher
func (h *HashingReader) Hash() hash.Hash {
	return h.hasher
}

// Count returns the total amount of read bytes so far.
func (h *HashingReader) Count() uint64 {
	return h.count
}
