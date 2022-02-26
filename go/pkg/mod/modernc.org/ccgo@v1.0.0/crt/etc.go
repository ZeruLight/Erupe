// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crt provides C-runtime services. (Work In Progress)
package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"modernc.org/ccir/libc/errno"
	"modernc.org/internal/buffer"
	"modernc.org/mathutil"
	"modernc.org/memory"
)

const (
	ptrSize   = mathutil.UintPtrBits / 8
	heapAlign = 2 * ptrSize
)

var (
	alloc         memory.Allocator
	allocMu       sync.Mutex
	brk           unsafe.Pointer
	heapAvailable int64
	ptr2u         = map[*uintptr]struct{}{}
	ptrMu         sync.Mutex
	threadID      uintptr
	u2ptr         = map[uintptr]*uintptr{}
)

func writeU8(p uintptr, v uint8) { *(*uint8)(unsafe.Pointer(p)) = v }

// TLS represents the C-thread local storage.
type TLS struct {
	threadID uintptr
	errno    int32
}

// NewTLS returns a newly created TLS.
func NewTLS() *TLS { return &TLS{threadID: atomic.AddUintptr(&threadID, 1)} }

func (t *TLS) setErrno(err interface{}) {
	switch x := err.(type) {
	case int:
		t.errno = int32(x)
	case *os.PathError:
		t.setErrno(x.Err)
	case syscall.Errno:
		t.errno = int32(x)
	default:
		panic(fmt.Errorf("TODO %T(%#v)", x, x))
	}
}

// TLS frees any resources used by t.
func (t *TLS) Close() {
	// nop
}

//TODO remove me.
func TODO(msg string, more ...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "%s:%d: %v\n", path.Base(fn), fl, fmt.Sprintf(msg, more...))
	os.Stderr.Sync()
	panic(fmt.Errorf("%s:%d: TODO %v", path.Base(fn), fl, fmt.Sprintf(msg, more...))) //TODOOK
}

type memWriter uintptr

func (m *memWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	*m += memWriter(movemem(unsafe.Pointer(*m), unsafe.Pointer(&b[0]), len(b)))
	return len(b), nil
}

func (m *memWriter) WriteByte(b byte) error {
	*(*byte)(unsafe.Pointer(*m)) = b
	*m++
	return nil
}

// Movemem copies n bytes from memory area src to memory area dest.  The memory
// areas may overlap: copying takes place as though the bytes in src are first
// copied into a temporary array that does not overlap src or dest, and the
// bytes are then copied from the temporary array to dest.
func Movemem(dest, src unsafe.Pointer, n int) int { return movemem(dest, src, n) }

func movemem(dst, src unsafe.Pointer, n int) int {
	return copy((*[math.MaxInt32]byte)(dst)[:n], (*[math.MaxInt32]byte)(src)[:n])
}

// GoString returns a string from a C char* null terminated string s.
func GoString(s *int8) string {
	if s == nil {
		return ""
	}

	var b buffer.Bytes
	for {
		ch := *s
		if ch == 0 {
			r := string(b.Bytes())
			b.Close()
			return r
		}

		b.WriteByte(byte(ch))
		*(*uintptr)(unsafe.Pointer(&s))++
	}
}

// GoStringLen returns a string from a C char* string s having length len bytes.
func GoStringLen(s *int8, len int) string {
	var b buffer.Bytes
	for ; len != 0; len-- {
		b.WriteByte(byte(*s))
		*(*uintptr)(unsafe.Pointer(&s))++
	}
	r := string(b.Bytes())
	b.Close()
	return r
}

// RegisterHeap registers the sbrk-style heap.
func RegisterHeap(h unsafe.Pointer, n int64) {
	brk = h
	heapAvailable = n
}

// if n%m != 0 { n += m-n%m }. m must be a power of 2.
func roundupI64(n, m int64) int64 { return (n + m - 1) &^ (m - 1) }

// CString allocates a C string initialized from s.
func CString(s string) unsafe.Pointer {
	n := len(s)
	var tls TLS
	p := malloc(&tls, n+1)
	if p == nil {
		return nil
	}

	copy((*[math.MaxInt32]byte)(p)[:n], s)
	(*[math.MaxInt32]byte)(p)[n] = 0
	return p
}

// Malloc allocates size bytes and returns a byte slice of the allocated
// memory. The memory is not initialized. Malloc panics for size < 0 and
// returns (nil, nil) for zero size. Malloc is safe for concurrent use by
// multiple goroutines.
func Malloc(size int) (unsafe.Pointer, error) {
	allocMu.Lock()
	b, err := alloc.Malloc(size)
	allocMu.Unlock()
	if err != nil {
		return nil, err
	}

	return unsafe.Pointer(&b[0]), nil
}

func malloc(tls *TLS, size int) unsafe.Pointer {
	allocMu.Lock()
	b, err := alloc.Malloc(size)
	allocMu.Unlock()
	if err != nil {
		tls.setErrno(errno.XENOMEM)
		return nil
	}

	return unsafe.Pointer(&b[0])
}

// Calloc is like Malloc except the allocated memory is zeroed. Calloc is safe
// for concurrent use by multiple goroutines.
func Calloc(size int) (unsafe.Pointer, error) {
	allocMu.Lock()
	b, err := alloc.Calloc(size)
	allocMu.Unlock()
	if err != nil {
		return nil, err
	}

	return unsafe.Pointer(&b[0]), nil
}

