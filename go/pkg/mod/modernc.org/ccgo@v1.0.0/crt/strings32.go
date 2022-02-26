// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

// int ffsl(long i);
func X__builtin_ffsl(tls *TLS, i int32) int32 { return X__builtin_ffs(tls, i) }

// int ffsl(long i);
func Xffsl(tls *TLS, i int32) int32 { return X__builtin_ffsl(tls, i) }
