// Copyright 2017 The File Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package file handles write-ahead logs and space management of os.File-like
// entities.
//
// Build status
//
// available at https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2ffile
//
// Changelog
//
// 2017-09-09: Write ahead log support - initial release.
package file // import "modernc.org/file"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"modernc.org/internal/buffer"
	ifile "modernc.org/internal/file"
	"modernc.org/mathutil"
)

const (
	// AllocAlign defines File offsets of allocations are 0 (mod AllocAlign).
	AllocAlign = 16
	// LowestAllocationOffset is the offset of the first allocation of an empty File.
	LowestAllocationOffset = szFile + szPage

	bufSize       = 1 << 20 // Calloc, Realloc.
	firstPageRank = maxSharedRank + 1
	maxSharedRank = slotRanks - 1
	maxSlot       = 1024
	oFilePages    = int64(unsafe.Offsetof(file{}.pages))
	oFileSkip     = int64(unsafe.Offsetof(file{}.skip))
	oFileSlots    = int64(unsafe.Offsetof(file{}.slots))
	oNodeNext     = int64(unsafe.Offsetof(node{}.next))
	oNodePrev     = int64(unsafe.Offsetof(node{}.prev))
	oPageBrk      = int64(unsafe.Offsetof(page{}.brk))
	oPageNext     = int64(unsafe.Offsetof(page{}.next))
	oPagePrev     = int64(unsafe.Offsetof(page{}.prev))
	oPageRank     = int64(unsafe.Offsetof(page{}.rank))
	oPageSize     = int64(unsafe.Offsetof(page{}.size))
	oPageUsed     = int64(unsafe.Offsetof(page{}.usedSlots))
	pageAvail     = pageSize - szPage - szInt64
	pageBits      = 12
	pageMask      = pageSize - 1
	pageSize      = 1 << pageBits
	ranks         = 23
	slotRanks     = 7
	szFile        = int64(unsafe.Sizeof(file{}))
	szInt64       = 8
	szNode        = int64(unsafe.Sizeof(node{}))
	szPage        = int64(unsafe.Sizeof(page{}))
)

var (
	_ File        = (*os.File)(nil)
	_ io.ReaderAt = File(nil)
	_ io.WriterAt = File(nil)

	zAllocator Allocator
	zMemNode   memNode
	zMemPage   memPage

	allocatorPool = sync.Pool{New: func() interface{} { return &Allocator{} }}
	memNodePool   = sync.Pool{New: func() interface{} { return &memNode{} }}
	memPagePool   = sync.Pool{New: func() interface{} { return &memPage{} }}
)

func init() {
	if szFile%AllocAlign != 0 || szFile != 256 || szNode%AllocAlign != 0 || szPage%AllocAlign != 0 {
		panic("internal error: invalid configuration")
	}
}

func aligned(off int64) bool { return off&(AllocAlign-1) == 0 }

// 0:     1 -      1
// 1:    10 -     10
// 2:    11 -    100
// 3:   101 -   1000
// 4:  1001 -  10000
// 5: 10001 - 100000
// ...
//
// 1<<log(n) is the rounded up to nearest power of 2 storage size required for
// n bytes.
func log(n int) int {
	if n <= 0 {
		panic(fmt.Errorf("internal error: log(%v)", n))
	}

	return mathutil.BitLen(n - 1)
}

//  7:       1025 - 1*4096
//  8:   1*4096+1 - 2*4096
//  9:   2*4096+1 - 3*4096
//  ...
func pageRank(n int64) int {
	if n <= maxSlot {
		panic(fmt.Errorf("internal error: pageRank(%v)", n))
	}

	r := int(roundup64(n, pageSize)>>pageBits) + 6
	if r >= ranks {
		r = ranks - 1
	}
	return r
}

// Get an int64 from b. len(b) must be >= 8.
func read(b []byte) int64 {
	var n int64
	for _, v := range b[:8] {
		n = n<<8 | int64(v)
	}
	return n
}

func rank(n int64) int {
	if n <= maxSlot {
		return slotRank(int(n))
	}

	return pageRank(n)
}

func roundup(n, m int) int       { return (n + m - 1) &^ (m - 1) }
func roundup64(n, m int64) int64 { return (n + m - 1) &^ (m - 1) }

//  0:      1 -   16
//  1:     17 -   32
//  2:     33 -   64
//  3:     65 -  128
//  4:    129 -  256
//  5:    257 -  512
//  6:    513 - 1024
func slotRank(n int) int {
	if n < 1 || n > 1024 {
		panic(fmt.Errorf("internal error: slotRank(%v)", n))
	}

	return log(roundup(n, AllocAlign)) - 4
}

