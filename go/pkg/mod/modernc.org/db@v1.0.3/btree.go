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
	"fmt"
	"math"

	"modernc.org/internal/buffer"
	"modernc.org/mathutil"
)

const (
	btND       = 256
	btNX       = 32
	maxCopyBuf = 64 << 20
)

var (
	_ btPage = btDPage(0)
	_ btPage = btXPage(0)
)

const (
	btTagDataPage = iota
	btTagIndexPage
)

const (
	oBTRoot  = 8 * iota // int64
	oBTLen              // int64
	oBTFirst            // int64
	oBTLast             // int64
	oBTKD               // int64
	oBTKX               // int64
	oBTSzKey            // int64
	oBTSzVal            // int64

	szBTree
)

const (
	oBTDPageTag   = 8 * iota // int32
	oBTDPageLen              // int32
	oBTDPagePrev             // int64
	oBTDPageNext             // int64
	oBTDPageItems            // [2*kd+1]struct{[szKey]byte, [szVal]byte}
)

const (
	oBTXPageTag   = 8 * iota // int32
	oBTXPageLen              // int32
	oBTXPageItems            // [2*kx+2]struct{int64,int64}
)

type btDPage int64

type btXPage int64

type btPage interface {
	//TODO off() int64
}

// BTree is a B+tree.
type BTree struct {
	*DB
	Off   int64 // Location in the database. R/O
	SzKey int64 // The szKey argument of NewBTree. R/O
	SzVal int64 // The szVal argument of NewBTree. R/O
	kd    int
	kx    int
}

// NewBTree allocates and returns a new, empty BTree or an error, if any.  The
// nd and nx arguments are the desired number of items in a data or index page.
// Passing zero will use default values. The szKey and szVal arguments are the
// sizes of the BTree keys and values.
func (db *DB) NewBTree(nd, nx int, szKey, szVal int64) (*BTree, error) {
	if nd < 0 || nd > (math.MaxInt32-1)/2 ||
		nx < 0 || nx > (math.MaxInt32-2)/2 ||
		szKey < 0 || szVal < 0 {
		panic(fmt.Errorf("%T.NewBTree: invalid argument", db))
	}

	if nd == 0 {
		nd = btND
	}
	kd := mathutil.Max(nd/2, 1)
	if nx == 0 {
		nx = btNX
	}
	kx := mathutil.Max(nx/2, 2)
	off, err := db.Calloc(szBTree)
	if err != nil {
		return nil, err
	}

	if err := db.w8(off+oBTKD, int64(kd)); err != nil {
		return nil, err
	}

	if err := db.w8(off+oBTKX, int64(kx)); err != nil {
		return nil, err
	}

	if err := db.w8(off+oBTSzKey, szKey); err != nil {
		return nil, err
	}

	if err := db.w8(off+oBTSzVal, szVal); err != nil {
		return nil, err
	}

	return &BTree{DB: db, Off: off, SzKey: szKey, SzVal: szVal, kd: kd, kx: kx}, nil
}

// OpenBTree returns an existing BTree found at offset off or an error, if any.
// The off argument must have been acquired from NewBTree.
func (db *DB) OpenBTree(off int64) (*BTree, error) {
	n, err := db.r8(off + oBTKD)
	if err != nil {
		return nil, err
	}

	if n < 0 || n > (mathutil.MaxInt-1)/2 {
		return nil, fmt.Errorf("%T.OpenBTree: corrupted database", db)
	}

	kd := int(n)
	if n, err = db.r8(off + oBTKX); err != nil {
		return nil, err
	}

	if n < 0 || n > (mathutil.MaxInt-2)/2 {
		return nil, fmt.Errorf("%T.OpenBTree: corrupted database", db)
	}

	kx := int(n)
	szKey, err := db.r8(off + oBTSzKey)
	if err != nil {
		return nil, err
	}

	szVal, err := db.r8(off + oBTSzVal)
	if err != nil {
		return nil, err
	}

	return &BTree{DB: db, Off: off, kd: kd, kx: kx, SzKey: szKey, SzVal: szVal}, nil
}

func (t *BTree) first() (int64, error)          { return t.r8(t.Off + oBTFirst) }
func (t *BTree) item(x btXPage, i int) int64    { return int64(x) + oBTXPageItems + int64(i)*16 }
func (t *BTree) last() (int64, error)           { return t.r8(t.Off + oBTLast) }
func (t *BTree) len(d btDPage) (int, error)     { return t.r4(int64(d) + oBTDPageLen) }
func (t *BTree) lenX(x btXPage) (int, error)    { return t.r4(int64(x) + oBTXPageLen) }
func (t *BTree) openDPage(off int64) btDPage    { return btDPage(off) } //TODO-
func (t *BTree) openXPage(off int64) btXPage    { return btXPage(off) } //TODO-
func (t *BTree) root() (int64, error)           { return t.r8(t.Off + oBTRoot) }
func (t *BTree) setFirst(d btDPage) error       { return t.w8(t.Off+oBTFirst, int64(d)) }
func (t *BTree) setLast(d btDPage) error        { return t.w8(t.Off+oBTLast, int64(d)) }
func (t *BTree) setLen(n int64) error           { return t.w8(t.Off+oBTLen, n) }
func (t *BTree) setLenD(d btDPage, n int) error { return t.w4(int64(d)+oBTDPageLen, n) }
func (t *BTree) setLenX(x btXPage, n int) error { return t.w4(int64(x)+oBTXPageLen, n) }
func (t *BTree) setNext(d, next btDPage) error  { return t.w8(int64(d)+oBTDPageNext, int64(next)) }
func (t *BTree) setPrev(d, prev btDPage) error  { return t.w8(int64(d)+oBTDPagePrev, int64(prev)) }
func (t *BTree) setRoot(n int64) error          { return t.w8(t.Off+oBTRoot, n) }
func (t *BTree) setTag(d btDPage) error         { return t.w4(int64(d)+oBTDPageTag, btTagDataPage) }
func (t *BTree) setTagX(x btXPage) error        { return t.w4(int64(x)+oBTXPageTag, btTagIndexPage) }
func (t *BTree) val(d btDPage, i int) int64     { return t.key(d, i) + t.SzVal }

