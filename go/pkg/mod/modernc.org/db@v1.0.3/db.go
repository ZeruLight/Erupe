// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//TODO https://blog.dgraph.io/post/badger-lmdb-boltdb/

// Package db implements selected data structures found in database
// implementations.
//
// Build status
//
// available at https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2fdb
package db // import "modernc.org/db"

import (
	"fmt"
	"os"

	"modernc.org/internal/buffer"
)

var (
	_ Storage = (*DB)(nil)
)

// Storage represents a database back end.
type Storage interface {
	// Alloc allocates a storage block large enough for storing size bytes
	// and returns its offset or an error, if any.
	Alloc(size int64) (int64, error)

	// Calloc is like Alloc but the allocated storage block is zeroed up
	// to size.
	Calloc(size int64) (int64, error)

	// Close finishes storage use.
	Close() error

	// Free recycles the allocated storage block at off.
	Free(off int64) error

	// ReadAt reads len(p) bytes into p starting at offset off in the
	// storage. It returns the number of bytes read (0 <= n <= len(p)) and
	// any error encountered.
	//
	// When ReadAt returns n < len(p), it returns a non-nil error
	// explaining why more bytes were not returned.
	//
	// Even if ReadAt returns n < len(p), it may use all of p as scratch
	// space during the call.
	//
	// If the n = len(p) bytes returned by ReadAt are at the end of the
	// storage, ReadAt may return either err == EOF or err == nil.
	ReadAt(p []byte, off int64) (n int, err error)

	// Realloc changes the size of the file block allocated at off, which
	// must have been returned from Alloc or Realloc, to size and returns
	// the offset of the relocated file block or an error, if any. The
	// contents will be unchanged in the range from the start of the region
	// up to the minimum of the old and new sizes. Realloc(off, 0) is equal
	// to Free(off). If the file block was moved, a Free(off) is done.
	Realloc(off, size int64) (int64, error)

	// Root returns the offset of the database root object or an error, if
	// any.  It's not an error if a newly created or empty database has no
	// root yet.  The returned offset in that case will be zero.
	Root() (int64, error)

	// SetRoot sets the offset of the database root object.
	SetRoot(root int64) error

	// Stat returns the os.FileInfo structure describing the storage. If
	// there is an error, it will be of type *os.PathError.
	Stat() (os.FileInfo, error)

	// Sync commits the current contents of the database to stable storage.
	// Typically, this means flushing the file system's in-memory copy of
	// recently written data to disk.
	Sync() error

	// Truncate changes the size of the storage. If there is an error, it
	// will be of type *os.PathError.
	Truncate(int64) error

	// WriteAt writes len(p) bytes from p to the storage at offset off. It
	// returns the number of bytes written from p (0 <= n <= len(p)) and
	// any error encountered that caused the write to stop early. WriteAt
	// must return a non-nil error if it returns n < len(p).
	WriteAt(p []byte, off int64) (n int, err error)
}

// DB represents a database.
type DB struct {
	Storage
}

// NewDB returns a newly created DB backed by s or an error, if any.
func NewDB(s Storage) (*DB, error) { return &DB{s}, nil }

func (db *DB) r4(off int64) (int, error)   { return r4(db, off) }
func (db *DB) r8(off int64) (int64, error) { return r8(db, off) }
func (db *DB) w4(off int64, n int) error   { return w4(db, off, n) }
func (db *DB) w8(off, n int64) error       { return w8(db, off, n) }

func r4(s Storage, off int64) (int, error) {
	p := buffer.Get(4)
	b := *p
	if n, err := s.ReadAt(b, off); n != 4 {
		if err == nil {
			err = fmt.Errorf("short storage read")
		}
		return 0, err
	}

	var n uint32
	for _, v := range b {
		n = n<<8 | uint32(v)
	}
	buffer.Put(p)
	return int(int32(n)), nil
}

func r8(s Storage, off int64) (int64, error) {
	p := buffer.Get(8)
	b := *p
	if n, err := s.ReadAt(b, off); n != 8 {
		if err == nil {
			err = fmt.Errorf("short storage read")
		}
		return 0, err
	}

	var n uint64
	for _, v := range b {
		n = n<<8 | uint64(v)
	}
	buffer.Put(p)
	return int64(n), nil
}

func r8b(b []byte) int64 {
	var n uint64
	for _, v := range b {
		n = n<<8 | uint64(v)
	}
	return int64(n)
}

func w4(s Storage, off int64, n int) error {
	p := buffer.Get(4)
	b := *p
	for i := range b {
		b[i] = byte(n >> 24)
		n <<= 8
	}
	_, err := s.WriteAt(b, off)
	buffer.Put(p)
	return err
}

func w8(s Storage, off, n int64) error {
	p := buffer.Get(8)
	b := *p
	for i := range b {
		b[i] = byte(n >> 56)
		n <<= 8
	}
	_, err := s.WriteAt(b, off)
	buffer.Put(p)
	return err
}
