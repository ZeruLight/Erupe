// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

const (
	oDListPrev = 8 * iota // int64		0	8
	oDListNext            // int64		8	8
	oDListData            // [dataSize]byte	16	dataSize
)

// DList is a node of a doubly linked list.
type DList struct {
	*DB
	Off int64 // Location in the database. R/O
}

// NewDList returns a newly allocated DList or an error, if any. The datasize
// parameter is the fixed size of data associated with the list node. To
// get/set the node data, use the ReadAt/WriteAt methods of db, using
// DList.DataOff() as the offset. Reading or writing more than datasize data at
// DataOff() is undefined behavior and may irreparably corrupt the database.
//
// The result of NewDList is not a part of any list.
func (db *DB) NewDList(dataSize int64) (DList, error) {
	off, err := db.Alloc(oDListData + dataSize)
	if err != nil {
		return DList{}, err
	}

	r, err := db.OpenDList(off)
	if err != nil {
		return DList{}, err
	}

	if err := r.setPrev(0); err != nil {
		return DList{}, err
	}

	return r, r.setNext(0)
}

// OpenDList returns an existing DList found at offset off or an error, if any.
// The off argument must have been acquired from NewDList.
func (db *DB) OpenDList(off int64) (DList, error) { return DList{db, off}, nil }

func (l DList) setNext(off int64) error { return l.w8(l.Off+oDListNext, off) }
func (l DList) setPrev(off int64) error { return l.w8(l.Off+oDListPrev, off) }

// DataOff returns the offset in db at which data of l are located.
func (l DList) DataOff() int64 { return l.Off + oDListData }

// Next returns the offset of the next node of l.
func (l DList) Next() (int64, error) { return l.r8(l.Off + oDListNext) }

// Prev returns the offset of the prev node of l.
func (l DList) Prev() (int64, error) { return l.r8(l.Off + oDListPrev) }

// InsertAfter inserts l after the DList node at off. Node l must not be
// already a part of any list.
func (l DList) InsertAfter(off int64) error {
	n, err := l.OpenDList(off)
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

	if err = l.setPrev(off); err != nil {
		return err
	}

	if afterNext != 0 {
		n, err := l.OpenDList(afterNext)
		if err != nil {
			return err
		}

		if err := n.setPrev(l.Off); err != nil {
			return err
		}
	}

	return l.setNext(afterNext)
}

// InsertBefore inserts l before the DList node at off. Node l must not be
// already a part of any list.
func (l DList) InsertBefore(off int64) error {
	n, err := l.OpenDList(off)
	if err != nil {
		return err
	}

	prev, err := n.Prev()
	if err != nil {
		return err
	}

	if n.setPrev(l.Off); err != nil {
		return err
	}

	if prev != 0 {
		p, err := l.OpenDList(prev)
		if err != nil {
			return err
		}

		if err := p.setNext(l.Off); err != nil {
			return err
		}
	}

	if l.setPrev(prev); err != nil {
		return err
	}

	return l.setNext(n.Off)
}

// Remove removes l from a list.
//
// The free function may be nil, otherwise it's called with the result of
// l.DataOff() before removing l.
func (l DList) Remove(free func(off int64) error) error {
	if free != nil {
		if err := free(l.DataOff()); err != nil {
			return err
		}
	}

	prev, err := l.Prev()
	if err != nil {
		return err
	}

	next, err := l.Next()
	if err != nil {
		return err
	}

	var p, n DList
	if prev != 0 {
		if p, err = l.OpenDList(prev); err != nil {
			return err
		}
	}
	if next != 0 {
		if n, err = l.OpenDList(next); err != nil {
			return err
		}
	}

	switch {
	case prev == 0:
		switch {
		case next == 0:
			// nop
		default:
			if err = n.setPrev(0); err != nil {
				return err
			}
		}
	default:
		switch {
		case next == 0:
			if err = p.setNext(0); err != nil {
				return err
			}
		default:
			if err := p.setNext(next); err != nil {
				return err
			}

			if err := n.setPrev(prev); err != nil {
				return err
			}
		}
	}
	return l.Free(l.Off)
}

// RemoveToLast removes all nodes from a list starting at l to the end of the
// list.
//
// The free function may be nil, otherwise it's called with the result of
// l.DataOff() before removing l.
func (l DList) RemoveToLast(free func(off int64) error) error {
	prev, err := l.Prev()
	if err != nil {
		return err
	}

	if prev != 0 {
		p, err := l.OpenDList(prev)
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

// RemoveToFirst removes all nodes from a list starting at first list node, up
// to and including l.
//
// The free function may be nil, otherwise it's called with the result of
// l.DataOff() before removing l.
func (l DList) RemoveToFirst(free func(off int64) error) error {
	next, err := l.Next()
	if err != nil {
		return err
	}

	if next != 0 {
		n, err := l.OpenDList(next)
		if err != nil {
			return err
		}

		if err := n.setPrev(0); err != nil {
			return err
		}
	}
	for l.Off != 0 {
		if free != nil {
			if err := free(l.DataOff()); err != nil {
				return err
			}
		}

		prev, err := l.Prev()
		if err != nil {
			return err
		}

		if err := l.Free(l.Off); err != nil {
			return err
		}

		l.Off = prev
	}
	return nil
}
