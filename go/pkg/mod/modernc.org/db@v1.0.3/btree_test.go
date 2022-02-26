// Copyright 2014 The b Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE-B file.

// Modifications are
//
// Copyright 2017 The DB Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db // import "modernc.org/db"

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"testing"

	"modernc.org/file"
	"modernc.org/strutil"
)

func (t *BTree) cmp(n int) func(off int64) (int, error) {
	return func(off int64) (int, error) {
		p, err := t.r8(off)
		if err != nil {
			return 0, err
		}

		m, err := t.r4(p)
		if err != nil {
			return 0, err
		}

		if n < m {
			return -1, nil
		}

		if n > m {
			return 1, nil
		}

		return 0, nil
	}
}

func (t *BTree) bcmp(n int) func(off int64) (int, error) {
	return func(off int64) (int, error) {
		m, err := t.r4(off)
		if err != nil {
			return 0, err
		}

		if n < m {
			return -1, nil
		}

		if n > m {
			return 1, nil
		}

		return 0, nil
	}
}

func (t *BTree) tlen(tb testing.TB) int64 {
	c, err := t.Len()
	if err != nil {
		tb.Fatal(err)
	}

	return c
}

func (t *BTree) get(tb testing.TB, k int) (y int, yy bool) {
	off, ok, err := t.Get(t.cmp(k))
	if err != nil {
		tb.Fatal(err)
	}

	if !ok {
		return 0, false
	}

	p, err := t.r8(off)
	if err != nil {
		tb.Fatal(err)
	}

	n, err := t.r4(p)
	if err != nil {
		tb.Fatal(err)
	}

	return n, true
}

func (t *BTree) bget(tb testing.TB, k int) {
	if _, _, err := t.Get(t.bcmp(k)); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) set(tb testing.TB, k, v int) {
	kalloc := true
	koff, voff, err := t.Set(t.cmp(k), func(off int64) error {
		p, err := t.r8(off)
		if err != nil {
			return err
		}

		kalloc = false
		return t.Free(p)
	})
	if err != nil {
		tb.Fatal(err)
	}

	var p, q int64
	if kalloc {
		if p, err = t.Alloc(4); err != nil {
			tb.Fatal(err)
		}

		if err := t.w4(p, k); err != nil {
			tb.Fatal(err)
		}

		if err := t.w8(koff, p); err != nil {
			tb.Fatal(err)
		}
	}

	if q, err = t.Alloc(4); err != nil {
		tb.Fatal(err)
	}

	if err := t.w4(q, v); err != nil {
		tb.Fatal(err)
	}

	if err := t.w8(voff, q); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) bset(tb testing.TB, k int) {
	koff, _, err := t.Set(t.bcmp(k), nil)
	if err != nil {
		tb.Fatal(err)
	}

	if err := t.w4(koff, k); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) clear(tb testing.TB) {
	if err := t.Clear(func(k, v int64) error {
		p, err := t.r8(k)
		if err != nil {
			return err
		}

		if err := t.Free(p); err != nil {
			return err
		}

		q, err := t.r8(v)
		if err != nil {
			return err
		}

		return t.Free(q)
	}); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) delete(tb testing.TB, k int) bool {
	ok, err := t.Delete(t.cmp(k), func(k, v int64) error {
		p, err := t.r8(k)
		if err != nil {
			return err
		}

		if err := t.Free(p); err != nil {
			return err
		}

		q, err := t.r8(v)
		if err != nil {
			return err
		}

		return t.Free(q)
	})
	if err != nil {
		tb.Fatal(err)
	}

	return ok
}

func (t *BTree) bdelete(tb testing.TB, k int) {
	if _, err := t.Delete(t.bcmp(k), nil); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) remove(tb testing.TB) {
	if err := t.Remove(func(k, v int64) error {
		p, err := t.r8(k)
		if err != nil {
			return err
		}

		if err := t.Free(p); err != nil {
			return err
		}

		q, err := t.r8(v)
		if err != nil {
			return err
		}

		return t.Free(q)
	}); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) bremove(tb testing.TB) {
	if err := t.Remove(nil); err != nil {
		tb.Fatal(err)
	}
}

func (t *BTree) seek(tb testing.TB, k int) (*BTreeCursor, bool) {
	en, hit, err := t.Seek(t.cmp(k))
	if err != nil {
		tb.Fatal(err)
	}

	return en, hit
}

