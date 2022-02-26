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

	"modernc.org/mathutil"
)

func TestTranslateSQLite(t *testing.T) {
	if runtime.GOOS == "netbsd" {
		t.Skip("TODO") //TODO
	}

	cfg := &Config{ABI: testABI, EnableAssignmentCompatibilityChecking: true}
	if isTestingMingw {
		cfg.DoNotTypecheckAsm = true
	}
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	t.Run("shell.c", func(t *testing.T) { testTranslate(t, cfg, testPredef, filepath.Join(root, "shell.c")) })
	t.Run("sqlite3.c", func(t *testing.T) { testTranslate(t, cfg, testPredef, filepath.Join(root, "sqlite3.c")) })
}

var (
	benchmarkTranslateSQLiteAST *AST
	testTranslateSQLiteAST      *AST
)

func testTranslate(t *testing.T, cfg *Config, predef string, files ...string) {
	testTranslateSQLiteAST = nil
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
	if testTranslateSQLiteAST, err = translate(ctx, testIncludes, testSysIncludes, sources); err != nil {
		t.Error(err)
	}
	d := time.Since(t0)
	debug.FreeOSMemory()
	runtime.ReadMemStats(&m1)
	t.Logf("sources %v, bytes %v, %v, %v B/s, mem %v",
		h(ctx.tuSources()), h(ctx.tuSize()), d, h(float64(time.Second)*float64(ctx.tuSize())/float64(d)), h(m1.Alloc-m0.Alloc))
}

func BenchmarkTranslateSQLite(b *testing.B) {
	cfg := &Config{ABI: testABI}
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	b.Run("shell.c", func(b *testing.B) { benchmarkTranslateSQLite(b, cfg, testPredef, filepath.Join(root, "shell.c")) })
	b.Run("sqlite3.c", func(b *testing.B) { benchmarkTranslateSQLite(b, cfg, testPredef, filepath.Join(root, "sqlite3.c")) })
}

func benchmarkTranslateSQLite(b *testing.B, cfg *Config, predef string, files ...string) {
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
	var err error
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if benchmarkTranslateSQLiteAST, err = Translate(cfg, testIncludes, testSysIncludes, sources); err != nil {
			b.Error(err)
		}
	}
	b.SetBytes(sz)
}

var (
	benchmarkTranslateAST *AST
	testTranslateAST      *AST
)

func TestTranslateTCC(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
		return
	}

	cfg := &Config{
		ABI:                                   testABI,
		ignoreUndefinedIdentifiers:            true,
		EnableAssignmentCompatibilityChecking: true,
	}
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	ok := 0
	const dir = "tests/tests2"
	t.Run(dir, func(t *testing.T) {
		ok += testTranslateDir(t, cfg, testPredef, filepath.Join(root, filepath.FromSlash(dir)), false)
	})
	t.Logf("ok %v", h(ok))
}

func TestTranslateGCC(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
		return
	}

	cfg := &Config{
		ABI:                                   testABI,
		ignoreUndefinedIdentifiers:            true,
		EnableAssignmentCompatibilityChecking: true,
	}
	root := filepath.Join(testWD, filepath.FromSlash(gccDir))
	ok := 0
	for _, v := range []string{
		"gcc/testsuite/gcc.c-torture/compile",
		"gcc/testsuite/gcc.c-torture/execute",
	} {
		t.Run(v, func(t *testing.T) {
			ok += testTranslateDir(t, cfg, testPredef, filepath.Join(root, filepath.FromSlash(v)), true)
		})
	}
	t.Logf("ok %v", h(ok))
}