// Put an int64 in b. len(b) must be >= 8.
func write(b []byte, n int64) {
	b = b[:8]
	for i := range b {
		b[i] = byte(n >> 56)
		n <<= 8
	}
}

// node is a double linked list item in File.
type node struct {
	prev, next int64
}

// memNode is the memory representation of a loaded node.
type memNode struct {
	*Allocator
	node
	off int64 // Offset of the node in File.

	dirty bool
}

func (m *memNode) close() {
	*m = zMemNode
	memNodePool.Put(m)
}

// flush stores/persists m if dirty.
func (m *memNode) flush() error {
	if !m.dirty {
		return nil
	}

	b := m.nodeBuf[:]
	write(b[oNodeNext:], m.next)
	write(b[oNodePrev:], m.prev)
	_, err := m.f.WriteAt(b, m.off)
	m.dirty = err == nil
	return err
}

// setNext sets m's next link.
func (m *memNode) setNext(n int64) { m.next = n; m.dirty = true }

// setPrev sets m's prev link.
func (m *memNode) setPrev(n int64) { m.prev = n; m.dirty = true }

// unlink removes m from the list it belongs to.
func (m *memNode) unlink(rank int) error {
	if m.prev != 0 {
		// Turn m.prev -> m -> m.next into m.prev -> m.next
		prev, err := m.openNode(m.prev)
		if err != nil {
			return err
		}

		prev.setNext(m.next)
		if err := prev.flush(); err != nil {
			return err
		}

		prev.close()
	}

	if m.next != 0 {
		// Turn m.prev <- m <- m.next into m.prev <- m.next
		next, err := m.openNode(m.next)
		if err != nil {
			return err
		}

		next.setPrev(m.prev)
		if err := next.flush(); err != nil {
			return err
		}

		next.close()
	}

	if m.slots[rank] == m.off { // m was the head of a slot list
		m.setSlot(rank, m.next)
	}

	return nil
}

// page is a page header in File.
type page struct {
	brk int64
	node
	rank      int64
	size      int64
	usedSlots int64
}

// memPage is the memory representation of a loaded page.
type memPage struct {
	*Allocator
	off int64
	page

	dirty bool
}

func (m *memPage) close() {
	*m = zMemPage
	memPagePool.Put(m)
}

// flush stores/persists m if dirty.
func (m *memPage) flush() error {
	if !m.dirty {
		return nil
	}

	b := m.pageBuf[:]
	write(b[oPageBrk:], m.brk)
	write(b[oPageNext:], m.next)
	write(b[oPagePrev:], m.prev)
	write(b[oPageRank:], m.rank)
	write(b[oPageSize:], m.size)
	write(b[oPageUsed:], m.usedSlots)
	_, err := m.f.WriteAt(b, m.off)
	m.dirty = err == nil
	return err
}

func (m *memPage) freeSlots() error {
	if m.usedSlots != 0 {
		return fmt.Errorf("internal error: %T.freeSlots: m.used %v", m, m.usedSlots)
	}

	for i := 0; i < int(m.brk); i++ {
		n, err := m.openNode(m.slot(i))
		if err != nil {
			return err
		}

		if err := n.unlink(int(m.rank)); err != nil {
			return err
		}

		if err := n.flush(); err != nil {
			return err
		}

		n.close()
	}
	return nil
}

func (m *memPage) setBrk(n int64)  { m.brk = n; m.dirty = true }
func (m *memPage) setNext(n int64) { m.next = n; m.dirty = true }
func (m *memPage) setPrev(n int64) { m.prev = n; m.dirty = true }
func (m *memPage) setRank(n int64) { m.rank = n; m.dirty = true }
func (m *memPage) setSize(n int64) { m.size = n; m.dirty = true }

func (m *memPage) getTail() (int64, error) {
	b := m.buf8[:]
	if n, err := m.f.ReadAt(b, m.off+m.size-szInt64); n != len(b) {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return -1, fmt.Errorf("%T.getTail: %v", m, err)
	}

	r := read(b)
	return r, nil
}

func (m *memPage) setTail(n int64) error {
	b := m.buf8[:]
	write(b, n)
	_, err := m.f.WriteAt(b, m.off+m.size-szInt64)
	return err
}

func (m *memPage) setUsed(n int64)  { m.usedSlots = n; m.dirty = true }
func (m *memPage) slot(i int) int64 { return m.off + szPage + int64(i)<<uint(m.rank+4) }