func (t *BTree) bseek(tb testing.TB, k int) (*BTreeCursor, bool) {
	en, hit, err := t.Seek(t.bcmp(k))
	if err != nil {
		tb.Fatal(err)
	}

	return en, hit
}

func (t *BTree) seekFirst(tb testing.TB) *BTreeCursor {
	en, err := t.SeekFirst()
	if err != nil {
		tb.Fatal(err)
	}

	return en
}

func (t *BTree) seekLast(tb testing.TB) *BTreeCursor {
	en, err := t.SeekLast()
	if err != nil {
		tb.Fatal(err)
	}

	return en
}

func (e *BTreeCursor) next(tb testing.TB) (int, int, bool) {
	if e.Next() {
		p, err := e.t.r8(e.K)
		if err != nil {
			tb.Fatal(err)
		}

		k, err := e.t.r4(p)
		if err != nil {
			tb.Fatal(err)
		}

		q, err := e.t.r8(e.V)
		if err != nil {
			tb.Fatal(err)
		}

		v, err := e.t.r4(q)
		if err != nil {
			tb.Fatal(err)
		}

		return k, v, true
	}

	return 0, 0, false
}

func (e *BTreeCursor) prev(tb testing.TB) (int, int, bool) {
	if e.Prev() {
		p, err := e.t.r8(e.K)
		if err != nil {
			tb.Fatal(err)
		}

		k, err := e.t.r4(p)
		if err != nil {
			tb.Fatal(err)
		}

		q, err := e.t.r8(e.V)
		if err != nil {
			tb.Fatal(err)
		}

		v, err := e.t.r4(q)
		if err != nil {
			tb.Fatal(err)
		}

		return k, v, true
	}

	return 0, 0, false
}

func (t *BTree) dump() (r string) {
	var buf bytes.Buffer

	defer func() {
		if err := recover(); err != nil {
			dbg("%q\n%s", err, debug.Stack())
			r = fmt.Sprint(err)
		}
	}()

	f := strutil.IndentFormatter(&buf, "\t")

	num := map[int64]int{}
	visited := map[int64]bool{}

	handle := func(off int64) int {
		if off == 0 {
			return 0
		}

		if n, ok := num[off]; ok {
			return n
		}

		n := len(num) + 1
		num[off] = n
		return n
	}

	var pagedump func(int64, string)
	pagedump = func(off int64, pref string) {
		if off == 0 || visited[off] {
			return
		}

		visited[off] = true
		p, err := t.openPage(off)
		if err != nil {
			panic(err)
		}

		switch x := p.(type) {
		case btDPage:
			c, err := t.len(x)
			if err != nil {
				panic(err)
			}

			p, err := t.prev(x)
			if err != nil {
				panic(err)
			}

			n, err := t.next(x)
			if err != nil {
				panic(err)
			}

			f.Format("%sD#%d(%#x) P#%d N#%d len %d {", pref, handle(off), off, handle(int64(p)), handle(int64(n)), c)
			for i := 0; i < c; i++ {
				if i != 0 {
					f.Format(" ")
				}
				koff := t.key(x, i)
				voff := t.val(x, i)
				p, err := t.r8(koff)
				if err != nil {
					panic(err)
				}

				k, err := t.r4(p)
				if err != nil {
					panic(err)
				}

				q, err := t.r8(voff)
				if err != nil {
					panic(err)
				}

				v, err := t.r4(q)
				if err != nil {
					panic(err)
				}

				f.Format("%v:%v", k, v)
			}
			f.Format("}\n")
		case btXPage:
			c, err := t.lenX(x)
			if err != nil {
				panic(err)
			}

			f.Format("%sX#%d(%#x) len %d {", pref, handle(off), off, c)
			a := []int64{}
			for i := 0; i <= c; i++ {
				ch, err := t.child(x, i)
				if err != nil {
					panic(err)
				}

				a = append(a, ch)
				if i != 0 {
					f.Format(" ")
				}
				f.Format("(C#%d(%#x)", handle(ch), ch)
				if i != c {
					ko, err := t.keyX(x, i)
					if err != nil {
						panic(err)
					}

					p, err := t.r8(ko)
					if err != nil {
						panic(err)
					}

					k, err := t.r4(p)
					if err != nil {
						panic(err)
					}

					f.Format(" K %v(%#x))", k, ko)
				}
				f.Format(")")
			}
			f.Format("}\n")
			for _, p := range a {
				pagedump(p, pref+". ")
			}
		default:
			panic(fmt.Errorf("%T", x))
		}
	}

	root, err := t.root()
	if err != nil {
		return err.Error()
	}

	pagedump(root, "")
	s := buf.String()
	if s != "" {
		s = s[:len(s)-1]
	}
	return s
}

