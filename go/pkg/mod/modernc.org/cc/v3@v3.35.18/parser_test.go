// Copyright 2019 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v3"

import (
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
)

func TestBOM(t *testing.T) {
	for i, v := range []struct {
		src string
		err string
	}{
		{"int main() {}", ""},
		{"\xEF\xBB\xBFint main() {}", ""},
	} {
		switch _, err := Parse(&Config{}, nil, nil, []Source{{Value: v.src, DoNotCache: true}}); {
		case v.err == "" && err != nil:
			t.Errorf("%v: unexpected error %v", i, err)
		case v.err != "" && err == nil:
			t.Errorf("%v: unexpected success, expected error matching %v", i, v.err)
		case v.err != "":
			if !regexp.MustCompile(v.err).MatchString(err.Error()) {
				t.Errorf("%v: error %v does not match %v", i, err, v.err)
			}
		}
	}
}

func TestStrCatSep(t *testing.T) {
	cfg := &Config{Config3: Config3{PreserveWhiteSpace: true}, PreprocessOnly: true}
	for i, v := range []struct {
		src         string
		lit         string
		sep         string
		trailingSep string
	}{
		{`int f() {  "a";}`, "a", "  ", "\n"},
		{`int f() { "a" "b";}`, "ab", "  ", "\n"},
		{`int f() { "a""b";}`, "ab", " ", "\n"},
		{`int f() { "a"` + "\n\t" + `"b"; }`, "ab", " \n\t", "\n"},
		{`int f() { "a";}`, "a", " ", "\n"},
		{`int f() { /*x*/ /*y*/ "a";}`, "a", " /*x*/ /*y*/ ", "\n"},
		{`int f() { /*x*/` + "\n" + `/*y*/ "a";}`, "a", " /*x*/\n/*y*/ ", "\n"},
		{`int f() { //x` + "\n" + ` "a";}`, "a", " //x\n ", "\n"},
		{`int f() { //x` + "\n" + `"a";}`, "a", " //x\n", "\n"},
		{`int f() { ` + "\n" + ` "a";}`, "a", " \n ", "\n"},
		{`int f() { ` + "\n" + `"a";}`, "a", " \n", "\n"},
		{`int f() {"a" "b";}`, "ab", " ", "\n"},
		{`int f() {"a"/*y*/"b";}`, "ab", "/*y*/", "\n"},
		{`int f() {"a";} /*x*/ `, "a", "", " /*x*/ \n"},
		{`int f() {"a";} /*x*/`, "a", "", " /*x*/\n"},
		{`int f() {"a";} /*x` + "\n" + `*/ `, "a", "", " /*x\n*/ \n"},
		{`int f() {"a";} `, "a", "", " \n"},
		{`int f() {"a";}/*x*/`, "a", "", "/*x*/\n"},
		{`int f() {"a";}` + "\n", "a", "", "\n"},
		{`int f() {"a";}`, "a", "", "\n"},
		{`int f() {/*x*/ /*y*/ "a";}`, "a", "/*x*/ /*y*/ ", "\n"},
		{`int f() {/*x*/"a""b";}`, "ab", "/*x*/", "\n"},
		{`int f() {/*x*/"a"/*y*/"b";}`, "ab", "/*x*//*y*/", "\n"},
		{`int f() {/*x*/"a";}`, "a", "/*x*/", "\n"},
		{`int f() {/*x*//*y*/ "a";}`, "a", "/*x*//*y*/ ", "\n"},
		{`int f() {/*x*//*y*/"a";}`, "a", "/*x*//*y*/", "\n"},
		{`int f() {//` + "\n" + `"a";}`, "a", "//\n", "\n"},
		{`int f() {//x` + "\n" + `"a";}`, "a", "//x\n", "\n"},
		{`int f() {` + "\n" + ` "a";}`, "a", "\n ", "\n"},
		{`int f() {` + "\n" + `"a";}`, "a", "\n", "\n"},
	} {
		ast, err := Parse(cfg, nil, nil, []Source{{Name: "test", Value: v.src, DoNotCache: true}})
		if err != nil {
			t.Errorf("%v: %v", i, err)
			continue
		}

		tok := ast.
			TranslationUnit.
			ExternalDeclaration.
			FunctionDefinition.
			CompoundStatement.
			BlockItemList.
			BlockItem.
			Statement.
			ExpressionStatement.
			Expression.
			AssignmentExpression.
			ConditionalExpression.
			LogicalOrExpression.
			LogicalAndExpression.
			InclusiveOrExpression.
			ExclusiveOrExpression.
			AndExpression.
			EqualityExpression.
			RelationalExpression.
			ShiftExpression.
			AdditiveExpression.
			MultiplicativeExpression.
			CastExpression.
			UnaryExpression.
			PostfixExpression.
			PrimaryExpression.
			Token
		if g, e := tok.String(), v.lit; g != e {
			t.Errorf("%v: %q %q", i, g, e)
		}
		if g, e := tok.Sep.String(), v.sep; g != e {
			t.Errorf("%v: %q %q", i, g, e)
		}
		if g, e := ast.TrailingSeperator.String(), v.trailingSep; g != e {
			t.Errorf("%v: %q %q", i, g, e)
		}
	}
}