func (t *BTree) cat(p btXPage, q, r btDPage, pc, qc, rc, pi int, free func(int64, int64) error) error {
	if err := t.mvL(q, r, qc, rc, rc); err != nil {
		return err
	}

	rn, err := t.next(r)
	if err != nil {
		return err
	}

	if rn != 0 {
		if err := t.setPrev(rn, q); err != nil {
			return err
		}
	} else if err := t.setLast(q); err != nil {
		return err
	}

	if err := t.setNext(q, rn); err != nil {
		return err
	}

	if err := t.Free(int64(r)); err != nil {
		return err
	}

	if pc > 1 {
		if err := t.extractX(p, pc, pi); err != nil {
			return err
		}

		return t.setChild(p, pi, int64(q))
	}

	root, err := t.root()
	if err != nil {
		return err
	}

	if err := t.Free(root); err != nil {
		return err
	}

	return t.setRoot(int64(q))
}

func (t *BTree) catX(p, q, r btXPage, pc, qc, rc, pi int) error {
	k, err := t.keyX(p, pi)
	if err != nil {
		return err
	}

	if err := t.setKey(q, qc, k); err != nil {
		return err
	}

	if err := t.copyX(q, r, qc+1, 0, rc); err != nil {
		return err
	}

	qc += rc + 1
	if err := t.setLenX(q, qc); err != nil {
		return err
	}

	ch, err := t.child(r, rc)
	if err != nil {
		return err
	}

	if err := t.setChild(q, qc, ch); err != nil {
		return err
	}

	if err := t.Free(int64(r)); err != nil {
		return err
	}

	if pc > 1 {
		pc--
		if err := t.setLenX(p, pc); err != nil {
			return err
		}

		if pi < pc {
			k, err := t.keyX(p, pi+1)
			if err != nil {
				return err
			}

			if err := t.setKey(p, pi, k); err != nil {
				return err
			}

			if err := t.copyX(p, p, pi+1, pi+2, pc-pi-1); err != nil {
				return err
			}

			ch, err := t.child(p, pc+1)
			if err != nil {
				return err
			}

			if err := t.setChild(p, pc, ch); err != nil {
				return err
			}
		}
		return nil
	}

	proot, err := t.root()
	if err != nil {
		return err
	}

	if err := t.Free(proot); err != nil {
		return err
	}

	return t.setRoot(int64(q))
}

func (t *BTree) child(x btXPage, i int) (y int64, yy error) {
	return t.r8(int64(x) + oBTXPageItems + int64(i)*16)
}

func (t *BTree) clr(off int64, free func(int64, int64) error) error {
	if off == 0 {
		return nil
	}

	p, err := t.openPage(off)
	if err != nil {
		return err
	}

	switch x := p.(type) {
	case btDPage:
		dc, err := t.len(x)
		if err != nil {
			return err
		}

		return t.clrD(x, dc, free)
	case btXPage:
		xc, err := t.lenX(x)
		if err != nil {
			return err
		}

		return t.clrX(x, xc, free)
	}
	panic("internal error")
}

func (t *BTree) clrD(d btDPage, dc int, free func(int64, int64) error) error {
	if free != nil {
		o := t.SzKey + t.SzVal
		koff := t.key(d, 0)
		voff := t.val(d, 0)
		for i := 0; i < dc; i++ {
			if err := free(koff, voff); err != nil {
				return err
			}

			koff += o
			voff += o
		}
	}
	return t.Free(int64(d))
}

func (t *BTree) clrX(x btXPage, xc int, free func(int64, int64) error) error {
	for i := 0; i <= xc; i++ {
		off, err := t.child(x, i)
		if err != nil {
			return err
		}

		if off == 0 {
			break
		}

		ch, err := t.openPage(off)
		if err != nil {
			return err
		}

		switch x := ch.(type) {
		case btDPage:
			dc, err := t.len(x)
			if err != nil {
				return err
			}

			if err := t.clrD(x, dc, free); err != nil {
				return err
			}
		case btXPage:
			xc, err := t.lenX(x)
			if err != nil {
				return err
			}

			if err := t.clrX(x, xc, free); err != nil {
				return err
			}
		}
	}
	return t.Free(int64(x))
}