func testBTreeGet0(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	g := func() {
		if g, e := bt.tlen(t), int64(0); g != e {
			t.Fatal(g, e)
		}

		_, ok := bt.get(t, 42)
		if g, e := ok, false; g != e {
			t.Fatal(g, e)
		}
	}

	g()
	if bt, err = db.OpenBTree(bt.Off); err != nil {
		t.Fatal(err)
	}
	g()
}

func TestBTreeGet0(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeGet0(t, v.f) }) {
			break
		}
	}
}

func testBTreeSetGet0(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	g := func() {
		bt.clear(t)
		bt.set(t, 42, 314)
		if g, e := bt.tlen(t), int64(1); g != e {
			t.Fatal(g, e)
		}

		v, ok := bt.get(t, 42)
		if !ok {
			t.Fatal(ok)
		}

		if g, e := v, 314; g != e {
			t.Fatal(g, e)
		}

		bt.set(t, 42, 278)
		if g, e := bt.tlen(t), int64(1); g != e {
			t.Fatal(g, e)
		}

		v, ok = bt.get(t, 42)
		if !ok {
			t.Fatal(ok)
		}

		if g, e := v, 278; g != e {
			t.Fatal(g, e)
		}

		bt.set(t, 420, 5)
		if g, e := bt.tlen(t), int64(2); g != e {
			t.Fatal(g, e)
		}

		v, ok = bt.get(t, 42)
		if !ok {
			t.Fatal(ok)
		}

		if g, e := v, 278; g != e {
			t.Fatal(g, e)
		}

		v, ok = bt.get(t, 420)
		if !ok {
			t.Fatal(ok)
		}

		if g, e := v, 5; g != e {
			t.Fatal(g, e)
		}
	}

	g()
	if bt, err = db.OpenBTree(bt.Off); err != nil {
		t.Fatal(err)
	}
	g()
}

func TestBTreeSetGet0(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSetGet0(t, v.f) }) {
			break
		}
	}
}

func testBTreeSetGet1(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	const N = 1 << 10
	for _, x := range []int{0, -1, 0x5555555, 0xaaaaaaa, 0x3333333, 0xccccccc, 0x31415926, 0x2718282} {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(16, 16, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			a := make([]int, N)
			for i := range a {
				a[i] = (i ^ x) << 1
			}
			for i, k := range a {
				bt.set(t, k, k^x)
				if g, e := bt.tlen(t), int64(i+1); g != e {
					t.Fatal(i, g, e)
				}
			}

			g := func() {
				for i, k := range a {
					v, ok := bt.get(t, k)
					if !ok {
						t.Fatal(i, k, ok)
					}

					if g, e := v, k^x; g != e {
						t.Fatal(i, g, e)
					}

					k |= 1
					_, ok = bt.get(t, k)
					if ok {
						t.Fatal(i, k)
					}
				}
			}

			g()
			if bt, err = db.OpenBTree(bt.Off); err != nil {
				t.Fatal(err)
			}
			g()

			for _, k := range a {
				bt.set(t, k, (k^x)+42)
			}

			g = func() {
				for i, k := range a {
					v, ok := bt.get(t, k)
					if !ok {
						t.Fatal(i, k, v, ok)
					}

					if g, e := v, k^x+42; g != e {
						t.Fatal(i, g, e)
					}

					k |= 1
					_, ok = bt.get(t, k)
					if ok {
						t.Fatal(i, k)
					}
				}
			}

			g()
			if bt, err = db.OpenBTree(bt.Off); err != nil {
				t.Fatal(err)
			}
			g()
		}()
	}
}

func TestBTreeSetGet1(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSetGet1(t, v.f) }) {
			break
		}
	}
}

