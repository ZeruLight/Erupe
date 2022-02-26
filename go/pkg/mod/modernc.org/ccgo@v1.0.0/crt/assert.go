// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

import (
	"fmt"
)

func X__builtin_assert_fail(tls *TLS, file *int8, line int32, fn, msg *int8) {
	panic(fmt.Errorf("%s:%s: assertion failure in %s: %s", GoString(file), line, GoString(fn), GoString(msg)))
}