func (m *memPage) split(need int64) (int64, error) {
	if m.rank <= maxSharedRank {
		return -1, fmt.Errorf("internal error: %T.split: m.rank %v", m, m.rank)
	}

	have := m.size
	m.setSize(need)
	m.setRank(int64(pageRank(m.size)))
	if err := m.flush(); err != nil {
		return -1, err
	}

	if err := m.setTail(0); err != nil {
		return -1, err
	}

	n := m.newMemPage(m.off + m.size)
	n.setSize(have - need)
	n.setRank(int64(pageRank(have - need)))
	m.npages++
	if err := m.insertPage(n); err != nil {
		return -1, err
	}

	if err := n.flush(); err != nil {
		return -1, err
	}

	if err := n.setTail(n.size); err != nil {
		return -1, err
	}

	r := m.off + szPage
	n.close()
	return r, m.Allocator.flush(m.autoflush)
}

func (m *memPage) unlink() error {
	if m.prev != 0 {
		prev, err := m.openPage(m.prev)
		if err != nil {
			return err
		}

		prev.setNext(m.next)
		if err := prev.flush(); err != nil {
			return err
		}

		prev.close()
	}

	if m.next != 0 {
		next, err := m.openPage(m.next)
		if err != nil {
			return err
		}

		next.setPrev(m.prev)
		if err := next.flush(); err != nil {
			return err
		}

		next.close()
	}

	if m.pages[m.rank] == m.off {
		m.setPage(int(m.rank), m.next)
	}

	m.setPrev(0)
	m.setNext(0)
	return nil
}

// File is an os.File-like entity.
//
// Note: *os.File implements File.
type File interface {
	Close() error
	ReadAt(p []byte, off int64) (n int, err error)
	Stat() (os.FileInfo, error)
	Sync() error
	Truncate(int64) error
	WriteAt(p []byte, off int64) (n int, err error)
}

// Mem returns a volatile File backed only by process memory or an error, if
// any. The Close method of the result must be eventually called to avoid
// resource leaks.
func Mem(name string) (File, error) { return ifile.OpenMem(name) }

// Map returns a File backed by memory mapping f or an error, if any. The Close
// method of the result must be eventually called to avoid resource leaks.
//
// This function returns (f, nil) on Windows because there some File methods,
// like Truncate, produce errors in the current implementation:
//
//	The requested operation cannot be performed on a file with a user-mapped section open.
//
// Windows expert needed to fix this.
func Map(f *os.File) (File, error) {
	if runtime.GOOS == "windows" {
		return f, nil
	}

	return ifile.Open(f)
}

type file struct {
	_ [16]byte // User area. Magic file number etc.

	// Persistent part.
	skip  [0]byte
	pages [ranks]int64
	slots [slotRanks]int64
}

type testStat struct {
	allocs int64
	bytes  int64
	npages int64
}

// Allocator manages allocation of file blocks within a File.
//
// Allocator methods are not safe for concurrent use by multiple goroutines.
// Callers must provide their own synchronization when it's used concurrently
// by multiple goroutines.
type Allocator struct {
	buf  [szFile - oFileSkip]byte
	buf8 [8]byte
	cap  [slotRanks]int
	f    File
	file
	fsize   int64
	nodeBuf [szNode]byte
	pageBuf [szPage]byte
	testStat

	autoflush bool
	dirty     bool
}

// NewAllocator returns a newly created Allocator managing f or an eror, if
// any. Allocator never touches the first 16 bytes within f.
func NewAllocator(f File) (*Allocator, error) {
	a := allocatorPool.Get().(*Allocator)
	a.autoflush = true
	for i := range a.cap {
		a.cap[i] = int(pageAvail) / (1 << uint(i+4))
	}
	if err := a.SetFile(f); err != nil {
		return nil, err
	}

	return a, nil
}

// SetFile sets the allocator's backing File.
func (a *Allocator) SetFile(f File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	fsize := fi.Size()
	switch {
	case fsize <= oFileSkip:
		for i := range a.buf {
			a.buf[i] = 0
		}
		a.file = file{}
		if _, err := f.WriteAt(a.buf[:], oFileSkip); err != nil {
			return err
		}

		fsize = oFileSkip + int64(len(a.buf))
	default:
		if n, err := f.ReadAt(a.buf[:], oFileSkip); n != len(a.buf) {
			if err == nil {
				err = fmt.Errorf("short read")
			}
			return err
		}

		max := fsize - pageSize
		if fsize == szFile {
			max = 0
		}
		for i := range a.pages {
			if a.pages[i], err = a.check(read(a.buf[int(oFilePages-oFileSkip)+8*i:]), 0, max); err != nil {
				return err
			}
		}
		for i := range a.slots {
			max := fsize - 16<<uint(i)
			if fsize == szFile {
				max = 0
			}
			if a.slots[i], err = a.check(read(a.buf[int(oFileSlots-oFileSkip)+8*i:]), 0, max); err != nil {
				return err
			}
		}
	}
	a.f = f
	a.fsize = fsize
	a.dirty = false
	return nil
}