func (t *BTree) copy(d, s btDPage, di, si, n int) error {
	if n <= 0 {
		return nil
	}

	dst := t.key(d, di)
	src := t.key(s, si)
	var rq int
	var p *[]byte
	var b []byte
	for rem := (t.SzKey + t.SzVal) * int64(n); rem != 0; rem -= int64(rq) {
		if rem <= maxCopyBuf {
			rq = int(rem)
		} else {
			rq = maxCopyBuf
		}

		if p == nil {
			p = buffer.Get(rq)
			b = *p
		}
		if nr, err := t.ReadAt(b[:rq], src); nr != rq {
			if err == nil {
				panic("internal error")
			}

			buffer.Put(p)
			return err
		}

		if nw, err := t.WriteAt(b[:rq], dst); nw != rq {
			if err == nil {
				panic("internal error")
			}

			buffer.Put(p)
			return err
		}
		src += int64(rq)
		dst += int64(rq)
	}
	buffer.Put(p)
	return nil
}

func (t *BTree) copyX(d, s btXPage, di, si, n int) error {
	nb := 16 * n
	p := buffer.Get(nb)
	if nr, err := t.ReadAt(*p, t.item(s, si)); nr != nb {
		if err == nil {
			panic("internal error")
		}

		buffer.Put(p)
		return err
	}

	if nw, err := t.WriteAt(*p, t.item(d, di)); nw != nb {
		if err == nil {
			panic("internal error")
		}

		buffer.Put(p)
		return err
	}

	buffer.Put(p)
	return nil
}

func (t *BTree) extract(d btDPage, dc, i int, free func(int64, int64) error) error {
	if free != nil {
		if err := free(t.key(d, i), t.val(d, i)); err != nil {
			return err
		}
	}

	dc--
	if err := t.setLenD(d, dc); err != nil {
		return err
	}

	if i < dc {
		if err := t.copy(d, d, i, i+1, dc-i); err != nil {
			return err
		}
	}
	tc, err := t.Len()
	if err != nil {
		return err
	}

	tc--
	return t.setLen(tc)
}

func (t *BTree) extractX(x btXPage, xc, i int) error {
	xc--
	err := t.setLenX(x, xc)
	if err != nil {
		return err
	}

	if i < xc {
		if err := t.copyX(x, x, i, i+1, xc-i); err != nil {
			return err
		}

		ch, err := t.child(x, xc+1)
		if err != nil {
			return err
		}

		if err := t.setChild(x, xc, ch); err != nil {
			return err
		}
	}

	return nil
}

func (t *BTree) find(d btDPage, dc int, cmp func(off int64) (int, error)) (int, bool, error) {
	var l int
	dc--
	for l <= dc {
		m := (l + dc) >> 1
		switch c, err := cmp(t.key(d, m)); {
		case err != nil:
			return 0, false, err
		case c > 0:
			l = m + 1
		case c == 0:
			return m, true, nil
		default:
			dc = m - 1
		}
	}
	return l, false, nil
}

func (t *BTree) findX(x btXPage, xc int, cmp func(off int64) (int, error)) (int, bool, error) {
	var l int
	xc--
	for l <= xc {
		m := (l + xc) >> 1
		k, err := t.keyX(x, m)
		if err != nil {
			return 0, false, err
		}

		switch c, err := cmp(k); {
		case err != nil:
			return 0, false, err
		case c > 0:
			l = m + 1
		case c == 0:
			return m, true, nil
		default:
			xc = m - 1
		}
	}
	return l, false, nil
}

func (t *BTree) insert(d btDPage, dc, i int) error {
	if i < dc {
		if err := t.copy(d, d, i+1, i, dc-i); err != nil {
			return err
		}
	}

	if err := t.setLenD(d, dc+1); err != nil {
		return err
	}

	n, err := t.Len()
	if err != nil {
		return err
	}

	return t.setLen(n + 1)
}

func (t *BTree) insertX(x btXPage, xc, i int, k, ch int64) error {
	if i < xc {
		ch, err := t.child(x, xc)
		if err != nil {
			return err
		}

		if err := t.setChild(x, xc+1, ch); err != nil {
			return err
		}

		if err := t.copyX(x, x, i+2, i+1, xc-i-1); err != nil {
			return err
		}

		k, err := t.keyX(x, i)
		if err != nil {
			return err
		}

		if err := t.setKey(x, i+1, k); err != nil {
			return err
		}
	}

	if err := t.setLenX(x, xc+1); err != nil {
		return err
	}

	if err := t.setKey(x, i, k); err != nil {
		return err
	}

	return t.setChild(x, i+1, ch)
}

func (t *BTree) key(d btDPage, i int) int64 {
	return int64(d) + oBTDPageItems + int64(i)*(t.SzKey+t.SzVal)
}

func (t *BTree) keyX(x btXPage, i int) (int64, error) {
	return t.r8(int64(x) + oBTXPageItems + int64(i)*16 + 8)
}

func (t *BTree) mvL(d, r btDPage, dc, rc, c int) error {
	if err := t.copy(d, r, dc, 0, c); err != nil {
		return err
	}

	if err := t.copy(r, r, 0, c, rc-c); err != nil {
		return err
	}

	if err := t.setLenD(d, dc+c); err != nil {
		return err
	}

	return t.setLenD(r, rc-c)
}

