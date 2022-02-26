// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

// int tolower(int c);
func Xtolower(tls *TLS, c int32) int32 {
	if c >= 'A' && c <= 'Z' {
		c |= ' '
	}
	return c
}

// int isprint(int c);
func X__builtin_isprint(tls *TLS, c int32) int32 {
	if c >= ' ' && c <= '~' {
		return 1
	}

	return 0
}

// int isprint(int c);
func Xisprint(tls *TLS, c int32) int32 { return X__builtin_isprint(tls, c) }
