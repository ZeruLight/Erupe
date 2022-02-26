// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32 arm64 mips64 mips64le mips64p32 mips64p32le ppc64 sparc64

// +build !windows

package crt // import "modernc.org/ccgo/crt"

const Ttm = "struct{tm_sec int32,tm_min int32,tm_hour int32,tm_mday int32,tm_mon int32,tm_year int32,tm_wday int32,tm_yday int32,tm_isdst int32,__tm_gmtoff int64,__tm_zone *int8}"

type Xtm struct {
	Xtm_sec      int32
	Xtm_min      int32
	Xtm_hour     int32
	Xtm_mday     int32
	Xtm_mon      int32
	Xtm_year     int32
	Xtm_wday     int32
	Xtm_yday     int32
	Xtm_isdst    int32
	X__tm_gmtoff int64
	X__tm_zone   *int8
} // t162 struct{tm_sec int32,tm_min int32,tm_hour int32,tm_mday int32,tm_mon int32,tm_year int32,tm_wday int32,tm_yday int32,tm_isdst int32,__tm_gmtoff int64,__tm_zone *int8}

// struct tm *localtime(const time_t *timep);
func Xlocaltime(tls *TLS, timep *int64) *Xtm {
	TODO("")
	panic("TODO")
}

// time_t time(time_t *tloc);
func Xtime(tls *TLS, tloc *int64) int64 {
	panic("TODO")
}