func TestParseJhjourdan(t *testing.T) {
	mustFail := map[string]struct{}{
		"dangling_else_misleading.fail.c": {},
		"atomic_parenthesis.c":            {}, // See [3], 3.5, pg. 20.
		//TODO Type checking needed to actually fail "bitfield_declaration_ambiguity.fail.c":  {},
	}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	cfg := &Config{
		ignoreIncludes:             true,
		ignoreUndefinedIdentifiers: true,
	}
	var ok, n int
	if err := filepath.Walk(filepath.Join(testWD, filepath.FromSlash("testdata/jhjourdan/")), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".c") {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		cache = newPPCache()
		n++
		_, expectFail := mustFail[filepath.Base(path)]
		switch _, err := Parse(cfg, nil, nil, []Source{{Name: path, DoNotCache: true}}); {
		case err != nil:
			if !expectFail {
				t.Errorf("FAIL: unexpected error: %v", err)
				return nil
			}
		default:
			if expectFail {
				t.Errorf("FAIL: %v: unexpected success", path)
				return nil
			}
		}

		ok++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("jhjourdan parse: ok %v/%v\n", ok, n)
}

func TestParseSQLite(t *testing.T) {
	cfg := &Config{}
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	t.Run("shell.c", func(t *testing.T) { testParse(t, cfg, testPredef, filepath.Join(root, "shell.c")) })
	t.Run("sqlite3.c", func(t *testing.T) { testParse(t, cfg, testPredef, filepath.Join(root, "sqlite3.c")) })
}

var testParseAST *AST

func testParse(t *testing.T, cfg *Config, predef string, files ...string) {
	testParseAST = nil
	sources := []Source{
		{Name: "<predefined>", Value: predef, DoNotCache: true},
		{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
	}
	for _, v := range files {
		sources = append(sources, Source{Name: v, DoNotCache: true})
	}
	ctx := newContext(cfg)
	var m0, m1 runtime.MemStats
	var err error
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	if testParseAST, err = parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	t.Logf("sources %v, bytes %v, %v, %v B/s, mem %v",
		h(ctx.tuSources()), h(ctx.tuSize()), d, h(float64(time.Second)*float64(ctx.tuSize())/float64(d)), h(m1.Alloc-m0.Alloc))
}

func BenchmarkParseSQLite(b *testing.B) {
	cfg := &Config{}
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	b.Run("shell.c", func(b *testing.B) { benchmarkParseSQLite(b, cfg, testPredef, filepath.Join(root, "shell.c")) })
	b.Run("sqlite3.c", func(b *testing.B) { benchmarkParseSQLite(b, cfg, testPredef, filepath.Join(root, "sqlite3.c")) })
}

func benchmarkParseSQLite(b *testing.B, cfg *Config, predef string, files ...string) {
	sources := []Source{
		{Name: "<predefined>", Value: predef, DoNotCache: true},
		{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
	}
	for _, v := range files {
		sources = append(sources, Source{Name: v, DoNotCache: true})
	}
	ctx := newContext(cfg)
	// Warm up the cache
	if _, err := parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
		b.Error(err)
		return
	}

	sz := ctx.tuSize()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Parse(cfg, testIncludes, testSysIncludes, sources); err != nil {
			b.Error(err)
			return
		}
	}
	b.SetBytes(sz)
}

func TestParseTCC(t *testing.T) {
	cfg := &Config{
		ignoreUndefinedIdentifiers: true,
	}
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	ok := 0
	const dir = "tests/tests2"
	t.Run(dir, func(t *testing.T) {
		ok += testParseDir(t, cfg, testPredef, filepath.Join(root, filepath.FromSlash(dir)), false, true)
	})
	t.Logf("ok %v", h(ok))
}

func TestParseGCC(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
		return
	}

	cfg := &Config{
		ignoreUndefinedIdentifiers: true,
	}
	root := filepath.Join(testWD, filepath.FromSlash(gccDir))
	ok := 0
	for _, v := range []string{
		"gcc/testsuite/gcc.c-torture/compile",
		"gcc/testsuite/gcc.c-torture/execute",
	} {
		t.Run(v, func(t *testing.T) {
			ok += testParseDir(t, cfg, testPredef, filepath.Join(root, filepath.FromSlash(v)), true, true)
		})
	}
	t.Logf("ok %v", h(ok))
}