func testTranslateDir(t *testing.T, cfg *Config, predef, dir string, hfiles bool) (ok int) {
	blacklist := map[string]struct{}{ //TODO-
		// TCC
		"34_array_assignment.c": {}, // gcc: 16:6: error: assignment to expression with array type
		"90_struct-init.c":      {}, //TODO [ x ... y ] designator
		"94_generic.c":          {},

		// GCC
		"20000120-2.c":                 {}, //TODO function redefinition
		"20000804-1.c":                 {}, //TODO 1: unsupported type: complex long long
		"20021120-1.c":                 {}, //TODO function redefinition
		"20021120-2.c":                 {}, //TODO function redefinition
		"20041124-1.c":                 {}, //TODO complex num
		"20041201-1.c":                 {}, //TODO complex num
		"20050122-2.c":                 {}, //TODO goto from nested function to outer function label
		"20050215-1.c":                 {}, //TODO function redefinition
		"20050215-2.c":                 {}, //TODO function redefinition
		"20050215-3.c":                 {}, //TODO function redefinition
		"920415-1.c":                   {}, //TODO label l undefined
		"920428-2.c":                   {}, //TODO goto from nested function to outer function label
		"920501-7.c":                   {}, //TODO goto from nested function to outer function label
		"920721-4.c":                   {}, //TODO label default_lab undefined
		"builtin-types-compatible-p.c": {}, //TODO
		"comp-goto-2.c":                {}, //TODO goto from nested function to outer function label
		"complex-1.c":                  {}, //TODO complex num
		"complex-5.c":                  {}, //TODO 9: unsupported type: complex int
		"complex-6.c":                  {}, //TODO complex num
		"nestfunc-5.c":                 {}, //TODO goto from nested function to outer function label
		"nestfunc-6.c":                 {}, //TODO goto from nested function to outer function label
		"pr21728.c":                    {}, //TODO goto from nested function to outer function label
		"pr24135.c":                    {}, //TODO goto from nested function to outer function label
		"pr27889.c":                    {}, //TODO 1: unsupported type: complex int
		"pr35431.c":                    {}, //TODO 3: unsupported type: complex int
		"pr38151.c":                    {}, //TODO 3: unsupported type: complex int
		"pr41987.c":                    {}, //TODO 3: unsupported type: complex char
		"pr51447.c":                    {}, //TODO goto from nested function to outer function label
		"pr56837.c":                    {}, //TODO 1: unsupported type: complex int
		"pr80692.c":                    {}, //TODO strconv.ParseFloat: parsing "0.DD": invalid syntax
		"pr86122.c":                    {}, //TODO 1: unsupported type: complex int
		"pr86123.c":                    {}, //TODO 6: unsupported type: complex unsigned
	}
	if isTestingMingw {
		blacklist["loop-2f.c"] = struct{}{} // sys/mman.h
		blacklist["loop-2g.c"] = struct{}{} // sys/mman.h
	}
	if cfg.ABI.Types[Ptr].Size == 4 {
		blacklist["pr70355.c"] = struct{}{} // /* { dg-require-effective-target int128 } */
	}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	var files, psources int
	var bytes int64
	var m0, m1 runtime.MemStats
	testTranslateAST = nil
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
			return skipDir(path)
		}

		if filepath.Ext(path) != ".c" && (!hfiles || filepath.Ext(path) != ".h") || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		cache = newPPCache()
		if strings.Contains(filepath.Base(path), "limits-") {
			debug.FreeOSMemory()
		}
		files++
		if re != nil && !re.MatchString(path) {
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
			fmt.Fprintln(os.Stderr, files, ok, path)
		}
		if testTranslateAST, err = parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
			t.Error(err)
			return nil
		}

		if err = testTranslateAST.Typecheck(); err != nil {
			t.Error(err)
			return nil
		}
		if strings.Contains(filepath.Base(path), "limits-") {
			debug.FreeOSMemory()
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
	if files != ok {
		t.Errorf("files %v, bytes %v, ok %v", files, bytes, ok)
	}
	return ok
}

func BenchmarkTranslateTCC(b *testing.B) {
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	cfg := &Config{
		ABI:                        testABI,
		ignoreUndefinedIdentifiers: true,
	}
	const dir = "tests/tests2"
	b.Run(dir, func(b *testing.B) {
		benchmarkTranslateDir(b, cfg, testPredef, filepath.Join(root, filepath.FromSlash(dir)), false)
	})
}

func BenchmarkTranslateGCC(b *testing.B) {
	root := filepath.Join(testWD, filepath.FromSlash(gccDir))
	cfg := &Config{
		ABI:                        testABI,
		ignoreUndefinedIdentifiers: true,
	}
	for _, v := range []string{
		"gcc/testsuite/gcc.c-torture/compile",
		"gcc/testsuite/gcc.c-torture/execute",
	} {
		b.Run(v, func(b *testing.B) {
			benchmarkTranslateDir(b, cfg, testPredef, filepath.Join(root, filepath.FromSlash(v)), true)
		})
	}
}

