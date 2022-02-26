// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package ccgo // import "modernc.org/ccgo"

import (
	"modernc.org/ccgo/crt"
	"modernc.org/ir"
)

var typeMap map[ir.TypeID]string = map[ir.TypeID]string{
	ir.TypeID(dict.SID(crt.TFILE)):                "crt.XFILE",
	ir.TypeID(dict.SID(crt.Tpthread_attr_t)):      "crt.Xpthread_attr_t",
	ir.TypeID(dict.SID(crt.Tpthread_cond_t)):      "crt.Xpthread_cond_t",
	ir.TypeID(dict.SID(crt.Tpthread_mutex_t)):     "crt.Xpthread_mutex_t",
	ir.TypeID(dict.SID(crt.Tpthread_mutexattr_t)): "crt.Xpthread_mutexattr_t",
	ir.TypeID(dict.SID(crt.Tstruct_stat64)):       "crt.Xstruct_stat64",
	ir.TypeID(dict.SID(crt.Tstruct_timeval)):      "crt.Xstruct_timeval",
	ir.TypeID(dict.SID(crt.Ttm)):                  "crt.Xtm",
}
