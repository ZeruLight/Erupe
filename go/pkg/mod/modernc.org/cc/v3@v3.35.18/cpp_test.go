// Copyright 2019 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v3"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestCPPExpand(t *testing.T) {
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	cfg := &Config{fakeIncludes: true, PreprocessOnly: true, RejectIncompatibleMacroRedef: true}
	if err := filepath.Walk(filepath.Join(testWD, filepath.FromSlash("testdata/cpp-expand/")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (!strings.HasSuffix(path, ".c") && !strings.HasSuffix(path, ".h")) {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		ctx := newContext(cfg)
		cf, err := cache.getFile(ctx, path, false, true)
		if err != nil {
			return err
		}

		cache = newPPCache()
		cpp := newCPP(ctx)
		var b strings.Builder
		expParth := path + ".expect"
		for line := range cpp.translationPhase4([]source{cf}) {
			for _, tok := range *line {
				b.WriteString(tok.String())
			}
			token4Pool.Put(line)
		}

		switch {
		case strings.Contains(filepath.ToSlash(path), "/mustfail/"):
			if err := ctx.Err(); err != nil {
				return nil
			}

			t.Fatalf("unexpected success: %s", path)
		default:
			if err := ctx.Err(); err != nil {
				t.Error(err)
			}
		}

		exp, err := ioutil.ReadFile(expParth)
		if err != nil {
			t.Error(err)
		}

		if g, e := b.String(), string(exp); g != e {
			a := strings.Split(g, "\n")
			b := strings.Split(e, "\n")
			n := len(a)
			if len(b) > n {
				n = len(b)
			}
			for i := 0; i < n; i++ {
				var x, y string
				if i < len(a) {
					x = a[i]
				}
				if i < len(b) {
					y = b[i]
				}
				x = strings.ReplaceAll(x, "\r", "")
				y = strings.ReplaceAll(y, "\r", "")
				if x != y {
					t.Errorf("%s:%v: %v", path, i+1, cmp.Diff(y, x))
				}
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestPreprocess(t *testing.T) {
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	if err := filepath.Walk(filepath.Join(testWD, filepath.FromSlash("testdata/preprocess")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (!strings.HasSuffix(path, ".c") && !strings.HasSuffix(path, ".h")) {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		var b strings.Builder
		if err = Preprocess(&Config{}, nil, nil, []Source{{Name: path}}, &b); err != nil {
			return err
		}

		expParth := path + ".expect"
		exp, err := ioutil.ReadFile(expParth)
		if err != nil {
			return err
		}

		if g, e := b.String(), string(exp); g != e {
			a := strings.Split(g, "\n")
			b := strings.Split(e, "\n")
			n := len(a)
			if len(b) > n {
				n = len(b)
			}
			for i := 0; i < n; i++ {
				var x, y string
				if i < len(a) {
					x = a[i]
				}
				if i < len(b) {
					y = b[i]
				}
				x = strings.ReplaceAll(x, "\r", "")
				y = strings.ReplaceAll(y, "\r", "")
				if x != y {
					t.Errorf("%s:%v: %v", path, i+1, cmp.Diff(y, x))
				}
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

}

func TestTCCExpand(t *testing.T) {
	blacklist := map[string]struct{}{}
	mustFail := map[string]string{
		"16.c": "redefinition",
	}
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	cfg := &Config{fakeIncludes: true, PreprocessOnly: true, RejectIncompatibleMacroRedef: true}
	files := 0
	if err := filepath.Walk(filepath.Join(root, filepath.FromSlash("tests/pp")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (!strings.HasSuffix(path, ".c") && !strings.HasSuffix(path, ".S")) {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		cache = newPPCache()
		files++
		if *oTrace {
			fmt.Fprintln(os.Stderr, files, path)
		}

		ctx := newContext(cfg)
		cf, err := cache.getFile(ctx, path, false, true)
		if err != nil {
			return err
		}

		cpp := newCPP(ctx)
		var b strings.Builder
		expParth := path[:len(path)-len(filepath.Ext(path))] + ".expect"
		for line := range cpp.translationPhase4([]source{cf}) {
			for _, tok := range *line {
				b.WriteString(tok.String())
			}
			token4Pool.Put(line)
		}

		if err := ctx.Err(); err != nil {
			if re := mustFail[filepath.Base(path)]; re != "" {
				if regexp.MustCompile(re).MatchString(err.Error()) {
					return nil
				}
			}

			t.Error(err)
		}
		exp, err := ioutil.ReadFile(expParth)
		if err != nil {
			t.Error(err)
		}

		g, e := b.String(), string(exp)
		a := strings.Split(g, "\n")
		w := 0
		for _, v := range a {
			if strings.TrimSpace(v) == "" {
				continue
			}

			a[w] = v
			w++
		}
		g = strings.Join(a[:w], "\n")
		switch filepath.Base(path) {
		case "02.c", "05.c":
			g = strings.ReplaceAll(g, " ", "")
			e = strings.ReplaceAll(e, " ", "")
		}
		//dbg("\ngot:\n%s\nexp:\n%s", g, e)
		if g != e {
			ok := true
			a := strings.Split(g, "\n")
			b := strings.Split(e, "\n")
			n := len(a)
			if len(b) > n {
				n = len(b)
			}
			for i := 0; i < n; i++ {
				var x, y string
				if i < len(a) {
					x = a[i]
				}
				if i < len(b) {
					y = b[i]
				}
				x = strings.TrimSpace(x)
				y = strings.TrimSpace(y)
				for n := len(x); ; {
					x = strings.ReplaceAll(x, "  ", " ")
					if len(x) == n {
						break
					}

					n = len(x)
				}
				for n := len(y); ; {
					y = strings.ReplaceAll(y, "  ", " ")
					if len(y) == n {
						break
					}

					n = len(y)
				}
				if x != y {
					ok = false
					t.Errorf("%s:%v: %v", path, i+1, cmp.Diff(y, x))
				}
			}
			if !ok {
				t.Errorf("\ngot:\n%s\nexp:\n%s", g, e)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTranslationPhase4(t *testing.T) {
	t.Run("shell.c", func(t *testing.T) { testTranslationPhase4(t, testPredefSource, testShellSource) })
	t.Run("sqlite3.c", func(t *testing.T) { testTranslationPhase4(t, testPredefSource, testSQLiteSource) })
}

func testTranslationPhase4(t *testing.T, predef, src source) {
	sources := []source{predef, testBuiltinSource, src}
	cfg := &Config{}
	ctx := newContext(cfg)
	ctx.includePaths = testIncludes
	ctx.sysIncludePaths = testSysIncludes
	cpp := newCPP(ctx)
	var m0, m1 runtime.MemStats
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	for line := range cpp.translationPhase4(sources) {
		token4Pool.Put(line)
	}
	if err := ctx.Err(); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	t.Logf("sources %v, bytes %v, %v, %v B/s, mem %v",
		h(ctx.tuSources()), h(ctx.tuSize()), d, h(float64(time.Second)*float64(ctx.tuSize())/float64(d)), h(m1.Alloc-m0.Alloc))
}

func BenchmarkTranslationPhase4(b *testing.B) {
	b.Run("shell.c", func(b *testing.B) { benchmarkTranslationPhase4(b, testPredefSource, testShellSource) })
	b.Run("sqlite3.c", func(b *testing.B) { benchmarkTranslationPhase4(b, testPredefSource, testSQLiteSource) })
}

func benchmarkTranslationPhase4(b *testing.B, predef, src source) {
	sources := []source{predef, testBuiltinSource, src}
	cfg := &Config{}
	var ctx *context
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx = newContext(cfg)
		ctx.includePaths = testIncludes
		ctx.sysIncludePaths = testSysIncludes
		cpp := newCPP(ctx)
		for line := range cpp.translationPhase4(sources) {
			token4Pool.Put(line)
		}
		if err := ctx.Err(); err != nil {
			b.Error(err)
		}
	}
	b.SetBytes(ctx.tuSize())
}

func TestMacroPosition(t *testing.T) {
	cfg := &Config{PreprocessOnly: true}
	ast, err := Parse(cfg, nil, nil, []Source{{Name: "test", Value: `
/* noise to make position more interesting */
 # define foo 42
`, DoNotCache: true}})
	if err != nil {
		t.Fatal(err)
	}
	m := ast.Macros[String("foo")]
	pos := m.Position()
	if g, e := pos.String(), `test:3:2`; g != e {
		t.Errorf("bad position: %q != %q: %#v", g, e, pos)
	}
}

// https://gitlab.com/cznic/cc/-/issues/127
func TestIssue127(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	if err := os.Chdir(filepath.FromSlash("testdata/issue127/")); err != nil {
		t.Error(err)
		return
	}

	cd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("working directory: %s", cd)
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Error(err)
		return
	}

	ast, err := Translate(
		&Config{ABI: abi},
		[]string{"include"},
		nil,
		[]Source{{Name: "main.c"}},
	)
	if err != nil {
		t.Error(err)
	}

	fd := ast.Scope[String("getopt")]
	if len(fd) == 0 {
		t.Errorf("cannot find getopt")
		return
	}

	switch x := fd[0].(type) {
	case *Declarator:
		t.Logf("getopt C type: %s", x.Type())
	default:
		t.Errorf("unexpected getopt Go type: %T", x)
	}
}