// SetAutoFlush turns on/off automatic flushing of allocator's metadata. When
// the argument is true Flush is called automatically whenever the metadata are
// chaneged. When the argument is false, Flush is called automatically only on
// Close.
//
// The default allocator state has auto flushing turned on.
func (a *Allocator) SetAutoFlush(v bool) { a.autoflush = v }

// Alloc allocates a file block large enough for storing size bytes and returns
// its offset or an error, if any.
func (a *Allocator) Alloc(size int64) (int64, error) {
	if size <= 0 {
		return -1, fmt.Errorf("invalid argument: %T.Alloc(%v)", a, size)
	}

	a.allocs++
	if size > maxSlot {
		return a.allocBig(size)
	}

	rank := slotRank(int(size))
	if off := a.pages[rank]; off != 0 {
		return a.sbrk(off, rank)
	}

	if off := a.pages[firstPageRank]; off != 0 {
		return a.sbrk2(off, rank)
	}

	if off := a.slots[rank]; off != 0 {
		return a.allocSlot(off, rank)
	}

	p, err := a.newSharedPage(rank)
	if err != nil {
		return -1, err
	}

	if err := a.insertPage(p); err != nil {
		return -1, err
	}

	p.setUsed(1)
	p.setBrk(1)
	if err := p.flush(); err != nil {
		return -1, err
	}

	r := p.slot(0)
	p.close()
	return r, a.flush(a.autoflush)
}

// Calloc is like Alloc but the allocated file block is zeroed up to size.
func (a *Allocator) Calloc(size int64) (int64, error) {
	if size <= 0 {
		return -1, fmt.Errorf("invalid argument: %T.Calloc(%v)", a, size)
	}

	off, err := a.Alloc(size)
	if err != nil {
		return -1, err
	}

	p := buffer.CGet(int(mathutil.MinInt64(bufSize, size)))
	b := *p
	dst := off
	for size != 0 {
		rq := len(b)
		if size < int64(rq) {
			rq = int(size)
		}
		if _, err := a.f.WriteAt(b[:rq], dst); err != nil {
			return -1, err
		}

		dst += int64(rq)
		size -= int64(rq)
	}

	buffer.Put(p)
	return off, nil
}

// Close flushes and closes the allocator and its underlying File.
func (a *Allocator) Close() error {
	if err := a.Flush(); err != nil {
		return err
	}

	if err := a.f.Close(); err != nil {
		return err
	}

	*a = zAllocator
	allocatorPool.Put(a)
	return nil
}

// Free recycles the allocated file block at off.
func (a *Allocator) Free(off int64) error {
	if off < szFile+szPage || !aligned(off) {
		return fmt.Errorf("invalid argument: %T.Free(%v)", a, off)
	}

	a.allocs--
	p, err := a.openPage((off-szFile)&^pageMask + szFile)
	if err != nil {
		return err
	}

	if p.rank > maxSharedRank {
		if err := a.freePage(p); err != nil {
			return err
		}

		p.close()
		return a.flush(a.autoflush)
	}

	p.setUsed(p.usedSlots - 1)
	if err := a.insertSlot(int(p.rank), off); err != nil {
		return err
	}

	if p.usedSlots == 0 {
		if err := a.freePage(p); err != nil {
			return err
		}

		p.close()
		return a.flush(a.autoflush)
	}

	if err := p.flush(); err != nil {
		return err
	}

	p.close()
	return a.flush(a.autoflush)
}

// Realloc changes the size of the file block allocated at off, which must have
// been returned from Alloc or Realloc, to size and returns the offset of the
// relocated file block or an error, if any. The contents will be unchanged in
// the range from the start of the region up to the minimum of the old and new
// sizes. Realloc(off, 0) is equal to Free(off). If the file block was moved, a
// Free(off) is done.
func (a *Allocator) Realloc(off, size int64) (int64, error) {
	if off < szFile+szPage || !aligned(off) {
		return -1, fmt.Errorf("invalid argument: %T.Realloc(%v)", a, off)
	}

	if size == 0 {
		return -1, a.Free(off)
	}

	oldSize, p, err := a.usableSize(off)
	if err != nil {
		return -1, err
	}

	if oldSize >= size {
		newRank := rank(size)
		if int(p.rank) == newRank {
			p.close()
			return off, nil
		}

		if newRank > maxSharedRank {
			if need := roundup64(szPage+size+szInt64, pageSize); p.size > need {
				off, err := p.split(need)
				p.close()
				return off, err
			}
		}
	}

	newOff, err := a.Alloc(size)
	if err != nil {
		return -1, err
	}

	rem := mathutil.MinInt64(oldSize, size)
	q := buffer.Get(int(mathutil.MinInt64(bufSize, rem)))
	b := *q
	src := off
	dst := newOff
	for rem != 0 {
		n, err := a.f.ReadAt(b, src)
		if n == 0 {
			if err == nil {
				err = fmt.Errorf("short read")
			}
			return -1, err
		}

		if _, err := a.f.WriteAt(b[:n], dst); err != nil {
			return -1, err
		}

		src += int64(n)
		dst += int64(n)
		rem -= int64(n)
	}
	buffer.Put(q)
	return newOff, a.Free(off)
}

