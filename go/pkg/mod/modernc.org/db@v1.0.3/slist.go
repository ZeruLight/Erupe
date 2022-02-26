// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

const (
	oSListNext = 8 * iota // int64		0	8
	oSListData            // [dataSize]byte	8	dataSize
)

// SList is a node of a single linked list.
type SList struct {
	*DB
	Off int64 // Location in the database. R/O
}

// NewSList returns a newly allocated SList or an error, if any. The datasize
// parameter is the fixed size of data associated with the list node. To
// get/set the node data, use the ReadAt/WriteAt methods of db, using
// SList.DataOff() as the offset. Reading or writing more than datasize data at
// DataOff() is undefined behavior and may irreparably corrupt the database.
//
// The result of NewSList is not a part of any list.
func (db *DB) NewSList(dataSize int64) (SList, error) {
	off, err := db.Alloc(oSListData + dataSize)
	if err != nil {
		return SList{}, err
	}

	r, err := db.OpenSList(off)
	if err != nil {
		return SList{}, err
	}

	return r, r.setNext(0)
}

func (l SList) setNext(off int64) error { return l.w8(l.Off+oSListNext, off) }

// OpenSList returns an existing SList found at offset off or an error, of any.
// The off argument must have been acquired from NewSList.
func (db *DB) OpenSList(off int64) (SList, error) { return SList{db, off}, nil }

// DataOff returns the offset in db at which data of l are located.
func (l SList) DataOff() int64 { return l.Off + oSListData }

// Next returns the offset of the next node of l.
func (l SList) Next() (int64, error) { return l.r8(l.Off + oSListNext) }

// InsertAfter inserts l after the SList node at off. Node l must not be
// already a part of any list.
func (l SList) InsertAfter(off int64) error {
	n, err := l.OpenSList(off)
	if err != nil {
		return err
	}

	afterNext, err := n.Next()
	if err != nil {
		return err
	}

	if err = n.setNext(l.Off); err != nil {
		return err
	}

	return l.setNext(afterNext)
}

// InsertBefore inserts l before the SList node at off. If the SList node at
// off is linked to from an SList node at prev, the prev argument must reflect
// that, otherwise prev must be zero. Node l must not be already a part of any
// list.
func (l SList) InsertBefore(prev, off int64) error {
	n, err := l.OpenSList(off)
	if err != nil {
		return err
	}

	if prev != 0 {
		p, err := l.OpenSList(prev)
		if err != nil {
			return err
		}

		if err := p.setNext(l.Off); err != nil {
			return err
		}
	}
	return l.setNext(n.Off)
}

// Remove removes l from a list. If l is linked to from an SList node at prev,
// the prev argument must reflect that, otherwise prev must be zero.
//
// The free function may be nil, otherwise it's called with the result of
// l.DataOff() before removing l.
func (l SList) Remove(prev int64, free func(off int64) error) error {
	if free != nil {
		if err := free(l.DataOff()); err != nil {
			return err
		}
	}

	if prev != 0 {
		next, err := l.Next()
		if err != nil {
			return err
		}

		p, err := l.OpenSList(prev)
		if err != nil {
			return err
		}

		if err := p.setNext(next); err != nil {
			return err
		}
	}
	return l.Free(l.Off)
}

// RemoveToLast removes all nodes from a list starting at l to the end of the
// list. If l is linked to from an SList node at prev, the prev argument must
// reflect that, otherwise prev must be zero.
//
// The free function may be nil, otherwise it's called with the result of
// l.DataOff() before removing l.
func (l SList) RemoveToLast(prev int64, free func(off int64) error) error {
	if prev != 0 {
		p, err := l.OpenSList(prev)
		if err != nil {
			return err
		}

		if err := p.setNext(0); err != nil {
			return err
		}
	}
	for l.Off != 0 {
		if free != nil {
			if err := free(l.DataOff()); err != nil {
				return err
			}
		}

		next, err := l.Next()
		if err != nil {
			return err
		}

		if err := l.Free(l.Off); err != nil {
			return err
		}

		l.Off = next
	}
	return nil
}
