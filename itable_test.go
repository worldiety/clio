package clio

import (
	"testing"
)

func Test_itable(t *testing.T) {
	table := &itable{}
	assertInt(t, table.Size(), 0)
	has, _ := table.Get([]byte("hello"))
	assertFalse(t, has)

	has, _ = table.Get([]byte(""))
	assertFalse(t, has)

	has, _ = table.Delete([]byte("hello"))
	assertFalse(t, has)

	table.Insert(inode{[]byte("hello"), hPtr{}}, true)
	assertInt(t, 1, table.Size())

	has, _ = table.Get([]byte(""))
	assertFalse(t, has)
	has, _ = table.Get([]byte("hella"))
	assertFalse(t, has)
	has, _ = table.Get([]byte("hello2"))
	assertFalse(t, has)
	has, _ = table.Get([]byte("hello"))
	assertTrue(t, has)

	table.Insert(inode{[]byte("hello"), hPtr{}}, true)
	assertInt(t, table.Size(), 1)

	table.Insert(inode{[]byte("hello"), hPtr{}}, false)
	assertInt(t, 2, table.Size())

	has, _ = table.Delete([]byte("hello"))
	assertTrue(t, has)

	table.Insert(inode{[]byte("hello"), hPtr{}}, false)
	assertInt(t, 2, table.Size())

	deleted := table.DeleteAll([]byte("hello"))
	assertInt(t, 2, len(deleted))

	assertInt(t, 0, table.Size())
}

func assertTrue(t *testing.T, b bool) {
	t.Helper()
	if !b {
		t.Fatal("expected true")
	}
}

func assertFalse(t *testing.T, b bool) {
	t.Helper()
	if b {
		t.Fatal("expected false")
	}
}

func assertInt(t *testing.T, expected int, current int) {
	t.Helper()
	if expected != current {
		t.Fatalf("expected %d but got %d", expected, current)
	}
}
