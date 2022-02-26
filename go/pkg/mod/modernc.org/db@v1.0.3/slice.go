// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

import (
	"fmt"

	"modernc.org/internal/buffer"
	"modernc.org/mathutil"
)

const (
	oSliceLen    = 8 * iota // int64	0	8
	oSliceSzItem            // int64	8	8
	oSliceTable             // [64]int64	16	512

	/*

		index	tabIndex	N	index+1
		0:	tab[0]		1	1:	2^0
		1-2:	tab[1]		2	2-3:	2^1
		3-6:	tab[2]		4	4-7:	2^2
		7-14:	tab[3]		8	8-15:	2^3
		...
			tab[63] 			2^63

	*/

	szSlice = oSliceTable + 64*8
)

// Slice is a numbered sequence of items.
type Slice struct {
	*DB
	Len    int64 // The number of items in the slice. R/O
	Off    int64 // Location in the database. R/O
	SzItem int64 // The szItem argument of NewSlice. R/O
	table  [64]int64
}

// NewSlice allocates and returns a new Slice or an error, if any.  The szItem
// argument is the size of an item. The resulting slice has zero length.
func (db *DB) NewSlice(szItem int64) (s *Slice, err error) {
	if szItem < 0 {
		panic(fmt.Errorf("%T.NewSlice: invalid argument", db))
	}

	off, err := db.Alloc(szSlice)
	if err != nil {
		return nil, fmt.Errorf("%T.NewSlice: %v", db, err)
	}

	if err = db.w8(off+oSliceSzItem, szItem); err != nil {
		db.Free(off)
		return nil, fmt.Errorf("%T.NewSlice: %v", db, err)
	}

	return &Slice{
		DB:     db,
		Off:    off,
		SzItem: szItem,
	}, nil
}

// OpenSlice returns an existing Slice found at offset off or an error, if any.
// The off argument must have been acquired from NewSlice.
func (db *DB) OpenSlice(off int64) (*Slice, error) {
	szItem, err := db.r8(off + oSliceSzItem)
	if err != nil {
		return nil, err
	}

	if szItem < 0 {
		return nil, fmt.Errorf("%T.OpenSlice: corrupted database", db)
	}

	len, err := db.r8(off + oSliceLen)
	if err != nil {
		return nil, err
	}

	if len < 0 {
		return nil, fmt.Errorf("%T.OpenSlice: corrupted database", db)
	}

	p := buffer.Get(64 * 8)

	defer buffer.Put(p)

	b := *p
	if _, err := db.ReadAt(b, off+oSliceTable); err != nil {
		return nil, fmt.Errorf("%T.OpenSlice: corrupted database", db)
	}

	r := &Slice{
		DB:     db,
		Len:    len,
		Off:    off,
		SzItem: szItem,
	}

	for i := range r.table {
		o := 8 * i
		r.table[i] = r8b(b[o : o+8])
	}

	return r, nil
}

// Append returns the offset of item s[s.Len] and increments s.Len. The storage
// space at the returned offset is not touched and may contain garbage. Callers
// are expected to initialize/set the item after calling Append.
func (s *Slice) Append() (off int64, err error) {
	si := s.slotIndex(s.Len)
	block := s.table[si]
	if block == 0 {
		bsz := s.SzItem << uint(si)
		if block, err = s.Alloc(bsz); err != nil {
			return 0, fmt.Errorf("%T.Append: %v", s, err)
		}

		if err = s.w8(s.Off+oSliceTable+int64(si)*8, block); err != nil {
			return 0, fmt.Errorf("%T.Append: %v", s, err)
		}

		s.table[si] = block
	}
	bItems := int64(1) << uint(si)
	ix := s.Len - bItems + 1
	s.Len++
	return block + ix*s.SzItem, s.w8(s.Off+oSliceLen, s.Len)
}

func (s *Slice) slotIndex(index int64) int { return mathutil.Log2Uint64(uint64(index + 1)) }

// RemoveLast removes the last item of s and decrements s.Len.
//
// The free function may be nil, otherwise it's called with offset of the last
// item in s before removing it.
func (s *Slice) RemoveLast(free func(off int64) error) (err error) {
	if s.Len == 0 {
		return fmt.Errorf("%T.RemoveLast: no items to remove", s)
	}

	freeBlock := s.Len&(s.Len-1) == 0
	s.Len--
	si := s.slotIndex(s.Len)
	block := s.table[si]
	if free != nil {
		bItems := int64(1) << uint(si)
		ix := s.Len - bItems + 1
		if err = free(block + ix*s.SzItem); err != nil {
			return fmt.Errorf("%T.RemoveLast: %v", s, err)
		}
	}
	if freeBlock {
		if err = s.Free(block); err != nil {
			return fmt.Errorf("%T.RemoveLast: %v", s, err)
		}

		if err = s.w8(s.Off+oSliceTable+int64(si)*8, block); err != nil {
			return fmt.Errorf("%T.RemoveLast: %v", s, err)
		}

		s.table[si] = block
	}
	return s.w8(s.Off+oSliceLen, s.Len)
}

// At returns the offset if item at index i.
func (s *Slice) At(i int64) (int64, error) {
	if i < 0 || i >= s.Len {
		return 0, fmt.Errorf("%T.At: index out of bounds", s)
	}

	si := s.slotIndex(i)
	block := s.table[si]
	bItems := int64(1) << uint(si)
	ix := i - bItems + 1
	return block + ix*s.SzItem, nil
}