// verify how splitX works when splitting X for k pointing directly at split edge
func testBTreeSplitXOnEdge(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	g := func() {
		bt.clear(t)
		kd := bt.kd
		kx := bt.kx

		// one index page with 2*kx+2 elements (last has .k=∞  so x.c=2*kx+1)
		// which will splitX on next Set
		for i := 0; i <= (2*kx+1)*2*kd; i++ {
			// odd keys are left to be filled in second test
			bt.set(t, 2*i, 2*i)
		}

		r, err := bt.root()
		if err != nil {
			t.Fatal(err)
		}

		x0 := bt.openXPage(r)
		x0c, err := bt.lenX(x0)
		if err != nil {
			t.Fatal(err)
		}

		if x0c != 2*kx+1 {
			t.Fatalf("x0.c: %v  ; expected %v", x0c, 2*kx+1)
		}

		// set element with k directly at x0[kx].k
		kedge := 2 * (kx + 1) * (2 * kd)
		pk, err := bt.keyX(x0, kx)
		if err != nil {
			t.Fatal(err)
		}

		if pk, err = bt.r8(pk); err != nil {
			t.Fatal(err)
		}

		k, err := bt.r4(pk)
		if err != nil {
			t.Fatal(err)
		}

		if k != kedge {
			t.Fatalf("edge key before splitX: %v  ; expected %v", k, kedge)
		}

		bt.set(t, kedge, 777)

		// if splitX was wrong kedge:777 would land into wrong place with Get failing
		v, ok := bt.get(t, kedge)
		if !(v == 777 && ok) {
			t.Fatalf("after splitX: Get(%v) -> %v, %v  ; expected 777, true", kedge, v, ok)
		}

		// now check the same when splitted X has parent
		if r, err = bt.root(); err != nil {
			t.Fatal(err)
		}

		xr := bt.openXPage(r)
		xrc, err := bt.lenX(xr)
		if err != nil {
			t.Fatal(err)
		}

		if xrc != 1 { // second x comes with k=∞ with .c index
			t.Fatalf("after splitX: xr.c: %v  ; expected 1", xrc)
		}

		xr0ch, err := bt.child(xr, 0)
		if err != nil {
			t.Fatal(err)
		}

		if xr0ch != int64(x0) {
			t.Fatal("xr[0].ch is not x0")
		}

		for i := 0; i <= (2*kx)*kd; i++ {
			bt.set(t, 2*i+1, 2*i+1)
		}

		// check x0 is in pre-splitX condition and still at the right place
		if x0c, err = bt.lenX(x0); err != nil {
			t.Fatal(err)
		}

		if x0c != 2*kx+1 {
			t.Fatalf("x0.c: %v  ; expected %v", x0c, 2*kx+1)
		}

		if xr0ch, err = bt.child(xr, 0); err != nil {
			t.Fatal(err)
		}

		if xr0ch != int64(x0) {
			t.Fatal("xr[0].ch is not x0")
		}

		// set element with k directly at x0[kx].k
		kedge = (kx + 1) * (2 * kd)
		if pk, err = bt.keyX(x0, kx); err != nil {
			t.Fatal(err)
		}

		if pk, err = bt.r8(pk); err != nil {
			t.Fatal(err)
		}

		x0kxk, err := bt.r4(pk)
		if err != nil {
			t.Fatal(err)
		}

		if x0kxk != kedge {
			t.Fatalf("edge key before splitX: %v  ; expected %v", x0kxk, kedge)
		}

		bt.set(t, kedge, 888)

		// if splitX was wrong kedge:888 would land into wrong place
		v, ok = bt.get(t, kedge)
		if !(v == 888 && ok) {
			t.Fatalf("after splitX: Get(%v) -> %v, %v  ; expected 888, true", kedge, v, ok)
		}
	}

	g()
	if bt, err = db.OpenBTree(bt.Off); err != nil {
		t.Fatal(err)
	}
	g()
}

func TestBTreeSplitXOnEdge(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSplitXOnEdge(t, v.f) }) {
			break
		}
	}
}

