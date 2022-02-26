// Copyright © 2005-2014 Rich Felker, et al. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE-MUSL file.

package crt // import "modernc.org/ccgo/crt"

import (
	"unsafe"
)

// int atoi(const char *nptr);
func Xatoi(tls *TLS, _s *int8) (r0 int32) {
	var _n int32
	_ = _n
	var _neg int32
	_ = _neg
	_n = 0
	_neg = 0
_0:
	if (int32(*(*uint16)(unsafe.Pointer(uintptr((unsafe.Pointer)(*X__ctype_b_loc(tls))) + 2*uintptr(int32(*_s))))) & int32(8192)) == 0 {
		goto _1
	}

	*(*uintptr)(unsafe.Pointer(&_s)) += uintptr(1)
	goto _0

_1:
	switch int32(*_s) {
	case 43:
		goto _4
	case 45:
		goto _3
	default:
		goto _5
	}

_3:
	_neg = 1
_4:
	*(*uintptr)(unsafe.Pointer(&_s)) += uintptr(1)
_5:
_6:
	if (int32(*(*uint16)(unsafe.Pointer(uintptr((unsafe.Pointer)(*X__ctype_b_loc(tls))) + 2*uintptr(int32(*_s))))) & int32(2048)) == 0 {
		goto _7
	}

	_n = (10 * _n) - (int32(*postInc0(&_s, 1)) - 48)
	goto _6

_7:
	return func() int32 {
		if _neg != 0 {
			return _n
		}
		return (-_n)
	}()
}

func postInc0(p **int8, d int) *int8 {
	q := (*uintptr)(unsafe.Pointer(p))
	v := *q
	*q += uintptr(d)
	return (*int8)(unsafe.Pointer(v))
}
