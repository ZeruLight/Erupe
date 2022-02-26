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

const Tstruct_stat64 = "struct{st_dev uint64,__pad1 uint32,__st_ino uint32,st_mode uint32,st_nlink uint32,st_uid uint32,st_gid uint32,st_rdev uint64,__pad2 uint32,st_size int64,st_blksize int32,st_blocks int64,st_atime int32,st_atimensec uint32,st_mtime int32,st_mtimensec uint32,st_ctime int32,st_ctimensec uint32,st_ino uint64}"

type Xstruct_stat64 struct {
	Xst_dev       uint64
	X__pad1       uint32
	X__st_ino     uint32
	Xst_mode      uint32
	Xst_nlink     uint32
	Xst_uid       uint32
	Xst_gid       uint32
	Xst_rdev      uint64
	X__pad2       uint32
	Xst_size      int64
	Xst_blksize   int32
	Xst_blocks    int64
	Xst_atime     int32
	Xst_atimensec uint32
	Xst_mtime     int32
	Xst_mtimensec uint32
	Xst_ctime     int32
	Xst_ctimensec uint32
	Xst_ino       uint64
} // t195 struct{st_dev uint64,__pad1 uint32,__st_ino uint32,st_mode uint32,st_nlink uint32,st_uid uint32,st_gid uint32,st_rdev uint64,__pad2 uint32,st_size int64,st_blksize int32,st_blocks int64,st_atime int32,st_atimensec uint32,st_mtime int32,st_mtimensec uint32,st_ctime int32,st_ctimensec uint32,st_ino uint64}

// extern int stat64(char *__file, struct stat64 *__buf);
func Xstat64(tls *TLS, file *int8, buf *Xstruct_stat64) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_STAT64, uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(buf)), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "stat(%q, %#x) %v %v\n", GoString(file), buf, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// int fstat64(int fildes, struct stat64 *buf);
func Xfstat64(tls *TLS, fildes int32, buf *Xstruct_stat64) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_FSTAT64, uintptr(fildes), uintptr(unsafe.Pointer(buf)), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "fstat(%v, %#x) %v %v\n", fildes, buf, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// extern int lstat64(char *__file, struct stat64 *__buf);
func Xlstat64(tls *TLS, file *int8, buf *Xstruct_stat64) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_LSTAT64, uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(buf)), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "lstat(%q, %#x) %v %v\n", GoString(file), buf, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}
