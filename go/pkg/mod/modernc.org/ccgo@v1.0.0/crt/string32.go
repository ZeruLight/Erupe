// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// char *strncpy(char *dest, const char *src, size_t n)
func Xstrncpy(tls *TLS, dest, src *int8, n uint32) *int8 {
	ret := dest
	for c := *src; c != 0 && n > 0; n-- {
		*dest = c
		*(*uintptr)(unsafe.Pointer(&dest))++
		*(*uintptr)(unsafe.Pointer(&src))++
		c = *src
	}
	for ; n > 0; n-- {
		*dest = 0
		*(*uintptr)(unsafe.Pointer(&dest))++
	}
	return ret
}

// size_t strlen(const char *s)
func X__builtin_strlen(tls *TLS, s *int8) uint32 {
	var n uint32
	for ; *s != 0; *(*uintptr)(unsafe.Pointer(&s))++ {
		n++
	}
	return n
}

// size_t strlen(const char *s)
func Xstrlen(tls *TLS, s *int8) uint32 { return X__builtin_strlen(tls, s) }

// int strncmp(const char *s1, const char *s2, size_t n)
func Xstrncmp(tls *TLS, s1, s2 *int8, n uint32) int32 {
	var ch1, ch2 byte
	for n != 0 {
		ch1 = byte(*s1)
		*(*uintptr)(unsafe.Pointer(&s1))++
		ch2 = byte(*s2)
		*(*uintptr)(unsafe.Pointer(&s2))++
		n--
		if ch1 != ch2 || ch1 == 0 || ch2 == 0 {
			break
		}
	}
	if n != 0 {
		return int32(ch1) - int32(ch2)
	}

	return 0
}

// void *memset(void *s, int c, size_t n)
func Xmemset(tls *TLS, s unsafe.Pointer, c int32, n uint32) unsafe.Pointer {
	return X__builtin_memset(tls, s, c, n)
}

// void *memset(void *s, int c, size_t n)
func X__builtin_memset(tls *TLS, s unsafe.Pointer, c int32, n uint32) unsafe.Pointer {
	for d := (*int8)(unsafe.Pointer(s)); n > 0; n-- {
		*d = int8(c)
		*(*uintptr)(unsafe.Pointer(&d))++
	}
	return s
}

// void *memcpy(void *dest, const void *src, size_t n)
func X__builtin_memcpy(tls *TLS, dest, src unsafe.Pointer, n uint32) unsafe.Pointer {
	movemem(dest, src, int(n))
	return dest
}

// void *memcpy(void *dest, const void *src, size_t n)
func Xmemcpy(tls *TLS, dest, src unsafe.Pointer, n uint32) unsafe.Pointer {
	return X__builtin_memcpy(tls, dest, src, n)
}

// int memcmp(const void *s1, const void *s2, size_t n)
func X__builtin_memcmp(tls *TLS, s1, s2 unsafe.Pointer, n uint32) int32 {
	var ch1, ch2 byte
	for n != 0 {
		ch1 = *(*byte)(unsafe.Pointer(s1))
		*(*uintptr)(unsafe.Pointer(&s1))++
		ch2 = *(*byte)(unsafe.Pointer(s2))
		*(*uintptr)(unsafe.Pointer(&s2))++
		if ch1 != ch2 {
			break
		}

		n--
	}
	if n != 0 {
		return int32(ch1) - int32(ch2)
	}

	return 0
}

// int memcmp(const void *s1, const void *s2, size_t n)
func Xmemcmp(tls *TLS, s1, s2 unsafe.Pointer, n uint32) int32 {
	return X__builtin_memcmp(tls, s1, s2, n)
}

// void *memmove(void *dest, const void *src, size_t n);
func Xmemmove(tls *TLS, dest, src unsafe.Pointer, n uint32) unsafe.Pointer {
	movemem(dest, src, int(n))
	return dest
}

// void *mempcpy(void *dest, const void *src, size_t n);
func Xmempcpy(tls *TLS, dest, src unsafe.Pointer, n uint32) unsafe.Pointer {
	movemem(dest, src, int(n))
	return unsafe.Pointer(uintptr(dest) + uintptr(n))
}

// int strerror_r(int errnum, char *buf, size_t buflen);
func Xstrerror_r(tls *TLS, errnum int32, buf *int8, buflen uint32) int32 {
	panic("TODO")
}
