// Copyright 2019 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v3"

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestScanner(t *testing.T) {
	const c = "a\nm\r\nz"
	//         12 34 5 6

	buf := []byte(c)
	file := tokenNewFile("x", len(buf))
	s := newScanner(newContext(&Config{}), bytes.NewReader(buf), file)
	var gtoks []token3
	for {
		s.lex()
		if s.tok.char < 0 {
			break
		}

		gtoks = append(gtoks, s.tok)
	}
	if err := s.ctx.Err(); err != nil {
		t.Fatal(err)
	}

	const I = IDENTIFIER
	etoks := toks(1, I, "a", 2, '\n', "\n", 3, I, "m", 4, '\n', "\r\n", 6, I, "z", 7, '\n', "\n")

	if g, e := len(gtoks), len(etoks); g != e {
		t.Fatal(g, e)
	}

	for i, g := range gtoks {
		g.src = 0        //TODO
		etoks[i].src = 0 //TODO
		if e := etoks[i]; g != e {
			t.Errorf("%v: %v %v", i, g.str(file), e.str(file))
		}
	}
}

func (t *token3) str(file *tokenFile) string {
	pos := ""
	switch {
	case t.pos <= 0 || int(t.pos) > file.Size()+1:
		pos = fmt.Sprintf("offset:%v", t.pos)
	default:
		pos = file.Position(t.Pos()).String()
	}
	return fmt.Sprintf("%v: %+q(%#x) %q", pos, string(t.char), t.char, t.value)
}

func toks(v ...interface{}) (r []token3) {
	if len(v)%3 != 0 {
		panic(internalError())
	}

	var t token3
	for i, v := range v {
		switch i % 3 {
		case 0:
			t.pos = int32(v.(int))
		case 1:
			switch x := v.(type) {
			case int:
				t.char = rune(x)
			case int32:
				t.char = x
			}
		case 2:
			t.value = dict.sid(v.(string))
			r = append(r, t)
		}
	}
	return r
}

func TestScannerTranslationPhase2(t *testing.T) {
	const c = `
#define x int[42] # foo
int ma\
in() {
	int x = ~3^4|5|6^7;
}
` // Must end in new-line for this test to pass.

	buf := []byte(c)
	for _, v := range trigraphs {
		buf = bytes.Replace(buf, v.to, v.from, -1)
	}
	file := tokenNewFile("x", len(buf))
	s := newScanner(newContext(&Config{Config3: Config3{PreserveWhiteSpace: true}}), bytes.NewReader(buf), file)
	out := ""
	for {
		s.lex()
		if s.tok.char < 0 {
			break
		}

		out += s.tok.String()
	}
	if err := s.ctx.Err(); err != nil {
		t.Fatal(err)
	}

	e := strings.Replace(c, "\\\n", "", -1)
	g := out
	if diff := cmp.Diff(e, g); diff != "" {
		t.Fatal(diff)
	}
}

func TestScannerTranslationPhase2b(t *testing.T) {
	for i, v := range []struct {
		in, out, err string
		cfg          *Config
	}{
		{"", "", "", nil},
		{"\\nb\n", "\\nb\n", "", nil},
		{"\\\nb\n", "b\n", "", nil},
		{"\n", "\n", "", nil},
		{"a\\\n", "a\n", "", nil},
		{"a\\\n", "a\n", "preceded by a backslash character", &Config{Config3: Config3{RejectFinalBackslash: true}}},
		{"a\n", "a\n", "", nil},
		{"a\nb", "a\nb\n", "", nil},
		{"a\nb", "a\nb\n", "shall end in a new-line character", &Config{Config3: Config3{RejectMissingFinalNewline: true}}},
		{"a\nb\n", "a\nb\n", "", nil},
		{"ma\\\nin\n", "main\n", "", nil},
	} {

		buf := []byte(v.in)
		file := tokenNewFile("x", len(buf))
		cfg := v.cfg
		if cfg == nil {
			cfg = &Config{}
		}
		s := newScanner(newContext(cfg), bytes.NewReader(buf), file)
		out := ""
		for {
			s.lex()
			if s.tok.char < 0 {
				break
			}

			out += s.tok.String()
		}
		switch err := s.ctx.Err(); {
		case v.err != "" && err != nil:
			if !regexp.MustCompile(v.err).MatchString(err.Error()) {
				t.Error(i, err)
			}
		case v.err == "" && err != nil:
			t.Errorf("unexpected error: %v", err)
		case v.err != "" && err == nil:
			t.Errorf("%v: expected error matching %s", i, v.err)
		}

		e := v.out
		g := out
		if diff := cmp.Diff(e, g); diff != "" {
			t.Error(i, diff)
		}
	}
}

