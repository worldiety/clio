package clio

import (
	"fmt"
)

// ErrTxnAlreadyClosed is a sentinel error when using an already closed transaction
var ErrTxnAlreadyClosed = fmt.Errorf("transaction already closed")

// Txn is a read or write transaction
type Txn struct {
	db       *DB
	readOnly bool
	alive    bool
}

func newTransaction(db *DB, readOnly bool) *Txn {
	t := &Txn{db: db, alive: true}
	t.lock(readOnly)

	return t
}

func (t *Txn) unlock() {
	if t.readOnly {
		t.db.mutex.RUnlock()
	} else {
		t.db.mutex.Unlock()
	}
}

func (t *Txn) lock(readOnly bool) {
	t.readOnly = readOnly
	if t.readOnly {
		t.db.mutex.RLock()
	} else {
		t.db.mutex.Lock()
	}
}

// Commit closes the transaction. A read-write transaction gives up the global write lock and makes the data
// visible for all subsequent transactions. It is not defined, if the data is actually persisted after the commit
// returns.
func (t *Txn) Commit() error {
	panic("implement me")
}

// Close frees
func (t *Txn) Close() error {
	if !t.alive {
		return ErrTxnAlreadyClosed
	}

	t.alive = false
	t.unlock()

	return nil
}