func benchmarkTranslateDir(b *testing.B, cfg *Config, predef, dir string, must bool) {
	blacklist := map[string]struct{}{ //TODO-
		// TCC
		"13_integer_literals.c":        {},
		"70_floating_point_literals.c": {},

		// GCC/exec
		"pr80692.c": {}, // Decimal64 literals
	}
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

			if _, ok := blacklist[filepath.Base(path)]; ok {
				return nil
			}

			sources := []Source{
				{Name: "<predefined>", Value: predef, DoNotCache: true},
				{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
				{Name: path, DoNotCache: true},
			}
			ctx := newContext(cfg)
			if benchmarkTranslateAST, err = parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
				if must {
					b.Error(err)
				}
				return nil
			}

			if err = benchmarkTranslateAST.Typecheck(); err != nil {
				if must {
					b.Error(err)
				}
				return nil
			}
			bytes += ctx.tuSize()
			return nil
		}); err != nil {
			b.Error(err)
		}
	}
	b.SetBytes(bytes)
}

func TestAbstractDeclarator(t *testing.T) { //TODO -> Example
	for i, test := range []struct{ src, typ string }{
		{"int i = sizeof(int);", "int"},                                                                                            // [0], 6.7.6, 3, (a)
		{"int i = sizeof(int*);", "pointer to int"},                                                                                // [0], 6.7.6, 3, (b)
		{"int i = sizeof(int*[3]);", "array of 3 pointer to int"},                                                                  // [0], 6.7.6, 3, (c)
		{"int i = sizeof(int(*)[3]);", "pointer to array of 3 int"},                                                                // [0], 6.7.6, 3, (d)
		{"int i = sizeof(int(*)[*]);", "pointer to array of int"},                                                                  // [0], 6.7.6, 3, (e)
		{"int i = sizeof(int *());", "function() returning pointer to int"},                                                        // [0], 6.7.6, 3, (f)
		{"int i = sizeof(int (*)(void));", "pointer to function(void) returning int"},                                              // [0], 6.7.6, 3, (g)
		{"int i = sizeof(int (*[])(unsigned int, ...));", "array of pointer to function(unsigned, ...) returning int"},             // [0], 6.7.6, 3, (h)
		{"int i = sizeof(int (*const [])(unsigned int, ...));", "array of const pointer to function(unsigned, ...) returning int"}, // [0], 6.7.6, 3, (h)
	} {
		letter := string(rune('a' + i))
		cfg := &Config{ABI: testABI, doNotSanityCheckComplexTypes: true}
		ast, err := Translate(cfg, nil, nil, []Source{
			{Name: "<built-in>", Value: "typedef long long unsigned size_t;", DoNotCache: true},
			{Name: "test", Value: test.src, DoNotCache: true},
		})
		if err != nil {
			t.Errorf("(%v): %v", letter, err)
			continue
		}

		var node Node
		depth := mathutil.MaxInt
		findNode("TypeName", ast.TranslationUnit, 0, &node, &depth)
		if node == nil {
			t.Errorf("(%v): %q, TypeName node not found", letter, test.src)
			continue
		}

		g, e := node.(*TypeName).Type().String(), test.typ
		if *oTrace {
			t.Logf("(%v): %q, %q", letter, test.src, g)
		}
		if g != e {
			t.Errorf("(%v): %q, got %q, expected %q", letter, test.src, g, e)
		}
	}
}

