// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

package crt // import "modernc.org/ccgo/crt"

const Ttm = "struct{int32,int32,int32,int32,int32,int32,int32,int32,int32}"

type Xtm struct {
	X0 int32
	X1 int32
	X2 int32
	X3 int32
	X4 int32
	X5 int32
	X6 int32
	X7 int32
	X8 int32
}

// struct tm *localtime(const time_t *timep);
func Xlocaltime(tls *TLS, timep *int64) *Xtm {
	TODO("")
	panic("TODO")
}

// time_t time(time_t *tloc);
func Xtime(tls *TLS, tloc *int64) int64 {
	panic("TODO")
}
