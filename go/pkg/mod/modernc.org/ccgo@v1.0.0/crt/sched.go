// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"runtime"
)

// int sched_yield(void);
func Xsched_yield(tls *TLS) int32 {
	runtime.Gosched()
	return 0
}