func (t *BTree) mvR(d, r btDPage, dc, rc, c int) error {
	if err := t.copy(r, r, c, 0, rc); err != nil {
		return err
	}

	if err := t.copy(r, d, 0, dc-c, c); err != nil {
		return err
	}
	if err := t.setLenD(r, rc+c); err != nil {
		return err
	}

	return t.setLenD(d, dc-c)
}

func (t *BTree) newBTDPage() (btDPage, error) {
	rq := oBTDPageItems + (2*int64(t.kd)+1)*(t.SzKey+t.SzVal)
	off, err := t.Alloc(rq)
	if err != nil {
		return 0, err
	}

	r := btDPage(off)
	if err := t.setTag(r); err != nil {
		return 0, err
	}

	if err := t.setLenD(r, 0); err != nil {
		return 0, err
	}

	if err := t.setNext(r, 0); err != nil {
		return 0, err
	}

	if err := t.setPrev(r, 0); err != nil {
		return 0, err
	}

	return r, nil
}

func (t *BTree) newBTXPage(ch0 int64) (r btXPage, err error) {
	off, err := t.Alloc(oBTXPageItems + 16*(2*int64(t.kx)+2))
	if err != nil {
		return 0, err
	}

	r = btXPage(off)
	if err := t.setTagX(r); err != nil {
		return 0, err
	}

	if ch0 != 0 {
		if err := t.setChild(r, 0, ch0); err != nil {
			return 0, err
		}
	}

	return r, nil
}

func (t *BTree) newEnumerator(d btDPage, dc, i int, hit bool) *BTreeCursor {
	return &BTreeCursor{
		btDPage: d,
		c:       dc,
		hit:     hit,
		i:       i,
		t:       t,
	}
}

func (t *BTree) next(d btDPage) (btDPage, error) {
	off, err := t.r8(int64(d) + oBTDPageNext)
	return btDPage(off), err
}

func (t *BTree) openPage(off int64) (btPage, error) {
	switch tag, err := t.r4(off); {
	case err != nil:
		return nil, err
	case tag == btTagDataPage:
		return t.openDPage(off), nil
	case tag == btTagIndexPage:
		return t.openXPage(off), nil
	default:
		return nil, fmt.Errorf("%T.openPage: corrupted database", t)
	}
}

func (t *BTree) overflow(d btDPage, p btXPage, dc, pc, pi, i int) (btDPage, int, error) {
	l, r, err := t.siblings(p, pc, pi)
	if err != nil {
		return 0, 0, err
	}

	if l != 0 {
		c, err := t.len(l)
		if err != nil {
			return 0, 0, err
		}

		if c < 2*t.kd && i != 0 {
			if err := t.mvL(l, d, c, dc, 1); err != nil {
				return 0, 0, err
			}

			if err := t.insert(d, dc-1, i-1); err != nil {
				return 0, 0, err
			}

			return d, i - 1, t.setKey(p, pi-1, t.key(d, 0))
		}
	}

	if r != 0 {
		rc, err := t.len(r)
		if err != nil {
			return 0, 0, err
		}

		if rc < 2*t.kd {
			if i < 2*t.kd {
				if err := t.mvR(d, r, dc, rc, 1); err != nil {
					return 0, 0, err
				}

				rc++
				if err := t.insert(d, dc-1, i); err != nil {
					return 0, 0, err
				}

				return 0, 0, t.setKey(p, pi, t.key(r, 0))
			}

			if err := t.insert(r, rc, 0); err != nil {
				return 0, 0, err
			}

			if err := t.setKey(p, pi, t.key(r, 0)); err != nil {
				return 0, 0, err
			}

			return r, 0, nil
		}
	}

	return t.split(d, p, pi, i)
}

func (t *BTree) prev(d btDPage) (btDPage, error) {
	off, err := t.r8(int64(d) + oBTDPagePrev)
	return btDPage(off), err
}

func (t *BTree) setChild(x btXPage, i int, c int64) error {
	return t.w8(int64(x)+oBTXPageItems+int64(i)*16, c)
}

func (t *BTree) setKey(x btXPage, i int, k int64) error {
	return t.w8(int64(x)+oBTXPageItems+int64(i)*16+8, k)
}

func (t *BTree) siblings(x btXPage, xc, i int) (l, r btDPage, err error) {
	if x == 0 {
		return 0, 0, nil
	}

	if i >= 0 {
		if i > 0 {
			ch, err := t.child(x, i-1)
			if err != nil {
				return l, r, err
			}

			l = t.openDPage(ch)
		}
		if i < xc {
			ch, err := t.child(x, i+1)
			if err != nil {
				return l, r, err
			}

			r = t.openDPage(ch)
		}
	}
	return l, r, nil
}

