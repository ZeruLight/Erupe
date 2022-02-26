// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// int close(int fd);
func Xclose(tls *TLS, fd int32) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_CLOSE, uintptr(fd), 0, 0)
	if strace {
		fmt.Fprintf(os.Stderr, "close(%v) %v %v\n", fd, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// int access(const char *path, int amode);
func Xaccess(tls *TLS, path *int8, amode int32) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_ACCESS, uintptr(unsafe.Pointer(path)), uintptr(amode), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "access(%q) %v %v\n", GoString(path), r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// int unlink(const char *path);
func Xunlink(tls *TLS, path *int8) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_UNLINK, uintptr(unsafe.Pointer(path)), 0, 0)
	if strace {
		fmt.Fprintf(os.Stderr, "unlink(%q) %v %v\n", GoString(path), r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// int rmdir(const char *pathname);
func Xrmdir(tls *TLS, pathname *int8) int32 {
	panic("TODO")
}

// int fchown(int fd, uid_t owner, gid_t group);
func Xfchown(tls *TLS, fd int32, owner, group uint32) int32 {
	panic("TODO")
}

// uid_t geteuid(void);
func Xgeteuid(tls *TLS) uint32 {
	r, _, _ := syscall.RawSyscall(syscall.SYS_GETEUID, 0, 0, 0)
	if strace {
		fmt.Fprintf(os.Stderr, "geteuid() %v\n", r)
	}
	return uint32(r)
}

// int fsync(int fildes);
func Xfsync(tls *TLS, fildes int32) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_FSYNC, uintptr(fildes), 0, 0)
	if strace {
		fmt.Fprintf(os.Stderr, "fsync(%v) %v %v\n", fildes, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// pid_t getpid(void);
func Xgetpid(tls *TLS) int32 {
	r, _, _ := syscall.RawSyscall(syscall.SYS_GETPID, 0, 0, 0)
	if strace {
		fmt.Fprintf(os.Stderr, "getpid() %v\n", r)
	}
	return int32(r)
}

// unsigned sleep(unsigned seconds);
func Xsleep(tls *TLS, seconds uint32) uint32 {
	time.Sleep(time.Duration(seconds) * time.Second)
	if strace {
		fmt.Fprintf(os.Stderr, "sleep(%#x)", seconds)
	}
	return 0
}

// off_t lseek64(int fildes, off_t offset, int whence);
func Xlseek64(tls *TLS, fildes int32, offset int64, whence int32) int64 {
	r, _, err := syscall.Syscall(syscall.SYS_LSEEK, uintptr(fildes), uintptr(offset), uintptr(whence))
	if strace {
		fmt.Fprintf(os.Stderr, "lseek(%v, %v, %v) %v %v\n", fildes, offset, whence, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int64(r)
}

// int ftruncate(int fildes, off_t length);
func Xftruncate64(tls *TLS, fildes int32, length int64) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_FTRUNCATE, uintptr(fildes), uintptr(length), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "ftruncate(%#x, %#x) %v, %v\n", fildes, length, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}

// int usleep(useconds_t usec);
func Xusleep(tls *TLS, usec uint32) int32 {
	time.Sleep(time.Duration(usec) * time.Microsecond)
	if strace {
		fmt.Fprintf(os.Stderr, "usleep(%#x)", usec)
	}
	return 0
}
