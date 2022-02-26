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

const Tstruct_timeval = "struct{tv_sec int32,tv_usec int32}"

type Xstruct_timeval struct {
	Xtv_sec  int32
	Xtv_usec int32
} // t195 struct{tv_sec int32,tv_usec int32}

// int gettimeofday(struct timeval *restrict tp, void *restrict tzp);
func Xgettimeofday(tls *TLS, tp *Xstruct_timeval, tzp unsafe.Pointer) int32 {
	r, _, err := syscall.Syscall(syscall.SYS_GETTIMEOFDAY, uintptr(tzp), uintptr(unsafe.Pointer(tp)), 0)
	if strace {
		fmt.Fprintf(os.Stderr, "gettimeofday(%#x, %#x) %v %v\n", tzp, tp, r, err)
	}
	return int32(r)
}

// int utimes(const char *filename, const struct timeval times[2]);
func Xutimes(tls *TLS, filename *int8, times *[2]Xstruct_timeval) int32 {
	panic("TODO")
}