func TestAbstractDeclarator2(t *testing.T) { //TODO -> Example
	for i, test := range []struct{ src, typ string }{
		{"void f(int);", "int"},                                                                                            // [0], 6.7.6, 3, (a)
		{"void f(int*);", "pointer to int"},                                                                                // [0], 6.7.6, 3, (b)
		{"void f(int*[3]);", "array of 3 pointer to int"},                                                                  // [0], 6.7.6, 3, (c)
		{"void f(int(*)[3]);", "pointer to array of 3 int"},                                                                // [0], 6.7.6, 3, (d)
		{"void f(int(*)[*]);", "pointer to array of int"},                                                                  // [0], 6.7.6, 3, (e)
		{"void f(int *());", "function() returning pointer to int"},                                                        // [0], 6.7.6, 3, (f)
		{"void f(int (*)(void));", "pointer to function(void) returning int"},                                              // [0], 6.7.6, 3, (g)
		{"void f(int (*[])(unsigned int, ...));", "array of pointer to function(unsigned, ...) returning int"},             // [0], 6.7.6, 3, (h)
		{"void f(int (*const [])(unsigned int, ...));", "array of const pointer to function(unsigned, ...) returning int"}, // [0], 6.7.6, 3, (h)
	} {
		letter := string(rune('a' + i))
		cfg := &Config{ABI: testABI, doNotSanityCheckComplexTypes: true}
		ast, err := Translate(cfg, nil, nil, []Source{{Name: "test", Value: test.src, DoNotCache: true}})
		if err != nil {
			t.Errorf("(%v): %v", letter, err)
			continue
		}

		var node Node
		depth := mathutil.MaxInt
		findNode("ParameterDeclaration", ast.TranslationUnit, 0, &node, &depth)
		if node == nil {
			t.Errorf("(%v): %q, ParameterDeclaration node not found", letter, test.src)
			continue
		}

		g, e := node.(*ParameterDeclaration).Type().String(), test.typ
		if *oTrace {
			t.Logf("(%v): %q, %q", letter, test.src, g)
		}
		if g != e {
			t.Errorf("(%v): %q, got %q, expected %q", letter, test.src, g, e)
		}
	}
}

func TestDeclarator(t *testing.T) { //TODO -> Example
	for i, test := range []struct{ src, typ string }{
		{"int x;", "int"},                                                                                           // [0], 6.7.6, 3, (a)
		{"int *x;", "pointer to int"},                                                                               // [0], 6.7.6, 3, (b)
		{"int *x[3];", "array of 3 pointer to int"},                                                                 // [0], 6.7.6, 3, (c)
		{"int (*x)[3];", "pointer to array of 3 int"},                                                               // [0], 6.7.6, 3, (d)
		{"int (*x)[*];", "pointer to array of int"},                                                                 // [0], 6.7.6, 3, (e)
		{"int *x();", "function() returning pointer to int"},                                                        // [0], 6.7.6, 3, (f)
		{"int (*x)(void);", "pointer to function(void) returning int"},                                              // [0], 6.7.6, 3, (g)
		{"int (*x[])(unsigned int, ...);", "array of pointer to function(unsigned, ...) returning int"},             // [0], 6.7.6, 3, (h)
		{"int (*const x[])(unsigned int, ...);", "array of const pointer to function(unsigned, ...) returning int"}, // [0], 6.7.6, 3, (h)
	} {
		letter := string(rune('a' + i))
		cfg := &Config{ABI: testABI, doNotSanityCheckComplexTypes: true}
		ast, err := Translate(cfg, nil, nil, []Source{{Name: "test", Value: test.src, DoNotCache: true}})
		if err != nil {
			t.Errorf("(%v): %v", letter, err)
			continue
		}

		var node Node
		depth := mathutil.MaxInt
		findNode("Declarator", ast.TranslationUnit, 0, &node, &depth)
		if node == nil {
			t.Errorf("(%v): %q, Declarator node not found", letter, test.src)
			continue
		}

		g, e := node.(*Declarator).Type().String(), test.typ
		if *oTrace {
			t.Logf("(%v): %q, %q", letter, test.src, g)
		}
		if g != e {
			t.Errorf("(%v): %q, got %q, expected %q", letter, test.src, g, e)
		}
	}
}

