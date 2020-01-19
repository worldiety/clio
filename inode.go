package clio

import (
	"bytes"
)

// A hPtr is a Sha256 hash used for pointing purposes. It is just a pure 32 byte array and is allocated at the stack.
type hPtr [32]byte

// An inode is an index node which describes where to find the actual data. By design it contains no pointers
// and can be purely allocated on the stack.
type inode struct {
	name []byte
	ptr  hPtr
}

// inodes is just a slice of inodes and implements the sortable interface for the name. Even if the memory region is
// allocated at the heap, it only requires a single pointer.
type inodes []inode

func (k inodes) Len() int {
	return len(k)
}

func (k inodes) Less(i, j int) bool {
	return bytes.Compare(k[i].name, k[j].name) < 0
}

func (k inodes) Swap(i, j int) {
	k[j], k[i] = k[i], k[j]
}
