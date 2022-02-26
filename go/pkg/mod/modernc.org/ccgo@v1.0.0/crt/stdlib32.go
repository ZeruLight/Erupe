// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"unsafe"

	"modernc.org/mathutil"
)

// void *calloc(size_t nmemb, size_t size);
func Xcalloc(tls *TLS, nmemb, size uint32) unsafe.Pointer {
	n := uint64(nmemb) * uint64(size)
	var p unsafe.Pointer
	if n <= mathutil.MaxInt {
		p = calloc(tls, int(n))
	}
	if strace {
		fmt.Fprintf(os.Stderr, "calloc(%#x) %#x\n", size, p)
	}
	return p
}

// void *malloc(size_t size);
func X__builtin_malloc(tls *TLS, size uint32) unsafe.Pointer {
	if int(size) < mathutil.MaxInt {
		return malloc(tls, int(size))
	}

	return nil
}

// void *malloc(size_t size);
func Xmalloc(tls *TLS, size uint32) unsafe.Pointer { return X__builtin_malloc(tls, size) }

// void *realloc(void *ptr, size_t size);
func Xrealloc(tls *TLS, ptr unsafe.Pointer, size uint32) unsafe.Pointer {
	return realloc(tls, ptr, int(size))
}

// void qsort(void *base, size_t nmemb, size_t size, int (*compar)(const void *, const void *));
func Xqsort(tls *TLS, base unsafe.Pointer, nmemb, size uint32, compar func(tls *TLS, a, b unsafe.Pointer) int32) {
	qsort(tls, base, uint64(nmemb), uint64(size), compar)
}