func testBTreeSetGet2(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	const N = 1 << 10
	for _, x := range []int{0, -1, 0x5555555, 0xaaaaaaa, 0x3333333, 0xccccccc, 0x31415926, 0x2718282} {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(16, 16, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			rng := rng()
			a := make([]int, N)
			for i := range a {
				a[i] = (rng.Next() ^ x) << 1
			}
			for i, k := range a {
				bt.set(t, k, k^x)
				if g, e := bt.tlen(t), int64(i)+1; g != e {
					t.Fatal(i, x, g, e)
				}
			}

			g := func() {
				for i, k := range a {
					v, ok := bt.get(t, k)
					if !ok {
						t.Fatal(i, k, v, ok)
					}

					if g, e := v, k^x; g != e {
						t.Fatal(i, g, e)
					}

					k |= 1
					_, ok = bt.get(t, k)
					if ok {
						t.Fatal(i, k)
					}
				}
			}

			g()
			if bt, err = db.OpenBTree(bt.Off); err != nil {
				t.Fatal(err)
			}
			g()

			for _, k := range a {
				bt.set(t, k, (k^x)+42)
			}

			g = func() {
				for i, k := range a {
					v, ok := bt.get(t, k)
					if !ok {
						t.Fatal(i, k, v, ok)
					}

					if g, e := v, k^x+42; g != e {
						t.Fatal(i, g, e)
					}

					k |= 1
					_, ok = bt.get(t, k)
					if ok {
						t.Fatal(i, k)
					}
				}
			}

			g()
			if bt, err = db.OpenBTree(bt.Off); err != nil {
				t.Fatal(err)
			}
			g()
		}()
	}
}

func TestBTreeSetGet2(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSetGet2(t, v.f) }) {
			break
		}
	}
}

func testBTreeSetGet3(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	var i int
	for i = 0; ; i++ {
		bt.set(t, i, -i)
		r, err := bt.root()
		if err != nil {
			t.Fatal(err)
		}

		p, err := bt.openPage(r)
		if err != nil {
			t.Fatal(err)
		}

		if _, ok := p.(btXPage); ok {
			break
		}
	}
	for j := 0; j <= i; j++ {
		bt.set(t, j, j)
	}

	for j := 0; j <= i; j++ {
		v, ok := bt.get(t, j)
		if !ok {
			t.Fatal(j)
		}

		if g, e := v, j; g != e {
			t.Fatal(g, e)
		}
	}
}

func TestBTreeSetGet3(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSetGet3(t, v.f) }) {
			break
		}
	}
}

func testBTreeDelete0(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	if ok := bt.delete(t, 0); ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(0); g != e {
		t.Fatal(g, e)
	}

	bt.set(t, 0, 0)
	if ok := bt.delete(t, 1); ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(1); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 0); !ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(0); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 0); ok {
		t.Fatal(ok)
	}

	bt.set(t, 0, 0)
	bt.set(t, 1, 1)
	if ok := bt.delete(t, 1); !ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(1); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 1); ok {
		t.Fatal(ok)
	}

	if ok := bt.delete(t, 0); !ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(0); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 0); ok {
		t.Fatal(ok)
	}

	bt.set(t, 0, 0)
	bt.set(t, 1, 1)
	if ok := bt.delete(t, 0); !ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(1); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 0); ok {
		t.Fatal(ok)
	}

	if ok := bt.delete(t, 1); !ok {
		t.Fatal(ok)
	}

	if g, e := bt.tlen(t), int64(0); g != e {
		t.Fatal(g, e)
	}

	if ok := bt.delete(t, 1); ok {
		t.Fatal(ok)
	}
}

func TestBTreeDelete0(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeDelete0(t, v.f) }) {
			break
		}
	}
}

func testBTreeDelete1(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	const N = 1 << 11
	for _, x := range []int{0, -1, 0x5555555, 0xaaaaaaa, 0x3333333, 0xccccccc, 0x31415926, 0x2718282} {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(16, 16, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			a := make([]int, N)
			for i := range a {
				a[i] = (i ^ x) << 1
			}
			for _, k := range a {
				bt.set(t, k, 0)
			}
			for i, k := range a {
				ok := bt.delete(t, k)
				if !ok {
					t.Fatal(i, x, k)
				}

				if g, e := bt.tlen(t), int64(N-i-1); g != e {
					t.Fatal(i, g, e)
				}
			}
		}()
	}
}

func TestBTreeDelete1(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeDelete1(t, v.f) }) {
			break
		}
	}
}

func testBTreeDelete2(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	const N = 1 << 11
	for _, x := range []int{0, -1, 0x5555555, 0xaaaaaaa, 0x3333333, 0xccccccc, 0x31415926, 0x2718282} {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(16, 16, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			a := make([]int, N)
			rng := rng()
			for i := range a {
				a[i] = (rng.Next() ^ x) << 1
			}
			for _, k := range a {
				bt.set(t, k, 0)
			}

			for i, k := range a {
				ok := bt.delete(t, k)
				if !ok {
					t.Fatal(i, x, k)
				}

				if g, e := bt.tlen(t), int64(N-i-1); g != e {
					t.Fatal(i, g, e)
				}
			}
		}()
	}
}

