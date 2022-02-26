// Copyright 2017 The CRT Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crt // import "modernc.org/ccgo/crt"

// extern int *__errno_location(void);
func X__errno_location(tls *TLS) *int32 { return &tls.errno }
