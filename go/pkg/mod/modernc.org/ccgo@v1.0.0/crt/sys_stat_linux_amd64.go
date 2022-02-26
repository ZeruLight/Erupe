// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const Tstruct_stat64 = "struct{st_dev uint64,st_ino uint64,st_nlink uint64,st_mode uint32,st_uid uint32,st_gid uint32,__pad0 int32,st_rdev uint64,st_size int64,st_blksize int64,st_blocks int64,st_atime int64,st_atimensec uint64,st_mtime int64,st_mtimensec uint64,st_ctime int64,st_ctimensec uint64,__glibc_reserved [3]int64}"

type Xstruct_stat64 struct {
	Xst_dev           uint64
	Xst_ino           uint64
	Xst_nlink         uint64
	Xst_mode          uint32
	Xst_uid           uint32
	Xst_gid           uint32
	X__pad0           int32
	Xst_rdev          uint64
	Xst_size          int64
	Xst_blksize       int64
	Xst_blocks        int64
	Xst_atime         int64
	Xst_atimensec     uint64
	Xst_mtime         int64
	Xst_mtimensec     uint64
	Xst_ctime         int64
	Xst_ctimensec     uint64
	X__glibc_reserved [3]int64
} // t196 struct{st_dev uint64,st_ino uint64,st_nlink uint64,st_mode uint32,st_uid uint32,st_gid uint32,__pad0 int32,st_rdev uint64,st_size int64,st_blksize int64,st_blocks int64,st_atime int64,st_atimensec uint64,st_mtime int64,st_mtimensec uint64,st_ctime int64,st_ctimensec uint64,__glibc_reserved [3]int64}

// extern int stat64(char *__file, struct stat64 *__buf);
func Xstat64(tls *TLS, file *int8, buf *Xstruct_stat64) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_STAT, uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(buf)), 0)
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
	r, _, err := syscall.Syscall(syscall.SYS_FSTAT, uintptr(fildes), uintptr(unsafe.Pointer(buf)), 0)
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
	r, _, err := syscall.Syscall(syscall.SYS_LSTAT, uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(buf)), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "lstat(%q, %#x) %v %v\n", GoString(file), buf, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}
