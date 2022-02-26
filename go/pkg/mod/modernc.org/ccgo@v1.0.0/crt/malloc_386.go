// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// size_t malloc_usable_size (void *ptr);
func Xmalloc_usable_size(tls *TLS, ptr unsafe.Pointer) uint32 { return uint32(UsableSize(ptr)) }
