// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

// int ffs(int i);
func X__builtin_ffs(tls *TLS, i int32) int32 {
	if i == 0 {
		return 0
	}

	var r int32
	for ; r < 32 && i&(1<<uint(r)) == 0; r++ {
	}
	return r + 1
}

// int ffs(int i);
func Xffs(tls *TLS, i int32) int32 { return X__builtin_ffs(tls, i) }

// int ffsll(long long i);
func X__builtin_ffsll(tls *TLS, i int64) int32 {
	if i == 0 {
		return 0
	}

	var r int32
	for ; r < 64 && i&(1<<uint(r)) == 0; r++ {
	}
	return r + 1
}

// int ffsll(long long i);
func Xffsll(tls *TLS, i int64) int32 { return X__builtin_ffsll(tls, i) }