func TestScanner2(t *testing.T) {
	const I = IDENTIFIER
	white := &Config{Config3: Config3{PreserveWhiteSpace: true}}
next:
	for i, v := range []struct {
		in, err string
		toks    []token3
		cfg     *Config
	}{
		{"  ", "", toks(1, ' ', "  ", 3, '\n', "\n"), white},
		{"  ", "", toks(1, ' ', " ", 3, '\n', "\n"), nil},
		{"  \n", "", toks(1, ' ', " ", 3, '\n', "\n"), nil},
		{"  a  ", "", toks(1, ' ', " ", 3, I, "a", 4, ' ', " ", 6, '\n', "\n"), nil},
		{"  a  \n", "", toks(1, ' ', " ", 3, I, "a", 4, ' ', " ", 6, '\n', "\n"), nil},
		{"  a ", "", toks(1, ' ', " ", 3, I, "a", 4, ' ', " ", 5, '\n', "\n"), nil},
		{"  a \n", "", toks(1, ' ', " ", 3, I, "a", 4, ' ', " ", 5, '\n', "\n"), nil},
		{"  a", "", toks(1, ' ', " ", 3, I, "a", 4, '\n', "\n"), nil},
		{"  a\n", "", toks(1, ' ', " ", 3, I, "a", 4, '\n', "\n"), nil},
		{" ", "", toks(1, ' ', " ", 2, '\n', "\n"), white},
		{" \n", "", toks(1, ' ', " ", 2, '\n', "\n"), nil},
		{" a  ", "", toks(1, ' ', " ", 2, I, "a", 3, ' ', " ", 5, '\n', "\n"), nil},
		{" a  \n", "", toks(1, ' ', " ", 2, I, "a", 3, ' ', " ", 5, '\n', "\n"), nil},
		{" a ", "", toks(1, ' ', " ", 2, I, "a", 3, ' ', " ", 4, '\n', "\n"), nil},
		{" a \n", "", toks(1, ' ', " ", 2, I, "a", 3, ' ', " ", 4, '\n', "\n"), nil},
		{" a", "", toks(1, ' ', " ", 2, I, "a", 3, '\n', "\n"), nil},
		{" a\n", "", toks(1, ' ', " ", 2, I, "a", 3, '\n', "\n"), nil},
		{"", "", nil, nil},
		{"", "", toks(), white},
		{"/*", "unterminated", toks(1, ' ', "/*", 3, '\n', "\n"), white},
		{"/**", "unterminated", toks(1, ' ', "/**", 4, '\n', "\n"), white},
		{"/***", "unterminated", toks(1, ' ', "/***", 5, '\n', "\n"), white},
		{"/**/", "", toks(1, ' ', "/**/", 5, '\n', "\n"), white},
		{"/*\n*/", "", toks(1, ' ', " ", 6, '\n', "\n"), nil},
		{"/*x", "unterminated", toks(1, ' ', "/*x", 4, '\n', "\n"), white},
		{"/*x*", "unterminated", toks(1, ' ', "/*x*", 5, '\n', "\n"), white},
		{"/*xy", "unterminated", toks(1, ' ', "/*xy", 5, '\n', "\n"), white},
		{"/*xy*", "unterminated", toks(1, ' ', "/*xy*", 6, '\n', "\n"), white},
		{"// a /* b */ c", "", toks(1, ' ', " ", 15, '\n', "\n"), nil},
		{"// a /* b */ c", "", toks(1, ' ', "// a /* b */ c", 15, '\n', "\n"), white},
		{"\n", "", toks(1, '\n', "\n"), nil},
		{"\r\n", "", toks(1, '\n', "\r\n"), nil},
		{"a  ", "", toks(1, I, "a", 2, ' ', " ", 4, '\n', "\n"), nil},
		{"a  \n", "", toks(1, I, "a", 2, ' ', " ", 4, '\n', "\n"), nil},
		{"a  b", "", toks(1, I, "a", 2, ' ', "  ", 4, I, "b", 5, '\n', "\n"), white},
		{"a  b", "", toks(1, I, "a", 2, ' ', " ", 4, I, "b", 5, '\n', "\n"), nil},
		{"a ", "", toks(1, I, "a", 2, ' ', " ", 3, '\n', "\n"), nil},
		{"a /**/b", "", toks(1, I, "a", 2, ' ', " ", 7, I, "b", 8, '\n', "\n"), nil},
		{"a /**/b", "", toks(1, I, "a", 2, ' ', " /**/", 7, I, "b", 8, '\n', "\n"), white},
		{"a /*x*/b", "", toks(1, I, "a", 2, ' ', " ", 8, I, "b", 9, '\n', "\n"), nil},
		{"a /*x*/b", "", toks(1, I, "a", 2, ' ', " /*x*/", 8, I, "b", 9, '\n', "\n"), white},
		{"a //", "", toks(1, I, "a", 2, ' ', " ", 5, '\n', "\n"), nil},
		{"a //", "", toks(1, I, "a", 2, ' ', " //", 5, '\n', "\n"), white},
		{"a //\nb", "", toks(1, I, "a", 2, ' ', " ", 5, '\n', "\n", 6, I, "b", 7, '\n', "\n"), nil},
		{"a //\nb", "", toks(1, I, "a", 2, ' ', " //", 5, '\n', "\n", 6, I, "b", 7, '\n', "\n"), white},
		{"a //x", "", toks(1, I, "a", 2, ' ', " ", 6, '\n', "\n"), nil},
		{"a //x", "", toks(1, I, "a", 2, ' ', " //x", 6, '\n', "\n"), white},
		{"a //x\nb", "", toks(1, I, "a", 2, ' ', " ", 6, '\n', "\n", 7, I, "b", 8, '\n', "\n"), nil},
		{"a //x\nb", "", toks(1, I, "a", 2, ' ', " //x", 6, '\n', "\n", 7, I, "b", 8, '\n', "\n"), white},
		{"a /b", "", toks(1, I, "a", 2, ' ', " ", 3, '/', "/", 4, I, "b", 5, '\n', "\n"), nil},
		{"a /b", "", toks(1, I, "a", 2, ' ', " ", 3, '/', "/", 4, I, "b", 5, '\n', "\n"), white},
		{"a \n", "", toks(1, I, "a", 2, ' ', " ", 3, '\n', "\n"), nil},
		{"a b", "", toks(1, I, "a", 2, ' ', " ", 3, I, "b", 4, '\n', "\n"), nil},
		{"a b", "", toks(1, I, "a", 2, ' ', " ", 3, I, "b", 4, '\n', "\n"), white},
		{"a", "", toks(1, I, "a", 2, '\n', "\n"), nil},
		{"a/*", "unterminated", toks(1, I, "a", 2, ' ', " ", 4, '\n', "\n"), nil},
		{"a/**/ b", "", toks(1, I, "a", 2, ' ', " ", 7, I, "b", 8, '\n', "\n"), nil},
		{"a/**/ b", "", toks(1, I, "a", 2, ' ', "/**/ ", 7, I, "b", 8, '\n', "\n"), white},
		{"a/**//", "", toks(1, I, "a", 2, ' ', " ", 6, '/', "/", 7, '\n', "\n"), nil},
		{"a/**//", "", toks(1, I, "a", 2, ' ', "/**/", 6, '/', "/", 7, '\n', "\n"), white},
		{"a/**//**/b", "", toks(1, I, "a", 2, ' ', " ", 10, I, "b", 11, '\n', "\n"), nil},
		{"a/**//**/b", "", toks(1, I, "a", 2, ' ', "/**//**/", 10, I, "b", 11, '\n', "\n"), white},
		{"a/**//*x*/b", "", toks(1, I, "a", 2, ' ', " ", 11, I, "b", 12, '\n', "\n"), nil},
		{"a/**//*x*/b", "", toks(1, I, "a", 2, ' ', "/**//*x*/", 11, I, "b", 12, '\n', "\n"), white},
		{"a/**///", "", toks(1, I, "a", 2, ' ', " ", 8, '\n', "\n"), nil},
		{"a/**///", "", toks(1, I, "a", 2, ' ', "/**///", 8, '\n', "\n"), white},
		{"a/**///x", "", toks(1, I, "a", 2, ' ', " ", 9, '\n', "\n"), nil},
		{"a/**///x", "", toks(1, I, "a", 2, ' ', "/**///x", 9, '\n', "\n"), white},
		{"a/**/b", "", toks(1, I, "a", 2, ' ', " ", 6, I, "b", 7, '\n', "\n"), nil},
		{"a/**/b", "", toks(1, I, "a", 2, ' ', "/**/", 6, I, "b", 7, '\n', "\n"), white},
		{"a/*x*/ b", "", toks(1, I, "a", 2, ' ', " ", 8, I, "b", 9, '\n', "\n"), nil},
		{"a/*x*/ b", "", toks(1, I, "a", 2, ' ', "/*x*/ ", 8, I, "b", 9, '\n', "\n"), white},
		{"a/*x*//", "", toks(1, I, "a", 2, ' ', " ", 7, '/', "/", 8, '\n', "\n"), nil},
		{"a/*x*//", "", toks(1, I, "a", 2, ' ', "/*x*/", 7, '/', "/", 8, '\n', "\n"), white},
		{"a/*x*//**/b", "", toks(1, I, "a", 2, ' ', " ", 11, I, "b", 12, '\n', "\n"), nil},
		{"a/*x*//**/b", "", toks(1, I, "a", 2, ' ', "/*x*//**/", 11, I, "b", 12, '\n', "\n"), white},
		{"a/*x*//*y*/b", "", toks(1, I, "a", 2, ' ', " ", 12, I, "b", 13, '\n', "\n"), nil},
		{"a/*x*//*y*/b", "", toks(1, I, "a", 2, ' ', "/*x*//*y*/", 12, I, "b", 13, '\n', "\n"), white},
		{"a/*x*///", "", toks(1, I, "a", 2, ' ', " ", 9, '\n', "\n"), nil},
		{"a/*x*///", "", toks(1, I, "a", 2, ' ', "/*x*///", 9, '\n', "\n"), white},
		{"a/*x*/b", "", toks(1, I, "a", 2, ' ', " ", 7, I, "b", 8, '\n', "\n"), nil},
		{"a/*x*/b", "", toks(1, I, "a", 2, ' ', "/*x*/", 7, I, "b", 8, '\n', "\n"), white},
		{"a//", "", toks(1, I, "a", 2, ' ', " ", 4, '\n', "\n"), nil},
		{"a//\nb", "", toks(1, I, "a", 2, ' ', " ", 4, '\n', "\n", 5, I, "b", 6, '\n', "\n"), nil},
		{"a//\nb", "", toks(1, I, "a", 2, ' ', "//", 4, '\n', "\n", 5, I, "b", 6, '\n', "\n"), white},
		{"a//x\nb", "", toks(1, I, "a", 2, ' ', " ", 5, '\n', "\n", 6, I, "b", 7, '\n', "\n"), nil},
		{"a//x\nb", "", toks(1, I, "a", 2, ' ', "//x", 5, '\n', "\n", 6, I, "b", 7, '\n', "\n"), white},
		{"a\\\nb", "", toks(1, I, "ab", 5, '\n', "\n"), nil},
		{"a\\\nb\nc", "", toks(1, I, "ab", 5, '\n', "\n", 6, I, "c", 7, '\n', "\n"), nil},
		{"a\n", "", toks(1, I, "a", 2, '\n', "\n"), nil},
		{"a\nb", "", toks(1, I, "a", 2, '\n', "\n", 3, I, "b", 4, '\n', "\n"), nil},
		{"a\nb\n", "", toks(1, I, "a", 2, '\n', "\n", 3, I, "b", 4, '\n', "\n"), nil},
		{"a\r\n", "", toks(1, I, "a", 2, '\n', "\r\n"), nil},
		{"a\r\nb", "", toks(1, I, "a", 2, '\n', "\r\n", 4, I, "b", 5, '\n', "\n"), nil},
		{"ab", "", toks(1, I, "ab", 3, '\n', "\n"), nil},
		{"ab\n", "", toks(1, I, "ab", 3, '\n', "\n"), nil},
		{"ař", "", toks(1, I, "ař", 4, '\n', "\n"), nil},
		{"ařb", "", toks(1, I, "ařb", 5, '\n', "\n"), nil},
		{"ř", "", toks(1, I, "ř", 3, '\n', "\n"), nil},
		{"řa", "", toks(1, I, "řa", 4, '\n', "\n"), nil},
	} {
		buf := []byte(v.in)
		cfg := v.cfg
		if cfg == nil {
			cfg = &Config{}
		}
		file := tokenNewFile("x", len(buf))
		s := newScanner(newContext(cfg), bytes.NewReader(buf), file)
		s.ctx = newContext(cfg)
		var toks []token3
		for {
			s.lex()
			if s.tok.char < 0 {
				break
			}

			toks = append(toks, s.tok)
		}

		switch err := s.ctx.Err(); {
		case v.err != "" && err != nil:
			if !regexp.MustCompile(v.err).MatchString(err.Error()) {
				t.Error(i, err)
			}
		case v.err == "" && err != nil:
			t.Errorf("unexpected error: %v", err)
		case v.err != "" && err == nil:
			t.Errorf("%v: expected error matching %s", i, v.err)
		}
		if g, e := len(toks), len(v.toks); g != e {
			t.Logf("%v: %q", i, v.in)
			for j, v := range toks {
				t.Logf("%v: got %v: %s", i, j, v.str(file))
			}
			for j, v := range v.toks {
				t.Logf("%v: exp %v: %s", i, j, v.str(file))
			}
			t.Error(i, g, e)
			continue next
		}

		for j, gtok := range toks {
			gtok.src = 0      //TODO
			v.toks[j].src = 0 //TODO
			if etok := v.toks[j]; gtok != etok {
				t.Logf("%q", v.in)
				t.Errorf("%v, %v: %v %v", i, j, gtok.str(file), etok.str(file))
			}
		}
	}
}