// UsableSize reports the size of the file block allocated at off, which must
// have been returned from Alloc or Realloc.  The allocated file block size can
// be larger than the size originally requested from Alloc or Realloc.
func (a *Allocator) UsableSize(off int64) (int64, error) {
	if off < szFile+szPage || !aligned(off) {
		return -1, fmt.Errorf("invalid argument: %T.UsableSize(%v)", a, off)
	}

	n, p, err := a.usableSize(off)
	p.close()
	return n, err
}

func (a *Allocator) allocBig(size int64) (int64, error) {
	need := roundup64(szPage+size+szInt64, pageSize)
	rank := pageRank(need)
	for i := rank; i < len(a.pages); i++ {
		off := a.pages[i]
		if off == 0 {
			continue
		}

		if i < ranks-1 {
			return a.allocBig2(off)
		}

		for j := 0; off != 0 && j < 2; j++ {
			p, err := a.openPage(off)
			if err != nil {
				return -1, err
			}

			if p.size >= need {
				r, err := a.allocMaxRank(p, need)
				p.close()
				return r, err
			}

			off = p.next
			p.close()
		}
	}

	p, err := a.newPage(size)
	if err != nil {
		return -1, err
	}

	if err := p.flush(); err != nil {
		return -1, err
	}

	r := p.off + szPage
	p.close()
	return r, a.flush(a.autoflush)
}

func (a *Allocator) allocBig2(off int64) (int64, error) {
	p, err := a.openPage(off)
	if err != nil {
		return -1, err
	}

	if err := p.unlink(); err != nil {
		return -1, err
	}

	if err := p.flush(); err != nil {
		return -1, err
	}

	if err := p.setTail(0); err != nil {
		return -1, err
	}

	r := p.off + szPage
	p.close()
	return r, a.flush(a.autoflush)
}

func (a *Allocator) allocMaxRank(p *memPage, need int64) (int64, error) {
	if err := p.unlink(); err != nil {
		return -1, err
	}

	rem := p.size - need
	p.setSize(need)
	p.setRank(int64(pageRank(p.size)))
	if err := p.flush(); err != nil {
		return -1, err
	}

	if err := p.setTail(0); err != nil {
		return -1, err
	}

	if rem != 0 {
		q := a.newMemPage(p.off + p.size)
		q.setSize(rem)
		q.setRank(int64(pageRank(rem)))
		a.npages++
		if err := a.insertPage(q); err != nil {
			return -1, err
		}

		if err := q.flush(); err != nil {
			return -1, err
		}

		if err := q.setTail(q.size); err != nil {
			return -1, err
		}

		q.close()
	}

	return p.off + szPage, a.flush(a.autoflush)
}

func (a *Allocator) allocSlot(off int64, rank int) (int64, error) {
	n, err := a.openNode(off)
	if err != nil {
		return -1, err
	}

	if err := n.unlink(rank); err != nil {
		return -1, err
	}

	n.close()
	p, err := a.openPage((off-szFile)&^pageMask + szFile)
	if err != nil {
		return -1, err
	}

	p.setUsed(p.usedSlots + 1)
	if err := p.flush(); err != nil {
		return -1, err
	}

	p.close()
	return off, a.flush(a.autoflush)
}

func (a *Allocator) check(n, min, max int64) (int64, error) {
	if n < min || n > max {
		return 0, fmt.Errorf("corrupted file: %#x not in [%#x, %#x]", n, min, max)
	}

	return n, nil
}

// Flush writes the allocator metadata to its backing File.
//
// Note: Close calls Flush automatically.
func (a *Allocator) Flush() error { return a.flush(true) }

func (a *Allocator) flush(v bool) error {
	if !v || !a.dirty {
		return nil
	}

	for i, v := range a.pages {
		write(a.buf[int(oFilePages-oFileSkip)+8*i:], v)
	}
	for i, v := range a.slots {
		write(a.buf[int(oFileSlots-oFileSkip)+8*i:], v)
	}
	_, err := a.f.WriteAt(a.buf[:], oFileSkip)
	a.dirty = err != nil
	return err
}

