// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"sort"
	"unsafe"

	"modernc.org/internal/buffer"
	"modernc.org/mathutil"
)

// void exit(int);
func Xexit(tls *TLS, n int32) { X__builtin_exit(tls, n) }

// void exit(int);
func X__builtin_exit(tls *TLS, n int32) {
	os.Exit(int(n))
}

// void free(void *ptr);
func Xfree(tls *TLS, ptr unsafe.Pointer) {
	free(ptr)
	if strace {
		fmt.Fprintf(os.Stderr, "free(%#x)\n", ptr)
	}
}

// void abort();
func Xabort(tls *TLS) { X__builtin_abort(tls) }

// void __builtin_trap();
func X__builtin_trap(tls *TLS) { os.Exit(1) }

// void abort();
func X__builtin_abort(tls *TLS) { X__builtin_trap(tls) }

// char *getenv(const char *name);
func Xgetenv(tls *TLS, name *int8) *int8 {
	nm := GoString(name)
	v := os.Getenv(nm)
	var p unsafe.Pointer
	if v != "" {
		p = CString(v) //TODO memory leak
	}
	return (*int8)(p)
}

// int abs(int j);
func X__builtin_abs(tks *TLS, j int32) int32 {
	if j < 0 {
		return -j
	}

	return j
}

// int abs(int j);
func Xabs(tls *TLS, j int32) int32 { return X__builtin_abs(tls, j) }

type sorter struct {
	base   unsafe.Pointer
	compar func(tls *TLS, a, b unsafe.Pointer) int32
	nmemb  int
	size   uintptr
	tls    *TLS
}

func (s *sorter) Len() int { return s.nmemb }

func (s *sorter) Less(i, j int) bool {
	return s.compar(s.tls, unsafe.Pointer(uintptr(s.base)+uintptr(i)*s.size), unsafe.Pointer(uintptr(s.base)+uintptr(j)*s.size)) < 0
}

func (s *sorter) Swap(i, j int) {
	p := unsafe.Pointer(uintptr(s.base) + uintptr(i)*s.size)
	q := unsafe.Pointer(uintptr(s.base) + uintptr(j)*s.size)
	switch s.size {
	case 1:
		v := *(*byte)(p)
		*(*byte)(p) = *(*byte)(q)
		*(*byte)(q) = v
	case 2:
		v := *(*int16)(p)
		*(*int16)(p) = *(*int16)(q)
		*(*int16)(q) = v
	case 4:
		v := *(*int32)(p)
		*(*int32)(p) = *(*int32)(q)
		*(*int32)(q) = v
	case 8:
		v := *(*int64)(p)
		*(*int64)(p) = *(*int64)(q)
		*(*int64)(q) = v
	default:
		size := int(s.size)
		bp := buffer.Get(size) //TODO static alloc
		buf := *bp
		movemem(unsafe.Pointer(&buf[0]), p, size)
		movemem(p, q, size)
		movemem(q, unsafe.Pointer(&buf[0]), size)
		buffer.Put(bp)
	}
}

// void qsort(void *base, size_t nmemb, size_t size, int (*compar)(const void *, const void *));
func qsort(tls *TLS, base unsafe.Pointer, nmemb, size uint64, compar func(tls *TLS, a, b unsafe.Pointer) int32) {
	if size > mathutil.MaxInt {
		panic("size overflow")
	}

	if nmemb > mathutil.MaxInt {
		panic("nmemb overflow")
	}

	s := sorter{base, compar, int(nmemb), uintptr(size), tls}
	sort.Sort(&s)
}
