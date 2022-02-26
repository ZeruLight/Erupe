// Copyright 2020 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "modernc.org/ccgo/v3"

import (
	"fmt"
	"os"

	ccgo "modernc.org/ccgo/v3/lib"
	_ "modernc.org/libc"
)

//TODO parallel

//TODO CPython
//TODO Cython
//TODO gmp
//TODO gofrontend
//TODO gsl
//TODO gtk
//TODO hdf5
//TODO minigmp
//TODO mpc
//TODO mpfr
//TODO pcre
//TODO pcre2
//TODO quickjs
//TODO redis
//TODO tcl/tk
//TODO wolfssl
//TODO zdat
//TODO zlib

func main() {
	if err := ccgo.NewTask(os.Args, os.Stdout, os.Stderr).Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
