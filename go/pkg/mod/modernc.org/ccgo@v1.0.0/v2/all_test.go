// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// linux_386
//
//	TCC0	cc 51 ccgo 51 build 51 run 51 ok 51
//	Other0	cc 18 ccgo 18 build 18 run 18 ok 18
//	GCC0	cc 1112 ccgo 1093 build 1089 run 1089 ok 1089
//	Shell0	cc 1 ccgo 1 build 1 run 1 ok 1
//	TCL0	tclsqlite build ok
//	--- FAIL: TestTCL0 (36.48s)
//		all_test.go:1586:
//			Test cases:        0
//			Pass:              0 (NaN%)
//			Fail:              0 (NaN%)
//		all_test.go:1594:
//			Test binary exit error: exit status 2
//			Last completed test file: ""
//			Last passed test: ""
//			Last line written to stdout: ""
//			Blacklisted test files: 107
//			btreefault.test
//			cffault.test
//			collate1.test
//			collate2.test
//			collate3.test
//			collate4.test
//			collate5.test
//			collate6.test
//			collate9.test
//			corruptC.test
//			crash.test
//			crash2.test
//			crash3.test
//			crash4.test
//			crash6.test
//			crash7.test
//			date.test
//			e_createtable.test
//			e_delete.test
//			e_insert.test
//			e_reindex.test
//			e_select.test
//			e_update.test
//			e_walauto.test
//			exists.test
//			func4.test
//			fuzz.test
//			fuzzerfault.test
//			ieee754.test
//			incrcorrupt.test
//			incrvacuum_ioerr.test
//			ioerr3.test
//			journal3.test
//			lock.test
//			lock4.test
//			lock5.test
//			malloc.test
//			minmax.test
//			misc1.test
//			misc3.test
//			misc7.test
//			mjournal.test
//			mmap1.test
//			mmap4.test
//			multiplex2.test
//			nan.test
//			pager1.test
//			pager4.test
//			pagerfault.test
//			pagerfault2.test
//			pagerfault3.test
//			pragma.test
//			printf.test
//			quota2.test
//			rbu.test
//			reindex.test
//			rollbackfault.test
//			rowallock.test
//			savepoint.test
//			savepoint4.test
//			savepointfault.test
//			schema3.test
//			select9.test
//			shared2.test
//			shared9.test
//			sharedA.test
//			sort2.test
//			sort3.test
//			sort4.test
//			sortfault.test
//			speed4.test
//			speed4p.test
//			statfault.test
//			superlock.test
//			symlink.test
//			syscall.test
//			tempfault.test
//			thread001.test
//			thread002.test
//			thread003.test
//			thread004.test
//			thread005.test
//			thread1.test
//			thread2.test
//			tkt-5d863f876e.test
//			tkt-fc62af4523.test
//			tkt3838.test
//			tkt3997.test
//			trans.test
//			unionvtabfault.test
//			unixexcl.test
//			vacuum2.test
//			vtabH.test
//			wal.test
//			wal2.test
//			wal3.test
//			wal4.test
//			wal5.test
//			walcrash.test
//			walcrash2.test
//			walcrash4.test
//			walro.test
//			walslow.test
//			walthread.test
//			where.test
//			whereD.test
//			writecrash.test
//	FAIL
//	exit status 1
//	FAIL	modernc.org/ccgo/v2	481.539s

//	go version go1.11 linux/amd64
//
//	Thu Sep 27 18:10:24 CEST 2018
//	TCC	cc 51 ccgo 51 build 51 run 51 ok 51 n 51
//	Other	cc 34 ccgo 34 build 34 run 34 ok 34 n 34
//	GCC Compat	cc 7 ccgo 7 build 7 run 7 ok 7 n 7
//	GCC Compile	cc 994 ccgo 994 build 994 ok 994 n 1708
//	GCC Execute	cc 1093 ccgo 1093 build 1093 run 1093 ok 1093 n 1413
//	Shell	cc 1 ccgo 1 build 1 run 1 ok 1 n 1
//	cc 1 ccgo 1 build 1 run 1 ok 1 (100.00%) csmith 1 (2.118180533s) -s 1352671588
//	cc 2 ccgo 2 build 2 run 2 ok 2 (100.00%) csmith 2 (4.275911183s) -s 1322542678
//	cc 3 ccgo 3 build 3 run 3 ok 3 (100.00%) csmith 3 (5.506741298s) -s 1901462611
//	cc 4 ccgo 4 build 4 run 4 ok 4 (100.00%) csmith 4 (8.529098616s) -s 2829621573
//	cc 5 ccgo 5 build 5 run 5 ok 5 (100.00%) csmith 5 (9.399702001s) -s 576413934
//	cc 6 ccgo 6 build 6 run 6 ok 6 (100.00%) csmith 6 (11.192800405s) -s 151038493
//	cc 7 ccgo 7 build 7 run 7 ok 7 (100.00%) csmith 7 (19.012242333s)
//	cc 8 ccgo 8 build 8 run 8 ok 8 (100.00%) csmith 8 (19.864834663s)
//	cc 9 ccgo 9 build 9 run 9 ok 9 (100.00%) csmith 9 (22.511054232s)
//	cc 10 ccgo 10 build 10 run 10 ok 10 (100.00%) csmith 10 (24.888215782s)
//	cc 11 ccgo 11 build 11 run 11 ok 11 (100.00%) csmith 11 (26.746089666s)
//	cc 12 ccgo 12 build 12 run 12 ok 12 (100.00%) csmith 12 (30.066075584s)
//	cc 13 ccgo 13 build 13 run 13 ok 13 (100.00%) csmith 13 (32.02813583s)
//	cc 14 ccgo 14 build 14 run 14 ok 14 (100.00%) csmith 14 (34.834296524s)
//	cc 15 ccgo 15 build 15 run 15 ok 15 (100.00%) csmith 15 (36.688226243s)
//	cc 16 ccgo 16 build 16 run 16 ok 16 (100.00%) csmith 16 (38.482662304s)
//	cc 17 ccgo 17 build 17 run 17 ok 17 (100.00%) csmith 17 (40.120553995s)
//	cc 18 ccgo 18 build 18 run 18 ok 18 (100.00%) csmith 18 (48.550537748s)
//	cc 19 ccgo 19 build 19 run 19 ok 19 (100.00%) csmith 19 (52.354074016s)
//	cc 20 ccgo 20 build 20 run 20 ok 20 (100.00%) csmith 20 (53.549131137s)
//	cc 21 ccgo 21 build 21 run 21 ok 21 (100.00%) csmith 21 (55.093100821s)
//	cc 22 ccgo 22 build 22 run 22 ok 22 (100.00%) csmith 22 (55.967966951s)
//	cc 23 ccgo 23 build 23 run 23 ok 23 (100.00%) csmith 23 (56.82333014s)
//	cc 24 ccgo 24 build 24 run 24 ok 24 (100.00%) csmith 24 (58.503562438s)
//	cc 25 ccgo 25 build 25 run 25 ok 25 (100.00%) csmith 25 (1m0.909371585s)
//	CSmith0	cc 25 ccgo 25 build 25 run 25 ok 25 (100.00%) csmith 25 (1m0.909416163s)
//	PASS
//	ok  	modernc.org/ccgo/v2	1084.208s

package ccgo // import "modernc.org/ccgo/v2"

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"modernc.org/cc/v2"
	"modernc.org/strutil"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("# TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO) //TODOOK
	flag.BoolVar(&traceLConsts, "lconst", false, "")
	flag.BoolVar(&traceOpt, "to", false, "")
	flag.BoolVar(&traceTODO, "todo", false, "")
	flag.BoolVar(&traceWrites, "tw", false, "")
	isTesting = true
}

// ============================================================================

const (
	testTimeout = 60 * time.Second
)

var (
	oBuild   = flag.Bool("build", false, "full build errors")
	oCC      = flag.Bool("cc", false, "full cc errors")
	oCCGO    = flag.Bool("ccgo", false, "full ccgo errors")
	oCSmith  = flag.Duration("csmith", time.Minute, "") // Use something like -timeout 25h -csmith 24h for real testing.
	oEdit    = flag.Bool("edit", false, "")
	oI       = flag.String("I", "", "")
	oNoCmp   = flag.Bool("nocmp", false, "")
	oRE      = flag.String("re", "", "")
	oTCLRace = flag.Bool("tclrace", false, "")
	oTmp     = flag.String("tmp", "", "")
	oTrace   = flag.Bool("trc", false, "")

	re          *regexp.Regexp
	searchPaths []string
	defCCGO     = cc.NewStringSource("<defines>", "#define __ccgo__ 1\n#define __FUNCTION__ __func__\n")
)

func init() {
	var err error
	if searchPaths, err = cc.Paths(true); err != nil {
		panic(err)
	}
}

// Command outputs a Go program generated from in to w.
//
// No package or import clause is generated.
func Command(w io.Writer, in []*cc.TranslationUnit) (err error) {
	returned := false

	defer func() {
		if e := recover(); !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, compact(string(debugStack()), compactStack))
		}
	}()

	err = newGen(w, in).gen(true)
	returned = true
	return err
}