func TestScanner3(t *testing.T) {
	var bytes int64
	var files int
	var m0, m1 runtime.MemStats
	cfg := &Config{}
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	limit := *oMaxFiles
	if err := filepath.Walk(sqliteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return skipDir(path)
		}

		if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if limit == 0 {
			return nil
		}

		cache = newPPCache()
		limit--
		files++
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}

		defer f.Close()

		sz := info.Size()
		bytes += sz
		s := newScanner(newContext(cfg), bufio.NewReader(f), tokenNewFile(path, int(sz)))
		for {
			s.lex()
			if s.tok.char < 0 {
				break
			}
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	dsz := 0
	for _, v := range dict.strings {
		dsz += len(v)
	}
	t.Logf("files %v, bytes %v, %v, %v B/s, mem %v, dict items %v, strings %v %v B/item", h(files),
		h(bytes), d, h(float64(time.Second)*float64(bytes)/float64(d)),
		h(m1.Alloc-m0.Alloc), h(len(dict.strings)), h(dsz), h(float64(dsz)/float64(len(dict.strings))))
	m := map[string]struct{}{}
	for _, v := range dict.strings {
		if _, ok := m[v]; ok {
			t.Fatalf("duplicate item %q", v)
		}

		m[v] = struct{}{}
	}
}

