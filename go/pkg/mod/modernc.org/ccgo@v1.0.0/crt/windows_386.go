// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package crt // import "modernc.org/ccgo/crt"

import "unsafe"

const (
	TCRITICAL_SECTION    = "struct{*struct{},int32,int32,*struct{},*struct{},uint32}"
	TFILETIME            = "struct{uint32,uint32}"
	TLARGE_INTEGER       = "union{struct{uint32,int32},struct{uint32,int32},int64}"
	TSECURITY_ATTRIBUTES = "struct{uint32,*struct{},int32}"
	TSYSTEM_INFO         = "struct{union{uint32,struct{uint16,uint16}},uint32,*struct{},*struct{},uint32,uint32,uint32,uint32,uint16,uint16}"
	TSYSTEMTIME          = "struct{uint16,uint16,uint16,uint16,uint16,uint16,uint16,uint16}"
	THMODULE             = "struct{int32}"
	TOSVERSIONINFOA      = "struct{uint32,uint32,uint32,uint32,uint32,[128]int8}"
	TOSVERSIONINFOW      = "struct{uint32,uint32,uint32,uint32,uint32,[128]uint16}"
	TOVERLAPPED          = "struct{uint32,uint32,union{struct{uint32,uint32},*struct{}},*struct{}}"
)

type XOSVERSIONINFOA struct {
	X0 uint32
	X1 uint32
	X2 uint32
	X3 uint32
	X4 uint32
	X5 [128]int8
}

type XHMODULE struct {
	X0 int32
}

type XCRITICAL_SECTION struct {
	X0 unsafe.Pointer
	X1 int32
	X2 int32
	X3 unsafe.Pointer
	X4 unsafe.Pointer
	X5 uint32
}

type XFILETIME struct {
	X0 uint32
	X1 uint32
}

type XLARGE_INTEGER struct {
	X [0]struct {
		X0 struct {
			X0 uint32
			X1 int32
		}
		X1 struct {
			X0 uint32
			X1 int32
		}
		X2 int64
	}
	U [8]byte
}

type XSYSTEM_INFO struct {
	X0 struct {
		X [0]struct {
			X0 uint32
			X1 struct {
				X0 uint16
				X1 uint16
			}
		}
		U [4]byte
	}
	X1 uint32
	X2 unsafe.Pointer
	X3 unsafe.Pointer
	X4 uint32
	X5 uint32
	X6 uint32
	X7 uint32
	X8 uint16
	X9 uint16
}

type XSYSTEMTIME struct {
	X0 uint16
	X1 uint16
	X2 uint16
	X3 uint16
	X4 uint16
	X5 uint16
	X6 uint16
	X7 uint16
}

type XOSVERSIONINFOW struct {
	X0 uint32
	X1 uint32
	X2 uint32
	X3 uint32
	X4 uint32
	X5 [128]uint16
}

type XOVERLAPPED struct {
	X0 uint32
	X1 uint32
	X2 struct {
		X [0]struct {
			X0 struct {
				X0 uint32
				X1 uint32
			}
			X1 unsafe.Pointer
		}
		U [8]byte
	}
	X3 unsafe.Pointer
}

type XSECURITY_ATTRIBUTES struct {
	X0 uint32
	X1 unsafe.Pointer
	X2 int32
}
