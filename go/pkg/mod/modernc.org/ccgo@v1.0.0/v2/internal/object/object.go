// Copyright 2018 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package object reads and writes object files.
package object // import "modernc.org/ccgo/v2/internal/object"

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	// BinVersion defines the binary object version
	BinVersion = 1

	// ObjVersion define the linker object version
	ObjVersion = 1
)

var (
	// BinMagic defines the binary object magic. R/O.
	BinMagic = []byte{0xfa, 0x67, 0x7b, 0x2d, 0xae, 0x0a, 0x88, 0x06}

	// ObjMagic defines the linker object magic. R/O.
	ObjMagic = []byte{0xc6, 0x1f, 0xd0, 0xb5, 0xc4, 0x39, 0xad, 0x56}
)

// Decode reads data in object format from in and writes the result to out
// while requiring goos, goarch, version and magic to match exactly those
// passed to Encode.
func Decode(out io.Writer, goos, goarch string, version uint64, magic []byte, in io.Reader) (err error) {
	r, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("error reading object file: %v", err)
	}

	if g, e := len(r.Header.Extra), len(magic); g < e {
		return fmt.Errorf("Decode: got file magic length %v, expected %v", g, e)
	}

	if g, e := r.Header.Extra[:len(magic)], magic; !bytes.Equal(g, e) {
		return fmt.Errorf("Decode: unrecognized file format: expected magic |% x| got |% x|", g, e)
	}

	buf := r.Header.Extra[len(magic):]
	a := bytes.Split(buf, []byte{'|'})
	if len(a) != 3 {
		return fmt.Errorf("corrupted file")
	}

	if s := string(a[0]); s != goos {
		return fmt.Errorf("invalid platform %q", s)
	}

	if s := string(a[1]); s != goarch {
		return fmt.Errorf("invalid architecture %q", s)
	}

	v, err := strconv.ParseUint(string(a[2]), 10, 64)
	if err != nil {
		return err
	}

	if v != version {
		return fmt.Errorf("invalid version number %v", v)
	}

	if _, err := io.Copy(out, r); err != nil {
		return err
	}

	return r.Close()
}

// Encode reads data from in and writes them in object format to out, tagged
// with goos, goarch, version and magic. The character '|' may not appear in
// goos or goarch.
func Encode(out io.Writer, goos, goarch string, version uint64, magic []byte, in io.Reader) (err error) {
	if strings.Contains(goos, "|") {
		return fmt.Errorf("invalid goos: %q", goos)
	}

	if strings.Contains(goarch, "|") {
		return fmt.Errorf("invalid goarch: %q", goarch)
	}

	w := gzip.NewWriter(out)
	w.Header.Comment = "ccgo object file"
	var buf bytes.Buffer
	buf.Write(magic)
	fmt.Fprintf(&buf, "%s|%s|%v", goos, goarch, version)
	w.Header.Extra = buf.Bytes()
	w.Header.ModTime = time.Now()
	w.Header.OS = 255 // Unknown OS.
	if _, err = io.Copy(w, in); err != nil {
		return err
	}

	return w.Close()
}
