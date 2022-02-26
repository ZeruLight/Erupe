// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo"

import (
	"modernc.org/ccgo/crt"
	"modernc.org/ir"
)

var typeMap map[ir.TypeID]string = map[ir.TypeID]string{
	ir.TypeID(dict.SID(crt.TCRITICAL_SECTION)):    "crt.XCRITICAL_SECTION",
	ir.TypeID(dict.SID(crt.TFILETIME)):            "crt.XFILETIME",
	ir.TypeID(dict.SID(crt.TLARGE_INTEGER)):       "crt.XLARGE_INTEGER",
	ir.TypeID(dict.SID(crt.TSECURITY_ATTRIBUTES)): "crt.XSECURITY_ATTRIBUTES",
	ir.TypeID(dict.SID(crt.TSYSTEM_INFO)):         "crt.XSYSTEM_INFO",
	ir.TypeID(dict.SID(crt.TSYSTEMTIME)):          "crt.XSYSTEMTIME",
	ir.TypeID(dict.SID(crt.THMODULE)):             "crt.XHMODULE",
	ir.TypeID(dict.SID(crt.TOSVERSIONINFOA)):      "crt.XOSVERSIONINFOA",
	ir.TypeID(dict.SID(crt.TOSVERSIONINFOW)):      "crt.XOSVERSIONINFOW",
	ir.TypeID(dict.SID(crt.TOVERLAPPED)):          "crt.XOVERLAPPED",
	ir.TypeID(dict.SID(crt.Tpthread_attr_t)):      "crt.Xpthread_attr_t",
	ir.TypeID(dict.SID(crt.Tpthread_mutex_t)):     "crt.Xpthread_mutex_t",
	ir.TypeID(dict.SID(crt.Tpthread_mutexattr_t)): "crt.Xpthread_mutexattr_t",
	ir.TypeID(dict.SID(crt.Ttm)):                  "crt.Xtm",
}
