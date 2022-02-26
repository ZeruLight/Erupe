// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64
// +build !windows

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"unsafe"

	"modernc.org/ccir/libc/unistd"
)

// void *sbrk(intptr_t increment);
func Xsbrk(tls *TLS, increment int64) unsafe.Pointer { return sbrk(tls, increment) }

// ssize_t read(int fd, void *buf, size_t count);
func Xread(tls *TLS, fd int32, buf unsafe.Pointer, count uint64) int64 { //TODO stdin
	r, _, err := syscall.Syscall(syscall.SYS_READ, uintptr(fd), uintptr(buf), uintptr(count))
	if strace {
		fmt.Fprintf(os.Stderr, "read(%v, %#x, %v) %v %v\n", fd, buf, count, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int64(r)
}

// char *getcwd(char *buf, size_t size);
func Xgetcwd(tls *TLS, buf *int8, size uint64) *int8 {
	r, _, err := syscall.Syscall(syscall.SYS_GETCWD, uintptr(unsafe.Pointer(buf)), uintptr(size), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "getcwd(%#x, %#x) %v %v %q\n", buf, size, r, err, GoString(buf))
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return (*int8)(unsafe.Pointer(r))
}

// ssize_t write(int fd, const void *buf, size_t count);
func Xwrite(tls *TLS, fd int32, buf unsafe.Pointer, count uint64) int64 {
	switch fd {
	case unistd.XSTDOUT_FILENO:
		n, err := os.Stdout.Write((*[math.MaxInt32]byte)(buf)[:count])
		if err != nil {
			tls.setErrno(err)
		}
		return int64(n)
	case unistd.XSTDERR_FILENO:
		n, err := os.Stderr.Write((*[math.MaxInt32]byte)(buf)[:count])
		if err != nil {
			tls.setErrno(err)
		}
		return int64(n)
	}
	r, _, err := syscall.Syscall(syscall.SYS_WRITE, uintptr(fd), uintptr(buf), uintptr(count))
	if strace {
		fmt.Fprintf(os.Stderr, "write(%v, %#x, %v) %v %v\n", fd, buf, count, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int64(r)
}

// ssize_t readlink(const char *pathname, char *buf, size_t bufsiz);
func Xreadlink(tls *TLS, pathname, buf *int8, bufsiz uint64) int64 {
	panic("TODO")
}

// long sysconf(int name);
func Xsysconf(tls *TLS, name int32) int64 {
	switch name {
	case unistd.X_SC_PAGESIZE:
		return int64(os.Getpagesize())
	default:
		panic(fmt.Errorf("%v(%#x)", name, name))
	}
}