func (a *Allocator) freeLastPage(p *memPage) error {
	p0 := p
	for {
		if p.rank <= maxSharedRank {
			if err := p.freeSlots(); err != nil {
				return err
			}
		}
		if err := p.unlink(); err != nil {
			return err
		}

		if err := p.flush(); err != nil {
			return err
		}

		if err := a.f.Truncate(p.off); err != nil {
			return err
		}

		a.fsize = p.off
		a.npages--
		a.bytes -= p.size
		if p.off > szFile {
			prevSize, err := a.read(p.off - szInt64)
			if err != nil {
				return err
			}

			if prevSize != 0 {
				q, err := a.openPage(p.off - prevSize)
				if err != nil {
					return err
				}

				if p != p0 {
					p.close()
				}
				p = q
				continue
			}
		}
		if p != p0 {
			p.close()
		}
		return nil
	}
}

func (a *Allocator) freePage(p *memPage) error {
	if p.usedSlots != 0 {
		return fmt.Errorf("internal error: %T.freePage: p.used %v", a, p.usedSlots)
	}

	if p.off+p.size == a.fsize {
		return a.freeLastPage(p)
	}

	if p.rank <= maxSharedRank {
		if err := p.freeSlots(); err != nil {
			return err
		}

		if err := p.unlink(); err != nil {
			return err
		}

		p.setBrk(0)
		p.setRank(firstPageRank)
	}
	if err := a.insertPage(p); err != nil {
		return err
	}

	if err := p.flush(); err != nil {
		return err
	}

	return p.setTail(p.size)
}

func (a *Allocator) insertPage(p *memPage) error {
	if p.prev != 0 || p.next != 0 {
		panic(fmt.Errorf("internal error: %T insertPage: p.prev %#x, p.next %#x", a, p.prev, p.next))
	}

	p.setNext(a.pages[p.rank])
	if p.next != 0 {
		next, err := a.openPage(p.next)
		if err != nil {
			return err
		}

		next.setPrev(p.off)
		if err := next.flush(); err != nil {
			return err
		}

		next.close()
	}
	a.setPage(int(p.rank), p.off)
	return nil
}

func (a *Allocator) insertSlot(rank int, off int64) error {
	m := memNode{Allocator: a, off: off}
	m.setNext(a.slots[rank])
	if m.next != 0 {
		next, err := a.openNode(m.next)
		if err != nil {
			return err
		}

		next.setPrev(off)
		if err := next.flush(); err != nil {
			return err
		}

		next.close()
	}
	a.setSlot(rank, off)
	return m.flush()
}

func (a *Allocator) newMemPage(off int64) *memPage {
	m := memPagePool.Get().(*memPage)
	m.Allocator = a
	m.off = off
	return m
}

func (a *Allocator) newPage(size int64) (*memPage, error) {
	off := roundup64(a.fsize-szFile, pageSize) + szFile
	size = roundup64(szPage+size+szInt64, pageSize)
	p := a.newMemPage(off)
	p.setRank(int64(pageRank(size)))
	p.setSize(size)
	a.bytes += size
	a.fsize = off + size
	a.npages++
	return p, p.setTail(0)
}

func (a *Allocator) newSharedPage(rank int) (*memPage, error) {
	off := roundup64(a.fsize-szFile, pageSize) + szFile
	p := a.newMemPage(off)
	p.setRank(int64(rank))
	p.setSize(pageSize)
	a.bytes += pageSize
	a.fsize = off + pageSize
	a.npages++
	return p, p.setTail(0)
}

// openNode returns a memNode from a node at File offset off.
func (a *Allocator) openNode(off int64) (*memNode, error) {
	b := a.nodeBuf[:]
	if n, err := a.f.ReadAt(b, off); n != len(b) {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return nil, err
	}

	m := memNodePool.Get().(*memNode)
	m.Allocator = a
	m.off = off
	m.node = node{
		next: read(b[oNodeNext:]),
		prev: read(b[oNodePrev:]),
	}
	return m, nil
}

func (a *Allocator) openPage(off int64) (*memPage, error) {
	b := a.pageBuf[:]
	if n, err := a.f.ReadAt(b, off); n != len(b) {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return nil, err
	}

	m := a.newMemPage(off)
	m.page = page{
		brk: read(b[oPageBrk:]),
		node: node{
			next: read(b[oPageNext:]),
			prev: read(b[oPagePrev:]),
		},
		rank:      read(b[oPageRank:]),
		size:      read(b[oPageSize:]),
		usedSlots: read(b[oPageUsed:]),
	}
	return m, nil
}