func testParseDir(t *testing.T, cfg *Config, predef, dir string, hfiles, must bool) (ok int) {
	blacklist := map[string]struct{}{ //TODO-
		"90_struct-init.c":  {}, //TODO [ x ... y ] designator
		"94_generic.c":      {},
		"95_bitfields.c":    {},
		"95_bitfields_ms.c": {},
		"99_fastcall.c":     {},
	}
	if isTestingMingw {
		blacklist["loop-2f.c"] = struct{}{} // sys/mman.h
		blacklist["loop-2g.c"] = struct{}{} // sys/mman.h
	}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	var files, psources int
	var bytes int64
	var m0, m1 runtime.MemStats
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return err
		}

		if info.IsDir() {
			if path == dir {
				return nil
			}

			return skipDir(path)
		}

		if filepath.Ext(path) != ".c" && (!hfiles || filepath.Ext(path) != ".h") || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		cache = newPPCache()
		files++
		if re != nil && !re.MatchString(path) {
			ok++
			return nil
		}

		sources := []Source{
			{Name: "<predefined>", Value: predef, DoNotCache: true},
			{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
			{Name: path, DoNotCache: true},
		}
		ctx := newContext(cfg)

		defer func() {
			psources += ctx.tuSources()
			bytes += ctx.tuSize()
		}()

		if *oTrace {
			fmt.Fprintln(os.Stderr, files, path)
		}
		if _, err := parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
			if must {
				t.Error(err)
			}
			return nil
		}

		ok++
		return nil
	}); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	t.Logf("files %v, sources %v, bytes %v, ok %v, %v, %v B/s, mem %v",
		h(files), h(psources), h(bytes), h(ok), d, h(float64(time.Second)*float64(bytes)/float64(d)), h(m1.Alloc-m0.Alloc))
	if files != ok && must {
		t.Errorf("files %v, bytes %v, ok %v", files, bytes, ok)
	}
	return ok
}

func BenchmarkParseTCC(b *testing.B) {
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	cfg := &Config{
		ignoreUndefinedIdentifiers: true,
	}
	const dir = "tests/tests2"
	b.Run(dir, func(b *testing.B) {
		benchmarkParseDir(b, cfg, testPredef, filepath.Join(root, filepath.FromSlash(dir)), false)
	})
}

func BenchmarkParseGCC(b *testing.B) {
	root := filepath.Join(testWD, filepath.FromSlash(gccDir))
	cfg := &Config{
		ignoreUndefinedIdentifiers: true,
	}
	for _, v := range []string{
		"gcc/testsuite/gcc.c-torture/compile",
		"gcc/testsuite/gcc.c-torture/execute",
	} {
		b.Run(v, func(b *testing.B) {
			benchmarkParseDir(b, cfg, testPredef, filepath.Join(root, filepath.FromSlash(v)), true)
		})
	}
}

