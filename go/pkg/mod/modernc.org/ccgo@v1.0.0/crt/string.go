// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// // void *memmove(void *dest, const void *src, size_t n);
// func (c *cpu) memmove() {
// 	sp, n := popLong(c.sp)
// 	sp, src := popPtr(sp)
// 	dest := readPtr(sp)
// 	movemem(dest, src, int(n))
// 	writePtr(c.rp, dest)
// }
//
// // void *mempcpy(void *dest, const void *src, size_t n);
// func (c *cpu) mempcpy() {
// 	sp, n := popLong(c.sp)
// 	sp, src := popPtr(sp)
// 	dest := readPtr(sp)
// 	movemem(dest, src, int(n))
// 	writePtr(c.rp, dest+uintptr(n))
// }

// char *strcat(char *dest, const char *src)
func Xstrcat(tls *TLS, dest, src *int8) *int8 {
	ret := dest
	for *dest != 0 {
		*(*uintptr)(unsafe.Pointer(&dest))++
	}
	for {
		c := *src
		*(*uintptr)(unsafe.Pointer(&src))++
		*dest = c
		*(*uintptr)(unsafe.Pointer(&dest))++
		if c == 0 {
			return ret
		}
	}
}

// char *index(const char *s, int c)
func Xindex(tls *TLS, s *int8, c int32) *int8 { return Xstrchr(tls, s, c) }

// char *strchr(const char *s, int c)
func Xstrchr(tls *TLS, s *int8, c int32) *int8 {
	for {
		ch2 := byte(*s)
		if ch2 == byte(c) {
			return s
		}

		if ch2 == 0 {
			return nil
		}

		*(*uintptr)(unsafe.Pointer(&s))++
	}
}

// int strcmp(const char *s1, const char *s2)
func X__builtin_strcmp(tls *TLS, s1, s2 *int8) int32 {
	for {
		ch1 := byte(*s1)
		*(*uintptr)(unsafe.Pointer(&s1))++
		ch2 := byte(*s2)
		*(*uintptr)(unsafe.Pointer(&s2))++
		if ch1 != ch2 || ch1 == 0 || ch2 == 0 {
			return int32(ch1) - int32(ch2)
		}
	}
}

// int strcmp(const char *s1, const char *s2)
func Xstrcmp(tls *TLS, s1, s2 *int8) int32 { return X__builtin_strcmp(tls, s1, s2) }

// char *strcpy(char *dest, const char *src)
func X__builtin_strcpy(tls *TLS, dest, src *int8) *int8 {
	r := dest
	for {
		c := *src
		*(*uintptr)(unsafe.Pointer(&src))++
		*dest = c
		*(*uintptr)(unsafe.Pointer(&dest))++
		if c == 0 {
			return r
		}
	}
}

// char *strcpy(char *dest, const char *src)
func Xstrcpy(tls *TLS, dest, src *int8) *int8 { return X__builtin_strcpy(tls, dest, src) }

// // char *strncpy(char *dest, const char *src, size_t n)
// func (c *cpu) strncpy() {
// 	sp, n := popLong(c.sp)
// 	sp, src := popPtr(sp)
// 	dest := readPtr(sp)
// 	ret := dest
// 	var ch int8
// 	for ch = readI8(src); ch != 0 && n > 0; n-- {
// 		writeI8(dest, ch)
// 		dest++
// 		src++
// 		ch = readI8(src)
// 	}
// 	for ; n > 0; n-- {
// 		writeI8(dest, 0)
// 		dest++
// 	}
// 	writePtr(c.rp, ret)
// }

// char *rindex(const char *s, int c)
func Xrindex(tls *TLS, s *int8, c int32) *int8 { return Xstrrchr(tls, s, c) }

// char *strrchr(const char *s, int c)
func Xstrrchr(tls *TLS, s *int8, c int32) *int8 {
	var ret *int8
	for {
		ch2 := byte(*s)
		if ch2 == 0 {
			return ret
		}

		if ch2 == byte(c) {
			ret = s
		}
		*(*uintptr)(unsafe.Pointer(&s))++
	}
}