func BenchmarkScanner(b *testing.B) {
	var bytes int64
	cfg := &Config{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes = 0
		limit := *oMaxFiles
		if err := filepath.Walk(sqliteDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return skipDir(path)
			}

			if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" || info.Mode()&os.ModeType != 0 {
				return nil
			}

			if limit == 0 {
				return nil
			}

			limit--
			f, err := os.Open(path)
			if err != nil {
				b.Fatal(err)
			}

			defer f.Close()

			sz := info.Size()
			bytes += sz
			s := newScanner(newContext(cfg), bufio.NewReader(f), tokenNewFile(path, int(sz)))
			for {
				s.lex()
				if s.tok.char < 0 {
					break
				}
			}
			return nil
		}); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(bytes)
}

func TestScannerTranslationPhase3(t *testing.T) {
	cfg := &Config{}
	var bytes int64
	var files int
	var m0, m1 runtime.MemStats
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	limit := *oMaxFiles
	if err := filepath.Walk(sqliteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return skipDir(path)
		}

		if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if limit == 0 {
			return nil
		}

		cache = newPPCache()
		limit--
		files++
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}

		defer f.Close()

		sz := info.Size()
		bytes += sz
		ctx := newContext(cfg)
		newScanner(newContext(cfg), bufio.NewReader(f), tokenNewFile(path, int(sz))).translationPhase3()
		if err := ctx.Err(); err != nil {
			t.Error(err)
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	t.Logf("files %v, bytes %v, %v, %v B/s, mem %v",
		h(files), h(bytes), d, h(float64(time.Second)*float64(bytes)/float64(d)), h(m1.Alloc-m0.Alloc))
}