func TestBTreeDelete2(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeDelete2(t, v.f) }) {
			break
		}
	}
}

func testBTreeEnumeratorNext(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(2, 4, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	// seeking within 5 keys: 10, 20, 30, 40, 50
	table := []struct {
		k    int
		hit  bool
		keys []int
	}{
		{5, false, []int{10, 20, 30, 40, 50}},
		{10, true, []int{10, 20, 30, 40, 50}},
		{15, false, []int{20, 30, 40, 50}},
		{20, true, []int{20, 30, 40, 50}},
		{25, false, []int{30, 40, 50}},
		// 5
		{30, true, []int{30, 40, 50}},
		{35, false, []int{40, 50}},
		{40, true, []int{40, 50}},
		{45, false, []int{50}},
		{50, true, []int{50}},
		// 10
		{55, false, nil},
	}

	for i, test := range table {
		keys := test.keys

		bt.set(t, 10, 100)
		bt.set(t, 20, 200)
		bt.set(t, 30, 300)
		bt.set(t, 40, 400)
		bt.set(t, 50, 500)

		en, hit := bt.seek(t, test.k)

		if g, e := hit, test.hit; g != e {
			t.Fatal(i, g, e)
		}

		j := 0
		for {
			k, v, ok := en.next(t)
			if !ok {
				if err := en.Err(); err != nil {
					t.Fatal(i, err)
				}

				break
			}

			if j >= len(keys) {
				t.Fatal(i, j, len(keys))
			}

			if g, e := k, keys[j]; g != e {
				t.Fatal(i, j, g, e)
			}

			if g, e := v, 10*keys[j]; g != e {
				t.Fatal(i, g, e)
			}

			j++

		}

		if g, e := j, len(keys); g != e {
			t.Fatal(i, j, g, e)
		}
	}
}

func TestBTreeEnumeratorNext(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeEnumeratorNext(t, v.f) }) {
			break
		}
	}
}

func testBTreeEnumeratorPrev(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(2, 4, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	// seeking within 5 keys: 10, 20, 30, 40, 50
	table := []struct {
		k    int
		hit  bool
		keys []int
	}{
		{5, false, nil},
		{10, true, []int{10}},
		{15, false, []int{10}},
		{20, true, []int{20, 10}},
		{25, false, []int{20, 10}},
		// 5
		{30, true, []int{30, 20, 10}},
		{35, false, []int{30, 20, 10}},
		{40, true, []int{40, 30, 20, 10}},
		{45, false, []int{40, 30, 20, 10}},
		{50, true, []int{50, 40, 30, 20, 10}},
		// 10
		{55, false, []int{50, 40, 30, 20, 10}},
	}

	for i, test := range table {
		keys := test.keys

		bt.set(t, 10, 100)
		bt.set(t, 20, 200)
		bt.set(t, 30, 300)
		bt.set(t, 40, 400)
		bt.set(t, 50, 500)

		en, hit := bt.seek(t, test.k)

		if g, e := hit, test.hit; g != e {
			t.Fatal(i, g, e)
		}

		j := 0
		for {
			k, v, ok := en.prev(t)
			if !ok {
				if err := en.Err(); err != nil {
					t.Fatal(i, err)
				}

				break
			}

			if j >= len(keys) {
				t.Fatal(i, j, len(keys), k, v)
			}

			if g, e := k, keys[j]; g != e {
				t.Fatal(i, j, g, e)
			}

			if g, e := v, 10*keys[j]; g != e {
				t.Fatal(i, g, e)
			}

			j++

		}

		if g, e := j, len(keys); g != e {
			t.Fatal(i, j, g, e)
		}
	}
}

func TestBTreeEnumeratorPrev(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeEnumeratorPrev(t, v.f) }) {
			break
		}
	}
}

