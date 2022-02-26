// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// int open64(const char *pathname, int flags, ...);
func Xopen64(tls *TLS, pathname *int8, flags int32, args ...interface{}) int32 {
	var mode int32
	if len(args) != 0 {
		mode = args[0].(int32)
	}
	r, _, err := syscall.Syscall(syscall.SYS_OPEN, uintptr(unsafe.Pointer(pathname)), uintptr(flags), uintptr(mode))
	if strace {
		fmt.Fprintf(os.Stderr, "open(%q, %v, %#o) %v %v\n", GoString(pathname), modeString(flags), mode, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}