func TestDeclarator2(t *testing.T) {
	for i, test := range []struct{ src, typ string }{
		{"struct { int i; } s;", "struct {i int; }"},                                                                       // (a)
		{"union { int i; double d; } u;", "union {i int; d double; }"},                                                     // (b)
		{"typedef struct { unsigned char __c[8]; double __d; } s;", "struct {__c array of 8 unsigned char; __d double; }"}, // (c)
		{"typedef union { unsigned char __c[8]; double __d; } u;", "union {__c array of 8 unsigned char; __d double; }"},   // (d)
		{"struct s { int i;}; typeof(struct s) x;", "struct s"},                                                            // (e)
		{"typeof(42) x;", "int"},       // (f)
		{"typeof(42L) x;", "long"},     // (g)
		{"typeof(42U) x;", "unsigned"}, // (h)
		{"typeof(42.) x;", "double"},   // (i)
		{"#define __GNUC__\ntypedef int x __attribute__ ((vector_size (16)));", "vector of 4 int __attribute__ ((vector_size (16)))"}, // (j)
	} {
		letter := string(rune('a' + i))
		cfg := &Config{ABI: testABI, doNotSanityCheckComplexTypes: true}
		ast, err := Translate(cfg, nil, nil, []Source{{Name: "test", Value: test.src, DoNotCache: true}})
		if err != nil {
			t.Errorf("(%v): %v", letter, err)
			continue
		}

		var node Node
		depth := mathutil.MaxInt
		findNode("Declarator", ast.TranslationUnit, 0, &node, &depth)
		if node == nil {
			t.Errorf("(%v): %q, Declarator node not found", letter, test.src)
			continue
		}

		g, e := node.(*Declarator).Type().String(), test.typ
		if *oTrace {
			t.Logf("(%v): %q, %q", letter, test.src, g)
		}
		if g != e {
			t.Errorf("(%v): %q, got %q, expected %q", letter, test.src, g, e)
		}
	}
}

func TestTranslateCSmith(t *testing.T) {
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

		cfg := &Config{ABI: testABI, Config3: Config3{MaxSourceLine: 1 << 20}, EnableAssignmentCompatibilityChecking: true}
		ctx := newContext(cfg)
		files++
		size += int64(len(out))
		sources := []Source{
			{Name: "<predefined>", Value: testPredef, DoNotCache: true},
			{Name: "<built-in>", Value: parserTestBuiltin, DoNotCache: true},
			{Name: "test.c", Value: string(out), DoNotCache: true},
		}
		if err := func() (err error) {
			defer func() {
				if e := recover(); e != nil && err == nil {
					err = fmt.Errorf("%v", e)
				}
			}()

			var ast *AST
			if ast, err = parse(ctx, testIncludes, testSysIncludes, sources); err != nil {
				return err
			}

			err = ast.Typecheck()
			return err

		}(); err != nil {
			t.Errorf("%s\n%s\nTypecheck: %v", extra, out, err)
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

// https://gitlab.com/cznic/cc/-/issues/116
func Test116(t *testing.T) {
	const filename = "lib.h"
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	_, err = Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: filename, Value: `
typedef struct {
 _Bool has_external_tokens;
 _Bool is_keyword;

 union {
   struct {
     unsigned int node_count;
     unsigned short production_id;
   };
 };
} A;

void foo() {
	A *v;
	unsigned production_id;
	v = (A) {
		.is_keyword = 0,
		{{
		  .node_count = 0,
		  .production_id = production_id,
		}}
	};
}
`},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/-/issues/116
func Test116b(t *testing.T) {
	const filename = "lib.h"
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: filename, Value: `
typedef struct outer {
 _Bool has_external_tokens;
 _Bool is_keyword;

 union inner1 {
   struct inner2 {
     unsigned int node_count;
     unsigned short production_id;
   } si2;
 } ui1;
} A;

void foo() {
	A *v;
	unsigned production_id;
	v = (A) {
		.is_keyword = 0,
		{{
		  .node_count = 0,
		  .production_id = production_id,
		}}
	};
}
`},
	})
	if err != nil {
		t.Fatal(err)
	}

	ta := ast.StructTypes[String("outer")]
	if ta != nil {
		t.Logf("\n%s", dumpLayout(ta))
	}

	m := map[*Initializer]struct{}{}
	Inspect(ast.TranslationUnit, func(n Node, entry bool) bool {
		if !entry {
			return true
		}

		if x, ok := n.(*Initializer); ok {
			if _, ok := m[x]; !ok {
				t.Logf("%v: %v", x.Position(), x.Type())
				m[x] = struct{}{}
			}
		}
		return true
	})

	for d := range ast.TLD {
		fd := d.FunctionDefinition()
		if fd == nil {
			continue
		}

		list := fd.CompoundStatement.BlockItemList
		list = list.BlockItemList
		list = list.BlockItemList
		list = list.BlockItemList
		st := list.BlockItem.Statement
		init1 := st.ExpressionStatement.Expression.
			AssignmentExpression.AssignmentExpression.ConditionalExpression.
			LogicalOrExpression.LogicalAndExpression.InclusiveOrExpression.
			ExclusiveOrExpression.AndExpression.EqualityExpression.
			RelationalExpression.ShiftExpression.AdditiveExpression.
			MultiplicativeExpression.CastExpression.UnaryExpression.
			PostfixExpression.InitializerList
		for it, i := init1, 0; it != nil; it, i = it.InitializerList, i+1 {
			if i != 1 {
				continue
			}

			d := it.Initializer.InitializerList.
				Initializer.InitializerList.
				Designation.DesignatorList.Designator
			typ := it.Initializer.Type()
			t.Logf("%s.%s", typ, d.Token2)
			if typ.Kind() == Bool {
				t.Errorf("cannot set fields on _Bool")
			}
		}
	}
}