func testBTreeSeekFirst(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	for i := 0; i < 10; i++ {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(2, 4, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			for j := 0; j < i; j++ {
				bt.set(t, 10*j, 100*j)
			}

			switch {
			case i == 0:
				en := bt.seekFirst(t)
				_, _, ok := en.prev(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}

				en = bt.seekFirst(t)
				_, _, ok = en.next(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}
			default:
				en := bt.seekFirst(t)
				k, v, ok := en.prev(t)
				if g, e := ok, true; g != e {
					t.Fatal(i, g, e)
				}

				if g, e := k, 0; g != e {
					t.Fatal(i, g, e)
				}

				if g, e := v, 0; g != e {
					t.Fatal(i, g, e)
				}

				_, _, ok = en.prev(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}

				en = bt.seekFirst(t)
				for j := 0; j < i; j++ {
					k, v, ok := en.next(t)
					if g, e := ok, true; g != e {
						t.Fatal(i, g, e)
					}

					if g, e := k, 10*j; g != e {
						t.Fatal(i, g, e)
					}

					if g, e := v, 100*j; g != e {
						t.Fatal(i, g, e)
					}
				}
				_, _, ok = en.next(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}
			}

		}()
	}
}

func TestBTreeSeekFirst(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSeekFirst(t, v.f) }) {
			break
		}
	}
}

func testBTreeSeekLast(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	for i := 0; i < 10; i++ {
		func() {
			db, f := tmpDB(t, ts)

			defer f()

			bt, err := db.NewBTree(2, 4, 8, 8)
			if err != nil {
				t.Fatal(err)
			}

			defer bt.remove(t)

			for j := 0; j < i; j++ {
				bt.set(t, 10*j, 100*j)
			}

			switch {
			case i == 0:
				en := bt.seekLast(t)
				_, _, ok := en.prev(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}

				en = bt.seekLast(t)
				_, _, ok = en.next(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}
			default:
				en := bt.seekLast(t)
				k, v, ok := en.next(t)

				if g, e := ok, true; g != e {
					t.Fatal(i, g, e)
				}

				if g, e := k, 10*(i-1); g != e {
					t.Fatal(i, g, e)
				}

				if g, e := v, 100*(i-1); g != e {
					t.Fatal(i, g, e)
				}

				_, _, ok = en.next(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}

				en = bt.seekLast(t)
				for j := i - 1; j >= 0; j-- {
					k, v, ok := en.prev(t)
					if g, e := ok, true; g != e {
						t.Fatal(i, g, e)
					}

					if g, e := k, 10*j; g != e {
						t.Fatal(i, g, e)
					}

					if g, e := v, 100*j; g != e {
						t.Fatal(i, g, e)
					}
				}
				_, _, ok = en.prev(t)
				if g, e := ok, false; g != e {
					t.Fatal(i, g, e)
				}
			}

		}()
	}
}

func TestBTreeSeekLast(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSeekLast(t, v.f) }) {
			break
		}
	}
}

func testBTreeSeek(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	const N = 1 << 10
	for i := 0; i < N; i++ {
		k := 2*i + 1
		bt.set(t, k, 0)
	}
	for i := 0; i < N; i++ {
		k := 2 * i
		e, ok := bt.seek(t, k)
		if ok {
			t.Fatal(i, k)
		}

		for j := i; j < N; j++ {
			k2, _, ok := e.next(t)
			if !ok {
				t.Fatal(i, k, err)
			}

			if g, e := k2, 2*j+1; g != e {
				t.Fatal(i, j, g, e)
			}
		}

		if _, _, ok = e.next(t); ok {
			t.Fatal(i)
		}
	}
}

func TestBTreeSeek(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeSeek(t, v.f) }) {
			break
		}
	}
}

// https://gitlab.com/cznic/b/pull/4
func testBTreeBPR4(t *testing.T, ts func(t testing.TB) (file.File, func())) {
	db, f := tmpDB(t, ts)

	defer f()

	bt, err := db.NewBTree(16, 16, 8, 8)
	if err != nil {
		t.Fatal(err)
	}

	defer bt.remove(t)

	kd := bt.kd
	for i := 0; i < 2*kd+1; i++ {
		k := 1000 * i
		bt.set(t, k, 0)
	}
	bt.delete(t, 1000*kd)
	for i := 0; i < kd; i++ {
		bt.set(t, 1000*(kd+1)-1-i, 0)
	}
	k := 1000*(kd+1) - 1 - kd
	bt.set(t, k, 0)
	if _, ok := bt.get(t, k); !ok {
		t.Fatalf("key lost: %v", k)
	}
}

func TestBTreeBPR4(t *testing.T) {
	for _, v := range ctors {
		if !t.Run(v.s, func(t *testing.T) { testBTreeBPR4(t, v.f) }) {
			break
		}
	}
}

