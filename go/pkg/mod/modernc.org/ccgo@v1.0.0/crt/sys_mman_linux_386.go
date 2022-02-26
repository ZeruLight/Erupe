// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// void *mmap(void *addr, size_t len, int prot, int flags, int fildes, off_t off);
func Xmmap64(tls *TLS, addr unsafe.Pointer, len uint32, prot, flags, fildes int32, off int64) unsafe.Pointer {
	r, _, err := syscall.Syscall6(syscall.SYS_MMAP, uintptr(addr), uintptr(len), uintptr(prot), uintptr(flags), uintptr(fildes), uintptr(off))
	if strace {
		fmt.Fprintf(os.Stderr, "mmap(%#x, %#x, %#x, %#x, %#x, %#x) (%#x, %v)\n", addr, len, prot, flags, fildes, off, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}

	return unsafe.Pointer(r)
}

// int munmap(void *addr, size_t len);
func Xmunmap(tls *TLS, addr unsafe.Pointer, len uint32) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_MUNMAP, uintptr(addr), uintptr(len), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "munmap(%#x, %#x) (%#x, %v)\n", addr, len, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}