func TestOpt(t *testing.T) {
	for _, v := range []struct{ in, out string }{
		{"var _ = (a(b))", "var _ = a(b)"},
		{"var _ = ((a)(b))", "var _ = a(b)"},
		{"var _ = *((*a)(b))", "var _ = *(*a)(b)"},
	} {
		in := bytes.NewBufferString(v.in)
		var out bytes.Buffer
		if err := newOpt().do(&out, in, "TestOp", 0); err != nil {
			t.Fatal(err)
		}

		if g, e := bytes.TrimSpace(out.Bytes()), []byte(v.out); !bytes.Equal(g, e) {
			t.Fatalf("got\n%s\nexp\n%s", g, e)
		}
	}
}

func trim(b []byte) []byte {
	a := bytes.Split(b, []byte{'\n'})
	for i, v := range a {
		a[i] = bytes.TrimRight(v, " ")
	}
	return bytes.Join(a, []byte{'\n'})
}

func translate(tweaks *cc.Tweaks, includePaths, sysIncludePaths []string, def string, sources ...cc.Source) (*cc.TranslationUnit, error) {
	in := []cc.Source{defCCGO, cc.MustBuiltin()}
	if def != "" {
		in = append(in, cc.NewStringSource("<defines>", def))
	}
	in = append(in, sources...)
	if *oTrace {
		fmt.Fprintln(os.Stderr, in)
	}
	return cc.Translate(tweaks, includePaths, sysIncludePaths, in...)
}

func test(t *testing.T, clean bool, c, ccgo, build, run *int, def, imp string, inc2 []string, dir string, pth []string, args ...string) ([]byte, error) {
	testFn = pth[len(pth)-1]
	if clean {
		m, err := filepath.Glob(filepath.Join(dir, "*.*"))
		if err != nil {
			t.Fatal(err)
		}

		for _, v := range m {
			if err := os.Remove(v); err != nil {
				t.Fatal(err)
			}
		}
	}

	tweaks := &cc.Tweaks{
		// TrackExpand:                 func(s string) { fmt.Print(s) }, //TODO-
		EnableAnonymousStructFields: true,
		EnableEmptyStructs:          true,
		EnableImplicitBuiltins:      true,
		EnableImplicitDeclarations:  true,
		EnableOmitFuncDeclSpec:      true,
		EnablePointerCompatibility:  true, // CSmith transparent_crc_bytes
		EnableReturnExprInVoidFunc:  true,
		IgnoreUnknownPragmas:        true,
		InjectFinalNL:               true,
	}
	inc := append([]string{"@"}, inc2...)

	crt0, err := cc.Translate(tweaks, inc, searchPaths, defCCGO, cc.MustBuiltin(), cc.MustCrt0())
	if err != nil {
		return nil, err
	}

	tus := []*cc.TranslationUnit{crt0}
	for _, v := range pth {
		tu, err := translate(tweaks, inc, searchPaths, def, cc.MustFileSource2(v, false))
		if err != nil {
			//dbg("cc: %v", errString(err)) //TODO-
			if !*oCC {
				err = nil
			}
			return nil, err
		}

		tus = append(tus, tu)
	}

	*c++
	f, err := os.Create(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}

	w := bufio.NewWriter(f)
	w.WriteString(`package main
	
import (
	"math"
	"os"
	"unsafe"

	"modernc.org/crt"
)

var _ = math.Inf
`)
	w.WriteString(imp)
	if err := Command(w, tus); err != nil {
		//dbg("ccgo: %v", errString(err)) //TODO-
		if !*oCCGO {
			err = nil
		}
		return nil, err
	}

	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	*ccgo++

	if out, err := exec.Command("go", "build", "-o", filepath.Join(dir, "main"), f.Name()).CombinedOutput(); err != nil {
		//dbg("build: %v", errString(err)) //TODO-
		if !*oBuild {
			return nil, nil
		}

		return nil, fmt.Errorf("%v: %s", err, out)
	}

	*build++

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)

	defer cancel()

	out, err := exec.CommandContext(ctx, filepath.Join(dir, "main"), args...).CombinedOutput()
	switch {
	case err != nil:
		//dbg("run: %v", errString(err)) //TODO-
	default:
		*run++
	}
	return out, err
}

func TestTCC(t *testing.T) {
	blacklist := map[string]struct{}{
		"13_integer_literals.c": {}, // 9:12: ExprInt strconv.ParseUint: parsing "0b010101010101": invalid syntax
		"31_args.c":             {},
		"34_array_assignment.c": {}, // gcc: main.c:16:6: error: incompatible types when assigning to type ‘int[4]’ from type ‘int *’
		"46_grep.c":             {}, // incompatible forward declaration type
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-tcc-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/tcc-0.9.26/tests/tests2/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := filepath.Glob(root)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	testGo := filepath.Join(dir, "main.go")
	testBin := filepath.Join(dir, "main")
	var c, ccgo, build, run, ok, n int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		if *oTrace {
			fmt.Fprintln(os.Stderr, pth)
		}
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", testGo, pth).CombinedOutput()
		if err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		c++
		ccgo++
		if out, err := exec.Command("go", "build", "-o", testBin, testGo).CombinedOutput(); err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		build++
		if out, err = exec.Command(testBin).CombinedOutput(); err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		run++
		expect, err := ioutil.ReadFile(pth[:len(pth)-len(filepath.Ext(pth))] + ".expect")
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		expect = trim(expect)
		if !bytes.Equal(out, expect) {
			t.Logf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(expect), out, expect)
			continue
		}

		ok++
	}
	if c != n || ccgo != n || build != n || run != n || ok != n {
		t.Fatalf("TCC cc %v ccgo %v build %v run %v ok %v n %v", c, ccgo, build, run, ok, n)
	}

	if *oEdit {
		fmt.Printf("TCC\tcc %v ccgo %v build %v run %v ok %v n %v\n", c, ccgo, build, run, ok, n)
	}
}

func compileCCGO(dir string) (string, error) {
	ip, err := strutil.ImportPath()
	if err != nil {
		return "", err
	}

	ip = filepath.Join(ip, "ccgo")
	compiler := filepath.Join(dir, "ccgo")
	out, err := exec.Command("go", "build", "-o", compiler, ip).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s\n%v", out, err)
	}

	return compiler, nil
}

func TestOther0(t *testing.T) { //TODO-
	return //TODO
	cc.FlushCache()
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir, err := ioutil.TempDir("", "test-ccgo-other-")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()

	m, err := filepath.Glob(filepath.FromSlash("testdata/bug/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	var cc, ccgo, build, run, ok, n int
	for _, pth := range m {
		if b := filepath.Base(pth); b == "log.c" && *oRE != "log.c" || re != nil && !re.MatchString(b) {
			continue
		}

		run0 := run
		n++
		out, err := test(t, false, &cc, &ccgo, &build, &run, "", "", strings.Split(*oI, ","), dir, []string{pth})
		if err != nil {
			t.Errorf("%v: %v", pth, err)
			continue
		}

		if run == run0 {
			continue
		}

		fn := pth[:len(pth)-len(filepath.Ext(pth))] + ".expect"
		s, err := ioutil.ReadFile(fn)
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		s = trim(s)
		if !bytes.Equal(out, s) {
			t.Errorf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(s), out, s)
			continue
		}

		ok++
	}
	if cc != n || ccgo != n || build != n || run != n || ok != n {
		t.Fatalf("cc %v ccgo %v build %v run %v ok %v", cc, ccgo, build, run, ok)
	}

	if *oEdit {
		fmt.Printf("Other0\tcc %v ccgo %v build %v run %v ok %v\n", cc, ccgo, build, run, ok)
	}
}

func TestOther(t *testing.T) {
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-other-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/bug/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := filepath.Glob(root)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	testGo := filepath.Join(dir, "main.go")
	testBin := filepath.Join(dir, "main")
	var c, ccgo, build, run, ok, n int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if *oTrace {
			fmt.Fprintln(os.Stderr, pth)
		}
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", testGo, pth).CombinedOutput()
		if err != nil {
			t.Logf("%s:\n%s\n%v", pth, out, err)
			continue
		}

		c++
		ccgo++
		if out, err := exec.Command("go", "build", "-o", testBin, testGo).CombinedOutput(); err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		build++
		if out, err = exec.Command(testBin).CombinedOutput(); err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		run++
		expect, err := ioutil.ReadFile(pth[:len(pth)-len(filepath.Ext(pth))] + ".expect")
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		expect = trim(expect)
		if !bytes.Equal(out, expect) {
			t.Logf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(expect), out, expect)
			continue
		}

		ok++
	}
	if c != n || ccgo != n || build != n || run != n || ok != n {
		t.Fatalf("Other\tcc %v ccgo %v build %v run %v ok %v n %v", c, ccgo, build, run, ok, n)
	}

	if *oEdit {
		fmt.Printf("Other\tcc %v ccgo %v build %v run %v ok %v n %v\n", c, ccgo, build, run, ok, n)
	}
}

