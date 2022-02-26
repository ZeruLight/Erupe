// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

// uint64_t __builtin_bswap64 (uint64_t x)
func X__builtin_bswap64(tls *TLS, x uint64) uint64 {
	return x&0x00000000000000ff<<56 |
		x&0x000000000000ff00<<40 |
		x&0x0000000000ff0000<<24 |
		x&0x00000000ff000000<<8 |
		x&0x000000ff00000000>>8 |
		x&0x0000ff0000000000>>24 |
		x&0x00ff000000000000>>40 |
		x&0xff00000000000000>>56
}