func benchmarkBTreeSetSeq(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			db, f := tmpDB(b, ts)

			defer f()

			bt, err := db.NewBTree(nd, nx, 4, 0)
			if err != nil {
				b.Fatal(err)
			}

			defer bt.bremove(b)

			b.StartTimer()
			for j := 0; j < n; j++ {
				bt.bset(b, j)
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkBTreeSetSeq(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeSetSeq(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeGetSeq(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	for i := 0; i < n; i++ {
		bt.bset(b, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			bt.bget(b, j)
		}
	}
	b.StopTimer()
}

func BenchmarkBTreeGetSeq(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeGetSeq(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeSetRnd(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	rng := rng()
	a := make([]int, n)
	for i := range a {
		a[i] = rng.Next()
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			db, f := tmpDB(b, ts)

			defer f()

			bt, err := db.NewBTree(nd, nx, 4, 0)
			if err != nil {
				b.Fatal(err)
			}

			defer bt.bremove(b)

			b.StartTimer()
			for _, v := range a {
				bt.bset(b, v)
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkBTreeSetRnd(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeSetRnd(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeGetRnd(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	rng := rng()
	a := make([]int, n)
	for i := range a {
		a[i] = rng.Next()
	}
	for _, v := range a {
		bt.bset(b, v)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range a {
			bt.bget(b, v)
		}
	}
	b.StopTimer()
}

func BenchmarkBTreeGetRnd(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeGetRnd(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeDeleteSeq(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			db, f := tmpDB(b, ts)

			defer f()

			bt, err := db.NewBTree(nd, nx, 4, 0)
			if err != nil {
				b.Fatal(err)
			}

			defer bt.bremove(b)

			for j := 0; j < n; j++ {
				bt.bset(b, j)
			}
			b.StartTimer()
			for j := 0; j < n; j++ {
				bt.bdelete(b, j)
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkBTreeDeleteSeq(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeDeleteSeq(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeDeleteRnd(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	rng := rng()
	a := make([]int, n)
	for i := range a {
		a[i] = rng.Next()
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			db, f := tmpDB(b, ts)

			defer f()

			bt, err := db.NewBTree(nd, nx, 4, 0)
			if err != nil {
				b.Fatal(err)
			}

			defer bt.bremove(b)

			for _, v := range a {
				bt.bset(b, v)
			}
			b.StartTimer()
			for _, v := range a {
				bt.bdelete(b, v)
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkBTreeDeleteRnd(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeDeleteRnd(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeSeekSeq(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	for i := 0; i < n; i++ {
		bt.bset(b, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			bt.bseek(b, j)
		}
	}
	b.StopTimer()
}

func BenchmarkBTreeSeekSeq(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeSeekSeq(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeSeekRnd(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	rng := rng()
	a := make([]int, n)
	for i := range a {
		a[i] = rng.Next()
	}

	for _, v := range a {
		bt.bset(b, v)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range a {
			bt.bseek(b, v)
		}
	}
	b.StopTimer()
}

func BenchmarkBTreeSeekRnd(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeSeekRnd(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreeNext(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	for i := 0; i < n; i++ {
		bt.bset(b, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en := bt.seekFirst(b)
		for j := 0; j < n; j++ {
			en.Next()
		}
	}
	b.StopTimer()
}

func BenchmarkBTreeNext(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreeNext(b, v.f, btND, btNX, n) })
		}
	}
}

func benchmarkBTreePrev(b *testing.B, ts func(t testing.TB) (file.File, func()), nd, nx, n int) {
	db, f := tmpDB(b, ts)

	defer f()

	bt, err := db.NewBTree(nd, nx, 4, 0)
	if err != nil {
		b.Fatal(err)
	}

	defer bt.bremove(b)

	for i := 0; i < n; i++ {
		bt.bset(b, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en := bt.seekLast(b)
		for j := 0; j < n; j++ {
			en.Prev()
		}
	}
	b.StopTimer()
}

func BenchmarkBTreePrev(b *testing.B) {
	for _, v := range ctors {
		var n int
		for _, e := range []int{2, 3, 4, 5} {
			n = 1
			for i := 0; i < e; i++ {
				n *= 10
			}
			b.Run(fmt.Sprintf("%s1e%d", v.s, e), func(b *testing.B) { benchmarkBTreePrev(b, v.f, btND, btNX, n) })
		}
	}
}