func TestScannerCSmith(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
		return
	}

	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	fixedBugs := []string{}
	ch := time.After(*oCSmith)
	t0 := time.Now()
	var files, ok int
	var size int64
out:
	for i := 0; ; i++ {
		extra := ""
		var args string
		switch {
		case i < len(fixedBugs):
			args += fixedBugs[i]
			a := strings.Split(fixedBugs[i], " ")
			extra = strings.Join(a[len(a)-2:], " ")
		default:
			select {
			case <-ch:
				break out
			default:
			}

			args += csmithArgs
		}
		out, err := exec.Command(csmith, strings.Split(args, " ")...).Output()
		if err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if fn := *oBlackBox; fn != "" {
			if err := ioutil.WriteFile(fn, out, 0660); err != nil {
				t.Fatal(err)
			}
		}

		cfg := &Config{Config3: Config3{MaxSourceLine: 1 << 20}}
		ctx := newContext(cfg)
		files++
		size += int64(len(out))
		newScanner(newContext(cfg), bytes.NewReader(out), tokenNewFile("test-scanner", len(out))).translationPhase3()
		if err := ctx.Err(); err != nil {
			t.Errorf("%s\n%s\n%v", extra, out, err)
			break
		}

		ok++
		if *oTrace {
			fmt.Fprintln(os.Stderr, time.Since(t0), files, ok)
		}
	}
	d := time.Since(t0)
	t.Logf("files %v, bytes %v, ok %v in %v", h(files), h(size), h(ok), d)
}

//TODO fuzz
