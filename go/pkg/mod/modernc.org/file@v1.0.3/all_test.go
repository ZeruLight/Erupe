// Copyright 2017 The File Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file // import "modernc.org/file"

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"modernc.org/internal/buffer"
	ifile "modernc.org/internal/file"
	"modernc.org/mathutil"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("# TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO) //TODOOK
}

// ============================================================================

const (
	big   = int(pageAvail) << 5
	quota = 64 << 20
	small = int(pageAvail)
)

var oK = flag.Int("k", 1, "")

func tmpMem(t testing.TB) (File, func()) {
	f, err := Mem("")
	if err != nil {
		t.Fatal(err)
	}

	return f, func() {}
}

func tmpFile(t testing.TB) (File, func()) {
	dir, err := ioutil.TempDir("", "file-test-")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(filepath.Join(dir, "foo"))
	if err != nil {
		t.Fatal(err)
	}

	fn := func() {
		os.RemoveAll(dir)
	}

	return f, fn
}

func tmpMap(t testing.TB) (File, func()) {
	dir, err := ioutil.TempDir("", "file-test-")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(filepath.Join(dir, "foo"))
	if err != nil {
		t.Fatal(err)
	}

	fi, err := Map(f)
	if err != nil {
		t.Fatal(err)
	}

	fn := func() {
		os.RemoveAll(dir)
	}

	return fi, fn
}

func TestLog(t *testing.T) {
	for i, v := range []struct{ n, e int }{
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 2},
		{5, 3},
		{6, 3},
		{7, 3},
		{8, 3},
		{9, 4},
	} {
		if g, e := log(v.n), v.e; g != e {
			t.Fatal(i, g, e)
		}
	}
}

func TestSlotRank(t *testing.T) {
	for i, v := range []struct{ n, e int }{
		{1, 0},
		{16, 0},
		{17, 1},
		{32, 1},
		{33, 2},
		{64, 2},
		{65, 3},
		{128, 3},
		{129, 4},
		{256, 4},
		{257, 5},
		{512, 5},
		{513, 6},
		{1024, 6},
	} {
		if g, e := slotRank(v.n), v.e; g != e {
			t.Fatal(i, g, e)
		}
	}
}

func TestPageRank(t *testing.T) {
	for i, v := range []struct {
		n int64
		e int
	}{
		{1025, 7},
		{1 * 4096, 7},
		{1*4096 + 1, 8},
		{2 * 4096, 8},
		{2*4096 + 1, 9},

		{3 * 4096, 9},
		{3*4096 + 1, 10},
		{4 * 4096, 10},
		{4*4096 + 1, 11},
		{5 * 4096, 11},

		{5*4096 + 1, 12},
		{6 * 4096, 12},
		{6*4096 + 1, 13},
		{7 * 4096, 13},
		{7*4096 + 1, 14},

		{8 * 4096, 14},
		{8*4096 + 1, 15},
		{9 * 4096, 15},
		{9*4096 + 1, 16},
		{10 * 4096, 16},

		{10*4096 + 1, 17},
		{11 * 4096, 17},
		{11*4096 + 1, 18},
		{12 * 4096, 18},
		{12*4096 + 1, 19},

		{13 * 4096, 19},
		{13*4096 + 1, 20},
		{14 * 4096, 20},
		{14*4096 + 1, 21},
		{15 * 4096, 21},

		{15*4096 + 1, 22},
		{1 << 16, 22},
		{1 << 20, 22},
		{1 << 30, 22},
		{1 << 40, 22},
	} {
		if g, e := pageRank(v.n), v.e; g != e {
			t.Fatal(i, g, e)
		}
	}
}

func TestCap(t *testing.T) {
	f0, fn := tmpMem(t)

	defer fn()

	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	defer func() {
		if err := f.Verify(nil); err != nil {
			t.Error(err)
		}
	}()

	if g, e := f.cap, [...]int{252, 126, 63, 31, 15, 7, 3}; g != e {
		t.Fatal(g, e)
	}
}

