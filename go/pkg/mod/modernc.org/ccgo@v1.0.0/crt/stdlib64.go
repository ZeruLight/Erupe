// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"unsafe"

	"modernc.org/mathutil"
)

// void *calloc(size_t nmemb, size_t size);
func Xcalloc(tls *TLS, nmemb, size uint64) (p unsafe.Pointer) {
	hi, lo := mathutil.MulUint128_64(nmemb, size)
	if hi == 0 && lo <= mathutil.MaxInt {
		p = calloc(tls, int(lo))
	}
	if strace {
		fmt.Fprintf(os.Stderr, "calloc(%#x) %#x\n", size, p)
	}
	return p
}

// void *malloc(size_t size);
func X__builtin_malloc(tls *TLS, size uint64) (p unsafe.Pointer) {
	if size < mathutil.MaxInt {
		p = malloc(tls, int(size))
	}
	if strace {
		fmt.Fprintf(os.Stderr, "malloc(%#x) %#x\n", size, p)
	}
	return p
}

// void *malloc(size_t size);
func Xmalloc(tls *TLS, size uint64) unsafe.Pointer { return X__builtin_malloc(tls, size) }

// void *realloc(void *ptr, size_t size);
func Xrealloc(tls *TLS, ptr unsafe.Pointer, size uint64) unsafe.Pointer {
	return realloc(tls, ptr, int(size))
}

// void qsort(void *base, size_t nmemb, size_t size, int (*compar)(const void *, const void *));
func Xqsort(tls *TLS, base unsafe.Pointer, nmemb, size uint64, compar func(tls *TLS, a, b unsafe.Pointer) int32) {
	qsort(tls, base, nmemb, size, compar)
}