func (a *Allocator) read(off int64) (int64, error) {
	b := a.buf8[:]
	if n, err := a.f.ReadAt(b, off); n != len(b) {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return -1, err
	}

	n := read(b)
	return n, nil
}

func (a *Allocator) sbrk(off int64, rank int) (int64, error) {
	p, err := a.openPage(off)
	if err != nil {
		return -1, err
	}

	if int64(rank) != p.rank {
		panic(fmt.Errorf("internal error: %T.sbrk: rank %v, p.rank %v", a, rank, p.rank))
	}

	p.setUsed(p.usedSlots + 1)
	p.setBrk(p.brk + 1)
	if int(p.brk) == a.cap[rank] {
		if err := p.unlink(); err != nil {
			return -1, err
		}
	}
	if err := p.flush(); err != nil {
		return -1, err
	}

	r := p.slot(int(p.brk) - 1)
	p.close()
	return r, a.flush(a.autoflush)
}

func (a *Allocator) sbrk2(off int64, rank int) (int64, error) {
	p, err := a.openPage(off)
	if err != nil {
		return -1, err
	}

	if err := p.unlink(); err != nil {
		return -1, err
	}

	p.setRank(int64(rank))
	p.setUsed(1)
	p.setBrk(1)
	if err := a.insertPage(p); err != nil {
		return -1, err
	}

	if err := p.flush(); err != nil {
		return -1, err
	}

	if err := p.setTail(0); err != nil {
		return -1, err
	}

	r := p.off + szPage
	p.close()
	return r, a.flush(a.autoflush)
}

func (a *Allocator) setPage(rank int, n int64) { a.pages[rank] = n; a.dirty = true }
func (a *Allocator) setSlot(rank int, n int64) { a.slots[rank] = n; a.dirty = true }

func (a *Allocator) usableSize(off int64) (int64, *memPage, error) {
	if off < szFile+szPage {
		return -1, nil, fmt.Errorf("invalid argument: %T.UsableSize(%v)", a, off)
	}

	p, err := a.openPage((off-szFile)&^pageMask + szFile)
	if err != nil {
		return -1, nil, err
	}

	if p.rank < 0 || p.rank >= ranks {
		panic(fmt.Errorf("internal error: %T.UsableSize: p.rank %v", a, p.rank))
	}

	if p.rank <= maxSharedRank {
		return int64(1 << uint(p.rank+4)), p, nil
	}

	return p.size - szPage - szInt64, p, nil
}

type bitmap struct {
	fn  string
	m   File
	buf []byte
}

func newBitmap() (*bitmap, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}

	fn := f.Name()
	m, err := Map(f)
	if err != nil {
		f.Close()
		os.Remove(fn)
		return nil, err
	}

	return &bitmap{
		buf: make([]byte, 1),
		fn:  fn,
		m:   m,
	}, nil
}

func (m *bitmap) close() error {
	if m.m == nil {
		return nil
	}

	r := m.m.Close()
	os.Remove(m.fn)
	m.m = nil
	return r
}

func (m *bitmap) set(bit int64) (bool, error) {
	if bit&15 != 0 {
		return false, fmt.Errorf("%T.set(%#x): != 0 (mod 16)", m, bit)
	}

	bit >>= 4
	off := bit >> 3
	mask := byte(1) << byte(bit&7)
	m.buf[0] = 0
	if _, err := m.m.ReadAt(m.buf, off); err != nil {
		if err != io.EOF {
			return false, fmt.Errorf("%T.read(%#x): %v", m, off, err)
		}
	}

	r := m.buf[0]&mask != 0
	m.buf[0] |= mask
	if _, err := m.m.WriteAt(m.buf, off); err != nil {
		return false, fmt.Errorf("%T.write(%#x): %v", m, off, err)
	}

	return r, nil
}

// VerifyOptions optionally provide more information from Verify.
type VerifyOptions struct {
	Allocs    int64 // Number of allocations in use.
	Pages     int64 // Number of pages.
	UsedPages int64 // Number of pages in use.
}

