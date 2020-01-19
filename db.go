package clio

import "sync"

// DB represents a clio database. The current implementation follows a KISS principle so that one can restore and
// inspect the data with your POSIX knife. On the other side, this results in a lot of performance penalty points like
// high lock contention, blocking write transactions, low object discover performance and a high cost for
// keeping a transaction file to verify the state of the filesystem.
// However the database performance is very good for many concurrent read transactions and handling very large files,
// because it is only a thin and simple layer on top of the filesystem. Also it deduplicates each value.
//
// The layout is like this:
//   root-0/00
//
//          ...
//          FF
type DB struct {
	mutex   sync.RWMutex
	options Options
}

// Open tries to read an existing Database using the given options. Only existing databases can be opened.
func Open(opts Options) (*DB, error) {
	return &DB{options: opts}, nil
}
