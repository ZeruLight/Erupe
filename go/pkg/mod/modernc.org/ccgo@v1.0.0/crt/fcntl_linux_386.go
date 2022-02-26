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

// int fcntl(int fildes, int cmd, ...);
func Xfcntl(tls *TLS, fildes, cmd int32, args ...interface{}) int32 {
	var arg uintptr
	if len(args) != 0 {
		switch x := args[0].(type) {
		case int32:
			arg = uintptr(x)
		case unsafe.Pointer:
			arg = uintptr(x)
		default:
			TODO("%T", x)
		}
	}
	r, _, err := syscall.Syscall(syscall.SYS_FCNTL64, uintptr(fildes), uintptr(cmd), uintptr(arg))
	if strace {
		fmt.Fprintf(os.Stderr, "fcntl(%v, %v, %#x) %v %v\n", fildes, cmdString(cmd), arg, r, err)
	}
	if err != 0 {
		tls.setErrno(err)
	}
	return int32(r)
}