func (t *BTree) split(d btDPage, p btXPage, pi, i int) (q btDPage, j int, err error) {
	var r btDPage
	if r, err = t.newBTDPage(); err != nil {
		return q, j, err
	}

	n, err := t.next(d)
	if err != nil {
		return q, j, err
	}

	if n != 0 {
		if err := t.setNext(r, n); err != nil {
			return q, j, err
		}

		if err = t.setPrev(n, r); err != nil {
			return q, j, err
		}
	} else {
		if err := t.setLast(r); err != nil {
			return q, j, err
		}
	}

	if err := t.setNext(d, r); err != nil {
		return q, j, err
	}

	if err := t.setPrev(r, d); err != nil {
		return q, j, err
	}

	if err := t.copy(r, d, 0, t.kd, 2*t.kd-t.kd); err != nil {
		return q, j, err
	}

	if err := t.setLenD(d, t.kd); err != nil {
		return q, j, err
	}

	if err := t.setLenD(r, t.kd); err != nil {
		return q, j, err
	}

	var done bool
	if i > t.kd {
		done = true
		q = r
		j = i - t.kd
		if err := t.insert(q, t.kd, j); err != nil {
			return 0, 0, err
		}
	}

	if pi >= 0 {
		pc, err := t.lenX(p)
		if err != nil {
			return 0, 0, err
		}

		if err := t.insertX(p, pc, pi, t.key(r, 0), int64(r)); err != nil {
			return 0, 0, err
		}
	} else {
		x, err := t.newBTXPage(int64(d))
		if err != nil {
			return 0, 0, err
		}

		if err := t.insertX(x, 0, 0, t.key(r, 0), int64(r)); err != nil {
			return 0, 0, err
		}

		if err := t.setRoot(int64(x)); err != nil {
			return 0, 0, err
		}
	}
	if done {
		return q, j, nil
	}

	return 0, 0, t.insert(d, t.kd, i)
}

func (t *BTree) splitX(p, q btXPage, pc, qc, pi, i int) (btXPage, int, error) {
	r, err := t.newBTXPage(0)
	if err != nil {
		return 0, 0, err
	}

	if err := t.copyX(r, q, 0, t.kx+1, qc-t.kx); err != nil {
		return 0, 0, err
	}

	if err := t.setLenX(q, t.kx); err != nil {
		return 0, 0, err
	}

	qc = t.kx
	if err := t.setLenX(r, t.kx); err != nil {
		return 0, 0, err
	}

	if pi >= 0 {
		k, err := t.keyX(q, t.kx)
		if err != nil {
			return 0, 0, err
		}

		if err := t.insertX(p, pc, pi, k, int64(r)); err != nil {
			return 0, 0, err
		}
	} else {
		nx, err := t.newBTXPage(int64(q))
		if err != nil {
			return 0, 0, err
		}

		k, err := t.keyX(q, t.kx)
		if err != nil {
			return 0, 0, err
		}

		if err := t.insertX(nx, 0, 0, k, int64(r)); err != nil {
			return 0, 0, err
		}

		if err := t.setRoot(int64(nx)); err != nil {
			return 0, 0, err
		}
	}

	if i > t.kx {
		q = r
		i -= t.kx + 1
	}

	return q, i, nil
}

func (t *BTree) underflow(d btDPage, p btXPage, dc, pi int, free func(int64, int64) error) error {
	pc, err := t.lenX(p)
	if err != nil {
		return err
	}

	l, r, err := t.siblings(p, pc, pi)
	if err != nil {
		return err
	}

	var lc, rc int
	if l != 0 {
		if lc, err = t.len(l); err != nil {
			return err
		}

		if lc+dc >= 2*t.kd {
			if err := t.mvR(l, d, lc, dc, 1); err != nil {
				return err
			}

			return t.setKey(p, pi-1, t.key(d, 0))
		}
	}

	if r != 0 {
		if rc, err = t.len(r); err != nil {
			return err
		}

		if dc+rc >= 2*t.kd {
			if err := t.mvL(d, r, dc, rc, 1); err != nil {
				return err
			}

			return t.setKey(p, pi, t.key(r, 0))
		}
	}

	if l != 0 {
		return t.cat(p, l, d, pc, lc, dc, pi-1, free)
	}

	return t.cat(p, d, r, pc, dc, rc, pi, free)
}

