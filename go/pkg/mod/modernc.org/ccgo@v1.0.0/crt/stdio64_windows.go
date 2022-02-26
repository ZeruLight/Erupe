// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

import (
	"math"
	"unsafe"

	"modernc.org/ccir/libc/errno"
	"modernc.org/mathutil"
)

const (
	longBits = 32
)

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func fread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *unsafe.Pointer) uint64 {
	hi, lo := mathutil.MulUint128_64(size, nmemb)
	if hi != 0 || lo > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.reader(unsafe.Pointer(stream)).Read((*[math.MaxInt32]byte)(ptr)[:lo])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint64(n) / size
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func fwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *unsafe.Pointer) uint64 {
	hi, lo := mathutil.MulUint128_64(size, nmemb)
	if hi != 0 || lo > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.writer(unsafe.Pointer(stream)).Write((*[math.MaxInt32]byte)(ptr)[:lo])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint64(n) / size
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *unsafe.Pointer) uint64 {
	return fwrite(tls, ptr, size, nmemb, stream)
}

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint64, stream *unsafe.Pointer) uint64 {
	return fread(tls, ptr, size, nmemb, stream)
}

// int fseek(FILE *stream, long offset, int whence);
func Xfseek(tls *TLS, stream *unsafe.Pointer, offset, whence int32) int32 {
	return fseek(tls, stream, int64(offset), whence)
}

// long ftell(FILE *stream);
func Xftell(tls *TLS, stream *unsafe.Pointer) int32 { return int32(ftell(tls, stream)) }
