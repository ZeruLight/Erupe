// Copyright 2018 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

import (
	"testing"

	"modernc.org/file"
)

func testSLice(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	const N = 223456

	db, f := tmpDB(t, ts)

	defer f()

	s, err := db.NewSlice(8)
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		for s.Len != 0 {
			if err := s.RemoveLast(nil); err != nil {
				t.Error(err)
				return
			}
		}
		if err := db.Free(s.Off); err != nil {
			t.Error(err)
		}
	}()

	for i := 0; i < N; i++ {
		off, err := s.Append()
		if err != nil {
			t.Error(err)
			return
		}

		if err = db.w8(off, int64(i)); err != nil {
			t.Error(err)
			return
		}
	}
	fi, err := s.Stat()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(N, fi.Size(), float64(fi.Size())/float64(N))

	for i := 0; i < N; i++ {
		off, err := s.At(int64(i))
		if err != nil {
			t.Error(err)
			return
		}

		n, err := s.r8(off)
		if err != nil {
			t.Error(err)
			return
		}

		if g, e := n, int64(i); g != e {
			t.Error(i, g, e)
			return
		}
	}

	if g, e := s.Len, int64(N); g != e {
		t.Error(g, e)
		return
	}

	if s, err = db.OpenSlice(s.Off); err != nil {
		t.Error(err)
		return
	}

	if g, e := s.Len, int64(N); g != e {
		t.Error(g, e)
		return
	}

	for i := 0; i < N; i++ {
		off, err := s.At(int64(i))
		if err != nil {
			t.Error(err)
			return
		}

		n, err := s.r8(off)
		if err != nil {
			t.Error(err)
			return
		}

		if g, e := n, int64(i); g != e {
			t.Error(i, g, e)
			return
		}
	}
}

func TestSlice(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testSLice(t, v.f) }) {
			break
		}
	}
}