func (t *BTree) underflowX(p, q btXPage, pc, qc, pi, i int) (btXPage, int, error) {
	var l, r btXPage
	var lc, rc int
	var err error
	if pi >= 0 {
		if pi > 0 {
			ch, err := t.child(p, pi-1)
			if err != nil {
				return 0, 0, err
			}

			l = t.openXPage(ch)
		}
		if pi < pc {
			ch, err := t.child(p, pi+1)
			if err != nil {
				return 0, 0, err
			}

			r = t.openXPage(ch)
		}
	}

	if l != 0 {
		if lc, err = t.lenX(l); err != nil {
			return 0, 0, err
		}

		if lc > t.kx {
			ch, err := t.child(q, qc)
			if err != nil {
				return 0, 0, err
			}

			if t.setChild(q, qc+1, ch); err != nil {
				return 0, 0, err
			}

			if err := t.copyX(q, q, 1, 0, qc); err != nil {
				return 0, 0, err
			}

			if ch, err = t.child(l, lc); err != nil {
				return 0, 0, err
			}

			if t.setChild(q, 0, ch); err != nil {
				return 0, 0, err
			}

			k, err := t.keyX(p, pi-1)
			if err != nil {
				return 0, 0, err
			}

			if err := t.setKey(q, 0, k); err != nil {
				return 0, 0, err
			}

			qc++
			if err := t.setLenX(q, qc); err != nil {
				return 0, 0, err
			}

			i++
			lc--
			if err := t.setLenX(l, lc); err != nil {
				return 0, 0, err
			}

			if k, err = t.keyX(l, lc); err != nil {
				return 0, 0, err
			}

			if err := t.setKey(p, pi-1, k); err != nil {
				return 0, 0, err
			}

			return q, i, nil
		}
	}

	if r != 0 {
		if rc, err = t.lenX(r); err != nil {
			return 0, 0, err
		}

		if rc > t.kx {
			k, err := t.keyX(p, pi)
			if err != nil {
				return 0, 0, err
			}

			if err := t.setKey(q, qc, k); err != nil {
				return 0, 0, err
			}

			qc++
			if err := t.setLenX(q, qc); err != nil {
				return 0, 0, err
			}

			ch, err := t.child(r, 0)
			if err != nil {
				return 0, 0, err
			}

			if t.setChild(q, qc, ch); err != nil {
				return 0, 0, err
			}

			if k, err = t.keyX(r, 0); err != nil {
				return 0, 0, err
			}

			if err := t.setKey(p, pi, k); err != nil {
				return 0, 0, err
			}

			if err := t.copyX(r, r, 0, 1, rc-1); err != nil {
				return 0, 0, err
			}

			rc--
			if err := t.setLenX(r, rc); err != nil {
				return 0, 0, err
			}

			if ch, err = t.child(r, rc+1); err != nil {
				return 0, 0, err
			}

			if err := t.setChild(r, rc, ch); err != nil {
				return 0, 0, err
			}

			return q, i, nil
		}
	}

	if l != 0 {
		i += lc + 1
		if err := t.catX(p, l, q, pc, lc, qc, pi-1); err != nil {
			return 0, 0, err
		}

		return l, i, nil
	}

	if err := t.catX(p, q, r, pc, qc, rc, pi); err != nil {
		return 0, 0, err
	}

	return q, i, nil
}

// Len returns the number of items in t or an error, if any.
func (t *BTree) Len() (int64, error) { return t.r8(t.Off + oBTLen) }

// Clear deletes all items of t.
//
// The free function may be nil, otherwise it's called with the offsets of the
// key and value of an item that is being deleted from the tree. Both koff and
// voff may be zero when appropriate.
func (t *BTree) Clear(free func(koff, voff int64) error) error {
	r, err := t.root()
	if err != nil {
		return err
	}

	if r == 0 {
		return nil
	}

	p, err := t.openPage(r)
	if err != nil {
		return err
	}

	switch x := p.(type) {
	case btDPage:
		dc, err := t.len(x)
		if err != nil {
			return err
		}

		if err := t.clrD(x, dc, free); err != nil {
			return err
		}
	case btXPage:
		xc, err := t.lenX(x)
		if err != nil {
			return err
		}

		if err := t.clrX(x, xc, free); err != nil {
			return err
		}
	}

	if err := t.setLen(0); err != nil {
		return err
	}

	if err := t.setFirst(0); err != nil {
		return err
	}

	if err := t.setLast(0); err != nil {
		return err
	}

	return t.setRoot(0)
}

// Delete removes an item from t and returns a boolean value indicating if the
// item was found.
//
// The item is searched for by calling the cmp function that gets the offset of
// a tree key to compare. It returns a positive value if the desired key
// collates after the tree key, a zero if the keys are equal and a negative
// value if the desired key collates before the tree key.
//
// For discussion of the free function see Clear.
func (t *BTree) Delete(cmp func(koff int64) (int, error), free func(koff, voff int64) error) (bool, error) {
	pi := -1
	var p btXPage
	pc := -1
	r, err := t.root()
	if err != nil {
		return false, err
	}

	if r == 0 {
		return false, nil
	}

	q, err := t.openPage(r)
	if err != nil {
		return false, err
	}
	for {
		switch x := q.(type) {
		case btXPage:
			xc, err := t.lenX(x)
			if err != nil {
				return false, err
			}

			i, ok, err := t.findX(x, xc, cmp)
			if err != nil {
				return false, err
			}

			if ok {
				r, err := t.root()
				if err != nil {
					return false, err
				}

				if xc < t.kx && int64(x) != r {
					if pi >= 0 {
						if pc, err = t.lenX(p); err != nil {
							return false, err
						}
					}
					if x, i, err = t.underflowX(p, x, pc, xc, pi, i); err != nil {
						return false, err
					}
				}
				pi = i + 1
				p = x
				ch, err := t.child(x, pi)
				if err != nil {
					return false, err
				}

				if q, err = t.openPage(ch); err != nil {
					return false, err
				}

				continue
			}

			r, err := t.root()
			if err != nil {
				return false, err
			}

			if xc < t.kx && int64(x) != r {
				if pi >= 0 {
					if pc, err = t.lenX(p); err != nil {
						return false, err
					}
				}
				if x, i, err = t.underflowX(p, x, pc, xc, pi, i); err != nil {
					return false, err
				}
			}
			pi = i
			p = x
			ch, err := t.child(x, i)
			if err != nil {
				return false, err
			}

			if q, err = t.openPage(ch); err != nil {
				return false, err
			}
		case btDPage:
			xc, err := t.len(x)
			if err != nil {
				return false, err
			}

			i, ok, err := t.find(x, xc, cmp)
			if err != nil {
				return false, err
			}

			if ok {
				if err := t.extract(x, xc, i, free); err != nil {
					return false, err
				}

				xc--
				if xc >= t.kd {
					return true, nil
				}

				r, err := t.root()
				if err != nil {
					return false, err
				}

				if int64(x) != r {
					if err := t.underflow(x, p, xc, pi, free); err != nil {
						return false, err
					}
				} else {
					tc, err := t.Len()
					if err != nil {
						return false, err
					}

					if tc == 0 {
						if err := t.Clear(free); err != nil {
							return false, err
						}
					}
				}
				return true, nil
			}

			return false, nil
		}
	}
}

