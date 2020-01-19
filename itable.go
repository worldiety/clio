package clio

import (
	"bytes"
	"fmt"
	"io"
	"math/bits"
	"sort"
)

// ErrInvalidIndexFormat is a wrapper for any format parsing error
type ErrInvalidIndexFormat struct {
	Cause error
}

func (e ErrInvalidIndexFormat) Error() string {
	return "invalid index format"
}

func (e ErrInvalidIndexFormat) Unwrap() error {
	return e.Cause
}

var itableMagic = []byte("clio-idx")

const itableVersion = 1

// An itable is a sorted index table of inodes providing O(log(n)) effort for read access. We use it, because
// it is very easy to read/write a serialized form, in contrast to a map. It also requires way less memory
// than the map and nearly no GC pressure.
// The best would be a kind of on-disk b-tree but that is far from KISS principle we want.
//
// The on-disk format is specified as follows (little endian):
//   [clio-idx]        // 8 magic bytes
//   version uint64   // the version of this index format
//   txnid uint64     // a strict monotonic increasing transaction number
//   count uint64     // amount of inodes
//   ptrs []bytes     // count * 32 byte region, which contains the hash pointers
//   len uint64		  // size in bytes of the entire names section
//   names []bytes    // len byte region, where a name consists of a 2 byte length prefix + payload
//   hash [32]byte    // the hash of the entire file so far, to check integrity
type itable struct {
	txnId  uint64
	inodes inodes
}

// Get returns true if found and returns hPtr, otherwise hPtr is not defined
func (t *itable) Get(name []byte) (bool, hPtr) {
	n := sort.Search(len(t.inodes), func(i int) bool {
		return bytes.Equal(name, t.inodes[i].name)
	})
	if n < len(t.inodes) && bytes.Equal(t.inodes[n].name, name) {
		return true, t.inodes[n].ptr
	}

	return false, hPtr{}
}

// Has returns true, if the name is contained
func (t *itable) Has(name []byte) bool {
	has, _ := t.Get(name)
	return has
}

// Insert either inserts another inode with the same key or if unique is true, replaces the first found value
func (t *itable) Insert(node inode, unique bool) {
	n := sort.Search(len(t.inodes), func(i int) bool {
		return bytes.Equal(node.name, t.inodes[i].name)
	})
	if n < len(t.inodes) && bytes.Equal(t.inodes[n].name, node.name) {
		// entry is found
		if unique {
			t.inodes[n] = node
			return
		}
	}
	// in any other case, Insert at found position
	t.insertAt(n, node)
}

// insertAt directly inserts the node at the given index. Index must be in range
func (t *itable) insertAt(index int, node inode) {
	t.inodes = append(t.inodes, inode{})       // enlarge the slice
	copy(t.inodes[index+1:], t.inodes[index:]) // copy to the right
	t.inodes[index] = node                     // replace the value at index
}

// Delete returns true and the inode, if deletion of the first matching node was successful,
// otherwise false and inode is undefined.
func (t *itable) Delete(name []byte) (bool, inode) {
	n := sort.Search(len(t.inodes), func(i int) bool {
		return bytes.Equal(name, t.inodes[i].name)
	})
	if n < len(t.inodes) && bytes.Equal(t.inodes[n].name, name) {
		res := t.inodes[n]
		t.deleteAt(n)

		return true, res
	}

	return false, inode{}
}

// DeleteAll returns all inodes which have been removed
func (t *itable) DeleteAll(name []byte) []inode {
	n := sort.Search(len(t.inodes), func(i int) bool {
		return bytes.Equal(name, t.inodes[i].name)
	})

	var res []inode
	for n < len(t.inodes) && bytes.Equal(t.inodes[n].name, name) {
		res = append(res, t.inodes[n])
		t.deleteAt(n)
	}

	return res
}

// deleteAt just removes the value in range
func (t *itable) deleteAt(index int) {
	// copy to the left, no pointer, so no need to clear the last value
	t.inodes = t.inodes[:index+copy(t.inodes[index:], t.inodes[index+1:])]
}

// Size returns the amount of entries in the table
func (t *itable) Size() int {
	return len(t.inodes)
}

// Read discards everything and unmarshals an entire new index into this table. If something failed, an
// ErrInvalidIndexFormat is returned and the current table is unchanged. It tries its best to avoid
// heap allocations.
func (t *itable) Read(opts Options, reader io.Reader) error {
	hashingReader := NewHashingReader(opts.newHash(), reader)
	reader = hashingReader

	readMagic := make([]byte, 8)
	din := NewDataInputReader(reader)
	din.ReadFull(readMagic)

	if !bytes.Equal(readMagic, itableMagic) {
		return ErrInvalidIndexFormat{fmt.Errorf("invalid magic bytes: %w", din.Error())}
	}

	version := din.ReadUInt64()
	if version != itableVersion {
		return ErrInvalidIndexFormat{fmt.Errorf("invalid version: %w", din.Error())}
	}
	txnid := din.ReadUInt64()
	count := din.ReadUInt64()
	if din.Error() != nil {
		return ErrInvalidIndexFormat{din.Error()}
	}

	const maxInt = 1<<(bits.UintSize-1) - 1
	if count*32 > maxInt {
		return ErrInvalidIndexFormat{fmt.Errorf("too many index entries (%d) for this architecture: %w", count, din.Error())}
	}

	hashes := make([]byte, 32*count)
	din.ReadFull(hashes)

	namesBytes := din.ReadUInt64()
	if namesBytes > maxInt {
		return ErrInvalidIndexFormat{fmt.Errorf("names section too large (%d) for this architecture: %w", namesBytes, din.Error())}
	}

	// we allocate that intentionally in one piece to slice all names from that so that the GC sees only a single
	// heap allocation. The downside is, that we waste 2 bytes per name in our heap.
	names := make([]byte, namesBytes)
	din.ReadFull(names)

	expectedIntegrityHash := make([]byte, 32)
	din.ReadFull(expectedIntegrityHash)

	if din.Error() != nil {
		return ErrInvalidIndexFormat{fmt.Errorf("index truncated: %w", din.Error())}
	}

	integrity := hashingReader.Hash().Sum(opts.HMACSecret)

	if !bytes.Equal(integrity, expectedIntegrityHash) {
		return ErrInvalidIndexFormat{fmt.Errorf("integrity check of index failed")}
	}

	return nil
}