func testPageAlloc(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	rng, err := mathutil.NewFC32(1, max, false)
	if err != nil {
		t.Fatal(err)
	}

	var a []int64
	var p *memPage
	for rem := quota; rem > 0; {
		size := rng.Next()
		switch {
		case size > maxSlot:
			if p, err = f.newPage(int64(size)); err != nil {
				t.Fatal(err)
			}
		default:
			if p, err = f.newSharedPage(slotRank(size)); err != nil {
				t.Fatal(err)
			}
		}
		if err := p.flush(); err != nil {
			t.Fatal(err)
		}

		a = append(a, p.off)
		rem -= size
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	for i := len(a) - 1; i >= 0; i-- {
		p, err := f.openPage(a[i])
		if err != nil {
			t.Fatal(err)
		}

		if err := f.freePage(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}

	a = a[:0]
	for rem := quota; rem > 0; {
		size := rng.Next()
		switch {
		case size > maxSlot:
			if p, err = f.newPage(int64(size)); err != nil {
				t.Fatal(err)
			}
		default:
			if p, err = f.newSharedPage(slotRank(size)); err != nil {
				t.Fatal(err)
			}
		}
		if err := p.flush(); err != nil {
			t.Fatal(err)
		}

		a = append(a, p.off)
		rem -= size
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	for _, off := range a {
		p, err := f.openPage(off)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.freePage(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}

	a = a[:0]
	for rem := quota; rem > 0; {
		size := rng.Next()
		switch {
		case size > maxSlot:
			if p, err = f.newPage(int64(size)); err != nil {
				t.Fatal(err)
			}
		default:
			if p, err = f.newSharedPage(slotRank(size)); err != nil {
				t.Fatal(err)
			}
		}
		if err := p.flush(); err != nil {
			t.Fatal(err)
		}

		a = append(a, p.off)
		rem -= size
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	for i := range a {
		j := rng.Next() % len(a)
		a[i], a[j] = a[j], a[i]
	}
	for _, off := range a {
		p, err := f.openPage(off)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.freePage(p); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testPageAlloc2(t *testing.T, tmp func(testing.TB) (File, func()), max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testPageAlloc(t, f, *oK*quota, max)
}

func TestPageAllocSmallMem(t *testing.T)  { testPageAlloc2(t, tmpMem, small) }
func TestPageAllocBigMem(t *testing.T)    { testPageAlloc2(t, tmpMem, big) }
func TestPageAllocSmallMap(t *testing.T)  { testPageAlloc2(t, tmpMap, small) }
func TestPageAllocBigMap(t *testing.T)    { testPageAlloc2(t, tmpMap, big) }
func TestPageAllocSmallFile(t *testing.T) { testPageAlloc2(t, tmpFile, small) }
func TestPageAllocBigFile(t *testing.T)   { testPageAlloc2(t, tmpFile, big) }

func benchmarkPageAlloc(b *testing.B, tmp func(testing.TB) (File, func()), quota, size int) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			f0, fn := tmp(b)

			defer fn()
			defer f0.Close()

			f, err := NewAllocator(f0)
			if err != nil {
				b.Fatal(err)
			}

			var p *memPage
			b.StartTimer()
			for rem := quota; rem > 0; {
				switch {
				case size > maxSlot:
					if p, err = f.newPage(int64(size)); err != nil {
						b.Fatal(err)
					}
				default:
					if p, err = f.newSharedPage(slotRank(size)); err != nil {
						b.Fatal(err)
					}
				}
				if err := p.flush(); err != nil {
					b.Fatal(err)
				}

				rem -= size
			}
			if err := f.Flush(); err != nil {
				b.Fatal(err)
			}

			b.StopTimer()
			if err := f.Close(); err != nil {
				b.Fatal(err)
			}
		}()
	}
	b.SetBytes(int64(quota))
}

func BenchmarkPageAllocMem(b *testing.B)  { benchmarkPageAlloc(b, tmpMem, quota, small) }
func BenchmarkPageAllocMap(b *testing.B)  { benchmarkPageAlloc(b, tmpMap, quota, small) }
func BenchmarkPageAllocFile(b *testing.B) { benchmarkPageAlloc(b, tmpFile, quota, small) }

func testAlloc(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	var a []int64
	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	// Allocate
	for rem := quota; rem > 0; {
		size := srng.Next()%max + 1
		off, err := f.Alloc(int64(size))
		if err != nil {
			t.Fatal(err)
		}

		g, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if e := int64(size); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(size)
		b := *p
		for i := range b {
			b[i] = byte(vrng.Next())
		}
		if _, err := f0.WriteAt(b, off); err != nil {
			t.Fatal(err)
		}

		buffer.Put(p)
		a = append(a, off)
		rem -= size
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	srng.Seek(0)
	vrng.Seek(0)
	// Verify
	for i, off := range a {
		size := srng.Next()%max + 1
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(size); g < e {
			t.Fatal(i, g, e)
		}

		p := buffer.Get(size)
		b := *p
		if n, err := f0.ReadAt(b, off); n != len(b) {
			t.Fatal(err)
		}

		for j, g := range b {
			if e := byte(vrng.Next()); g != e {
				t.Fatalf("%v: %#x+%#x: %#02x %#02x", i, off, j, g, e)
			}
		}
		buffer.Put(p)
	}
	// Shuffle
	for i := range a {
		j := srng.Next() % len(a)
		a[i], a[j] = a[j], a[i]
	}
	// Free
	for _, off := range a {
		sz, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		p := buffer.CGet(int(sz))
		_, err = f0.WriteAt(*p, off)
		buffer.Put(p)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testAllocB(t *testing.T, tmp func(testing.TB) (File, func()), quota, max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testAlloc(t, f, quota, max)
}

func TestAllocSmallMem(t *testing.T)  { testAllocB(t, tmpMem, *oK*quota, small) }
func TestAllocBigMem(t *testing.T)    { testAllocB(t, tmpMem, *oK*quota, big) }
func TestAllocSmallMap(t *testing.T)  { testAllocB(t, tmpMap, *oK*quota, small) }
func TestAllocBigMap(t *testing.T)    { testAllocB(t, tmpMap, *oK*quota, big) }
func TestAllocSmallFile(t *testing.T) { testAllocB(t, tmpFile, *oK*quota, small) }
func TestAllocBigFile(t *testing.T)   { testAllocB(t, tmpFile, *oK*quota, big) }

func testAlloc2(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	var a []int64
	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	// Allocate
	for rem := quota; rem > 0; {
		size := srng.Next()%max + 1
		off, err := f.Alloc(int64(size))
		if err != nil {
			t.Fatal(err)
		}

		g, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if e := int64(size); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(size)
		b := *p
		for i := range b {
			b[i] = byte(vrng.Next())
		}
		if _, err := f0.WriteAt(b, off); err != nil {
			t.Fatal(err)
		}

		buffer.Put(p)
		a = append(a, off)
		rem -= size
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	srng.Seek(0)
	vrng.Seek(0)
	// Verify & free
	for i, off := range a {
		size := srng.Next()%max + 1
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(size); g < e {
			t.Fatal(i, g, e)
		}

		p := buffer.Get(size)
		b := *p
		if n, err := f0.ReadAt(b, off); n != len(b) {
			t.Fatal(err)
		}

		for i, g := range b {
			if e := byte(vrng.Next()); g != e {
				t.Fatalf("%#x: %#02x %#02x", i, g, e)
			}
		}
		buffer.Put(p)

		p = buffer.CGet(int(u))
		_, err = f0.WriteAt(*p, off)
		buffer.Put(p)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testAlloc2B(t *testing.T, tmp func(testing.TB) (File, func()), quota, max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testAlloc2(t, f, quota, max)
}

func TestAlloc2SmallMem(t *testing.T)  { testAlloc2B(t, tmpMem, *oK*quota, small) }
func TestAlloc2BigMem(t *testing.T)    { testAlloc2B(t, tmpMem, *oK*quota, big) }
func TestAlloc2SmallMap(t *testing.T)  { testAlloc2B(t, tmpMap, *oK*quota, small) }
func TestAlloc2BigMap(t *testing.T)    { testAlloc2B(t, tmpMap, *oK*quota, big) }
func TestAlloc2SmallFile(t *testing.T) { testAlloc2B(t, tmpFile, *oK*quota, small) }
func TestAlloc2BigFile(t *testing.T)   { testAlloc2B(t, tmpFile, *oK*quota, big) }

func testAlloc3(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	m := map[int64][]byte{}
	rem := quota
	for rem > 0 {
		switch srng.Next() % 3 {
		case 0, 1: // 2/3 allocate
			size := srng.Next()%max + 1
			rem -= size
			off, err := f.Alloc(int64(size))
			if err != nil {
				t.Fatal(err)
			}

			g, err := f.UsableSize(off)
			if err != nil {
				t.Fatal(err)
			}

			if e := int64(size); g < e {
				t.Fatalf("%#x, %#x: %#x %#x", off, size, g, e)
			}

			b := make([]byte, size)
			for i := range b {
				b[i] = byte(vrng.Next())
			}
			if _, err := f0.WriteAt(b, off); err != nil {
				t.Fatal(err)
			}

			m[off] = b
		default: // 1/3 free
			for off, b := range m {
				u, err := f.UsableSize(off)
				if err != nil {
					t.Fatal(err)
				}

				if g, e := u, int64(len(b)); g < e {
					t.Fatal(g, e)
				}

				p := buffer.Get(len(b))
				b2 := *p
				if n, err := f0.ReadAt(b2, off); n != len(b2) {
					t.Fatal(err)
				}

				if !bytes.Equal(b, b2) {
					t.Fatal("corrupted data")
				}

				buffer.Put(p)

				p = buffer.CGet(int(u))
				_, err = f0.WriteAt(*p, off)
				buffer.Put(p)
				if err != nil {
					t.Fatal(err)
				}

				rem += len(b)
				if err := f.Free(off); err != nil {
					t.Fatal(err)
				}

				delete(m, off)
				break
			}
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	// Verify & free
	for off, b := range m {
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(len(b)); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(len(b))
		b2 := *p
		if n, err := f0.ReadAt(b2, off); n != len(b2) {
			t.Fatal(err)
		}

		if !bytes.Equal(b, b2) {
			t.Fatal("corrupted data")
		}

		buffer.Put(p)

		p = buffer.CGet(int(u))
		_, err = f0.WriteAt(*p, off)
		buffer.Put(p)
		if err != nil {
			t.Fatal(err)
		}

		rem += len(b)
		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testAlloc3B(t *testing.T, tmp func(testing.TB) (File, func()), quota, max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testAlloc3(t, f, quota, max)
}

func TestAlloc3SmallMem(t *testing.T)  { testAlloc3B(t, tmpMem, *oK*quota, small) }
func TestAlloc3BigMem(t *testing.T)    { testAlloc3B(t, tmpMem, *oK*quota, big) }
func TestAlloc3SmallMap(t *testing.T)  { testAlloc3B(t, tmpMap, *oK*quota, small) }
func TestAlloc3BigMap(t *testing.T)    { testAlloc3B(t, tmpMap, *oK*quota, big) }
func TestAlloc3SmallFile(t *testing.T) { testAlloc3B(t, tmpFile, *oK*quota, small) }
func TestAlloc3BigFile(t *testing.T)   { testAlloc3B(t, tmpFile, *oK*quota, big) }

func testReopen(t *testing.T, quota, max int) {
	dir, err := ioutil.TempDir("", "file-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)

	f0, err := os.Create(filepath.Join(dir, "foo"))
	if err != nil {
		t.Fatal(err)
	}

	nm := f0.Name()
	f1, err := ifile.Open(f0)
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewAllocator(f1)
	if err != nil {
		t.Fatal(err)
	}

	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	m := map[int64][]byte{}
	rem := quota
	for rem > 0 {
		switch srng.Next() % 3 {
		case 0, 1: // 2/3 allocate
			size := srng.Next()%max + 1
			rem -= size
			off, err := f.Alloc(int64(size))
			if err != nil {
				t.Fatal(err)
			}

			g, err := f.UsableSize(off)
			if err != nil {
				t.Fatal(err)
			}

			if e := int64(size); g < e {
				t.Fatal(g, e)
			}

			b := make([]byte, size)
			for i := range b {
				b[i] = byte(vrng.Next())
			}
			if _, err := f0.WriteAt(b, off); err != nil {
				t.Fatal(err)
			}

			m[off] = b
		default: // 1/3 free
			for off, b := range m {
				u, err := f.UsableSize(off)
				if err != nil {
					t.Fatal(err)
				}

				if g, e := u, int64(len(b)); g < e {
					t.Fatal(g, e)
				}

				p := buffer.Get(len(b))
				b2 := *p
				if n, err := f0.ReadAt(b2, off); n != len(b2) {
					t.Fatal(err)
				}

				if !bytes.Equal(b, b2) {
					t.Fatal("corrupted data")
				}

				buffer.Put(p)

				p = buffer.CGet(int(u))
				_, err = f0.WriteAt(*p, off)
				buffer.Put(p)
				if err != nil {
					t.Fatal(err)
				}

				rem += len(b)
				if err := f.Free(off); err != nil {
					t.Fatal(err)
				}

				delete(m, off)
				break
			}
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	ts := f.testStat
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if f0, err = os.OpenFile(nm, os.O_RDWR, 0600); err != nil {
		t.Fatal(err)
	}

	if f1, err = ifile.Open(f0); err != nil {
		t.Fatal(err)
	}

	if f, err = NewAllocator(f1); err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	f.testStat = ts
	// Verify & free
	for off, b := range m {
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(len(b)); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(len(b))
		b2 := *p
		if n, err := f0.ReadAt(b2, off); n != len(b2) {
			t.Fatal(err)
		}

		if !bytes.Equal(b, b2) {
			t.Fatal("corrupted data")
		}

		buffer.Put(p)

		p = buffer.CGet(int(u))
		_, err = f0.WriteAt(*p, off)
		buffer.Put(p)
		if err != nil {
			t.Fatal(err)
		}

		rem += len(b)
		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func TestReopenSmallFile(t *testing.T) { testReopen(t, *oK*quota, small) }
func TestReopenBigFile(t *testing.T)   { testReopen(t, *oK*quota, big) }

func testCalloc(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	m := map[int64][]byte{}
	rem := quota
	for rem > 0 {
		switch srng.Next() % 3 {
		case 0, 1: // 2/3 allocate
			size := srng.Next()%max + 1
			rem -= size
			off, err := f.Calloc(int64(size))
			if err != nil {
				t.Fatal(err)
			}

			g, err := f.UsableSize(off)
			if err != nil {
				t.Fatal(err)
			}

			if e := int64(size); g < e {
				t.Fatal(g, e)
			}

			b := make([]byte, size)
			if n, err := f0.ReadAt(b, off); n != len(b) {
				t.Fatal(err)
			}

			for _, v := range b {
				if v != 0 {
					t.Fatal(v)
				}
			}

			for i := range b {
				b[i] = byte(vrng.Next())
			}
			if _, err := f0.WriteAt(b, off); err != nil {
				t.Fatal(err)
			}

			m[off] = b
		default: // 1/3 free
			for off, b := range m {
				u, err := f.UsableSize(off)
				if err != nil {
					t.Fatal(err)
				}

				if g, e := u, int64(len(b)); g < e {
					t.Fatal(g, e)
				}

				p := buffer.Get(len(b))
				b2 := *p
				if n, err := f0.ReadAt(b2, off); n != len(b2) {
					t.Fatal(err)
				}

				if !bytes.Equal(b, b2) {
					t.Fatal("corrupted data")
				}

				buffer.Put(p)
				rem += len(b)
				if err := f.Free(off); err != nil {
					t.Fatal(err)
				}

				delete(m, off)
				break
			}
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	// Verify & free
	for off, b := range m {
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(len(b)); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(len(b))
		b2 := *p
		if n, err := f0.ReadAt(b2, off); n != len(b2) {
			t.Fatal(err)
		}

		if !bytes.Equal(b, b2) {
			t.Fatal("corrupted data")
		}

		buffer.Put(p)
		rem += len(b)
		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testCalloc2(t *testing.T, tmp func(testing.TB) (File, func()), quota, max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testCalloc(t, f, quota, max)
}

func TestCallocSmallMem(t *testing.T)  { testCalloc2(t, tmpMem, *oK*quota, small) }
func TestCallocBigMem(t *testing.T)    { testCalloc2(t, tmpMem, *oK*quota, big) }
func TestCallocSmallMap(t *testing.T)  { testCalloc2(t, tmpMap, *oK*quota, small) }
func TestCallocBigMap(t *testing.T)    { testCalloc2(t, tmpMap, *oK*quota, big) }
func TestCallocSmallFile(t *testing.T) { testCalloc2(t, tmpFile, *oK*quota, small) }
func TestCallocBigFile(t *testing.T)   { testCalloc2(t, tmpFile, *oK*quota, big) }

func testRealloc(t *testing.T, f0 File, quota, max int) {
	f, err := NewAllocator(f0)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	srng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	vrng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	m := map[int64][]byte{}
	rem := quota
	for rem > 0 {
		switch srng.Next() % 4 {
		case 0, 1: // 2/4 allocate
			size := srng.Next()%max + 1
			rem -= size
			off, err := f.Alloc(int64(size))
			if err != nil {
				t.Fatal(err)
			}

			g, err := f.UsableSize(off)
			if err != nil {
				t.Fatal(err)
			}

			if e := int64(size); g < e {
				t.Fatal(g, e)
			}

			b := make([]byte, size)
			for i := range b {
				b[i] = byte(vrng.Next())
			}
			if _, err := f0.WriteAt(b, off); err != nil {
				t.Fatal(err)
			}

			m[off] = b
		case 2: // 1/4 realloc
			for off, b := range m {
				u, err := f.UsableSize(off)
				if err != nil {
					t.Fatal(err)
				}

				if g, e := u, int64(len(b)); g < e {
					t.Fatalf("%#x, %#x: %#x, %#x", off, len(b), g, e)
				}

				p := buffer.Get(len(b))
				b2 := *p
				if n, err := f0.ReadAt(b2, off); n != len(b2) {
					t.Fatal(err)
				}

				if !bytes.Equal(b, b2) {
					t.Fatal("corrupted data")
				}

				buffer.Put(p)

				size := srng.Next()%max + 1
				off2, err := f.Realloc(off, int64(size))
				if err != nil {
					t.Fatal(err)
				}

				min := mathutil.Min(len(b), size)
				b3 := make([]byte, size)
				if n, err := f0.ReadAt(b3[:min], off2); n != min {
					t.Fatal(err)
				}

				if !bytes.Equal(b[:min], b3[:min]) {
					t.Fatal("corrupted data")
				}

				for i := range b3 {
					b3[i] = byte(vrng.Next())
				}
				if _, err := f0.WriteAt(b3, off2); err != nil {
					t.Fatal(err)
				}

				delete(m, off)
				m[off2] = b3
				break
			}
		default: // 1/4 free
			for off, b := range m {
				u, err := f.UsableSize(off)
				if err != nil {
					t.Fatal(err)
				}

				if g, e := u, int64(len(b)); g < e {
					t.Fatalf("%#x, %#x: %#x, %#x", off, len(b), g, e)
				}

				p := buffer.Get(len(b))
				b2 := *p
				if n, err := f0.ReadAt(b2, off); n != len(b2) {
					t.Fatal(err)
				}

				if !bytes.Equal(b, b2) {
					t.Fatal("corrupted data")
				}

				buffer.Put(p)

				p = buffer.CGet(int(u))
				_, err = f0.WriteAt(*p, off)
				buffer.Put(p)
				if err != nil {
					t.Fatal(err)
				}

				rem += len(b)
				if err := f.Free(off); err != nil {
					t.Fatal(err)
				}

				delete(m, off)
				break
			}
		}
	}
	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	t.Logf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize)
	// Verify & free
	for off, b := range m {
		u, err := f.UsableSize(off)
		if err != nil {
			t.Fatal(err)
		}

		if g, e := u, int64(len(b)); g < e {
			t.Fatal(g, e)
		}

		p := buffer.Get(len(b))
		b2 := *p
		if n, err := f0.ReadAt(b2, off); n != len(b2) {
			t.Fatal(err)
		}

		if !bytes.Equal(b, b2) {
			t.Fatal("corrupted data")
		}

		buffer.Put(p)

		p = buffer.CGet(int(u))
		_, err = f0.WriteAt(*p, off)
		buffer.Put(p)
		if err != nil {
			t.Fatal(err)
		}

		rem += len(b)
		if err := f.Free(off); err != nil {
			t.Fatal(err)
		}
	}

	if err := f.Verify(nil); err != nil {
		t.Fatal(err)
	}

	if f.allocs != 0 || f.bytes != 0 || f.npages != 0 || f.fsize != szFile || f.slots != [slotRanks]int64{} || f.pages != [ranks]int64{} {
		t.Fatalf("quota %v, max %v, allocs %v, bytes %v, pages %v, fsize %v, slots %v, pages %v", quota, max, f.allocs, f.bytes, f.npages, f.fsize, f.slots, f.pages)
	}
}

func testReallocB(t *testing.T, tmp func(testing.TB) (File, func()), quota, max int) {
	f, fn := tmp(t)

	defer fn()
	defer f.Close()

	testRealloc(t, f, quota, max)
}

func TestReallocSmallMem(t *testing.T)  { testReallocB(t, tmpMem, *oK*quota, small) }
func TestReallocBigMem(t *testing.T)    { testReallocB(t, tmpMem, *oK*quota, big) }
func TestReallocSmallMap(t *testing.T)  { testReallocB(t, tmpMap, *oK*quota, small) }
func TestReallocBigMap(t *testing.T)    { testReallocB(t, tmpMap, *oK*quota, big) }
func TestReallocSmallFile(t *testing.T) { testReallocB(t, tmpFile, *oK*quota, small) }
func TestReallocBigFile(t *testing.T)   { testReallocB(t, tmpFile, *oK*quota, big) }

func benchmarkAlloc(b *testing.B, tmp func(testing.TB) (File, func()), quota, max int) {
	b.SetBytes(int64(quota))
	rng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		b.Fatal(err)
	}

	a := []int64{}
	for quota > 0 {
		size := rng.Next()%max + 1
		a = append(a, int64(size))
		quota -= size
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			f0, fn := tmp(b)

			defer fn()
			defer f0.Close()

			f, err := NewAllocator(f0)
			if err != nil {
				b.Fatal(err)
			}

			defer f.Close()

			b.StartTimer()
			for _, v := range a {
				if _, err := f.Alloc(v); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkAllocSmallMem(b *testing.B)  { benchmarkAlloc(b, tmpMem, quota, small) }
func BenchmarkAllocBigMem(b *testing.B)    { benchmarkAlloc(b, tmpMem, quota, big) }
func BenchmarkAllocSmallMap(b *testing.B)  { benchmarkAlloc(b, tmpMap, quota, small) }
func BenchmarkAllocBigMap(b *testing.B)    { benchmarkAlloc(b, tmpMap, quota, big) }
func BenchmarkAllocSmallFile(b *testing.B) { benchmarkAlloc(b, tmpFile, quota, small) }
func BenchmarkAllocBigFile(b *testing.B)   { benchmarkAlloc(b, tmpFile, quota, big) }

func benchmarkCalloc(b *testing.B, tmp func(testing.TB) (File, func()), quota, max int) {
	b.SetBytes(int64(quota))
	rng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		b.Fatal(err)
	}

	a := []int64{}
	for quota > 0 {
		size := rng.Next()%max + 1
		a = append(a, int64(size))
		quota -= size
	}
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			f0, fn := tmp(b)

			defer fn()
			defer f0.Close()

			f, err := NewAllocator(f0)
			if err != nil {
				b.Fatal(err)
			}

			defer f.Close()

			b.StartTimer()
			for _, v := range a {
				if _, err := f.Calloc(v); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkCallocSmallMem(b *testing.B)  { benchmarkCalloc(b, tmpMem, quota, small) }
func BenchmarkCallocBigMem(b *testing.B)    { benchmarkCalloc(b, tmpMem, quota, big) }
func BenchmarkCallocSmallMap(b *testing.B)  { benchmarkCalloc(b, tmpMap, quota, small) }
func BenchmarkCallocBigMap(b *testing.B)    { benchmarkCalloc(b, tmpMap, quota, big) }
func BenchmarkCallocSmallFile(b *testing.B) { benchmarkCalloc(b, tmpFile, quota, small) }
func BenchmarkCallocBigFile(b *testing.B)   { benchmarkCalloc(b, tmpFile, quota, big) }

func benchmarkFree(b *testing.B, tmp func(testing.TB) (File, func()), quota, max int) {
	b.SetBytes(int64(quota))
	rng, err := mathutil.NewFC32(0, math.MaxInt32, true)
	if err != nil {
		b.Fatal(err)
	}

	a := []int64{}
	for q := quota; q > 0; {
		size := rng.Next()%max + 1
		a = append(a, int64(size))
		q -= size
	}
	c := make([]int64, len(a))
	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		func() {
			f0, fn := tmp(b)

			defer fn()
			defer f0.Close()

			f, err := NewAllocator(f0)
			if err != nil {
				b.Fatal(err)
			}

			defer f.Close()

			for i, v := range a {
				if c[i], err = f.Alloc(v); err != nil {
					b.Fatal(err)
				}
			}
			for i := range c {
				j := rng.Next() % len(c)
				c[i], c[j] = c[j], c[i]
			}

			b.StartTimer()
			for _, v := range c {
				if err := f.Free(v); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
		}()
	}
}

func BenchmarkFreeSmallMem(b *testing.B)  { benchmarkFree(b, tmpMem, quota, small) }
func BenchmarkFreeBigMem(b *testing.B)    { benchmarkFree(b, tmpMem, quota, big) }
func BenchmarkFreeSmallMap(b *testing.B)  { benchmarkFree(b, tmpMap, quota, small) }
func BenchmarkFreeBigMap(b *testing.B)    { benchmarkFree(b, tmpMap, quota, big) }
func BenchmarkFreeSmallFile(b *testing.B) { benchmarkFree(b, tmpFile, quota, small) }
func BenchmarkFreeBigFile(b *testing.B)   { benchmarkFree(b, tmpFile, quota, big) }

func equal(f, g File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	p := buffer.Get(int(fi.Size()))
	defer buffer.Put(p)
	b := *p
	p = buffer.Get(int(fi.Size()))
	defer buffer.Put(p)
	b2 := *p

	if n, _ := f.ReadAt(b, 0); n != len(b) {
		return fmt.Errorf("read returned %v, expected %v", n, len(b))
	}

	if n, _ := g.ReadAt(b2, 0); n != len(b) {
		return fmt.Errorf("read returned %v, expected %v", n, len(b2))
	}

	if !bytes.Equal(b, b2) {
		return fmt.Errorf("files are different, expected equal")
	}

	return nil
}

func testWAL(t *testing.T, tmp func(testing.TB) (File, func())) {
	const (
		sz       = 1 << 24
		pageLog  = 16
		pageSize = 1 << pageLog
		wsz      = 2 * pageSize
		N        = 1000
	)

	f, fn := tmp(t)
	defer fn()
	g, fn := tmp(t)
	defer fn()
	w, fn := tmp(t)
	defer fn()

	fd := make([]byte, sz)
	rng, err := mathutil.NewFC32(1, math.MaxInt32, true)
	if err != nil {
		t.Fatal(err)
	}

	for i := range fd {
		fd[i] = byte(rng.Next())
	}

	if _, err := f.WriteAt(fd, 0); err != nil {
		t.Fatal(err)
	}

	if _, err := g.WriteAt(fd, 0); err != nil {
		t.Fatal(err)
	}

	wal, err := NewWAL(f, w, 0, pageLog)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, wsz)
	for j := 0; j < 3; j++ {
		for i := 0; i < N; i++ {
			fi, err := wal.Stat()
			if err != nil {
				t.Fatal(err)
			}

			fsz := fi.Size()
			off := int64(rng.Next() % int(fsz))
			nw := rng.Next() % wsz
			for i := range buf[:nw] {
				buf[i] = byte(rng.Next())
			}

			if _, err := g.WriteAt(buf[:nw], off); err != nil {
				t.Fatal(err)
			}

			if _, err := wal.WriteAt(buf[:nw], off); err != nil {
				t.Fatal(err)
			}

			if err := equal(g, wal); err != nil {
				t.Fatal(err)
			}

			if i%10 != 0 {
				continue
			}

			off = sz/2 + int64(rng.Next())%sz/2
			if err := g.Truncate(off); err != nil {
				t.Fatal(err)
			}

			if err := wal.Truncate(off); err != nil {
				t.Fatal(err)
			}
		}
		if err := equal(g, wal); err != nil {
			t.Fatal(err)
		}

		if err := wal.Commit(); err != nil {
			t.Fatal(err)
		}

		if err := equal(f, g); err != nil {
			t.Fatal(err)
		}
	}

	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	fsz := fi.Size()
	off := int64(rng.Next() % int(fsz))
	nw := rng.Next() % wsz
	for i := range buf[:nw] {
		buf[i] = byte(rng.Next())
	}
	if _, err := g.WriteAt(buf[:nw], off); err != nil {
		t.Fatal(err)
	}

	if _, err := wal.WriteAt(buf[:nw], off); err != nil {
		t.Fatal(err)
	}

	if err := equal(g, wal); err != nil {
		t.Fatal(err)
	}

	// corrupt f
	if err := f.Truncate(off + 2*wsz); err != nil {
		t.Fatal(err)
	}

	// corrupt f more.
	for i := range buf[:nw] {
		buf[i] = 0
	}
	if _, err := f.WriteAt(buf[:nw], off); err != nil {
		t.Fatal(err)
	}

	if err := equal(f, wal); err == nil {
		t.Fatal("unexpected success")
	}

	crash = true
	defer func() { crash = false }()

	if err := wal.Commit(); err != nil {
		t.Fatal(err)
	}

	if wal, err = NewWAL(f, w, 0, pageLog); err != nil {
		t.Fatal(err)
	}

	if err := equal(f, g); err != nil {
		t.Fatal(err)
	}
}

func TestWALMem(t *testing.T)  { testWAL(t, tmpMem) }
func TestWALMap(t *testing.T)  { testWAL(t, tmpMap) }
func TestWALFile(t *testing.T) { testWAL(t, tmpFile) }

func benchmarkWALWrite(b *testing.B, tmp func(testing.TB) (File, func()), wsz int) {
	const (
		sz      = 1 << 24
		pageLog = 16
	)

	f, fn := tmp(b)
	defer fn()
	w, fn := tmp(b)
	defer fn()

	fd := make([]byte, sz)
	rng, err := mathutil.NewFC32(1, math.MaxInt32, true)
	if err != nil {
		b.Fatal(err)
	}

	for i := range fd {
		fd[i] = byte(rng.Next())
	}

	if _, err := f.WriteAt(fd, 0); err != nil {
		b.Fatal(err)
	}

	wal, err := NewWAL(f, w, 0, pageLog)
	if err != nil {
		b.Fatal(err)
	}

	buf := make([]byte, wsz)
	for i := range buf {
		buf[i] = byte(rng.Next())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		off := int64(rng.Next() % sz)
		if _, err := wal.WriteAt(buf, off); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.SetBytes(int64(wsz))
	if err := wal.Commit(); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkWALWriteMem1(b *testing.B)   { benchmarkWALWrite(b, tmpMem, 1<<0) }
func BenchmarkWALWriteMem16(b *testing.B)  { benchmarkWALWrite(b, tmpMem, 1<<4) }
func BenchmarkWALWriteMem256(b *testing.B) { benchmarkWALWrite(b, tmpMem, 1<<8) }
func BenchmarkWALWriteMem4K(b *testing.B)  { benchmarkWALWrite(b, tmpMem, 1<<12) }
func BenchmarkWALWriteMem64K(b *testing.B) { benchmarkWALWrite(b, tmpMem, 1<<16) }
func BenchmarkWALWriteMem1M(b *testing.B)  { benchmarkWALWrite(b, tmpMem, 1<<20) }

func BenchmarkWALWriteMap1(b *testing.B)   { benchmarkWALWrite(b, tmpMap, 1<<0) }
func BenchmarkWALWriteMap16(b *testing.B)  { benchmarkWALWrite(b, tmpMap, 1<<4) }
func BenchmarkWALWriteMap256(b *testing.B) { benchmarkWALWrite(b, tmpMap, 1<<8) }
func BenchmarkWALWriteMap4K(b *testing.B)  { benchmarkWALWrite(b, tmpMap, 1<<12) }
func BenchmarkWALWriteMap64K(b *testing.B) { benchmarkWALWrite(b, tmpMap, 1<<16) }
func BenchmarkWALWriteMap1M(b *testing.B)  { benchmarkWALWrite(b, tmpMap, 1<<20) }

func BenchmarkWALWriteFile1(b *testing.B)   { benchmarkWALWrite(b, tmpFile, 1<<0) }
func BenchmarkWALWriteFile16(b *testing.B)  { benchmarkWALWrite(b, tmpFile, 1<<4) }
func BenchmarkWALWriteFile256(b *testing.B) { benchmarkWALWrite(b, tmpFile, 1<<8) }
func BenchmarkWALWriteFile4K(b *testing.B)  { benchmarkWALWrite(b, tmpFile, 1<<12) }
func BenchmarkWALWriteFile64K(b *testing.B) { benchmarkWALWrite(b, tmpFile, 1<<16) }
func BenchmarkWALWriteFile1M(b *testing.B)  { benchmarkWALWrite(b, tmpFile, 1<<20) }

func ExampleWAL() {
	const pageLog = 1

	dir, err := ioutil.TempDir("", "file-example-")
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(dir)

	f, err := os.Create(filepath.Join(dir, "f"))
	if err != nil {
		panic(err)
	}

	w, err := os.Create(filepath.Join(dir, "w"))
	if err != nil {
		panic(err)
	}

	db, err := NewWAL(f, w, 0, pageLog)
	if err != nil {
		panic(err)
	}

	write := func(b []byte, off int64) {
		if _, err := db.WriteAt(b, off); err != nil {
			panic(err)
		}

		fi, err := w.Stat()
		if err != nil {
			panic(err)
		}

		fmt.Printf("---- db.WriteAt(%#v, %v)\n", b, off)
		fmt.Printf("journal pages %v\n", len(db.m))
		fmt.Printf("journal size %v\n", fi.Size())
	}

	read := func() {
		b := make([]byte, 1024)
		n, err := f.ReadAt(b, 0)
		if n == 0 && err != io.EOF {
			panic(err.Error())
		}

		fmt.Printf("Read   committed: |% x|\n", b[:n])
		if n, err = db.ReadAt(b, 0); n == 0 && err != io.EOF {
			panic(err.Error())
		}

		fmt.Printf("Read uncommitted: |% x|\n", b[:n])
	}

	commit := func() {
		if err := db.Commit(); err != nil {
			panic(err)
		}

		fmt.Printf("---- Commit()\n")
		fmt.Printf("journal pages %v\n", len(db.m))
		read()
	}

	fmt.Printf("logical page size  %v\n", 1<<pageLog)
	fmt.Printf("journal page size %v\n", 1<<pageLog+8)
	write([]byte{1, 2, 3}, 0)
	read()
	commit()
	write([]byte{0xff}, 1)
	read()
	commit()
	// Output:
	// logical page size  2
	// journal page size 10
	// ---- db.WriteAt([]byte{0x1, 0x2, 0x3}, 0)
	// journal pages 2
	// journal size 20
	// Read   committed: ||
	// Read uncommitted: |01 02 03|
	// ---- Commit()
	// journal pages 0
	// Read   committed: |01 02 03|
	// Read uncommitted: |01 02 03|
	// ---- db.WriteAt([]byte{0xff}, 1)
	// journal pages 1
	// journal size 10
	// Read   committed: |01 02 03|
	// Read uncommitted: |01 ff 03|
	// ---- Commit()
	// journal pages 0
	// Read   committed: |01 ff 03|
	// Read uncommitted: |01 ff 03|
}
