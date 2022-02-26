// Copyright 2018 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"go/scanner"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"modernc.org/sortutil"
)

var (
	bNL    = []byte{'\n'}
	bPanic = []byte("panic")
)

func env(key, value string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}

	return value
}

func expandError(err error) error {
	switch x := err.(type) {
	case scanner.ErrorList:
		var a []string
		for _, v := range x {
			a = append(a, v.Error())
		}
		return fmt.Errorf("%s", strings.Join(a[:sortutil.Dedupe(sort.StringSlice(a))], "\n"))

	default:
		return err
	}
}

func toExt(nm, new string) string { return nm[:len(nm)-len(filepath.Ext(nm))] + new }

func debugStack2() []byte {
	b := debug.Stack()
	b = b[bytes.Index(b, bPanic)+1:]
	b = b[bytes.Index(b, bPanic):]
	b = b[bytes.Index(b, bNL)+1:]
	a := bytes.SplitN(b, bNL, 3)
	if len(a) > 2 {
		a = a[:2]
	}
	if len(a) > 1 {
		a = a[1:]
	}
	return bytes.Join(a, bNL)
}

func errs(out, in error) error {
	if out == nil {
		out = in
	}
	return out
}

type arHeader struct { // https://en.wikipedia.org/wiki/Ar_(Unix)#File_header
	FileName  [16]byte
	TimeStamp [12]byte
	Oid       [6]byte
	Gid       [6]byte
	Mode      [8]byte
	Size      [10]byte
	End       [2]byte
}

func (h *arHeader) fn(ext []byte) (string, error) {
	r := strings.TrimSpace(string(h.FileName[:]))
	if r != "//" {
		r = strings.TrimSuffix(r, "/")
	}
	if len(ext) != 0 && len(r) > 1 && r[0] == '/' && r[1] >= '0' && r[1] <= '9' {
		n, err := strconv.ParseUint(r[1:], 10, 31)
		if err != nil {
			return "", err
		}

		if int(n) >= len(ext) {
			return "", fmt.Errorf("invalid extended filename offset")
		}

		i := bytes.IndexByte(ext[n:], '\n')
		switch {
		case i < 0:
			i = len(ext)
		default:
			i += int(n)
		}
		r = strings.TrimSuffix(string(ext[n:i]), "/")
	}
	return r, nil
}

func (h *arHeader) sz() (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(string(h.Size[:])), 10, 63)
}

func init() {
	if unsafe.Sizeof(arHeader{}) != 60 {
		panic("internal error")
	}
}

type arReader struct {
	err error
	ext []byte
	fn  string
	r   *bufio.Reader
	rem int64

	odd bool
}

func newArReader(f *os.File) (*arReader, error) {
	r := bufio.NewReader(f)
	hdr, err := r.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("reading archive header: %v", err)
	}

	if !bytes.Equal(hdr, []byte("!<arch>\n")) {
		return nil, fmt.Errorf("unrecognized archive header %q", hdr)
	}

	return &arReader{r: r}, nil
}

func (r *arReader) Next() bool {
	if r.odd {
		var b [1]byte
		if n, err := r.r.Read(b[:]); n != 1 {
			if err == nil {
				err = fmt.Errorf("internal error")
			}
			r.err = err
			return false
		}
	}
	var h arHeader
	if r.err = binary.Read(r.r, binary.BigEndian, &h); r.err != nil {
		if r.err == io.EOF {
			r.err = nil
		}
		return false
	}

	r.rem, r.err = h.sz()
	if r.err != nil {
		return false
	}

	r.odd = r.rem&1 != 0
	if r.fn, r.err = h.fn(r.ext); r.err != nil {
		return false
	}

	if r.fn == "//" {
		if r.ext != nil {
			r.err = fmt.Errorf("multiple extended filenames section")
			return false
		}

		r.ext = make([]byte, r.rem)
		want := r.rem
		if n, _ := r.Read(r.ext); int64(n) != want {
			r.err = fmt.Errorf("error reading extended filenames section")
			return false
		}

		return r.Next()
	}
	return r.err == nil
}

func (r *arReader) Read(b []byte) (int, error) {
	rq := len(b)
	if int64(rq) > r.rem {
		rq = int(r.rem)
	}
	n, err := r.r.Read(b[:rq])
	r.rem -= int64(n)
	if n == 0 && err == nil {
		err = io.EOF
	}
	return n, err
}

func isArgumentMismatchError(s string) bool {
	a := strings.Split(s, "\n")
	for _, v := range a {
		switch {
		case
			strings.HasPrefix(v, "#"),
			strings.HasPrefix(v, "\thave"),
			strings.HasPrefix(v, "\twant"),
			strings.HasPrefix(v, "exit status "),
			v == "":

			// ok
		default:
			i := strings.Index(v, ": ")
			if i < 0 {
				return false
			}

			switch v = v[i+2:]; {
			case
				strings.HasPrefix(v, "cannot use"),
				strings.HasPrefix(v, "not enough arguments in call to"):

				// ok
			default:
				return false
			}
		}
	}
	return true
}
