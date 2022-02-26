// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// void *dlopen(const char *filename, int flags);
func Xdlopen(tls *TLS, filename *int8, flags int32) unsafe.Pointer {
	panic("TODO")
}

// char *dlerror(void);
func Xdlerror(tls *TLS) *int8 {
	panic("TODO")
}

// int dlclose(void *handle);
func Xdlclose(tls *TLS, handle unsafe.Pointer) int32 {
	panic("TODO")
}

// void *dlsym(void *handle, const char *symbol);
func Xdlsym(tls *TLS, handle unsafe.Pointer, symbol *int8) unsafe.Pointer {
	panic("TODO")
}
