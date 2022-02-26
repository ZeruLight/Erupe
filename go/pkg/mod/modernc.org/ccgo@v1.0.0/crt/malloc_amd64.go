// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// size_t malloc_usable_size (void *ptr);
func Xmalloc_usable_size(tls *TLS, ptr unsafe.Pointer) uint64 { return uint64(UsableSize(ptr)) }