// Get searches for a key in the tree and returns the offset of its associated
// value and a boolean value indicating success.
//
// For discussion of the cmp function see Delete.
func (t *BTree) Get(cmp func(koff int64) (int, error)) (int64, bool, error) {
	r, err := t.root()
	if err != nil {
		return 0, false, err
	}

	if r == 0 {
		return 0, false, nil
	}

	q, err := t.openPage(r)
	if err != nil {
		return 0, false, err
	}

	for {
		switch x := q.(type) {
		case btXPage:
			xc, err := t.lenX(x)
			if err != nil {
				return 0, false, err
			}

			i, ok, err := t.findX(x, xc, cmp)
			if err != nil {
				return 0, false, err
			}

			if ok {
				ch, err := t.child(x, i+1)
				if err != nil {
					return 0, false, err
				}

				if q, err = t.openPage(ch); err != nil {
					return 0, false, err
				}

				continue
			}

			ch, err := t.child(x, i)
			if err != nil {
				return 0, false, err
			}

			if q, err = t.openPage(ch); err != nil {
				return 0, false, err
			}
		case btDPage:
			xc, err := t.len(x)
			if err != nil {
				return 0, false, err
			}

			i, ok, err := t.find(x, xc, cmp)
			if err != nil {
				return 0, false, err
			}

			if ok {
				return t.val(x, i), true, nil
			}

			return 0, false, nil
		}
	}
}

// Remove frees all space used by t.
//
// For discussion of the free function see Clear.
func (t *BTree) Remove(free func(koff, voff int64) error) (err error) {
	r, err := t.root()
	if err != nil {
		return err
	}

	if err := t.clr(r, free); err != nil {
		return err
	}

	if err := t.Free(t.Off); err != nil {
		return err
	}

	t.Off = 0
	return nil
}

// Seek searches the tree for a key collating after the key used by the cmp
// function and a boolean value indicating the desired and found keys are
// equal.
//
// For discussion of the cmp function see Delete.
func (t *BTree) Seek(cmp func(int64) (int, error)) (*BTreeCursor, bool, error) {
	r, err := t.root()
	if err != nil {
		return nil, false, err
	}

	if r == 0 {
		return &BTreeCursor{}, false, nil
	}

	q, err := t.openPage(r)
	if err != nil {
		return nil, false, err
	}

	for {
		switch x := q.(type) {
		case btXPage:
			xc, err := t.lenX(x)
			if err != nil {
				return nil, false, err
			}

			i, ok, err := t.findX(x, xc, cmp)
			if err != nil {
				return nil, false, err
			}

			if ok {
				ch, err := t.child(x, i+1)
				if err != nil {
					return nil, false, err
				}

				if q, err = t.openPage(ch); err != nil {
					return nil, false, err
				}
				continue
			}

			ch, err := t.child(x, i)
			if err != nil {
				return nil, false, err
			}

			if q, err = t.openPage(ch); err != nil {
				return nil, false, err
			}
		case btDPage:
			xc, err := t.len(x)
			if err != nil {
				return nil, false, err
			}

			i, ok, err := t.find(x, xc, cmp)
			if err != nil {
				return nil, false, err
			}

			if ok {
				return t.newEnumerator(x, xc, i, true), true, nil
			}

			return t.newEnumerator(x, xc, i, false), false, nil
		}
	}
}

// SeekFirst returns an Enumerator position on the first item of t or an error,
// if any.
func (t *BTree) SeekFirst() (*BTreeCursor, error) {
	p, err := t.first()
	if err != nil {
		return nil, err
	}

	if p == 0 {
		return &BTreeCursor{}, nil
	}

	d := t.openDPage(p)
	dc, err := t.len(d)
	if err != nil {
		return &BTreeCursor{}, err
	}

	return t.newEnumerator(d, dc, 0, true), nil
}

// SeekLast returns an Enumerator position on the last item of t or an error,
// if any.
func (t *BTree) SeekLast() (*BTreeCursor, error) {
	p, err := t.last()
	if err != nil {
		return nil, err
	}

	if p == 0 {
		return &BTreeCursor{}, nil
	}

	d := t.openDPage(p)
	dc, err := t.len(d)
	if err != nil {
		return &BTreeCursor{}, err
	}

	e := t.newEnumerator(d, dc, 0, true)
	e.i = e.c - 1
	return e, nil
}