func TestGCC0(t *testing.T) { //TODO-
	return //TODO
	cc.FlushCache()
	const def = `
#define SIGNAL_SUPPRESS // gcc.c-torture/execute/20101011-1.c
`
	blacklist := map[string]struct{}{
		"20010904-1.c":    {}, // __attribute__((aligned(32)))
		"20010904-2.c":    {}, // __attribute__((aligned(32)))
		"20021127-1.c":    {}, // non standard GCC behavior
		"eeprof-1.c":      {}, // Need profiler code instrumentation
		"pr23467.c":       {}, // __attribute__ ((aligned (8)))
		"pr67037.c":       {}, // void f(); f(); f(42)
		"pushpop_macro.c": {}, // #pragma push_macro("_")
		"zerolen-2.c":     {}, // The Go translation makes the last zero items array to have size 1.

		"20000703-1.c":                 {}, //TODO statement expression
		"20040411-1.c":                 {}, //TODO VLA
		"20040423-1.c":                 {}, //TODO VLA
		"20040629-1.c":                 {}, //TODO bits, arithmetic precision
		"20040705-1.c":                 {}, //TODO bits, arithmetic precision
		"20040705-2.c":                 {}, //TODO bits, arithmetic precision
		"20041218-2.c":                 {}, //TODO VLA
		"20101011-1.c":                 {}, //TODO Needs sigfpe on int division by zero
		"921016-1.c":                   {}, //TODO bits, arithmetic precision
		"970217-1.c":                   {}, //TODO VLA
		"bitfld-1.c":                   {}, //TODO bits, arithmetic precision
		"bitfld-3.c":                   {}, //TODO bits, arithmetic precision
		"builtin-types-compatible-p.c": {}, //TODO must track type qualifiers
		"pr32244-1.c":                  {}, //TODO bits, arithmetic precision
		"pr34971.c":                    {}, //TODO bits, arithmetic precision
		"pr77767.c":                    {}, //TODO VLA

		//TODO bit field arithmetic
		// 20040709-1.c
		// 20040709-2.c

	}

	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-gcc-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	m, err := filepath.Glob(filepath.FromSlash("testdata/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	var cc, ccgo, build, run, ok int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		run0 := run
		out, err := test(t, false, &cc, &ccgo, &build, &run, def, "", nil, dir, []string{pth})
		if err != nil {
			t.Errorf("%v: %v", pth, err)
			continue
		}

		if run == run0 {
			continue
		}

		fn := pth[:len(pth)-len(filepath.Ext(pth))] + ".expect"
		s, err := ioutil.ReadFile(fn)
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		s = trim(s)
		if !bytes.Equal(out, s) {
			t.Errorf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(s), out, s)
			continue
		}

		ok++
	}
	if run == 0 || run != build || ok != build {
		t.Fatalf("cc %v ccgo %v build %v run %v ok %v", cc, ccgo, build, run, ok)
	}

	if *oEdit {
		fmt.Printf("GCC0\tcc %v ccgo %v build %v run %v ok %v\n", cc, ccgo, build, run, ok)
	}
}

func TestGCCCompat(t *testing.T) {
	blacklist := map[string]struct{}{}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-gcc-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/compat/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := filepath.Glob(root)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	testGo := filepath.Join(dir, "main.go")
	testBin := filepath.Join(dir, "main")
	var c, ccgo, build, run, ok, n int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		if *oTrace {
			fmt.Fprintln(os.Stderr, pth)
		}
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", testGo, pth).CombinedOutput()
		if err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		c++
		ccgo++
		if out, err := exec.Command("go", "build", "-o", testBin, testGo).CombinedOutput(); err != nil {
			t.Errorf("%s: %s\n%v", pth, out, err)
			continue
		}

		build++
		if out, err = exec.Command(testBin).CombinedOutput(); err != nil {
			t.Errorf("%s: %s\n%v", pth, out, err)
			continue
		}

		run++
		expect, err := ioutil.ReadFile(pth[:len(pth)-len(filepath.Ext(pth))] + ".expect")
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		expect = trim(expect)
		if !bytes.Equal(out, expect) {
			t.Errorf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(expect), out, expect)
			continue
		}

		ok++
	}

	if run == 0 || run != build || ok != build {
		t.Fatalf("cc %v ccgo %v build %v run %v ok %v n %v", c, ccgo, build, run, ok, n)
	}

	if *oEdit {
		fmt.Printf("GCC Compat\tcc %v ccgo %v build %v run %v ok %v n %v\n", c, ccgo, build, run, ok, n)
	}
}

func TestGCCCompile(t *testing.T) {
	blacklist := map[string]struct{}{}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-gcc-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/compile/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := filepath.Glob(root)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	testGo := filepath.Join(dir, "test.go")
	var c, ccgo, build, ok, n int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		if *oTrace {
			fmt.Fprintln(os.Stderr, pth)
		}
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", testGo, pth).CombinedOutput()
		if err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		c++
		ccgo++
		build++
		ok++
	}

	if ok != build {
		t.Fatalf("cc %v ccgo %v build %v ok %v n %v", c, ccgo, build, ok, n)
	}

	if *oEdit {
		fmt.Printf("GCC Compile\tcc %v ccgo %v build %v ok %v n %v\n", c, ccgo, build, ok, n)
	}
}