// https://gitlab.com/cznic/cc/-/issues/117
func TestIssue117(t *testing.T) {
	cfg := &Config{ABI: testABI, EnableAssignmentCompatibilityChecking: true}
	if _, err := Translate(
		cfg,
		nil,
		nil,
		[]Source{
			{Name: "117.c", Value: `
struct s {
	struct 
	{
		int i;	// ok
		int k;	// ambiguous
	};
	struct
	{
		int j;	// ok
		int k;	// ambiguous
	};
};

int main (void) {
	struct s v;
	v.i = 0;
	v.j = 0;
	v.k = 0;
}
`},
		}); err == nil {
		t.Fatal("should have failed")
	}

	if _, err := Translate(
		cfg,
		nil,
		nil,
		[]Source{
			{Name: "117b.c", Value: `
typedef char int8_t;
typedef short int16_t;
typedef int int32_t;
typedef long long int64_t;
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned uint32_t;

struct S0 {
   signed f0 : 20;
   unsigned f1 : 25;
   unsigned f2 : 23;
   signed f3 : 7;
   signed f4 : 26;
};

struct S1 {
   signed f0 : 10;
   struct S0  f1;
   unsigned : 0;
   struct S0  f2;
   uint32_t  f3;
   signed f4 : 18;
};

struct S2 {
   uint8_t  f0;
   struct S1  f1;
   uint32_t  f2;
   struct S1  f3;
   struct S0  f4;
   struct S0  f5;
};

struct S3 {
   uint16_t  f0;
   struct S0  f1;
   uint16_t  f2;
   struct S2  f3;
   struct S0  f4;
   uint32_t  f5;
   int32_t  f6;
};

struct S4 {
   int8_t  f0;
   int16_t  f1;
   int64_t  f2;
   struct S3  f3;
   struct S3  f4;
   int64_t  f5;
   struct S0  f6;
   int16_t  f7;
   uint32_t  f8;
};

static struct S4 g_431;
static uint16_t *g_739 = &g_431.f4.f2;

int main() {}
`},
		}); err != nil {
		t.Fatal(err)
	}

	if _, err := Translate(
		cfg,
		nil,
		nil,
		[]Source{
			{Name: "117c.c", Value: `

typedef unsigned DWORD;
typedef int LONG;
typedef long long LONGLONG;
typedef union _LARGE_INTEGER {
  struct {
    DWORD LowPart;
    LONG  HighPart;
  } DUMMYSTRUCTNAME;
  struct {
    DWORD LowPart;
    LONG  HighPart;
  } u;
  LONGLONG QuadPart;
} LARGE_INTEGER;

LARGE_INTEGER l;
DWORD d = l.LowPart;

int main() {}
`},
		}); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/-/issues/120
func TestIssue120(t *testing.T) {
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: "x.c", Value: `
void (*x)(void*);
void foo(void *bar){}

int main () {
    if (x == foo) {
        return 1;
    }

    return 0;
}
`},
	})
	if err != nil {
		t.Fatal(err)
	}

	ta := ast.StructTypes[String("outer")]
	if ta != nil {
		t.Logf("\n%s", dumpLayout(ta))
	}

	Inspect(ast.TranslationUnit, func(n Node, entry bool) bool {
		if !entry {
			return true
		}

		if x, ok := n.(*EqualityExpression); ok {
			if typ := x.Promote(); typ != nil {
				if g, e := typ.String(), "pointer to function(pointer to void)"; g != e {
					t.Errorf("%q %q", g, e)
				}
				typ = typ.Elem()
				if g, e := typ.String(), "function(pointer to void)"; g != e {
					t.Errorf("%q %q", g, e)
				}
			}
		}
		return true
	})
}

