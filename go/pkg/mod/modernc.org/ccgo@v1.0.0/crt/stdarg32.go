// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm arm64be armbe mips mipsle ppc ppc64le s390 s390x sparc

package crt // import "modernc.org/ccgo/crt"

func VALong(ap *[]interface{}) int64   { return int64(VAInt32(ap)) }
func VAULong(ap *[]interface{}) uint64 { return uint64(VAUint32(ap)) }
