// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"math"
	"unsafe"

	"modernc.org/ccir/libc/errno"
)

const (
	longBits = 32
)

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func fread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint32, stream *unsafe.Pointer) uint32 {
	req := uint64(size) * uint64(nmemb)
	if req > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.reader(unsafe.Pointer(stream)).Read((*[math.MaxInt32]byte)(ptr)[:req])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint32(n) / size
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func fwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint32, stream *unsafe.Pointer) uint32 {
	req := uint64(nmemb) * uint64(size)
	if req > math.MaxInt32 {
		tls.setErrno(errno.XE2BIG)
		return 0
	}

	n, err := files.writer(unsafe.Pointer(stream)).Write((*[math.MaxInt32]byte)(ptr)[:req])
	if err != nil {
		tls.setErrno(errno.XEIO)
	}
	return uint32(n) / size
}

// size_t fwrite(const void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfwrite(tls *TLS, ptr unsafe.Pointer, size, nmemb uint32, stream *unsafe.Pointer) uint32 {
	return fwrite(tls, ptr, size, nmemb, stream)
}

// size_t fread(void *ptr, size_t size, size_t nmemb, FILE *stream);
func Xfread(tls *TLS, ptr unsafe.Pointer, size, nmemb uint32, stream *unsafe.Pointer) uint32 {
	return fread(tls, ptr, size, nmemb, stream)
}

// int fseek(FILE *stream, long offset, int whence);
func Xfseek(tls *TLS, stream *unsafe.Pointer, offset, whence int32) int32 {
	return fseek(tls, stream, int64(offset), whence)
}

// long ftell(FILE *stream);
func Xftell(tls *TLS, stream *unsafe.Pointer) int32 { return int32(ftell(tls, stream)) }
