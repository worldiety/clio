package clio

import (
	"encoding/binary"
	"io"
)

// DataInput is a simplified contract to read various data formats in little endian
type DataInput interface {
	// ReadUInt64 reads the next 8 bytes and interprets them accordingly
	ReadUInt64() uint64

	// ReadFull reads exactly len(b) bytes
	ReadFull(b []byte) int

	// Error returns the first occurred error. Each call to any Read* method may cause an error.
	Error() error
}

type DataInputReader struct {
	buf8     []byte
	in       io.Reader
	firstErr error
}

func NewDataInputReader(in io.Reader) *DataInputReader {
	return &DataInputReader{
		buf8: make([]byte, 8),
		in:   in,
	}
}

func (r *DataInputReader) hasErr(err error) bool {
	if err != nil && r.firstErr == nil {
		r.firstErr = err
	}
	if r.firstErr != nil {
		return true
	}
	return false
}

func (r *DataInputReader) ReadUInt64() uint64 {
	_, err := io.ReadFull(r.in, r.buf8)
	if r.hasErr(err) {
		return 0
	}
	return binary.LittleEndian.Uint64(r.buf8)
}

func (r *DataInputReader) ReadFull(b []byte) int {
	n, err := io.ReadFull(r.in, r.buf8)
	if r.hasErr(err) {
		return n
	}
	return n
}

func (r *DataInputReader) Error() error {
	return r.firstErr
}
