// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
	"unsafe"
)

func VAPointer(ap *[]interface{}) (r unsafe.Pointer) {
	s := *ap
	switch x := s[0].(type) {
	case int32:
		r = unsafe.Pointer(uintptr(x))
	case uint32:
		r = unsafe.Pointer(uintptr(x))
	case int64:
		r = unsafe.Pointer(uintptr(x))
	case uint64:
		r = unsafe.Pointer(uintptr(x))
	case unsafe.Pointer:
		r = x
	case nil:
		// nop
	default:
		panic(fmt.Errorf("%T", x))
	}
	*ap = s[1:]
	return r
}

func VAFloat64(ap *[]interface{}) float64 {
	s := *ap
	v := s[0].(float64)
	*ap = s[1:]
	return v
}

func VAInt32(ap *[]interface{}) (v int32) {
	s := *ap
	switch x := s[0].(type) {
	case int32:
		v = x
	case uint32:
		v = int32(x)
	case int64:
		v = int32(x)
	case uint64:
		v = int32(x)
	default:
		panic(fmt.Errorf("%T", x))
	}
	*ap = s[1:]
	return v
}

func VAUint32(ap *[]interface{}) (v uint32) {
	s := *ap
	switch x := s[0].(type) {
	case int32:
		v = uint32(x)
	case uint32:
		v = x
	case int64:
		v = uint32(x)
	case uint64:
		v = uint32(x)
	default:
		panic(fmt.Errorf("%T", x))
	}
	*ap = s[1:]
	return v
}

func VAInt64(ap *[]interface{}) (v int64) {
	s := *ap
	switch x := s[0].(type) {
	case int32:
		v = int64(x)
	case uint32:
		v = int64(x)
	case int64:
		v = x
	case uint64:
		v = int64(x)
	default:
		panic(fmt.Errorf("%T", x))
	}
	*ap = s[1:]
	return v
}

func VAUint64(ap *[]interface{}) (v uint64) {
	s := *ap
	switch x := s[0].(type) {
	case int32:
		v = uint64(x)
	case uint32:
		v = uint64(x)
	case int64:
		v = uint64(x)
	case uint64:
		v = x
	default:
		panic(fmt.Errorf("%T", x))
	}
	*ap = s[1:]
	return v
}

func VAOther(ap *[]interface{}) (v interface{}) {
	s := *ap
	v = s[0]
	*ap = s[1:]
	return v
}