// Verify audits the correctness of the allocator and its backing File.
func (a *Allocator) Verify(opt *VerifyOptions) error {
	// Ensure disk and allocator are in sync.
	if err := a.Flush(); err != nil {
		return fmt.Errorf("cannot flush: %v", err)
	}

	fi, err := a.f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat: %v", err)
	}

	if g, e := fi.Size(), szFile; g < e {
		return fmt.Errorf("invalid file size, got %#x, expected at least %#x", g, e)
	}

	if g, e := fi.Size(), a.fsize; g != e {
		return fmt.Errorf("invalid file size, got %#x, expected %#x", g, e)
	}

	if g, e := a.cap, [...]int{252, 126, 63, 31, 15, 7, 3}; g != e {
		return fmt.Errorf("invalid page capacities, got %v, expected %v", g, e)
	}

	// Check file header.
	pm := map[int64]struct{}{}
	sm := map[int64]struct{}{}
	buf := make([]byte, int(szFile-oFileSkip))
	if n, err := a.f.ReadAt(buf, oFileSkip); n != len(a.buf) {
		if err == nil {
			err = fmt.Errorf("short read")
		}
		return fmt.Errorf("cannot read file header: %v", err)
	}

	max := a.fsize - pageSize
	if a.fsize == szFile {
		max = 0
	}
	for i, e := range a.pages {
		g, err := a.check(read(buf[int(oFilePages-oFileSkip)+8*i:]), 0, max)
		if err != nil {
			return fmt.Errorf("cannot read pages list[%v]: %v", i, err)
		}
		if g != e {
			return fmt.Errorf("invalid pages list[%v], got %v, expected %v", i, g, e)
		}

		if g == 0 {
			continue
		}

		if (g-szFile)&((1<<pageBits)-1) != 0 {
			return fmt.Errorf("invalid pages lists[%v] head: %#x", i, g)
		}

		if _, ok := pm[g]; ok && g != 0 {
			return fmt.Errorf("pages list[%v] reused: %#x", i, g)
		}

		pm[g] = struct{}{}
	}
	for i, e := range a.slots {
		max := a.fsize - 16<<uint(i)
		if a.fsize == szFile {
			max = 0
		}
		g, err := a.check(read(buf[int(oFileSlots-oFileSkip)+8*i:]), 0, max)
		if err != nil {
			return fmt.Errorf("cannot read slots list[%v]: %v", i, err)
		}

		if g != e {
			return fmt.Errorf("invalid slots list[%v], got %v, expected %v", i, g, e)
		}

		if g == 0 {
			continue
		}

		if _, ok := sm[g]; ok && g != 0 {
			return fmt.Errorf("slots list[%v] reused: %#x", i, g)
		}

		sm[g] = struct{}{}
	}

	// Check pages.
	var npages, usedPages, allocs int64
	off := szFile
	for off < a.fsize {
		if !aligned(off) {
			return fmt.Errorf("unaligned offset")
		}

		off2 := off - szFile
		if off2&((1<<pageBits)-1) != 0 {
			return fmt.Errorf("invalid page boundary %#x", off)
		}

		p, err := a.openPage(off)
		if err != nil {
			return fmt.Errorf("cannot read page at %#x: %v", off, err)
		}

		tailSize, err := p.getTail()
		if err != nil {
			return fmt.Errorf("cannot read tail of page %#x: %v", off, err)
		}

		if tailSize == 0 { // page in use
			usedPages++
			switch {
			case p.rank <= maxSharedRank:
				allocs += p.usedSlots
			default:
				allocs++
			}

		}

		delete(pm, off)
		npages++
		off += p.size
		p.close()
	}
	if len(pm) != 0 {
		return fmt.Errorf("invalid pages lists heads: %#x", pm)
	}

	if g, e := off, a.fsize; g != e {
		return fmt.Errorf("last file page does not end at file end, got %v, expected %v", g, e)
	}

	// Check page lists linking.
	bits, err := newBitmap()
	if err != nil {
		return fmt.Errorf("cannot create bitmap: %v", err)
	}

	defer bits.close()

	for i, off := range a.pages {
		// Walk the single list.
		for off != 0 {
			if !aligned(off) {
				return fmt.Errorf("unaligned offset")
			}

			v, err := bits.set(off)
			if err != nil {
				return fmt.Errorf("%v: bitmap.set(%#x): %v", i, off, err)
			}

			if v {
				return fmt.Errorf("%v: page listed multiple times %#x", i, off)
			}

			p, err := a.openPage(off)
			if err != nil {
				return fmt.Errorf("cannot read page at %#x: %v", off, err)
			}

			off = p.next
			p.close()
		}
	}

	// Check slots lists linking.
	for i, off := range a.slots {
		// Walk the single list.
		for off != 0 {
			if !aligned(off) {
				return fmt.Errorf("unaligned offset")
			}

			v, err := bits.set(off)
			if err != nil {
				return fmt.Errorf("%v: bitmap.set(%#x): %v", i, off, err)
			}

			if v {
				return fmt.Errorf("%v: slot listed multiple times %#x", i, off)
			}

			p, err := a.openNode(off)
			if err != nil {
				return fmt.Errorf("cannot read node at %#x: %v", off, err)
			}

			off = p.next
			p.close()
		}
	}

	if opt != nil {
		opt.Allocs = allocs
		opt.Pages = npages
		opt.UsedPages = usedPages
	}
	return nil
}
