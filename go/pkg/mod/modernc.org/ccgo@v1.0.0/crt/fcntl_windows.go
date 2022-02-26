// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"syscall"
)

// int open(const char *pathname, int flags, ...);
func Xopen(tls *TLS, pathname *int8, flags int32, args ...interface{}) int32 {
	return Xopen64(tls, pathname, flags, args)
}

// int close(int fd);
func Xclose(tls *TLS, fd int32) int32 {
	err := syscall.Close(syscall.Handle(fd))
	if strace {
		fmt.Fprintf(os.Stderr, "close(%v) %v\n", fd, err)
	}
	if err != nil {
		tls.setErrno(err)
	}
	return 0
}

// int open64(const char *pathname, int flags, ...);
func Xopen64(tls *TLS, pathname *int8, flags int32, args ...interface{}) int32 {
	var mode uint32
	if len(args) != 0 {
		mode = args[0].(uint32)
	}

	path := GoString(pathname)
	h, err := syscall.Open(path, int(flags), mode)
	if err != nil {
		tls.setErrno(err)
	}

	if strace {
		fmt.Fprintf(os.Stderr, "open(%q, %v, %#o) %v %v\n", path, modeString(flags), mode, h, err)
	}
	// For compatibility reasons a HANDLE (atleast for file types) is always 32-bits
	// so this truncating from uintptr -> int32 is safe.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa384203%28v=vs.85%29.aspx
	return int32(h)
}
