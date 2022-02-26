// Copyright 2019 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc

import (
	"io"
	"os"
	"testing"
)

type logPathFS struct {
	fs    Filesystem
	stats []string
	opens []string
}

func strSliceEqual(t testing.TB, exp, got []string) {
	if len(exp) != len(got) {
		t.Fatalf("unexpected output:\n%q\nvs\n%q", exp, got)
		return
	}
	for i := range exp {
		if exp[i] != got[i] {
			t.Fatalf("unexpected output:\n%q\nvs\n%q", exp, got)
			break
		}
	}
}

func (fs *logPathFS) Stat(path string, sys bool) (os.FileInfo, error) {
	fs.stats = append(fs.stats, path)
	return fs.fs.Stat(path, sys)
}

func (fs *logPathFS) Open(path string, sys bool) (io.ReadCloser, error) {
	fs.opens = append(fs.opens, path)
	return fs.fs.Open(path, sys)
}

func TestParseVFS(t *testing.T) {
	if isTestingMingw {
		t.Skip("mingw")
		return
	}

	isTesting = false
	defer func() {
		isTesting = true
	}()
	cfg := &Config{ABI: testABI}
	const (
		pref   = "/fake"
		usrDir = pref + "/usr"
		sysDir = pref + "/sys"
	)
	sys := StaticFS(map[string]string{
		sysDir + "/sys.h": `void bar() {}`,
	})
	usr := StaticFS(map[string]string{
		usrDir + "/file.h": `void foo();`,
		usrDir + "/file.c": `
#include "file.h"
#include <sys.h>

void foo() {
	bar();
}
`,
	})
	lfs := &logPathFS{fs: WorkingDir(usrDir, Overlay(usr, sys))}
	cfg.Filesystem = lfs
	_, err := Parse(cfg, []string{usrDir}, []string{sysDir}, []Source{{Name: "file.c", DoNotCache: true}})
	if err != nil {
		t.Logf("\nstats: %q\nopens: %q", lfs.stats, lfs.opens)
		t.Fatal(err)
	}
	strSliceEqual(t, []string{
		"file.c",
		usrDir + "/file.h",
		usrDir + "/file.h",
		sysDir + "/sys.h",
		sysDir + "/sys.h",
	}, lfs.stats)
	strSliceEqual(t, []string{
		"file.c",
		usrDir + "/file.h",
		sysDir + "/sys.h",
	}, lfs.opens)
}