func TestGCCExecute(t *testing.T) {
	blacklist := map[string]struct{}{
		"20010904-1.c": {}, // __attribute__((aligned(32)))
		"20010904-2.c": {}, // __attribute__((aligned(32)))
		"20021127-1.c": {}, // non standard GCC behavior
		"20040520-1.c": {}, // Nested func
		"eeprof-1.c":   {}, // requires profiler instrumentation
		"pr23467.c":    {}, // __attribute__ ((aligned (8)))

		"20101011-1.c":                 {}, //TODO Needs sigfpe on int division by zero
		"970217-1.c":                   {}, //TODO
		"alias-3.c":                    {}, //TODO attr alias for data
		"bitfld-1.c":                   {}, //TODO bits, arithmetic precision
		"bitfld-3.c":                   {}, //TODO bits arithcmetic precision
		"builtin-types-compatible-p.c": {}, //TODO must track type qualifiers
		"medce-1.c":                    {}, //TODO Handle dead code
		"pr32244-1.c":                  {}, //TODO bits, arithmetic precision
		"pr34971.c":                    {}, //TODO bits, arithmetic precision
		"pr67037.c":                    {}, //TODO void f(); f(); f(42)
		"pr77767.c":                    {}, //TODO
		"zero-struct-2.c":              {}, //TODO The Go translation makes the last zero items array to have size 1.
		"zerolen-2.c":                  {}, //TODO The Go translation makes the last zero items array to have size 1.
	}
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-gcc-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/*.c"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := filepath.Glob(root)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	testGo := filepath.Join(dir, "main.go")
	testBin := filepath.Join(dir, "main")
	var c, ccgo, build, run, ok, n int
	for _, pth := range m {
		if re != nil && !re.MatchString(filepath.Base(pth)) {
			continue
		}

		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		if *oTrace {
			fmt.Fprintln(os.Stderr, pth)
		}
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", testGo, pth).CombinedOutput()
		if err != nil {
			t.Logf("%s: %s\n%v", pth, out, err)
			continue
		}

		//dbg("%s", out) //TODO- DBG
		c++
		ccgo++
		if out, err := exec.Command("go", "build", "-o", testBin, testGo).CombinedOutput(); err != nil {
			t.Errorf("%s: %s\n%v", pth, out, err)
			continue
		}

		build++
		if out, err = exec.Command(testBin).CombinedOutput(); err != nil {
			t.Errorf("%s: %s\n%v", pth, out, err)
			continue
		}

		run++
		expect, err := ioutil.ReadFile(pth[:len(pth)-len(filepath.Ext(pth))] + ".expect")
		if err != nil {
			if os.IsNotExist(err) {
				ok++
				continue
			}
		}

		out = trim(out)
		expect = trim(expect)
		if !bytes.Equal(out, expect) {
			t.Errorf("%s\ngot\n%s\nexp\n%s----\ngot\n%s\nexp\n%s", pth, hex.Dump(out), hex.Dump(expect), out, expect)
			continue
		}

		ok++
	}

	if run == 0 || run != build || ok != build {
		t.Fatalf("cc %v ccgo %v build %v run %v ok %v n %v", c, ccgo, build, run, ok, n)
	}

	if *oEdit {
		fmt.Printf("GCC Execute\tcc %v ccgo %v build %v run %v ok %v n %v\n", c, ccgo, build, run, ok, n)
	}
}

func TestSQLiteShell0(t *testing.T) { //TODO-
	return //TODO
	cc.FlushCache()
	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-shell-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	var cc, ccgo, build, run, ok int
	root := filepath.FromSlash("testdata/_sqlite/sqlite-amalgamation-3210000")
	if out, err := test(t, false, &cc, &ccgo, &build, &run, `
		#define HAVE_FDATASYNC 1
		#define HAVE_ISNAN 1
		#define HAVE_LOCALTIME_R 1
		#define HAVE_USLEEP 1
		#define SQLITE_DEBUG 1
		#define SQLITE_MEMDEBUG 1
		/* #define HAVE_MALLOC_USABLE_SIZE 1 */
`,
		"",
		nil,
		dir,
		[]string{
			filepath.Join(root, "shell.c"),
			filepath.Join(root, "sqlite3.c"),
		},
		"foo", "create table t(i)",
	); err != nil {
		t.Fatalf("%s: %v", out, errString(err))
	}

	if run == 1 {
		ok++
	}
	if ok != 1 {
		t.Fatalf("cc %v ccgo %v build %v run %v ok %v", cc, ccgo, build, run, ok)
	}

	if *oEdit {
		fmt.Printf("Shell0\tcc %v ccgo %v build %v run %v ok %v\n", cc, ccgo, build, run, ok)
	}
}

func TestSQLiteShell(t *testing.T) {
	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-sqlite-shell-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	root, err := filepath.Abs(filepath.FromSlash("testdata/_sqlite/sqlite-amalgamation-3210000"))
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	var c, ccgo, build, run, ok, n int
	for {
		n++
		out, err := exec.Command(compiler, "--ccgo-struct-checks", filepath.Join(root, "shell.c"), filepath.Join(root, "sqlite3.c")).CombinedOutput()
		if err != nil {
			t.Logf("%s: %s\n%v", dir, out, err)
			break
		}

		os.Remove(compiler)
		c++
		ccgo++
		build++
		m, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil {
			t.Log(err)
			break

		}

		if len(m) != 1 {
			t.Fatalf("%v\n%v", len(m), m)
		}

		if out, err = exec.Command(m[0], "foo", "create table t(i)").CombinedOutput(); err != nil {
			t.Logf("\n%s\n%v", out, err)
			break
		}

		run++
		ok++
		break
	}
	if c != n || ccgo != n || build != n || run != n || ok != n {
		t.Fatalf("Shell cc %v ccgo %v build %v run %v ok %v n %v", c, ccgo, build, run, ok, n)
	}

	if *oEdit {
		fmt.Printf("Shell\tcc %v ccgo %v build %v run %v ok %v n %v\n", c, ccgo, build, run, ok, n)
	}
}

func TestTCL0(t *testing.T) { //TODO-
	return //TODO
	cc.FlushCache()
	const (
		allDefs = `// Output of gcc features.c && ./a.out in modernc.org/sqlite2go/internal/c99/headers on linux_amd64.
			#define _POSIX_SOURCE 1
			#define _POSIX_C_SOURCE 200809
			#define _DEFAULT_SOURCE 1
`
		sqlite     = "_sqlite"
		sqliteDefs = `
			#define HAVE_FDATASYNC 1
			#define HAVE_ISNAN 1
			#define HAVE_LOCALTIME_R 1
			#define HAVE_USLEEP 1
			#define SQLITE_CORE 1 // Must be defined for TCL extensions to work.
			#define SQLITE_DEBUG 1
			#define SQLITE_ENABLE_RBU 1
			#define SQLITE_PRIVATE
			#define SQLITE_TEST 1
			#define TCLSH_INIT_PROC sqlite3TestInit
			/* #define HAVE_MALLOC_USABLE_SIZE 1 */
			// #define SQLITE_MEMDEBUG 1 //TODO wants execinfo.backtrace*
`
		tcl      = "_tcl8.6.8"
		tclDefs0 = `
			#define BUILD_tcl
			#define CFG_INSTALL_BINDIR "/usr/local/bin"
			#define CFG_INSTALL_DOCDIR "/usr/local/man"
			#define CFG_INSTALL_INCDIR "/usr/local/include"
			#define CFG_INSTALL_LIBDIR "/usr/local/lib64"
			#define CFG_INSTALL_SCRDIR "/usr/local/lib/tcl8.6"
			#define CFG_RUNTIME_BINDIR "/usr/local/bin"
			#define CFG_RUNTIME_DOCDIR "/usr/local/man"
			#define CFG_RUNTIME_INCDIR "/usr/local/include"
			#define CFG_RUNTIME_LIBDIR "/usr/local/lib64"
			#define CFG_RUNTIME_SCRDIR "/usr/local/lib/tcl8.6"
			#define HAVE_BLKCNT_T 1
			#define HAVE_CAST_TO_UNION 1
			#define HAVE_FREEADDRINFO 1
			#define HAVE_FTS 1
			#define HAVE_GAI_STRERROR 1
			#define HAVE_GETADDRINFO 1
			#define HAVE_GETCWD 1
			#define HAVE_GETGRGID_R 1
			#define HAVE_GETGRGID_R_5 1
			#define HAVE_GETGRNAM_R 1
			#define HAVE_GETGRNAM_R_5 1
			#define HAVE_GETHOSTBYADDR_R 1
			#define HAVE_GETNAMEINFO 1
			#define HAVE_GETPWNAM_R 1
			#define HAVE_GETPWNAM_R_5 1
			#define HAVE_GETPWUID_R 1
			#define HAVE_GETPWUID_R_5 1
			#define HAVE_GMTIME_R 1
			#define HAVE_HIDDEN 1
			#define HAVE_INTPTR_T 1
			#define HAVE_INTTYPES_H 1
			#define HAVE_LOCALTIME_R 1
			#define HAVE_MEMORY_H 1
			#define HAVE_MKSTEMP 1
			#define HAVE_MKSTEMPS 1
			#define HAVE_MKTIME 1
			#define HAVE_OPENDIR 1
			#define HAVE_SIGNED_CHAR 1
			#define HAVE_STDINT_H 1
			#define HAVE_STDLIB_H 1
			#define HAVE_STRINGS_H 1
			#define HAVE_STRING_H 1
			#define HAVE_STRTOL 1
			#define HAVE_STRUCT_ADDRINFO 1
			#define HAVE_STRUCT_IN6_ADDR 1
			#define HAVE_STRUCT_SOCKADDR_IN6 1
			#define HAVE_STRUCT_SOCKADDR_STORAGE 1
			#define HAVE_STRUCT_STAT_ST_BLKSIZE 1
			#define HAVE_STRUCT_STAT_ST_BLOCKS 1
			#define HAVE_SYS_IOCTL_H 1
			#define HAVE_SYS_IOCTL_H 1
			#define HAVE_SYS_PARAM_H 1
			#define HAVE_SYS_STAT_H 1
			#define HAVE_SYS_TIME_H 1
			#define HAVE_SYS_TYPES_H 1
			#define HAVE_TERMIOS_H 1
			#define HAVE_TIMEZONE_VAR 1
			#define HAVE_TM_GMTOFF 1
			#define HAVE_UINTPTR_T 1
			#define HAVE_UNISTD_H 1
			#define HAVE_WAITPID 1
			#define MODULE_SCOPE extern
			#define MP_PREC 4
			#define PACKAGE_BUGREPORT ""
			#define PACKAGE_NAME "tcl"
			#define PACKAGE_STRING "tcl 8.6"
			#define PACKAGE_TARNAME "tcl"
			#define PACKAGE_VERSION "8.6"
			#define STDC_HEADERS 1
			#define TCL_CFGVAL_ENCODING "iso8859-1"
			#define TCL_CFG_OPTIMIZED 1
			#define TCL_COMPILE_DEBUG 1 //TODO-
			#define TCL_LIBRARY "/usr/local/lib/tcl8.6"
			#define TCL_PACKAGE_PATH "/usr/local/lib64 /usr/local/lib "
			#define TCL_SHLIB_EXT ".so"
			#define TCL_THREADS 1
			#define TCL_TOMMATH 1
			#define TCL_UNLOAD_DLLS 1
			#define TCL_WIDE_INT_TYPE long long //TODO ?386?
			#define TIME_WITH_SYS_TIME 1
			#define _REENTRANT 1
			#define _THREAD_SAFE 1
			// #define HAVE_CPUID 1
			// #define HAVE_GETHOSTBYADDR_R_8 1 // uses identifier h_errno -> UB
			// #define HAVE_GETHOSTBYNAME_R 1   // ../../_tcl8.6.8/unix/tclUnixCompat.c:580:20: undefined "compatLock"
			// #define HAVE_GETHOSTBYNAME_R_6 1  // uses identifier h_errno -> UB
			// #define HAVE_PTHREAD_ATFORK 1
			// #define HAVE_PTHREAD_ATTR_SETSTACKSIZE 1
			// #define HAVE_ZLIB 1
			// #define NDEBUG 1
			// #define USE_THREAD_ALLOC 1

			/* Rename the global symbols in libtommath to avoid linkage conflicts */

			#define KARATSUBA_MUL_CUTOFF TclBNKaratsubaMulCutoff
			#define KARATSUBA_SQR_CUTOFF TclBNKaratsubaSqrCutoff
			#define TOOM_MUL_CUTOFF TclBNToomMulCutoff
			#define TOOM_SQR_CUTOFF TclBNToomSqrCutoff

			#define bn_reverse TclBN_reverse
			#define fast_s_mp_mul_digs TclBN_fast_s_mp_mul_digs
			#define fast_s_mp_sqr TclBN_fast_s_mp_sqr
			#define mp_add TclBN_mp_add
			#define mp_add_d TclBN_mp_add_d
			#define mp_and TclBN_mp_and
			#define mp_clamp TclBN_mp_clamp
			#define mp_clear TclBN_mp_clear
			#define mp_clear_multi TclBN_mp_clear_multi
			#define mp_cmp TclBN_mp_cmp
			#define mp_cmp_d TclBN_mp_cmp_d
			#define mp_cmp_mag TclBN_mp_cmp_mag
			#define mp_cnt_lsb TclBN_mp_cnt_lsb
			#define mp_copy TclBN_mp_copy
			#define mp_count_bits TclBN_mp_count_bits
			#define mp_div TclBN_mp_div
			#define mp_div_2 TclBN_mp_div_2
			#define mp_div_2d TclBN_mp_div_2d
			#define mp_div_3 TclBN_mp_div_3
			#define mp_div_d TclBN_mp_div_d
			#define mp_exch TclBN_mp_exch
			#define mp_expt_d TclBN_mp_expt_d
			#define mp_grow TclBN_mp_grow
			#define mp_init TclBN_mp_init
			#define mp_init_copy TclBN_mp_init_copy
			#define mp_init_multi TclBN_mp_init_multi
			#define mp_init_set TclBN_mp_init_set
			#define mp_init_set_int TclBN_mp_init_set_int
			#define mp_init_size TclBN_mp_init_size
			#define mp_karatsuba_mul TclBN_mp_karatsuba_mul
			#define mp_karatsuba_sqr TclBN_mp_karatsuba_sqr
			#define mp_lshd TclBN_mp_lshd
			#define mp_mod TclBN_mp_mod
			#define mp_mod_2d TclBN_mp_mod_2d
			#define mp_mul TclBN_mp_mul
			#define mp_mul_2 TclBN_mp_mul_2
			#define mp_mul_2d TclBN_mp_mul_2d
			#define mp_mul_d TclBN_mp_mul_d
			#define mp_neg TclBN_mp_neg
			#define mp_or TclBN_mp_or
			#define mp_radix_size TclBN_mp_radix_size
			#define mp_read_radix TclBN_mp_read_radix
			#define mp_rshd TclBN_mp_rshd
			#define mp_s_rmap TclBNMpSRmap
			#define mp_set TclBN_mp_set
			#define mp_set_int TclBN_mp_set_int
			#define mp_shrink TclBN_mp_shrink
			#define mp_sqr TclBN_mp_sqr
			#define mp_sqrt TclBN_mp_sqrt
			#define mp_sub TclBN_mp_sub
			#define mp_sub_d TclBN_mp_sub_d
			#define mp_to_unsigned_bin TclBN_mp_to_unsigned_bin
			#define mp_to_unsigned_bin_n TclBN_mp_to_unsigned_bin_n
			#define mp_toom_mul TclBN_mp_toom_mul
			#define mp_toom_sqr TclBN_mp_toom_sqr
			#define mp_toradix_n TclBN_mp_toradix_n
			#define mp_unsigned_bin_size TclBN_mp_unsigned_bin_size
			#define mp_xor TclBN_mp_xor
			#define mp_zero TclBN_mp_zero
			#define s_mp_add TclBN_s_mp_add
			#define s_mp_mul_digs TclBN_s_mp_mul_digs
			#define s_mp_sqr TclBN_s_mp_sqr
			#define s_mp_sub TclBN_s_mp_sub
`
		tclDefs32 = `
			#define mp_digit unsigned long
`
		tclDefs64 = "\n#define mp_digit unsigned long long\n"
	)

	var tclDefs string
	switch arch := env("GOARCH", runtime.GOARCH); arch {
	case "386":
		tclDefs = allDefs + tclDefs0 + tclDefs32
	case "amd64":
		tclDefs = allDefs + tclDefs0 + tclDefs64
	default:
		panic(arch)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-tcl-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}
	testdir := filepath.Join(dir, "test")
	if err := mkdir(testdir); err != nil {
		t.Fatal(err)
	}

	g := newGen(nil, nil)
	g.escAllTLDs = true
	root := "testdata"

	sqliteTweaks := &cc.Tweaks{
		// TrackExpand:                 func(s string) { fmt.Print(s) }, //TODO-
		// TrackIncludes:               func(s string) { fmt.Printf("#include %s\n", s) }, //TODO-
		EnableAnonymousStructFields: true,
		EnableEmptyStructs:          true,
		InjectFinalNL:               true,
	}
	inc := append([]string{
		"@",
		filepath.FromSlash(filepath.Join(root, sqlite, "sqlite-amalgamation-3210000")),
		filepath.FromSlash(filepath.Join(root, tcl, "generic")),
	}, searchPaths...)
	sysInc := append(searchPaths, filepath.FromSlash(filepath.Join(root, sqlite, "sqlite-amalgamation-3210000")))

	for _, v := range []string{
		"sqlite-amalgamation-3210000/sqlite3.c", // Keep this first

		"ext/rbu/test_rbu.c",
		"ext/fts5/fts5_tcl.c",
		"ext/misc/amatch.c",
		"ext/misc/carray.c",
		"ext/misc/closure.c",
		"ext/misc/csv.c",
		"ext/misc/eval.c",
		"ext/misc/fileio.c",
		"ext/misc/fuzzer.c",
		"ext/misc/ieee754.c",
		"ext/misc/mmapwarm.c",
		"ext/misc/nextchar.c",
		"ext/misc/percentile.c",
		"ext/misc/regexp.c",
		"ext/misc/remember.c",
		"ext/misc/series.c",
		"ext/misc/spellfix.c",
		"ext/misc/totype.c",
		"ext/misc/unionvtab.c",
		"ext/misc/wholenumber.c",
	} {
		in := []cc.Source{cc.MustFileSource2(filepath.FromSlash(filepath.Join(root, sqlite, v)), false)}
		tu, err := translate(sqliteTweaks, inc, sysInc, sqliteDefs+tclDefs, in...)
		if err != nil {
			t.Fatal(err)
		}

		if err := g.file(testdir, v, tu); err != nil {
			t.Fatal(err)
		}
	}

	tclTweaks := &cc.Tweaks{
		// TrackExpand:                 func(s string) { fmt.Print(s) },                   //TODO-
		// TrackIncludes:               func(s string) { fmt.Printf("#include %s\n", s) }, //TODO-
		EnableAnonymousStructFields: true,
		EnableEmptyStructs:          true,
		EnableImplicitBuiltins:      true,
		InjectFinalNL:               true,
	}
	inc = []string{
		"@",
		filepath.FromSlash(filepath.Join(root, tcl, "unix")), //TODO Windows
		filepath.FromSlash(filepath.Join(root, tcl, "generic")),
		filepath.FromSlash(filepath.Join(root, tcl, "libtommath")),
	}
	sysInc = append(searchPaths, inc...)
	for _, v := range []string{
		"generic/regcomp.c",
		"generic/regerror.c",
		"generic/regexec.c",
		"generic/regfree.c",
		"generic/tclAlloc.c",
		"generic/tclAssembly.c",
		"generic/tclAsync.c",
		"generic/tclBasic.c",
		"generic/tclBinary.c",
		"generic/tclCkalloc.c",
		"generic/tclClock.c",
		"generic/tclCmdAH.c",
		"generic/tclCmdIL.c",
		"generic/tclCmdMZ.c",
		"generic/tclCompCmds.c",
		"generic/tclCompCmdsGR.c",
		"generic/tclCompCmdsSZ.c",
		"generic/tclCompExpr.c",
		"generic/tclCompile.c",
		"generic/tclConfig.c",
		"generic/tclDate.c",
		"generic/tclDictObj.c",
		"generic/tclDisassemble.c",
		"generic/tclEncoding.c",
		"generic/tclEnsemble.c",
		"generic/tclEnv.c",
		"generic/tclEvent.c",
		"generic/tclExecute.c",
		"generic/tclFCmd.c",
		"generic/tclFileName.c",
		"generic/tclGet.c",
		"generic/tclHash.c",
		"generic/tclHistory.c",
		"generic/tclIO.c",
		"generic/tclIOCmd.c",
		"generic/tclIORChan.c",
		"generic/tclIORTrans.c",
		"generic/tclIOSock.c",
		"generic/tclIOUtil.c",
		"generic/tclIndexObj.c",
		"generic/tclInterp.c",
		"generic/tclLink.c",
		"generic/tclListObj.c",
		"generic/tclLiteral.c",
		"generic/tclLoad.c",
		"generic/tclLoadNone.c", // TclGuessPackageName
		"generic/tclMain.c",
		"generic/tclNamesp.c",
		"generic/tclNotify.c",
		"generic/tclOO.c",
		"generic/tclOOBasic.c",
		"generic/tclOOCall.c",
		"generic/tclOODefineCmds.c",
		"generic/tclOOInfo.c",
		"generic/tclOOMethod.c",
		"generic/tclObj.c",
		"generic/tclOptimize.c",
		"generic/tclPanic.c",
		"generic/tclParse.c",
		"generic/tclPathObj.c",
		"generic/tclPipe.c",
		"generic/tclPkg.c",
		"generic/tclPkgConfig.c",
		"generic/tclPosixStr.c",
		"generic/tclPreserve.c",
		"generic/tclProc.c",
		"generic/tclRegexp.c",
		"generic/tclResolve.c",
		"generic/tclResult.c",
		"generic/tclScan.c",
		"generic/tclStrToD.c",
		"generic/tclStringObj.c",
		"generic/tclThread.c",
		"generic/tclThreadStorage.c",
		"generic/tclTimer.c",
		"generic/tclTomMathInterface.c",
		"generic/tclTrace.c",
		"generic/tclUtf.c",
		"generic/tclUtil.c",
		"generic/tclVar.c",
		"unix/tclUnixChan.c",
		"unix/tclUnixCompat.c",
		"unix/tclUnixEvent.c",
		"unix/tclUnixFCmd.c",
		"unix/tclUnixFile.c",
		"unix/tclUnixInit.c",
		"unix/tclUnixNotfy.c",
		"unix/tclUnixPipe.c",
		"unix/tclUnixSock.c",
		"unix/tclUnixThrd.c",
		"unix/tclUnixTime.c",
	} {
		tu, err := translate(tclTweaks, inc, sysInc, tclDefs, cc.MustFileSource2(filepath.FromSlash(filepath.Join(root, tcl, v)), false))
		if err != nil {
			t.Fatal(err)
		}

		if err := g.file(testdir, v, tu); err != nil {
			t.Fatal(err)
		}
	}

	m, err := filepath.Glob(filepath.FromSlash(filepath.Join(root, tcl, "libtommath/*.c")))
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range m {
		tu, err := translate(tclTweaks, inc, sysInc, tclDefs, cc.MustFileSource2(v, false))
		if err != nil {
			t.Fatal(err)
		}

		if err := g.file(testdir, v, tu); err != nil {
			t.Fatal(err)
		}
	}

	inc = append([]string{
		"@",
		filepath.FromSlash(filepath.Join(root, sqlite, "sqlite-amalgamation-3210000")),
		filepath.FromSlash(filepath.Join(root, tcl, "generic")),
	}, searchPaths...)

	if m, err = filepath.Glob(filepath.FromSlash(filepath.Join(root, sqlite, "src/test*.c"))); err != nil {
		t.Fatal(err)
	}

	for _, v := range m {
		tu, err := translate(sqliteTweaks, inc, sysInc, allDefs+sqliteDefs, cc.MustFileSource2(v, false))
		if err != nil {
			t.Fatal(err)
		}

		if err := g.file(testdir, v, tu); err != nil {
			t.Fatal(err)
		}
	}

	inc = append([]string{
		"@",
		filepath.FromSlash(filepath.Join(root, sqlite, "sqlite-amalgamation-3210000")),
		filepath.FromSlash(filepath.Join(root, tcl, "generic")),
	}, searchPaths...)

	// file with main must be last
	v := "src/tclsqlite.c"
	in := []cc.Source{cc.MustFileSource2(filepath.FromSlash(filepath.Join(root, sqlite, v)), false), cc.MustCrt0()}
	tu, err := translate(sqliteTweaks, inc, searchPaths, allDefs+sqliteDefs, in...)
	if err != nil {
		t.Fatal(err)
	}

	if err := g.file(testdir, v, tu); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	}()

	src, err := filepath.Abs(filepath.Join(cwd, "testdata"))
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(testdir); err != nil {
		t.Fatal(err)
	}

	cc.FlushCache()
	build := "build -o test"
	if *oTCLRace {
		build += " -race"
	}
	out, err := exec.Command("go", strings.Split(build, " ")...).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	if *oEdit {
		fmt.Printf("TCL0\ttclsqlite build ok\n")
	}

	if err := cpDir(filepath.Join(dir, "ext"), filepath.Join(src, filepath.FromSlash("_sqlite/ext")), nil); err != nil {
		t.Fatal(err)
	}

	if err := cpDir(filepath.Join(dir, "library"), filepath.Join(src, filepath.FromSlash("_tcl8.6.8/library")), nil); err != nil {
		t.Fatal(err)
	}

	if err := cpDir(filepath.Join(dir, "test"), filepath.Join(src, filepath.FromSlash("_sqlite/test")), nil); err != nil {
		t.Fatal(err)
	}

	blacklist := []string{
		"btreefault.test",
		"cffault.test",
		"collate1.test",
		"collate2.test",
		"collate3.test",
		"collate4.test",
		"collate5.test",
		"collate6.test",
		"collate9.test",
		"corruptC.test",
		"crash.test",
		"crash2.test",
		"crash3.test",
		"crash4.test",
		"crash6.test",
		"crash7.test",
		"date.test", // crt.Xselect
		"e_createtable.test",
		"e_delete.test",
		"e_insert.test",
		"e_reindex.test",
		"e_select.test",
		"e_update.test",
		"e_walauto.test",
		"exists.test",
		"func4.test",
		"fuzz.test",
		"fuzzerfault.test",
		"ieee754.test",
		"incrcorrupt.test", // crt.Xftruncate
		"incrvacuum_ioerr.test",
		"ioerr3.test",
		"journal3.test", // crt.Xfchmod
		"lock.test",
		"lock4.test", // crt.Xselect
		"lock5.test", // crt.Xutimes
		"malloc.test",
		"minmax.test",
		"misc1.test",
		"misc3.test",
		"misc7.test",    // crt.Xatof
		"mjournal.test", // crt.Xfopen
		"mmap1.test",
		"mmap4.test",
		"multiplex2.test",
		"nan.test",
		"pager1.test",
		"pager4.test", // crt.Xrename
		"pagerfault.test",
		"pagerfault2.test",
		"pagerfault3.test",
		"pragma.test", // crt.X__assert_fail
		"printf.test",
		"quota2.test", // crt.Xfopen
		"rbu.test",    // crt.Xrename
		"reindex.test",
		"rollbackfault.test",
		"rowallock.test",
		"savepoint.test",
		"savepoint4.test",
		"savepointfault.test",
		"schema3.test",
		"select9.test",
		"shared2.test",
		"shared9.test",
		"sharedA.test", // crt.Xpthread_attr_init
		"sort2.test",
		"sort3.test",
		"sort4.test", // crt.Xpthread_create
		"sortfault.test",
		"speed4.test",
		"speed4p.test",
		"statfault.test",
		"superlock.test",
		"symlink.test", // crt.Xsymlink
		"syscall.test",
		"tempfault.test",
		"thread001.test", // crt.Xpthread_attr_init
		"thread002.test", // crt.Xpthread_attr_init
		"thread003.test", // crt.Xpthread_attr_init
		"thread004.test", // crt.Xpthread_attr_init
		"thread005.test", // crt.Xpthread_attr_init
		"thread1.test",   // crt.Xpthread_create
		"thread2.test",   // crt.Xpthread_create
		"tkt-5d863f876e.test",
		"tkt-fc62af4523.test",
		"tkt3838.test",
		"tkt3997.test",
		"trans.test",
		"unionvtabfault.test",
		"unixexcl.test",
		"vacuum2.test",
		"vtabH.test", // crt.Xreaddir64_r
		"wal.test",
		"wal2.test",
		"wal3.test",
		"wal4.test",
		"wal5.test",
		"walcrash.test",
		"walcrash2.test",
		"walcrash4.test",
		"walro.test",
		"walslow.test",
		"walthread.test", // crt.Xpthread_attr_init
		"where.test",
		"whereD.test",
		"writecrash.test",
	}

	for _, v := range blacklist {
		if err := os.Remove(filepath.Join(testdir, v)); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("./test", "all.test")
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	const maxFails = 10
	var green, red int
	var fail []string
	sc := bufio.NewScanner(pipe)
	var lastStdout, lastOk, lastFinished string
	for sc.Scan() {
		lastStdout = sc.Text()
		if *oTrace {
			fmt.Fprintln(os.Stderr, lastStdout)
		}
		switch {
		case strings.HasPrefix(lastStdout, "!") && strings.Contains(lastStdout, "expected"):
			if len(fail) < maxFails {
				fail = append(fail, lastStdout)
			}
		case
			strings.HasPrefix(lastStdout, "Error:"),
			strings.HasPrefix(lastStdout, "!") && strings.Contains(lastStdout, "got"):

			red++
			if len(fail) < maxFails {
				fail = append(fail, lastStdout)
			}
		case strings.HasPrefix(lastStdout, "Time: "):
			lastFinished = lastStdout
		case strings.HasSuffix(lastStdout, "... Ok") && !strings.Contains(lastStdout, "-closeallfiles") && !strings.Contains(lastStdout, "-sharedcachesetting"):
			green++
			lastOk = lastStdout
		}
	}
	total := green + red
	if red > len(fail) {
		fail = append(fail, "... too many fails")
	}
	t.Logf(`
Test cases: %8d
Pass:       %8d (%3.2f%%)
Fail:       %8d (%3.2f%%)
%s`,
		total, green, 100*float64(green)/float64(total), red, 100*float64(red)/float64(total), strings.Join(fail, "\n"),
	)
	if err := cmd.Wait(); err != nil || len(blacklist) != 0 {
		t.Fatalf(`
Test binary exit error: %v
Last completed test file: %q
Last passed test: %q
Last line written to stdout: %q
Blacklisted test files: %d
%s`, err, lastFinished, lastOk, lastStdout, len(blacklist), strings.Join(blacklist, "\n"))
	}
}

func (g *gen) define2(n *cc.Declarator) {
more:
	n = g.normalizeDeclarator(n)
	nm := n.Name()
	done := false
	if n.Linkage == cc.LinkageExternal {
		_, produced := g.producedExterns[nm]
		_, initialized := g.initializedExterns[nm]
		hasInitializer := n.Initializer != nil
		switch {
		case !produced && !initialized && !hasInitializer:
			switch x := underlyingType(n.Type, true).(type) {
			case *cc.ArrayType:
				if x.Size.Value == nil {
					g.incompleteExternArrays[nm] = n
					done = true // Must resolve later
				}
			}
		case !produced && !initialized && hasInitializer:

			// nop here
		case
			produced && !initialized && !hasInitializer,
			produced && initialized && hasInitializer,
			produced && initialized && !hasInitializer:

			done = true
		case produced && !initialized && hasInitializer:
			g.escapedTLD2(n)
			done = true
		default:
			panic(fmt.Errorf("%v: %q produced %v, initialized %v, hasInitializer %v", g.position(n), dict.S(nm), produced, initialized, hasInitializer))
		}
	}

	if !done {
		_, done = g.producedDeclarators[n]
	}
	if !done {
		if n.Linkage == cc.LinkageExternal {
			_, done = g.producedExterns[nm]
			//dbg("%v: %q, %v", g.position(n), dict.S(nm), done)
		}

		if n.Type.Kind() == cc.Function && n.FunctionDefinition == nil {
			done = true
		}
		if !done {
			switch underlyingType(n.Type, true).(type) {
			case
				*cc.ArrayType,
				*cc.EnumType,
				*cc.FunctionType,
				*cc.PointerType,
				*cc.StructType,
				*cc.UnionType,
				cc.TypeKind:

				if n.Linkage == cc.LinkageExternal {
					g.producedExterns[nm] = struct{}{}
				}
				g.producedDeclarators[n] = struct{}{}
				g.tld(n)
			default:
				//dbg("%v: %q %v (%v)", g.position(n), dict.S(nm), n.Type, underlyingType(n.Type, true))
				todo("")
			}
		}
	}

	for g.queue.Front() != nil {
		m := g.queue.Front()
		g.queue.Remove(m)
		switch x := m.Value.(type) {
		case *cc.Declarator:
			n = x
			goto more
		case *cc.EnumType:
			g.defineEnumType(x)
		case *cc.FunctionType:
			// nop
		case *cc.NamedType:
			g.enqueue(x.Type)
		case *cc.PointerType:
			if !x.IsVoidPointerType() {
				g.enqueue(x.Item)
			}
		case *cc.StructType:
			// nop
		case *cc.TaggedEnumType:
			g.defineTaggedEnumType(x)
		case *cc.TaggedStructType:
			switch {
			case x.Type == nil || x.Type == x:
				g.opaqueStructTags[x.Tag] = struct{}{}
			default:
				g.defineTaggedStructType(x)
			}
		case *cc.TaggedUnionType:
			switch {
			case x.Type == nil || x.Type == x:
				g.opaqueStructTags[x.Tag] = struct{}{}
			default:
				g.defineTaggedUnionType(x)
			}
		case cc.TypeKind:
			// nop
		case *cc.UnionType:
			// nop
		default:
			todo("%T %v", x, x)
		}
	}
}

// produced && !initialized && hasInitializer
func (g *gen) escapedTLD2(n *cc.Declarator) {
	switch x := n.Type.(type) {
	case
		*cc.NamedType,
		*cc.TaggedStructType,
		*cc.TaggedUnionType:

		g.enqueue(x)
	}

	nm := n.Name()

	defer func() { g.initializedExterns[nm] = struct{}{} }()

	if g.isConstInitializer(n.Type, n.Initializer) {
		g.w("\n\nfunc init() {")
		g.w("%sCopy(%s, ds+%d, %d)", crt, g.mangleDeclarator(n), g.allocDS(n.Type, n.Initializer), g.model.Sizeof(n.Type))
		g.w("}")
		return
	}

	switch x := cc.UnderlyingType(n.Type).(type) {
	case *cc.ArrayType:
		if x.Item.Kind() == cc.Char && n.Initializer.Expr.Operand.Value != nil {
			todo("%v:", g.position(n))
			g.w("\nvar %s = ds + %d\n", g.mangleDeclarator(n), g.allocDS(n.Type, n.Initializer))
			return
		}
	}

	g.w("\n\nfunc init() {")
	g.w("*(*%s)(unsafe.Pointer(%s)) = ", g.typ(n.Type), g.mangleDeclarator(n))
	g.literal(n.Type, n.Initializer)
	g.w("}")
}

func (g *gen) file(dir, fn0 string, tu *cc.TranslationUnit) error {
	g.enqueued = map[interface{}]struct{}{}
	g.externs = map[int]*cc.Declarator{}
	g.producedDeclarators = map[*cc.Declarator]struct{}{}
	g.staticDeclarators = map[int]*cc.Declarator{}
	fn, err := g.file0(dir, fn0, tu)
	if err != nil {
		return err
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}

	defer func() {
		if e := f.Close(); e != nil && err == nil {
			err = e
		}
	}()

	w := bufio.NewWriter(f)

	defer func() {
		if e := w.Flush(); e != nil && err == nil {
			err = e
		}
	}()

	o := newOpt()
	o.noBool2int = true
	s := ""
	os := ""
	_, crt0 := tu.FileScope.Idents[idStart]
	if crt0 {
		o.forceBool2int = true
		if err := g.crt0(); err != nil {
			return err
		}
		s = `const null = uintptr(0)
`
		os = "\n\t\"os\""
	}

	w.WriteString(fmt.Sprintf(`package main

import (
	"math"
	"unsafe"%s

	"modernc.org/crt"
)

var (
	_ = crt.Malloc
	_ = math.Pi
	_ unsafe.Pointer
)

`, os))
	w.WriteString(s)
	if crt0 {
		g.w(mainSrc, crt)
	}
	return o.do(w, &g.out0, fn, 0)
}

func (g *gen) crt0() error {
	if err := g.errs.Err(); err != nil {
		return fmt.Errorf("%s", errString(err))
	}

	if g.needNZ64 {
		g.w("\n\nfunc init() { nz64 = -nz64 }")
	}
	if g.needNZ32 {
		g.w("\n\nfunc init() { nz32 = -nz32 }")
	}

	var a []string
	for k := range g.opaqueStructTags {
		a = append(a, string(dict.S(k)))
	}
	sort.Strings(a)
	for _, k := range a {
		tag := dict.SID(k)
		if _, ok := g.producedStructTags[tag]; !ok {
			g.w("\ntype S%s struct{ uintptr }\n", k)
		}
	}

	if g.needPreInc {
		g.w("\n\nfunc preinc(p *uintptr, n uintptr) uintptr { *p += n; return *p }")
	}
	if g.needAlloca {
		g.w("\n\nfunc alloca(p *[]uintptr, n int) uintptr { r := %sMustMalloc(n); *p = append(*p, r); return r }", crt)
	}

	g.genHelpers()

	g.w("\n\nvar (\n")
	if g.bss != 0 {
		g.w("bss = %sBSS(&bssInit[0])\n", crt)
		g.w("bssInit [%d]byte\n", g.bss)
	}
	if n := len(g.ds); n != 0 {
		if n < 16 {
			g.ds = append(g.ds, make([]byte, 16-n)...)
		}
		g.w("ds = %sDS(dsInit)\n", crt)
		g.w("dsInit = []byte{")
		if isTesting {
			g.w("\n")
		}
		for i, v := range g.ds {
			g.w("%#02x, ", v)
			if isTesting && i&15 == 15 {
				g.w("// %#x\n", i&^15)
			}
		}
		g.w("}\n")
	}
	if g.needNZ64 {
		g.w("nz64 float64\n")
	}
	if g.needNZ32 {
		g.w("nz32 float32\n")
	}
	g.w("ts = %sTS(\"", crt)
	for _, v := range g.text {
		s := fmt.Sprintf("%q", dict.S(v))
		g.w("%s\\x00", s[1:len(s)-1])
	}
	g.w("\")\n)\n")
	return nil
}

func (g *gen) file0(dir, fn string, tu *cc.TranslationUnit) (f string, err error) {
	returned := false

	defer func() {
		if e := recover(); !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, compact(string(debugStack()), compactStack))
		}
	}()

	bn := filepath.Base(fn)
	bn = bn[:len(bn)-len(filepath.Ext(bn))]
	s := ""
	i := -1
	for {
		fn0 := filepath.Join(dir, bn+s+".go")
		if _, ok := g.filenames[fn0]; !ok {
			f = fn0
			break
		}
		i++
		s = fmt.Sprintf("_%d", i)
	}
	g.in = []*cc.TranslationUnit{tu}
	g.out0 = bytes.Buffer{}
	switch {
	case g.model == nil:
		g.model = tu.Model
		g.fset = tu.FileSet
	default:
		if !g.model.Equal(tu.Model) {
			return "", fmt.Errorf("translation units use different memory models")
		}
	}

	var a []string
	for nm := range tu.FileScope.Idents {
		a = append(a, string(dict.S(nm)))
	}
	sort.Strings(a)
	var a2 []int
next:
	for _, s := range a {
		nm := dict.SID(s)
		n := tu.FileScope.Idents[nm]
		switch x := n.(type) {
		case *cc.Declarator:
			switch x.Linkage {
			case cc.LinkageExternal:
				p := g.position(x).Filename
				for _, v := range searchPaths {
					if strings.HasPrefix(p, v) {
						switch nm := x.Name(); nm {
						case idStdin, idStdout, idStderr:
						default:
							g.producedExterns[nm] = struct{}{}
							continue next
						}
					}
				}

				if x.Type.Kind() == cc.Function && x.Name() == idBacktrace {
					continue
				}

				g.externs[nm] = x
				a2 = append(a2, nm)
			case cc.LinkageInternal:
				g.staticDeclarators[nm] = x
			}
		case *cc.EnumerationConstant:
			// nop
		default:
			todo("%v: %T %s", g.position0(n), x, nm)
		}
	}
	for _, nm := range a2 {
		g.define2(g.externs[nm])
	}
	returned = true
	return f, nil
}

func TestCSmith0(t *testing.T) { //TODO-
	return //TODO
	cc.FlushCache()
	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	gcc, err := exec.LookPath("gcc")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	var inc string
	switch runtime.GOOS {
	case "linux":
		inc = "/usr/include"
	default:
		t.Logf("unsupported OS")
		return
	}
	if _, err := os.Stat(filepath.Join(inc, "csmith.h")); err != nil {
		if os.IsNotExist(err) {
			t.Logf("%s not found: skipping test", inc)
			return
		}

		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "test-ccgo-csmith-")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	const (
		gccBin = "gcc"
		mainC  = "main.c"
	)

	ch := time.After(*oCSmith)
	var cs, cc, ccgo, build, run, ok int
	t0 := time.Now()
out:
	for {
		select {
		case <-ch:
			break out
		default:
		}

		out, err := exec.Command(
			csmith,
			"-o", mainC,
			"--bitfields",            // --bitfields | --no-bitfields: enable | disable full-bitfields structs (disabled by default).
			"--no-const-pointers",    // --const-pointers | --no-const-pointers: enable | disable const pointers (enabled by default).
			"--no-consts",            // --consts | --no-consts: enable | disable const qualifier (enabled by default).
			"--no-packed-struct",     // --packed-struct | --no-packed-struct: enable | disable packed structs by adding #pragma pack(1) before struct definition (disabled by default).
			"--no-volatile-pointers", // --volatile-pointers | --no-volatile-pointers: enable | disable volatile pointers (enabled by default).
			"--no-volatiles",         // --volatiles | --no-volatiles: enable | disable volatiles (enabled by default).
			"--paranoid",             // --paranoid | --no-paranoid: enable | disable pointer-related assertions (disabled by default).
		).Output()
		if err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := exec.Command(gcc, "-w", "-o", gccBin, mainC).CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		var gccOut []byte
		var gccT0 time.Time
		var gccT time.Duration
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout/10)

			defer cancel()

			gccT0 = time.Now()
			gccOut, err = exec.CommandContext(ctx, filepath.Join(dir, gccBin)).CombinedOutput()
			gccT = time.Since(gccT0)
		}()
		if err != nil {
			continue
		}

		cs++
		build0 := build
		os.Remove("main.go")
		ccgoOut, err := test(t, false, &cc, &ccgo, &build, &run, "", "", []string{inc}, dir, []string{mainC})
		if err != nil {
			t.Log(err)
			csmithFatal(t, mainC, gccOut, ccgoOut, cc, ccgo, build, run, ok, cs, gccT)
		}

		if build == build0 {
			continue
		}

		if bytes.Equal(gccOut, ccgoOut) {
			ok++
			if *oEdit {
				fmt.Printf("cc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)\n", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, time.Since(t0))
			}
			continue
		}

		if *oNoCmp {
			continue
		}

		csmithFatal(t, mainC, gccOut, ccgoOut, cc, ccgo, build, run, ok, cs, gccT)
	}
	d := time.Since(t0)
	t.Logf("cc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, d)
	if *oEdit {
		fmt.Printf("CSmith0\tcc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)\n", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, d)
	}
}

func TestCSmith(t *testing.T) {
	regressionTests := []string{
		//gcc bug "--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2117898882",

		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1352671588",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1322542678",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1901462611",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2829621573",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 576413934",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 151038493",
	}

	defaultArgs := strings.Join([]string{
		"--bitfields",                     // --bitfields | --no-bitfields: enable | disable full-bitfields structs (disabled by default).
		"--no-const-pointers",             // --const-pointers | --no-const-pointers: enable | disable const pointers (enabled by default).
		"--no-consts",                     // --consts | --no-consts: enable | disable const qualifier (enabled by default).
		"--no-packed-struct",              // --packed-struct | --no-packed-struct: enable | disable packed structs by adding #pragma pack(1) before struct definition (disabled by default).
		"--no-volatile-pointers",          // --volatile-pointers | --no-volatile-pointers: enable | disable volatile pointers (enabled by default).
		"--no-volatiles",                  // --volatiles | --no-volatiles: enable | disable volatiles (enabled by default).
		"--paranoid",                      // --paranoid | --no-paranoid: enable | disable pointer-related assertions (disabled by default).
		"--max-nested-struct-level", "10", // --max-nested-struct-level <num>: limit maximum nested level of structs to <num>(default 0). Only works in the exhaustive mode.
	}, " ")

	cc.FlushCache()
	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	gcc, err := exec.LookPath("gcc")
	if err != nil {
		t.Logf("%v: skipping test", err)
		return
	}

	var inc string
	switch runtime.GOOS {
	case "linux":
		inc = "/usr/include"
	default:
		t.Logf("unsupported OS")
		return
	}
	if _, err := os.Stat(filepath.Join(inc, "csmith.h")); err != nil {
		if os.IsNotExist(err) {
			t.Logf("%s not found: skipping test", inc)
			return
		}

		t.Fatal(err)
	}

	dir := *oTmp
	if dir == "" {
		var err error
		if dir, err = ioutil.TempDir("", "test-ccgo-smith-"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Fatal(err)
			}
		}()
	}

	compiler, err := compileCCGO(dir)
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	const (
		gccBin = "gcc"
		mainC  = "main.c"
	)

	for _, v := range []string{
		"csmith.h",
		"random_inc.h",
		"safe_math.h",
		"platform_generic.h",
	} {
		if err := cpFile(v, filepath.Join(inc, v), nil); err != nil {
			t.Fatal(v)
		}
	}

	ch := time.After(*oCSmith)
	var cs, cc, ccgo, build, run, ok int
	t0 := time.Now()
	argsHead := "-o " + mainC + " "
out:
	for i := 0; ; i++ {
		extra := ""
		args := argsHead
		switch {
		case i < len(regressionTests):
			args += regressionTests[i]
			a := strings.Split(regressionTests[i], " ")
			extra = strings.Join(a[len(a)-2:], " ")
		default:
			select {
			case <-ch:
				break out
			default:
			}

			args += defaultArgs
		}

		os.Remove(mainC)
		out, err := exec.Command(csmith, strings.Split(args, " ")...).Output()
		if err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := exec.Command(gcc, "-w", "-o", gccBin, mainC).CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		var gccOut []byte
		var gccT0 time.Time
		var gccT time.Duration
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout/10)

			defer cancel()

			gccT0 = time.Now()
			gccOut, err = exec.CommandContext(ctx, filepath.Join(dir, gccBin)).CombinedOutput()
			gccT = time.Since(gccT0)
		}()
		if err != nil {
			continue
		}

		cs++
		os.Remove("main.go")
		if out, err := exec.Command(compiler, "--ccgo-struct-checks", "-o", "main.go", mainC).CombinedOutput(); err != nil {
			t.Log(err)
			csmithFatal(t, mainC, gccOut, out, cc, ccgo, build, run, ok, cs, gccT)
			continue
		}

		cc++
		ccgo++
		if out, err := exec.Command("go", "build", "-o", "test.bin", "main.go").CombinedOutput(); err != nil {
			t.Log(err)
			csmithFatal(t, mainC, gccOut, out, cc, ccgo, build, run, ok, cs, gccT)
			continue
		}

		build++
		ccgoOut, err := exec.Command("./test.bin").CombinedOutput()
		if err != nil {
			t.Log(err)
			csmithFatal(t, mainC, gccOut, ccgoOut, cc, ccgo, build, run, ok, cs, gccT)
		}

		run++
		if bytes.Equal(gccOut, ccgoOut) {
			ok++
			if *oEdit {
				fmt.Printf("cc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v) %s\n", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, time.Since(t0), extra)
			}
			continue
		}

		if *oNoCmp {
			continue
		}

		csmithFatal(t, mainC, gccOut, ccgoOut, cc, ccgo, build, run, ok, cs, gccT)
	}
	d := time.Since(t0)
	t.Logf("cc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, d)
	if *oEdit {
		fmt.Printf("CSmith0\tcc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)\n", cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, d)
	}
}

func csmithFatal(t *testing.T, mainC string, gccOut, ccgoOut []byte, cc, ccgo, build, run, ok, cs int, gccT time.Duration) {
	b, err := ioutil.ReadFile(mainC)
	if err != nil {
		t.Fatal(err)
	}

	b2, err := ioutil.ReadFile("main.go")
	if err != nil {
		b2 = nil
	}

	t.Fatalf(`
==== CSmith code ==============================================================
%s
==== Go code (if any ) ========================================================
%s
===============================================================================
 GCC   time: %v
 GCC output: %s
CCGO output: %s
cc %v ccgo %v build %v run %v ok %v (%.2f%%) csmith %v (%v)
`,
		b, b2, gccT, bytes.TrimSpace(gccOut), bytes.TrimSpace(ccgoOut),
		cc, ccgo, build, run, ok, 100*float64(ok)/float64(cs), cs, *oCSmith)
}
