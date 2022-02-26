// Copyright 2021 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package ccgo // import "modernc.org/ccgo/v3/lib"

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testSingle(t *testing.T, main, path string, ccgoArgs []string, runargs []string) (r bool) {
	defer func() {
		if err := recover(); err != nil {
			if *oStackTrace {
				fmt.Printf("%s\n", stack())
			}
			if *oTrace {
				fmt.Println(err)
			}
			t.Errorf("%s: %v", path, err)
			r = false
		}
	}()

	ccgoArgs = append(ccgoArgs, "-D__ccgo_test__")
	ccgoArgs = append(ccgoArgs, path)
	if err := NewTask(ccgoArgs, nil, nil).Main(); err != nil {
		if *oTrace {
			fmt.Println(err)
		}
		err = cpp(*oCpp, ccgoArgs, err)
		t.Errorf("%s: %v", path, err)
		return false
	}

	out, err := func() ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		return exec.CommandContext(ctx, "go", append([]string{"run", main}, runargs...)...).CombinedOutput()
	}()
	if err != nil {
		if *oTrace {
			fmt.Println(err)
		}
		b, _ := ioutil.ReadFile(main)
		t.Errorf("\n%s\n%v: %s\n%v", b, path, out, err)
		return false
	}

	if *oTraceF {
		b, _ := ioutil.ReadFile(main)
		fmt.Printf("\n----\n%s\n----\n", b)
	}
	if *oTraceO {
		fmt.Printf("%s\n", out)
	}
	exp, err := ioutil.ReadFile(noExt(path) + ".expect")
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}

		return false
	}

	out = trim(out)
	exp = trim(exp)
	switch base := filepath.Base(path); base {
	case "70_floating_point_literals.c": //TODO TCC binary extension
		a := strings.Split(string(exp), "\n")
		exp = []byte(strings.Join(a[:35], "\n"))
	}

	if !bytes.Equal(out, exp) {
		if *oTrace {
			fmt.Println(err)
		}
		t.Errorf("%v: out\n%s\nexp\n%s", path, out, exp)
		return false
	}

	return true
}