// Set adds or overwrites an item in t and returns the offsets if its key and value or an error, if any.
//
// For discussion of the cmp function see Delete.
//
// For discussion of the free function see Clear.
func (t *BTree) Set(cmp func(koff int64) (int, error), free func(koff int64) error) (int64, int64, error) {
	pi := -1
	r, err := t.root()
	if err != nil {
		return 0, 0, err
	}

	if r == 0 {
		z, err := t.newBTDPage()
		if err != nil {
			return 0, 0, err
		}

		if err := t.insert(z, 0, 0); err != nil {
			return 0, 0, err
		}

		if err := t.setRoot(int64(z)); err != nil {
			return 0, 0, err
		}

		if err := t.setFirst(z); err != nil {
			return 0, 0, err
		}

		if err := t.setLast(z); err != nil {
			return 0, 0, err
		}

		return t.key(z, 0), t.val(z, 0), nil
	}

	q, err := t.openPage(r)
	if err != nil {
		return 0, 0, err
	}

	var p btXPage
	pc := -1
	for {
		switch x := q.(type) {
		case btXPage:
			xc, err := t.lenX(x)
			if err != nil {
				return 0, 0, err
			}

			i, ok, err := t.findX(x, xc, cmp)
			if err != nil {
				return 0, 0, err
			}

			if p != 0 {
				if pc, err = t.lenX(p); err != nil {
					return 0, 0, err
				}
			}

			if ok {
				i++
				if xc > 2*t.kx {
					if x, i, err = t.splitX(p, x, pc, xc, pi, i); err != nil {
						return 0, 0, err
					}
				}
				pi = i
				p = x
				ch, err := t.child(x, i)
				if err != nil {
					return 0, 0, err
				}

				if q, err = t.openPage(ch); err != nil {
					return 0, 0, err
				}

				continue
			}

			if xc > 2*t.kx {
				if x, i, err = t.splitX(p, x, pc, xc, pi, i); err != nil {
					return 0, 0, err
				}
			}
			pi = i
			p = x
			ch, err := t.child(x, i)
			if err != nil {
				return 0, 0, err
			}

			if q, err = t.openPage(ch); err != nil {
				return 0, 0, err
			}
		case btDPage:
			xc, err := t.len(x)
			if err != nil {
				return 0, 0, err
			}

			i, ok, err := t.find(x, xc, cmp)
			if err != nil {
				return 0, 0, err
			}

			if ok {
				koff := t.key(x, i)
				voff := t.val(x, i)
				if free != nil {
					if err := free(voff); err != nil {
						return 0, 0, err
					}
				}

				return koff, voff, nil
			}

			switch {
			case xc < 2*t.kd:
				if err := t.insert(x, xc, i); err != nil {
					return 0, 0, err
				}
			default:
				pc, err := t.lenX(p)
				if err != nil {
					return 0, 0, err
				}

				q, j, err := t.overflow(x, p, xc, pc, pi, i)
				if err != nil {
					return 0, 0, err
				}

				if q != 0 {
					x = q
					i = j
				}
			}
			return t.key(x, i), t.val(x, i), nil
		}
	}
}

// BTreeCursor provides enumerating BTree items.
type BTreeCursor struct {
	K int64 // Item key offset. Not valid before calling Next or Prev.
	V int64 // Item value offset. Not valid before calling Next or Prev.
	btDPage
	c        int
	err      error
	hasMoved bool
	hit      bool
	i        int
	t        *BTree
}

// Err returns the error, if any, that was encountered during iteration.
func (e *BTreeCursor) Err() error { return e.err }

// Next moves the cursor to the next item in the tree and sets the K and V
// fields accordingly. It returns true on success, or false if there is no next
// item or an error happened while moving the cursor. Err should be consulted
// to distinguish between the two cases.
//
// Every use of the K/V fields, even the first one, must be preceded by a call
// to Next or Prev.
func (e *BTreeCursor) Next() bool {
	if e.err != nil || e.btDPage == 0 {
		return false
	}

	if e.hasMoved {
		e.i++
	}

	e.hasMoved = true
	if e.i < e.c {
		e.K = e.t.key(e.btDPage, e.i)
		e.V = e.K + e.t.SzKey
		return true
	}

	if e.btDPage, e.err = e.t.next(e.btDPage); e.err != nil || e.btDPage == 0 {
		return false
	}

	if e.c, e.err = e.t.len(e.btDPage); e.err != nil {
		return false
	}

	e.i = 0
	e.K = e.t.key(e.btDPage, 0)
	e.V = e.K + e.t.SzKey
	return true
}

// Prev moves the cursor to the previous item in the tree and sets the K and V
// fields accordingly. It returns true on success, or false if there is no
// previous item or an error happened while moving the cursor. Err should be
// consulted to distinguish between the two cases.
//
// Every use of the K/V fields, even the first one, must be preceded by a call
// to Next or Prev.
func (e *BTreeCursor) Prev() bool {
	if e.err != nil || e.btDPage == 0 {
		return false
	}

	if e.hasMoved || !e.hit {
		e.i--
	}

	e.hasMoved = true
	if e.i >= 0 {
		e.K = e.t.key(e.btDPage, e.i)
		e.V = e.K + e.t.SzKey
		return true
	}

	if e.btDPage, e.err = e.t.prev(e.btDPage); e.err != nil || e.btDPage == 0 {
		return false
	}

	if e.c, e.err = e.t.len(e.btDPage); e.err != nil {
		return false
	}

	e.i = e.c - 1
	e.K = e.t.key(e.btDPage, e.i)
	e.V = e.K + e.t.SzKey
	return true
}