func benchmarkParseDir(b *testing.B, cfg *Config, predef, dir string, must bool) {
	var bytes int64
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytes = 0
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					err = nil
				}
				return err
			}

			if info.IsDir() {
				return skipDir(path)
			}

			if filepath.Ext(path) != ".c" && filepath.Ext(path) != ".h" || info.Mode()&os.ModeType != 0 {
				return nil
			}

			sources := []Source{
				{Name: "<predefined>", Value: predef, DoNotCache: true},
				{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
				{Name: path, DoNotCache: true},
			}
			ctx := newContext(cfg)
			if _, err := parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
				if must {
					b.Error(err)
				}
			}
			bytes += ctx.tuSize()
			return nil
		}); err != nil {
			b.Error(err)
		}
	}
	b.SetBytes(bytes)
}

func ExampleInitDeclaratorList_uCN() {
	fmt.Println(exampleAST(0, `int a·z, a\u00b7z;`))
	// Output:
	// &cc.InitDeclaratorList{
	// · InitDeclarator: &cc.InitDeclarator{
	// · · Case: InitDeclaratorDecl,
	// · · Declarator: &cc.Declarator{
	// · · · DirectDeclarator: &cc.DirectDeclarator{
	// · · · · Case: DirectDeclaratorIdent,
	// · · · · Token: example.c:1:5: IDENTIFIER "a·z",
	// · · · },
	// · · },
	// · },
	// · InitDeclaratorList: &cc.InitDeclaratorList{
	// · · InitDeclarator: &cc.InitDeclarator{
	// · · · Case: InitDeclaratorDecl,
	// · · · Declarator: &cc.Declarator{
	// · · · · DirectDeclarator: &cc.DirectDeclarator{
	// · · · · · Case: DirectDeclaratorIdent,
	// · · · · · Token: example.c:1:11: IDENTIFIER "a·z",
	// · · · · },
	// · · · },
	// · · },
	// · · Token: example.c:1:9: ',' ",",
	// · },
	// }
}

func ExampleDirectDeclarator_line() {
	fmt.Println(exampleAST(0, "#line 1234\nint i;"))
	// Output:
	// &cc.DirectDeclarator{
	// · Case: DirectDeclaratorIdent,
	// · Token: example.c:1234:5: IDENTIFIER "i",
	// }
}

func ExampleDirectDeclarator_line2() {
	fmt.Println(exampleAST(0, "#line 1234 \"foo.c\"\nint i;"))
	// Output:
	// &cc.DirectDeclarator{
	// · Case: DirectDeclaratorIdent,
	// · Token: foo.c:1234:5: IDENTIFIER "i",
	// }
}

func ExampleDirectDeclarator_line3() {
	fmt.Println(exampleAST(0, "#line 1234\r\nint i;"))
	// Output:
	// &cc.DirectDeclarator{
	// · Case: DirectDeclaratorIdent,
	// · Token: example.c:1234:5: IDENTIFIER "i",
	// }
}

func ExampleDirectDeclarator_line4() {
	fmt.Println(exampleAST(0, "#line 1234 \"foo.c\"\r\nint i;"))
	// Output:
	// &cc.DirectDeclarator{
	// · Case: DirectDeclaratorIdent,
	// · Token: foo.c:1234:5: IDENTIFIER "i",
	// }
}

func ExamplePrimaryExpression_stringLiteral() {
	fmt.Println(exampleAST(0, "char s[] = \"a\"\n\"b\"\n\"c\";"))
	// Output:
	// &cc.PrimaryExpression{
	// · Case: PrimaryExpressionString,
	// · Token: example.c:1:12: STRINGLITERAL "abc",
	// }
}

func TestParserCSmith(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
		return
	}

	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	fixedBugs := []string{
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 1236173074", //TODO fails on darwin/amd64
	}
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

		cfg := &Config{Config3: Config3{MaxSourceLine: 1 << 19}}
		ctx := newContext(cfg)
		files++
		size += int64(len(out))
		sources := []Source{
			{Name: "<predefined>", Value: testPredef, DoNotCache: true},
			{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
			{Name: "test.c", Value: string(out), DoNotCache: true},
		}
		if _, err := parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
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
