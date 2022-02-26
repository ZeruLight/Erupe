// Copyright (c) 2014 ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql // import "modernc.org/ql"

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func TestWALRemoval(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ql-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	tempDBName := filepath.Join(tmpDir, "test_was_removal.db")
	wName := WalName(tempDBName)

	db, err := OpenFile(tempDBName, &Options{CanCreate: true})
	if err != nil {
		t.Fatalf("Cannot open db %s: %s\n", tempDBName, err)
	}
	db.Close()
	if !fileExists(wName) {
		t.Fatalf("Expect WAL file %s to exist but it doesn't", wName)
	}

	db, err = OpenFile(tempDBName, &Options{CanCreate: true, RemoveEmptyWAL: true})
	if err != nil {
		t.Fatalf("Cannot open db %s: %s\n", tempDBName, err)
	}
	db.Close()
	if fileExists(wName) {
		t.Fatalf("Expect WAL file %s to be removed but it still exists", wName)
	}
}