func calloc(tls *TLS, size int) unsafe.Pointer {
	allocMu.Lock()
	b, err := alloc.Calloc(size)
	allocMu.Unlock()
	if err != nil {
		tls.setErrno(errno.XENOMEM)
		return nil
	}

	return unsafe.Pointer(&b[0])
}

// Realloc changes the size of the memory allocated at ptr to size bytes or
// returns an error, if any.  The contents will be unchanged in the range from
// the start of the region up to the minimum of the old and new  sizes.   If
// the new size is larger than the old size, the added memory will not be
// initialized.  If ptr is nil, then the call is equivalent to Malloc(size),
// for all values of size; if size is equal to zero, and ptr is not nil, then
// the call is equivalent to Free(ptr).  Unless ptr is nil, it must have been
// returned by an earlier call to Malloc, Calloc or Realloc.  If the area
// pointed to was moved, a Free(ptr) is done. Relloc is safe for concurrent use
// by multiple goroutines.
func Realloc(tls *TLS, ptr unsafe.Pointer, size int) (unsafe.Pointer, error) {
	old := memory.UsableSize((*byte)(ptr))
	var b []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(ptr)
	sh.Len = old
	sh.Cap = old
	allocMu.Lock()
	r, err := alloc.Realloc(b, size)
	allocMu.Unlock()
	if err != nil {
		return nil, err
	}

	return unsafe.Pointer(&r[0]), nil
}

func realloc(tls *TLS, ptr unsafe.Pointer, size int) unsafe.Pointer {
	old := memory.UsableSize((*byte)(ptr))
	var b []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(ptr)
	sh.Len = old
	sh.Cap = old
	allocMu.Lock()
	r, err := alloc.Realloc(b, size)
	allocMu.Unlock()
	if err != nil {
		tls.setErrno(errno.XENOMEM)
		return nil
	}

	return unsafe.Pointer(&r[0])
}

// Free deallocates memory. The argument of Free must have been acquired from
// Calloc or Malloc or Realloc. Free is safe for concurrent use by multiple
// goroutines.
func Free(ptr unsafe.Pointer) error {
	var b []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(ptr)
	sh.Len = 1
	sh.Cap = 1
	allocMu.Lock()
	err := alloc.Free(b)
	allocMu.Unlock()
	return err
}

func free(ptr unsafe.Pointer) {
	var b []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh.Data = uintptr(ptr)
	sh.Len = 1
	sh.Cap = 1
	allocMu.Lock()
	alloc.Free(b)
	allocMu.Unlock()
}

// UsableSize reports the size of the memory block allocated at p, which must
// have been acquired from Calloc, Malloc or Realloc.  The allocated memory
// block size can be larger than the size originally requested from Calloc,
// Malloc or Realloc.
func UsableSize(p unsafe.Pointer) int { return memory.UsableSize((*byte)(p)) }

// CopyString copies src to dest, optionally adding a zero byte at the end.
func CopyString(dst unsafe.Pointer, src string, addNull bool) {
	copy((*[math.MaxInt32]byte)(dst)[:len(src)], src)
	if addNull {
		writeU8(uintptr(dst)+uintptr(len(src)), 0)
	}
}

// CopyBytes copies src to dest, optionally adding a zero byte at the end.
func CopyBytes(dst unsafe.Pointer, src []byte, addNull bool) {
	copy((*[math.MaxInt32]byte)(dst)[:len(src)], src)
	if addNull {
		writeU8(uintptr(dst)+uintptr(len(src)), 0)
	}
}

// GoBytesLen returns a []byte copied from a C char* string s having length len bytes.
func GoBytesLen(s *int8, len int) []byte {
	var b buffer.Bytes
	for ; len != 0; len-- {
		b.WriteByte(byte(*s))
		*(*uintptr)(unsafe.Pointer(&s))++
	}
	return b.Bytes()
}

// U2P returns unsafe.Pointer(&u).
//
// The function is safe for concurrent use by multiple goroutines.
func U2P(u uintptr) unsafe.Pointer {
	if u == 0 {
		return nil
	}

	ptrMu.Lock()
	if p := u2ptr[u]; p != nil {
		ptrMu.Unlock()
		return unsafe.Pointer(p)
	}

	allocMu.Lock()
	b, err := alloc.Malloc(ptrSize)
	allocMu.Unlock()
	if err != nil {
		panic("OOM")
	}

	p := (*uintptr)(unsafe.Pointer(&b[0]))
	*p = u
	u2ptr[u] = p
	ptr2u[p] = struct{}{}
	ptrMu.Unlock()
	return unsafe.Pointer(p)
}

// P2U returns the uintptr value represented by p. If p was obtained by U2P
// then the uintptr value originally passed to U2P is returned. Otherwise the
// return value is uintptr(p).
//
// The function is safe for concurrent use by multiple goroutines.
func P2U(p unsafe.Pointer) uintptr {
	q := (*uintptr)(p)
	ptrMu.Lock()
	if _, ok := ptr2u[q]; ok {
		ptrMu.Unlock()
		return *q
	}

	ptrMu.Unlock()
	return uintptr(p)
}
