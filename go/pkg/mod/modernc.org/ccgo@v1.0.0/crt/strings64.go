// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

// int ffsl(long i);
func X__builtin_ffsl(tls *TLS, i int64) int32 { return X__builtin_ffsll(tls, i) }

// int ffsl(long i);
func Xffsl(tls *TLS, i int64) int32 { return X__builtin_ffsl(tls, i) }