// https://gitlab.com/cznic/cc/-/issues/121
func TestIssue121(t *testing.T) {
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: "x.c", Value: `
struct xxx {
    unsigned ub:3;
    unsigned u:32;
    unsigned long long ullb:35;
    unsigned long long ull:64;
    unsigned char c;
} s;

`},
	})
	if err != nil {
		t.Fatal(err)
	}

	ta := ast.StructTypes[String("xxx")]
	t.Logf("\n%s", dumpLayout(ta))
	fld, _ := ta.FieldByName(String("c"))
	if g, e := fld.Offset(), uintptr(24); g != e {
		t.Fatalf("xxx.c offset: got %v, exp %v", g, e)
	}
}

func isConstInitializer(n *Initializer) bool {
	if n.IsConst() {
		return true
	}

	if e := n.AssignmentExpression; e != nil {
		if x, ok := e.Operand.Value().(*InitializerValue); ok && x.IsConst() {
			return true
		}
	}

	for list := n.InitializerList; list != nil; list = list.InitializerList {
		if !isConstInitializer(list.Initializer) {
			return false
		}
	}

	e := n.AssignmentExpression
	if e == nil {
		return true
	}

	switch t := e.Operand.Type().Decay(); t.Kind() {
	case Function:
		return true
	case Ptr:
		if d := e.Operand.Declarator(); d != nil && d.StorageClass == Static {
			return true
		}
	}

	// trc("%v: %T %v, %p", e.Position(), e.Operand.Value(), e.Operand.Type(), e.Operand.Declarator())
	return false
}

func TestConstInitializer(t *testing.T) {
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: "x.c", Value: `
void *x2 = (char*)50;
void *x3 = &((char*)50)[2];
void *x4 = &((char*)50)[-2];
int *x5[] = {&x2, &x2, (int *) &x3};
int **x6[] = {&x5[1], x5 + 2};
int x7[] = {1, 2, 3, 4, 5};
int *x8 = x7;
char *x9 = {"Hello" + 1};

//TODO typedef float vec_t[3];
//TODO 
//TODO typedef struct {
//TODO   vec_t loc;
//TODO } point_t;
//TODO 
//TODO static int x17[] = {(unsigned long) &((point_t *) 0)->loc[0], (unsigned long) &((point_t *) 0)->loc[2], (unsigned long) &((vec_t *) 0)[1]};

`},
	})
	if err != nil {
		t.Fatal(err)
	}

	Inspect(ast.TranslationUnit, func(n Node, entry bool) bool {
		if !entry {
			return true
		}

		if x, ok := n.(*Initializer); ok {
			if x.Parent() != nil {
				return true
			}

			if isConstInitializer(x) {
				return true
			}

			t.Errorf("%v: not constant", x.Position())
		}
		return true
	})
}

// https://gitlab.com/cznic/cc/-/issues/122
func TestTranslateBug(t *testing.T) {
	cfg := &Config{
		ABI:                                   testABI,
		ignoreUndefinedIdentifiers:            true,
		EnableAssignmentCompatibilityChecking: true,
	}
	root := filepath.Join(testWD, filepath.FromSlash("testdata"))
	ok := 0
	const dir = "bug"
	t.Run(dir, func(t *testing.T) {
		ok += testTranslateDir(t, cfg, testPredef, filepath.Join(root, filepath.FromSlash(dir)), false)
	})
	t.Logf("ok %v", h(ok))
}

// https://gitlab.com/cznic/ccgo/-/issues/20
func TestEscE(t *testing.T) {
	abi, err := NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ast, err := Translate(&Config{ABI: abi}, nil, nil, []Source{
		{Name: "x.c", Value: `

char *s = "abc\edef";

`},
	})
	if err != nil {
		t.Fatal(err)
	}

	ok := false
	Inspect(ast.TranslationUnit, func(n Node, entry bool) bool {
		if !entry {
			return true
		}

		switch x := n.(type) {
		case *Initializer:
			switch y := x.AssignmentExpression.Operand.Value().(type) {
			case StringValue:
				ok = true
				if g, e := StringID(y).String(), "abc\x1bdef"; g != e {
					t.Fatalf("got %q, expected %q", g, e)
				}
			}
		}
		return true
	})
	if !ok {
		t.Fatal(ast.Scope[String("s")])
	}
}
