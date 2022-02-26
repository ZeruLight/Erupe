// Copyright 2020 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v3/lib"

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
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/dustin/go-humanize"
	"github.com/pmezard/go-difflib/difflib"
	"modernc.org/cc/v3"
	"modernc.org/ccorpus"
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

func stack() string { return string(debug.Stack()) }

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO, stack) //TODOOK
}

// ----------------------------------------------------------------------------

var (
	fs = ccorpus.FileSystem()

	oBlackBox   = flag.String("blackbox", "", "Record CSmith file to this file")
	oCSmith     = flag.Duration("csmith", 15*time.Minute, "")
	oCpp        = flag.Bool("cpp", false, "Amend compiler errors with preprocessor output")
	oDebug      = flag.Bool("debug", false, "")
	oFullPaths  = flag.Bool("full-paths", false, "")
	oGCC        = flag.String("gcc", "", "")
	oKeep       = flag.Bool("keep", false, "keep temp directories")
	oKeepTmp    = flag.Bool("keep-tmp", false, "")
	oO          = flag.Int("O", 1, "")
	oRE         = flag.String("re", "", "")
	oStackTrace = flag.Bool("trcstack", false, "")
	oTrace      = flag.Bool("trc", false, "Print tested paths.")
	oTraceF     = flag.Bool("trcf", false, "Print test file content")
	oTraceO     = flag.Bool("trco", false, "Print test output")
	oTrc2       = flag.Bool("trc2", false, "")
	oXTags      = flag.String("xtags", "", "passed to go build of TestSQLite")
	writeFailed = flag.Bool("write-failed", false, "Write all failed tests into a file called FAILED in the cwd, in the format of go maps for easy copy-pasting.")

	gccDir    = filepath.FromSlash("testdata/gcc-9.1.0")
	sqliteDir = filepath.FromSlash("testdata/sqlite-amalgamation-3330000")
	tccDir    = filepath.FromSlash("testdata/tcc-0.9.27")

	overlayDir           string
	re                   *regexp.Regexp
	systemCC             string
	systemCCVersion      string
	tempDir              string
	testWD               string
	initIncludePathsOnce sync.Once
	includePaths         []string
	predefined           string
	sysIncludePaths      []string

	keep = map[string]struct{}{
		"go.mod": {},
		"go.sum": {},
	}

	csmithDefaultArgs = strings.Join([]string{
		"--bitfields",                     // --bitfields | --no-bitfields: enable | disable full-bitfields structs (disabled by default).
		"--max-nested-struct-level", "10", // --max-nested-struct-level <num>: limit maximum nested level of structs to <num>(default 0). Only works in the exhaustive mode.
		"--no-const-pointers",    // --const-pointers | --no-const-pointers: enable | disable const pointers (enabled by default).
		"--no-consts",            // --consts | --no-consts: enable | disable const qualifier (enabled by default).
		"--no-packed-struct",     // --packed-struct | --no-packed-struct: enable | disable packed structs by adding #pragma pack(1) before struct definition (disabled by default).
		"--no-volatile-pointers", // --volatile-pointers | --no-volatile-pointers: enable | disable volatile pointers (enabled by default).
		"--no-volatiles",         // --volatiles | --no-volatiles: enable | disable volatiles (enabled by default).
		"--paranoid",             // --paranoid | --no-paranoid: enable | disable pointer-related assertions (disabled by default).
	}, " ")
)

func TestMain(m *testing.M) {
	var rc int
	defer func() {
		if err := recover(); err != nil {
			rc = 1
			fmt.Fprintf(os.Stderr, "PANIC: %v\n%s\n", err, debug.Stack())
		}
		os.Exit(rc)
	}()

	// fmt.Printf("test binary compiled for %s/%s\n", runtime.GOOS, runtime.GOARCH)
	// fmt.Printf("temp dir: %s\n", os.TempDir()) //TODO-
	// if s := os.Getenv("CCGO_CPP"); s != "" {
	// 	fmt.Printf("CCGO_CPP=%s\n", os.Getenv("CCGO_CPP"))
	// }

	flag.BoolVar(&oTraceW, "trcw", false, "Print generator writes")
	flag.BoolVar(&oTraceG, "trcg", false, "Print generator output")
	flag.BoolVar(&oTracePin, "trcpin", false, "Print pinning")
	flag.Parse()
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
	var err error
	if testWD, err = os.Getwd(); err != nil {
		panic("Cannot determine working dir: " + err.Error())
	}
	s := filepath.FromSlash("testdata/overlay")
	if overlayDir, err = filepath.Abs(s); err != nil {
		panic(err) //TODOOK
	}
	if *oGCC == "" {
		var err error
		initIncludePathsOnce.Do(func() { err = initIncludePaths("") })
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if systemCC, err = exec.LookPath(env("CC", "gcc")); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "CC=%s\n", systemCC)
		out, err := exec.Command(systemCC, "--version").CombinedOutput()
		if err == nil {
			if a := strings.Split(string(out), "\n"); len(a) > 0 {
				systemCCVersion = a[0]
				fmt.Fprintf(os.Stderr, "%s\n", systemCCVersion)
			}
		}

		os.Exit(testMain(m))
	}

	var args []string
	for i, v := range os.Args {
		if v == "-gcc" {
			args = append(os.Args[:i], os.Args[i+2:]...)
		}
	}
	a := strings.Split(*oGCC, ",")
	for _, suffix := range a {
		systemCC = fmt.Sprintf("gcc-%s", suffix)
		systemCPP := fmt.Sprintf("cpp-%s", suffix)
		var err error
		if systemCC, err = exec.LookPath(systemCC); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", systemCC, err)
			continue
		}

		if systemCPP, err = exec.LookPath(systemCPP); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", systemCPP, err)
			continue
		}

		os.Setenv("CC", systemCC)
		os.Setenv("CCGO_CPP", systemCPP)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			rc = 1
		}
	}
	os.Exit(rc)
}

func initGoMod() error {
	switch os.Getenv("GO111MODULE") {
	case "off":
		return nil
	}

	dummy := filepath.Join(tempDir, "dummy.go")
	if err := ioutil.WriteFile(dummy, []byte(`
package main

import (
	"modernc.org/libc"
)

var (
	_ libc.TLS
)
func main() {}
`), 0600); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer os.Chdir(wd)

	if err := os.Chdir(tempDir); err != nil {
		return err
	}

	if b, err := exec.Command("go", "mod", "init", "example.com/ccgotest").CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init: %s\nFAIL: %v", b, err)
	}

	if b, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %s\nFAIL: %v", b, err)
	}

	return nil
}

func testMain(m *testing.M) int {
	var err error
	tempDir, err = ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		panic(err) //TODOOK
	}

	if err = initGoMod(); err != nil {
		panic(err) //TODOOK
	}

	switch {
	case *oKeepTmp:
		fmt.Fprintf(os.Stderr, "keeping temporary directory %s\n", tempDir)
	default:
		defer os.RemoveAll(tempDir)
	}

	s := filepath.FromSlash("testdata/overlay")
	if overlayDir, err = filepath.Abs(s); err != nil {
		panic(err) //TODOOK
	}

	return m.Run()
}

func initIncludePaths(cpp string) error {
	var err error
	predefined, includePaths, sysIncludePaths, err = cc.HostConfig(cpp)
	if err != nil {
		return err
	}

	includePaths = append(includePaths, "@")
	includePaths = append(includePaths, sysIncludePaths...)
	return nil
}

type golden struct {
	t *testing.T
	f *os.File
	w *bufio.Writer
}

func newGolden(t *testing.T, fn string) *golden {
	if *oRE != "" {
		return &golden{w: bufio.NewWriter(ioutil.Discard)}
	}

	f, err := os.Create(filepath.FromSlash(fn))
	if err != nil { // Possibly R/O fs in a VM
		base := filepath.Base(filepath.FromSlash(fn))
		f, err = ioutil.TempFile("", base)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("writing results to %s\n", f.Name())
	}

	w := bufio.NewWriter(f)
	return &golden{t, f, w}
}

func (g *golden) close() {
	if g.f == nil {
		return
	}

	if err := g.w.Flush(); err != nil {
		g.t.Fatal(err)
	}

	if err := g.f.Close(); err != nil {
		g.t.Fatal(err)
	}
}

func h(v interface{}) string {
	switch x := v.(type) {
	case int:
		return humanize.Comma(int64(x))
	case int64:
		return humanize.Comma(x)
	case uint64:
		return humanize.Comma(int64(x))
	case float64:
		return humanize.CommafWithDigits(x, 0)
	default:
		panic(fmt.Errorf("%T", x)) //TODOOK
	}
}

type runResult struct {
	ccTime    time.Duration
	csmithSrc []byte
	ccgoTime  time.Duration
	err       error
	name      string
	out       []byte
}

type skipErr string

func (e skipErr) Error() string { return "skipped: " + string(e) }

type runTask struct {
	args      []string
	c         chan *runResult
	cmd       string
	csmithSrc []byte
	opts      []string
	src       string

	ccCanFail       bool
	doNotExec       bool
	hasBinaryOutput bool
}

func (t *runTask) run() {
	r := &runResult{name: t.src}
	r.out, r.err, r.ccTime, r.ccgoTime = t.run0()
	t.c <- r
}

func (t *runTask) run0() (_ []byte, err error, ccTime, ccgoTime time.Duration) {
	const outLimit = 1 << 16
	defer func() {
		if e := recover(); e != nil {
			switch {
			case err == nil:
				err = fmt.Errorf("PANIC: %v\n%s", e, debug.Stack())
			default:
				err = fmt.Errorf("%v\nPANIC: %v\n%s", err, e, debug.Stack())
			}
		}
	}()

	overlay := filepath.Join(overlayDir, t.src)
	b, err := ioutil.ReadFile(overlay)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err, ccTime, ccgoTime
		}

		f, err := fs.Open(t.src)
		if err != nil {
			return nil, err, ccTime, ccgoTime
		}

		if b, err = ioutil.ReadAll(f); err != nil {
			return nil, err, ccTime, ccgoTime
		}

		if err = f.Close(); err != nil {
			return nil, err, ccTime, ccgoTime
		}
	}

	overlay = filepath.Join(overlayDir, t.src+".expectrc")
	b2, err := ioutil.ReadFile(overlay)
	if err != nil {
		f, err := fs.Open(t.src + ".expectrc")
		if err == nil {
			if b2, err = ioutil.ReadAll(f); err != nil {
				return nil, err, ccTime, ccgoTime
			}

			if err = f.Close(); err != nil {
				return nil, err, ccTime, ccgoTime
			}
		}
	}
	var expectRC int
	if len(b2) != 0 {
		s := strings.TrimSpace(string(b2))
		n, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err, ccTime, ccgoTime
		}

		expectRC = int(n)
	}

	baseName := filepath.Base(t.src)
	if err := ioutil.WriteFile(baseName, b, 0600); err != nil {
		return nil, err, ccTime, ccgoTime
	}

	args, err := getArgs(t.src)
	if err != nil {
		return nil, err, ccTime, ccgoTime
	}

	ccArgs := append([]string{"-lm"}, t.opts...)
	ok := true
	for _, v := range t.opts {
		if strings.HasPrefix(v, "-O") {
			ok = false
			break
		}
	}
	if ok {
		if o := *oO; o >= 0 {
			ccArgs = append(ccArgs, fmt.Sprintf("-O%d", o))
		}
	}
	if t.doNotExec {
		ccArgs = append(ccArgs, "-c")
	}
	binary, err := makeCCBinary(baseName, t.doNotExec, ccArgs...)
	if err != nil {
		return nil, skipErr(err.Error()), ccTime, ccgoTime
	}

	const (
		ccOut   = "cc.out"
		ccgoOut = "ccgo.out"
	)
	var binaryBytes, binaryBytes2 int
	var expected []byte
	if !t.doNotExec {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		if t.cmd != "" {
			binary = t.cmd
		}
		if len(t.args) != 0 {
			args = t.args
		}
		t0 := time.Now()
		if *oTrc2 {
			fmt.Fprintf(os.Stderr, "%v: started CC binary for %s: %v %v\n", t0, baseName, binary, args)
		}
		switch {
		case t.hasBinaryOutput:
			binaryBytes, err = execute(ctx, binary, ccOut, args)
			defer os.Remove(ccOut)
		default:
			expected, err = exec.CommandContext(ctx, binary, args...).CombinedOutput()
			if len(expected) > outLimit {
				panic(todo("", t.src, len(expected)))
			}
		}
		ccTime = time.Since(t0)
		if *oTrc2 {
			switch {
			case t.hasBinaryOutput:
				fmt.Fprintf(os.Stderr, "%v: CC binary for %s returned: %v bytes, err %v\n", time.Now(), baseName, binaryBytes, err)
			default:
				fmt.Fprintf(os.Stderr, "%v: CC binary for %s returned: err %v\n%s\n", time.Now(), baseName, err, expected)
			}
		}
		if err != nil {
			switch {
			case t.ccCanFail:
				expected = nil
				expectRC = 0
			default:
				rc := err.(*exec.ExitError).ProcessState.ExitCode()
				if rc != expectRC {
					return nil, skipErr(fmt.Sprintf("executing CC binary %v %v: %v (rc %v, expected %v)\n%s", binary, args, err, rc, expectRC, expected)), ccTime, ccgoTime
				}

				err = nil
			}
		}

		if *oTraceO {
			switch {
			case t.hasBinaryOutput:
				fmt.Fprintf(os.Stderr, "%v %q: %d bytes\n", ccTime, args, binaryBytes)
			default:
				fmt.Fprintf(os.Stderr, "%v %q: %s\n", ccTime, args, expected)
			}
		}
	}

	if t.cmd == "" {
		if err := os.Remove(binary); err != nil {
			return nil, fmt.Errorf("removing %v: %v", binary, err), ccTime, ccgoTime
		}
	}

	ccgoArgs := append([]string(nil), t.opts...)
	if *oFullPaths {
		ccgoArgs = append(ccgoArgs, "-full-paths-comments")
	}
	if binary, err = makeBinary(t.src, t.doNotExec, ccgoArgs...); err != nil {
		return nil, err, ccTime, ccgoTime
	}

	var got []byte
	if !t.doNotExec {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		if t.cmd != "" {
			binary = t.cmd
		}
		if len(t.args) != 0 {
			args = t.args
		}
		t0 := time.Now()
		if *oTrc2 {
			fmt.Fprintf(os.Stderr, "%v: started ccgo binary for %s: %v %v\n", t0, baseName, binary, args)
		}
		switch {
		case t.hasBinaryOutput:
			binaryBytes2, err = execute(ctx, binary, ccgoOut, args)
			defer os.Remove(ccgoOut)
		default:
			got, err = exec.CommandContext(ctx, binary, args...).CombinedOutput()
			if len(got) > outLimit {
				panic(todo("", t.src, len(expected)))
			}
		}
		ccgoTime = time.Since(t0)
		if *oTrc2 {
			switch {
			case t.hasBinaryOutput:
				fmt.Fprintf(os.Stderr, "%v: ccgo binary for %s returned: %v bytes, err %v\n", time.Now(), baseName, binaryBytes2, err)
			default:
				fmt.Fprintf(os.Stderr, "%v: ccgo binary for %s returned: err %v\n%s\n", time.Now(), baseName, err, got)
			}
		}
		if err != nil {
			rc := err.(*exec.ExitError).ProcessState.ExitCode()
			if rc != expectRC {
				return nil, fmt.Errorf("executing ccgo binary %v %v: %v (rc %v, expected %v)\n%s", binary, args, err, rc, expectRC, got), ccTime, ccgoTime
			}

			err = nil
		}

		if *oTraceO {
			switch {
			case t.hasBinaryOutput:
				fmt.Fprintf(os.Stderr, "%v %q: %d bytes\n", ccgoTime, args, binaryBytes2)
			default:
				fmt.Fprintf(os.Stderr, "%v %q: %s\n", ccgoTime, args, got)
			}
		}
		switch {
		case t.hasBinaryOutput:
			if err := fileEqual(ccgoOut, ccOut); err != nil {
				return nil, fmt.Errorf("binary output: %s", err), ccTime, ccgoTime
			}
		default:
			got := string(got)
			expected := string(expected)
			got = strings.ReplaceAll(got, "\r", "")
			got = lineTrim(strings.TrimSpace(got))
			expected = strings.ReplaceAll(expected, "\r", "")
			expected = lineTrim(strings.TrimSpace(expected))
			if got != expected {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(expected),
					B:        difflib.SplitLines(got),
					FromFile: "expected",
					ToFile:   "got",
					Context:  3,
				}
				text, _ := difflib.GetUnifiedDiffString(diff)
				return nil, fmt.Errorf(
					"%v: text output differs:\n%s\n---- x.c\ngot\n%s\nexp\n%s\ngot\n%s\nexp\n%s",
					t.src, text,
					hex.Dump([]byte(got)), hex.Dump([]byte(expected)),
					got, expected,
				), ccTime, ccgoTime
			}
		}
	}
	return got, err, ccTime, ccgoTime
}

func lineTrim(s string) string {
	a := strings.Split(s, "\n")
	for i, v := range a {
		a[i] = strings.TrimSpace(v)
	}
	return strings.Join(a, "\n")
}

func fileEqual(g, e string) error {
	fig, err := os.Stat(g)
	if err != nil {
		return err
	}

	fie, err := os.Stat(e)
	if err != nil {
		return err
	}

	if g, e := fig.Size(), fie.Size(); g != e {
		return fmt.Errorf("files sizes differ, got %v, expected %v", g, e)
	}

	rem := fig.Size()
	if rem == 0 {
		return nil
	}

	var bg, be [4096]byte
	fg, err := os.Open(g)
	if err != nil {
		return err
	}

	defer fg.Close()

	fe, err := os.Open(e)
	if err != nil {
		return err
	}

	defer fe.Close()

	for rem != 0 {
		n, err := io.ReadFull(fg, bg[:])
		if n == 0 {
			if err == io.EOF {
				err = nil
			}
			return err
		}

		n2, err := io.ReadFull(fe, be[:])
		if n == 0 {
			if err == io.EOF {
				err = nil
			}
			return err
		}

		if n != n2 {
			panic(todo("", n, n2))
		}

		if !bytes.Equal(bg[:n], be[:n]) {
			return fmt.Errorf("files are different")
		}

		rem -= int64(n)
	}
	return nil
}

var ftoken uint32

func newID() uint32 { return atomic.AddUint32(&ftoken, 1) }

func makeBinary(src string, obj bool, args ...string) (executable string, err error) {
	defer func() {
		if err != nil {
			if *oTrace {
				fmt.Println(err)
			}
			err = cpp(*oCpp, args, err)
			err = fmt.Errorf("%s: %v", src, err)
		}
	}()

	pkg := "main"
	if obj {
		pkg = "foo"
	}

	main := fmt.Sprintf("main%d.go", newID())
	src = filepath.Base(src)
	if err := NewTask(append([]string{"ccgo", "-o", main, "-pkgname", pkg, "-nocapi", src}, args...), nil, nil).Main(); err != nil {
		return "", err
	}

	if *oTraceF {
		b, _ := ioutil.ReadFile(main)
		fmt.Printf("\n----\n%s\n----\n", b)
	}

	executable = fmt.Sprintf("./%s%d", src[:len(src)-len(filepath.Ext(src))], newID())
	var ext string
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	executable += ext
	os.Remove(executable)
	var b []byte
	switch {
	case obj:
		b, err = exec.Command("go", "build", main).CombinedOutput()
	default:
		b, err = exec.Command("go", "build", "-o", executable, main).CombinedOutput()
	}
	if err != nil {
		err = fmt.Errorf("%s\n\tFAIL: %v", b, err)
	}
	return executable, err
}

type countingWriter struct {
	written int
	w       *bufio.Writer
}

func (c *countingWriter) Write(b []byte) (int, error) {
	n, err := c.w.Write(b)
	c.written += n
	return n, err
}

var _ io.Writer = (*countingWriter)(nil)

// err = execute(ctx, executable, args, ccOut)
func execute(ctx context.Context, executable, out string, args []string) (n int, err error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	f, err := os.Create(out)
	if err != nil {
		return 0, err
	}

	defer func() {
		if e := f.Close(); e != nil && err == nil {
			err = e
		}
	}()

	w := &countingWriter{w: bufio.NewWriter(f)}

	defer func() {
		if e := w.w.Flush(); e != nil && err == nil {
			err = e
		}
	}()

	cmd.Stdout = w
	err = cmd.Run()
	return w.written, err
}

func makeCCBinary(src string, obj bool, args ...string) (executable string, err error) {
	ext := ""
	if obj {
		ext = ".o"
	}
	src = filepath.Base(src)
	executable = "./" + src[:len(src)-len(filepath.Ext(src))]
	if runtime.GOOS == "windows" && !obj {
		ext = ".exe"
	}
	executable += ext
	os.Remove(executable)
	b, err := exec.Command(systemCC, append([]string{"-o", executable, src}, args...)...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v %v -o %v %v: system C compiler: %v\n%s", systemCC, args, executable, src, err, b)
	}

	return executable, nil
}

func getArgs(src string) (args []string, err error) {
	src = src[:len(src)-len(filepath.Ext(src))] + ".arg"
	overlay := filepath.Join(overlayDir, src)
	b, err := ioutil.ReadFile(overlay)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		f, err := fs.Open(src)
		if err != nil {
			return nil, nil
		}

		if b, err = ioutil.ReadAll(f); err != nil {
			return nil, err
		}

		if err = f.Close(); err != nil {
			return nil, err
		}
	}

	a := strings.Split(strings.TrimSpace(string(b)), "\n")
	for _, v := range a {
		switch {
		case strings.HasPrefix(v, "\"") || strings.HasPrefix(v, "`"):
			w, err := strconv.Unquote(v)
			if err != nil {
				return nil, fmt.Errorf("%s: %v: %v", src, v, err)
			}

			args = append(args, w)
		default:
			args = append(args, v)
		}
	}
	return args, nil
}

func TestTCC(t *testing.T) {
	const root = "/tcc-0.9.27/tests/tests2"
	g := newGolden(t, fmt.Sprintf("testdata/tcc_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	needFiles(t, root, []string{
		"18_include.h",
		"95_bitfields.c",
	})
	blacklist := map[string]struct{}{
		"60_errors_and_warnings.c":    {}, // no main
		"73_arm64.c":                  {}, // does not work properly on any gcc tested (7-11)
		"76_dollars_in_identifiers.c": {}, // `int $ = 10;` etc.
		"77_push_pop_macro.c":         {}, // unsupported push/pop macro
		"78_vla_label.c":              {}, //MAYBE
		"79_vla_continue.c":           {}, //MAYBE
		"80_flexarray.c":              {}, //MAYBE
		"83_utf8_in_identifiers.c":    {}, // No support before gcc 10.
		"85_asm-outside-function.c":   {}, // asm
		"90_struct-init.c":            {}, // 90_struct-init.c:168:25: `...`: expected ]
		"94_generic.c":                {}, // 94_generic.c:36:18: `int`: expected primary-expression
		"95_bitfields.c":              {}, // Included from 95_bitfields_ms.c
		"96_nodata_wanted.c":          {}, // no main
		"98_al_ax_extend.c":           {}, // asm
		"99_fastcall.c":               {}, // asm

		"95_bitfields_ms.c": {}, //TODO
	}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "linux/s390x":
		blacklist["function forward decl 2.c"] = struct{}{} //TODO
		blacklist["if bool eq int 0.c"] = struct{}{}        //TODO
	case "freebsd/amd64":
		blacklist["40_stdio.c"] = struct{}{} //TODO
	case "netbsd/amd64":
		blacklist["40_stdio.c"] = struct{}{} //TODO
	case "darwin/amd64":
		blacklist["40_stdio.c"] = struct{}{} //TODO
	}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, false, false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func run(src string, binaryOut, ccCanFail, doNotExec bool, c chan *runResult, opts ...string) {
	(&runTask{
		c:               c,
		ccCanFail:       ccCanFail,
		doNotExec:       doNotExec,
		hasBinaryOutput: binaryOut,
		opts:            opts,
		src:             src,
	}).run()
}

func walk(dir string, f func(pth string, fi os.FileInfo) error) error {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	root, err := fs.Open(dir)
	if err != nil {
		return err
	}

	fi, err := root.Stat()
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s: not a directory", fi.Name())
	}

	fis, err := root.Readdir(-1)
	if err != nil {
		return err
	}

	for _, v := range fis {
		switch {
		case v.IsDir():
			if err = walk(v.Name(), f); err != nil {
				return err
			}
		default:
			if err = f(v.Name(), v); err != nil {
				return err
			}
		}
	}
	return nil
}

func needFiles(t *testing.T, root string, a []string) {
	for _, v := range a {
		overlay := filepath.Join(overlayDir, filepath.FromSlash(root), v)
		b, err := ioutil.ReadFile(overlay)
		if err != nil {
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}

			f, err := fs.Open(path.Join(root, v))
			if err != nil {
				t.Fatal(err)
			}

			if b, err = ioutil.ReadAll(f); err != nil {
				t.Fatal(err)
			}

			if err = f.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if dir, _ := filepath.Split(v); dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
		}

		if err := ioutil.WriteFile(v, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func mustEmptyDir(t *testing.T, s string, keep map[string]struct{}) {
	if err := emptyDir(s, keep); err != nil {
		t.Fatal(err)
	}
}

func emptyDir(s string, keep map[string]struct{}) error {
	m, err := filepath.Glob(filepath.FromSlash(s + "/*"))
	if err != nil {
		return err
	}

	for _, v := range m {
		fi, err := os.Stat(v)
		if err != nil {
			return err
		}

		switch {
		case fi.IsDir():
			if err = os.RemoveAll(v); err != nil {
				return err
			}
		default:
			if _, ok := keep[filepath.Base(v)]; ok {
				break
			}

			if err = os.Remove(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func cpp(enabled bool, args []string, err0 error) error {
	if !enabled {
		return err0
	}

	args = append(args, "-E")
	var out bytes.Buffer
	if err := NewTask(args, &out, &out).Main(); err != nil {
		return fmt.Errorf("error while acquiring preprocessor output: %v\n%v", err, err0)
	}

	return fmt.Errorf("preprocessor output:\n%s\n%v", out.Bytes(), err0)
}

func trim(b []byte) (r []byte) {
	b = bytes.ReplaceAll(b, []byte{'\r'}, nil)
	b = bytes.TrimLeft(b, "\n")
	b = bytes.TrimRight(b, "\n")
	a := bytes.Split(b, []byte("\n"))
	for i, v := range a {
		a[i] = bytes.TrimRight(v, " ")
	}
	return bytes.Join(a, []byte("\n"))
}

func noExt(s string) string {
	ext := filepath.Ext(s)
	if ext == "" {
		panic("internal error") //TODOOK
	}
	return s[:len(s)-len(ext)]
}

func copyFile(src, dst string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, b, 0660)
}

func skipDir(path string) error {
	if strings.HasPrefix(filepath.Base(path), ".") {
		return filepath.SkipDir
	}

	return nil
}

func TestCAPI(t *testing.T) {
	task := NewTask(nil, nil, nil)
	pkgName, capi, err := task.capi("modernc.org/libc")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := capi["printf"]; !ok {
		t.Fatal("default libc does not export printf")
	}

	t.Log(pkgName, capi)
}

const text = "abcd\nefgh\x00ijkl"

var (
	text1 = text
	text2 = (*reflect.StringHeader)(unsafe.Pointer(&text1)).Data
)

func TestText(t *testing.T) {
	p := text2
	var b []byte
	for i := 0; i < len(text); i++ {
		b = append(b, *(*byte)(unsafe.Pointer(p)))
		p++
	}
	if g, e := string(b), text; g != e {
		t.Fatalf("%q %q", g, e)
	}
}

func TestMirBenchmarks(t *testing.T) {
	const root = "/github.com/vnmakarov/mir/c-benchmarks"
	g := newGolden(t, fmt.Sprintf("testdata/mir_c_benchmarks_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	needFiles(t, root, []string{
		"simple_hash.h",
	})
	blacklist := map[string]struct{}{
		"except.c": {}, // longjmp
	}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "darwin/amd64":
		blacklist["method-call.c"] = struct{}{} //TODO
	case "windows/amd64":
		blacklist["except.c"] = struct{}{}     //TODO
		blacklist["mandelbrot.c"] = struct{}{} //TODO
	case "windows/386":
		blacklist["mandelbrot.c"] = struct{}{} //TODO
	}
	binary := map[string]bool{
		"mandelbrot.c": true,
	}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestMirAndrewChambers(t *testing.T) {
	const root = "/github.com/vnmakarov/mir/c-tests/andrewchambers_c"
	g := newGolden(t, fmt.Sprintf("testdata/mir_andrew_chambers_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	blacklist := map[string]struct{}{
		"0011-switch1.c": {}, //TODO
		"0025-duff.c":    {}, //TODO
		"0028-inits06.c": {}, //TODO
		"0028-inits10.c": {}, //TODO
		"0028-inits11.c": {}, //TODO
		"0028-inits12.c": {}, //TODO
		"0028-inits13.c": {}, //TODO
		"0028-inits15.c": {}, //TODO
	}
	binary := map[string]bool{}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestMirLacc(t *testing.T) {
	const root = "/github.com/vnmakarov/mir/c-tests/lacc"
	g := newGolden(t, fmt.Sprintf("testdata/mir_lacc_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	needFiles(t, root, []string{
		"hello.c",
		"header.h",
	})
	blacklist := map[string]struct{}{
		"anonymous-struct.c":       {}, //TODO
		"array-registers.c":        {}, //TODO
		"bitfield-basic.c":         {}, //TODO
		"bitfield-extend.c":        {}, //TODO
		"bitfield-pack-next.c":     {}, //TODO
		"bitfield-trailing-zero.c": {}, //TODO
		"bitfield-types-init.c":    {}, //TODO
		"bitfield.c":               {}, //TODO
		"conditional-void.c":       {}, //TODO
		"declarator-complex.c":     {}, //TODO
		"duffs-device.c":           {}, //TODO
		"function-incomplete.c":    {}, //TODO
		"function-pointer-call.c":  {}, //TODO
		"function-pointer.c":       {}, //TODO
		"function.c":               {}, //TODO
		"hello.c":                  {}, //TODO
		"immediate-expr.c":         {}, //TODO
		"include.c":                {}, //TODO
		"ldouble-load-direct.c":    {}, //TODO
		"long-double-arithmetic.c": {}, //TODO
		"long-double-compare.c":    {}, //TODO
		"long-double-function.c":   {}, //TODO
		"long-double-load.c":       {}, //TODO
		"long-double-struct.c":     {}, //TODO
		"long-double-union.c":      {}, //TODO
		"macro-paste.c":            {}, //TODO
		"macro.c":                  {}, //TODO
		"pointer.c":                {}, //TODO
		"string-addr.c":            {}, //TODO
		"string-concat.c":          {}, //TODO
		"string-escape.c":          {}, //TODO
		"string-index.c":           {}, //TODO
		"string-length.c":          {}, //TODO
		"stringify.c":              {}, //TODO
		"strings.c":                {}, //TODO
		"token.c":                  {}, //TODO
		"typedef.c":                {}, //TODO
		"union-bitfield.c":         {}, //TODO
		"vararg-complex-1.c":       {}, //TODO
		"vararg-complex-2.c":       {}, //TODO
		"vararg.c":                 {}, //TODO
		"whitespace.c":             {}, //TODO
	}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "linux/386":
		blacklist["convert-unsigned-float.c"] = struct{}{} //TODO  Go1.18 https://github.com/golang/go/issues/48807 ?
	case "windows/amd64":
		blacklist["convert-unsigned-float.c"] = struct{}{} //TODO
		blacklist["immediate-pointer.c"] = struct{}{}      //TODO
		blacklist["unsigned-sign-extend.c"] = struct{}{}   //TODO
	case "windows/386":
		blacklist["convert-unsigned-float.c"] = struct{}{} //TODO
		blacklist["unsigned-sign-extend.c"] = struct{}{}   //TODO
	}
	binary := map[string]bool{}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestMirNew(t *testing.T) {
	const root = "/github.com/vnmakarov/mir/c-tests/new"
	g := newGolden(t, fmt.Sprintf("testdata/mir_new_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	blacklist := map[string]struct{}{
		// 1: /github.com/vnmakarov/mir/c-tests/new/endif.c
		//     all_test.go:1045: /github.com/vnmakarov/mir/c-tests/new/endif.c: /usr/bin/gcc: system C compiler: exit status 1
		//         endif.c:1:2: error: #endif without #if
		//          #endif
		//           ^~~~~
		"endif.c": {}, // No intent to support.

		// 1: /github.com/vnmakarov/mir/c-tests/new/fermian-2.c
		//     all_test.go:1051: /github.com/vnmakarov/mir/c-tests/new/fermian-2.c: /usr/bin/gcc: system C compiler: exit status 1
		//         fermian-2.c:1:3: error: expected ‘=’, ‘,’, ‘;’, ‘asm’ or ‘__attribute__’ before ‘{’ token
		//          a {
		//            ^
		"fermian-2.c": {}, // No intent to support.
		"fermian.c":   {}, // No main.

		"issue142.c": {}, //TODO
		"issue18.c":  {}, //TODO
		"issue23.c":  {}, //TODO
		"setjmp.c":   {}, //TODO
		"setjmp2.c":  {}, //TODO
	}
	binary := map[string]bool{}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestCompCert(t *testing.T) {
	const root = "/github.com/AbsInt/CompCert/test/c/"
	g := newGolden(t, fmt.Sprintf("testdata/compcert_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	needFiles(t, root, []string{
		"Results/knucleotide-input.txt",
		"endian.h",
	})
	blacklist := map[string]struct{}{}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "windows/386":
		blacklist["bisect.c"] = struct{}{}     //TODO
		blacklist["fftw.c"] = struct{}{}       //TODO
		blacklist["mandelbrot.c"] = struct{}{} //TODO
		blacklist["perlin.c"] = struct{}{}     //TODO
	case "windows/amd64":
		blacklist["mandelbrot.c"] = struct{}{} //TODO
	}
	binary := map[string]bool{
		"mandelbrot.c": true,
	}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestGCCExecute(t *testing.T) {
	const root = "/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute"
	g := newGolden(t, fmt.Sprintf("testdata/gcc_exec_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	// Prepare testdata
	needFiles(t, root, []string{
		"20040709-2.c",
	})
	blacklist := map[string]struct{}{
		// assembler
		"20001009-2.c": {},

		// Nested function
		"20000822-1.c": {},
		"20010209-1.c": {},
		"20010605-1.c": {},

		// Alignment > 8
		"20010904-1.c": {},
		"20010904-2.c": {},

		// Variable sized type
		"20040423-1.c": {},
		"20040411-1.c": {},

		// Relies on SIGFPE
		"20101011-1.c": {},

		// Relies on gcc instrumentation
		"eeprof-1.c": {},

		//TODO back-end: undefined: __builtin_return_address
		"20010122-1.c": {},

		//TODO crash in cc
		"20010605-2.c": {},

		//TODO flexible array members not supported
		"20010924-1.c": {},

		//TODO TODO in go.go
		"20000113-1.c": {},

		//TODO relies on link order, libc first
		"20021127-1.c": {},

		//TODO #pragma push/pop macro
		"pushpop_macro.c": {},

		//TODO not yet classified
		"20000801-3.c":                 {}, //TODO
		"20020107-1.c":                 {}, //TODO
		"20020206-2.c":                 {}, //TODO
		"20020227-1.c":                 {}, //TODO
		"20020314-1.c":                 {}, //TODO
		"20020320-1.c":                 {}, //TODO
		"20020411-1.c":                 {}, //TODO
		"20020412-1.c":                 {}, //TODO
		"20021113-1.c":                 {}, //TODO
		"20021120-1.c":                 {}, //TODO
		"20030109-1.c":                 {}, //TODO
		"20030128-1.c":                 {}, //TODO
		"20030222-1.c":                 {}, //TODO
		"20030501-1.c":                 {}, //TODO
		"20030910-1.c":                 {}, //TODO
		"20031003-1.c":                 {}, //TODO
		"20040223-1.c":                 {}, //TODO
		"20040302-1.c":                 {}, //TODO
		"20040308-1.c":                 {}, //TODO
		"20040520-1.c":                 {}, //TODO
		"20040629-1.c":                 {}, //TODO
		"20040705-1.c":                 {}, //TODO
		"20040705-2.c":                 {}, //TODO
		"20040707-1.c":                 {}, //TODO
		"20040709-1.c":                 {}, //TODO
		"20040709-2.c":                 {}, //TODO
		"20040709-3.c":                 {}, //TODO
		"20041011-1.c":                 {}, //TODO
		"20041124-1.c":                 {}, //TODO
		"20041201-1.c":                 {}, //TODO
		"20041214-1.c":                 {}, //TODO
		"20041218-2.c":                 {}, //TODO
		"20050121-1.c":                 {}, //TODO
		"20050316-1.c":                 {}, //TODO
		"20050316-2.c":                 {}, //TODO
		"20050316-3.c":                 {}, //TODO
		"20050604-1.c":                 {}, //TODO
		"20050607-1.c":                 {}, //TODO
		"20050613-1.c":                 {}, //TODO
		"20050929-1.c":                 {}, //TODO
		"20051012-1.c":                 {}, //TODO
		"20060420-1.c":                 {}, //TODO
		"20061220-1.c":                 {}, //TODO
		"20070614-1.c":                 {}, //TODO
		"20070824-1.c":                 {}, //TODO
		"20070919-1.c":                 {}, //TODO
		"20071210-1.c":                 {}, //TODO
		"20071211-1.c":                 {}, //TODO
		"20071220-1.c":                 {}, //TODO
		"20071220-2.c":                 {}, //TODO
		"20080502-1.c":                 {}, //TODO
		"20090219-1.c":                 {}, //TODO
		"20100430-1.c":                 {}, //TODO
		"20121108-1.c":                 {}, //TODO
		"20180921-1.c":                 {}, //TODO
		"920302-1.c":                   {}, //TODO
		"920415-1.c":                   {}, //TODO
		"920428-2.c":                   {}, //TODO
		"920501-1.c":                   {}, //TODO
		"920501-3.c":                   {}, //TODO
		"920501-4.c":                   {}, //TODO
		"920501-5.c":                   {}, //TODO
		"920501-7.c":                   {}, //TODO
		"920612-2.c":                   {}, //TODO
		"920625-1.c":                   {}, //TODO
		"920721-4.c":                   {}, //TODO
		"920908-1.c":                   {}, //TODO
		"921017-1.c":                   {}, //TODO
		"921202-1.c":                   {}, //TODO
		"921208-2.c":                   {}, //TODO
		"921215-1.c":                   {}, //TODO
		"930406-1.c":                   {}, //TODO
		"931002-1.c":                   {}, //TODO
		"931004-10.c":                  {}, //TODO
		"931004-12.c":                  {}, //TODO
		"931004-14.c":                  {}, //TODO
		"931004-2.c":                   {}, //TODO
		"931004-4.c":                   {}, //TODO
		"931004-6.c":                   {}, //TODO
		"931004-8.c":                   {}, //TODO
		"941014-1.c":                   {}, //TODO
		"941202-1.c":                   {}, //TODO
		"960312-1.c":                   {}, //TODO
		"960416-1.c":                   {}, //TODO
		"960512-1.c":                   {}, //TODO
		"970217-1.c":                   {}, //TODO
		"980526-1.c":                   {}, //TODO
		"990130-1.c":                   {}, //TODO
		"990208-1.c":                   {}, //TODO
		"990413-2.c":                   {}, //TODO
		"990524-1.c":                   {}, //TODO
		"991014-1.c":                   {}, //TODO
		"991112-1.c":                   {}, //TODO
		"991227-1.c":                   {}, //TODO
		"alias-2.c":                    {}, //TODO
		"alias-3.c":                    {}, //TODO
		"alias-4.c":                    {}, //TODO
		"align-3.c":                    {}, //TODO
		"align-nest.c":                 {}, //TODO
		"alloca-1.c":                   {}, //TODO
		"anon-1.c":                     {}, //TODO
		"bitfld-3.c":                   {}, //TODO
		"built-in-setjmp.c":            {}, //TODO
		"builtin-bitops-1.c":           {}, //TODO
		"builtin-constant.c":           {}, //TODO
		"builtin-prefetch-3.c":         {}, //TODO
		"builtin-types-compatible-p.c": {}, //TODO
		"call-trap-1.c":                {}, //TODO
		"comp-goto-1.c":                {}, //TODO
		"comp-goto-2.c":                {}, //TODO
		"complex-1.c":                  {}, //TODO
		"complex-2.c":                  {}, //TODO
		"complex-4.c":                  {}, //TODO
		"complex-5.c":                  {}, //TODO
		"complex-6.c":                  {}, //TODO
		"complex-7.c":                  {}, //TODO
		"ffs-1.c":                      {}, //TODO
		"ffs-2.c":                      {}, //TODO
		"fprintf-2.c":                  {}, //TODO
		"frame-address.c":              {}, //TODO
		"medce-1.c":                    {}, //TODO
		"nest-align-1.c":               {}, //TODO
		"nest-stdar-1.c":               {}, //TODO
		"nestfunc-1.c":                 {}, //TODO
		"nestfunc-2.c":                 {}, //TODO
		"nestfunc-3.c":                 {}, //TODO
		"nestfunc-5.c":                 {}, //TODO
		"nestfunc-6.c":                 {}, //TODO
		"nestfunc-7.c":                 {}, //TODO
		"pr17377.c":                    {}, //TODO
		"pr22061-1.c":                  {}, //TODO
		"pr22061-3.c":                  {}, //TODO
		"pr22061-4.c":                  {}, //TODO
		"pr23135.c":                    {}, //TODO
		"pr23324.c":                    {}, //TODO
		"pr23467.c":                    {}, //TODO
		"pr24135.c":                    {}, //TODO
		"pr28289.c":                    {}, //TODO
		"pr28865.c":                    {}, //TODO
		"pr33382.c":                    {}, //TODO
		"pr34154.c":                    {}, //TODO
		"pr35456.c":                    {}, //TODO
		"pr36321.c":                    {}, //TODO
		"pr37780.c":                    {}, //TODO
		"pr38151.c":                    {}, //TODO
		"pr38533.c":                    {}, //TODO
		"pr38969.c":                    {}, //TODO
		"pr39228.c":                    {}, //TODO
		"pr40022.c":                    {}, //TODO
		"pr40657.c":                    {}, //TODO
		"pr41239.c":                    {}, //TODO
		"pr41935.c":                    {}, //TODO
		"pr42248.c":                    {}, //TODO
		"pr43385.c":                    {}, //TODO
		"pr43560.c":                    {}, //TODO
		"pr44575.c":                    {}, //TODO
		"pr45695.c":                    {}, //TODO
		"pr46309.c":                    {}, //TODO
		"pr47237.c":                    {}, //TODO
		"pr49279.c":                    {}, //TODO
		"pr49390.c":                    {}, //TODO
		"pr49644.c":                    {}, //TODO
		"pr51447.c":                    {}, //TODO
		"pr51877.c":                    {}, //TODO
		"pr51933.c":                    {}, //TODO
		"pr52286.c":                    {}, //TODO
		"pr53160.c":                    {}, //TODO
		"pr53645-2.c":                  {}, //TODO
		"pr53645.c":                    {}, //TODO
		"pr56205.c":                    {}, //TODO
		"pr56837.c":                    {}, //TODO
		"pr56866.c":                    {}, //TODO
		"pr56982.c":                    {}, //TODO
		"pr57344-1.c":                  {}, //TODO
		"pr57344-2.c":                  {}, //TODO
		"pr57344-3.c":                  {}, //TODO
		"pr57344-4.c":                  {}, //TODO
		"pr60003.c":                    {}, //TODO
		"pr60960.c":                    {}, //TODO
		"pr61725.c":                    {}, //TODO
		"pr63641.c":                    {}, //TODO
		"pr64006.c":                    {}, //TODO
		"pr64242.c":                    {}, //TODO
		"pr65053-2.c":                  {}, //TODO
		"pr65427.c":                    {}, //TODO
		"pr65648.c":                    {}, //TODO
		"pr65956.c":                    {}, //TODO
		"pr66556.c":                    {}, //TODO
		"pr67037.c":                    {}, //TODO
		"pr68249.c":                    {}, //TODO
		"pr68328.c":                    {}, //TODO
		"pr68381.c":                    {}, //TODO
		"pr69320-2.c":                  {}, //TODO
		"pr70460.c":                    {}, //TODO
		"pr70903.c":                    {}, //TODO
		"pr71494.c":                    {}, //TODO
		"pr71554.c":                    {}, //TODO
		"pr71626-1.c":                  {}, //TODO
		"pr71626-2.c":                  {}, //TODO
		"pr71631.c":                    {}, //TODO
		"pr77767.c":                    {}, //TODO
		"pr78438.c":                    {}, //TODO
		"pr78726.c":                    {}, //TODO
		"pr79354.c":                    {}, //TODO
		"pr79737-2.c":                  {}, //TODO
		"pr80421.c":                    {}, //TODO
		"pr80692.c":                    {}, //TODO
		"pr81588.c":                    {}, //TODO
		"pr82210.c":                    {}, //TODO
		"pr82954.c":                    {}, //TODO
		"pr84478.c":                    {}, //TODO
		"pr84521.c":                    {}, //TODO
		"pr84524.c":                    {}, //TODO
		"pr85156.c":                    {}, //TODO
		"pr85169.c":                    {}, //TODO
		"pr85331.c":                    {}, //TODO
		"pr85529-1.c":                  {}, //TODO
		"pr86528.c":                    {}, //TODO
		"pr89195.c":                    {}, //TODO
		"pr89434.c":                    {}, //TODO
		"pr90311.c":                    {}, //TODO
		"pr91450-1.c":                  {}, //TODO
		"pr91450-2.c":                  {}, //TODO
		"pr91635.c":                    {}, //TODO
		"pr92618.c":                    {}, //TODO
		"pr92904.c":                    {}, //TODO
		"pr93213.c":                    {}, //TODO
		"pr93249.c":                    {}, //TODO
		"pr93434.c":                    {}, //TODO
		"pr93494.c":                    {}, //TODO
		"pr93744-1.c":                  {}, //TODO
		"pr93945.c":                    {}, //TODO
		"pr94130.c":                    {}, //TODO
		"pr94412.c":                    {}, //TODO
		"pr94524-1.c":                  {}, //TODO
		"pr94524-2.c":                  {}, //TODO
		"pr94591.c":                    {}, //TODO
		"pr97325.c":                    {}, //TODO
		"pr98366.c":                    {}, //TODO
		"pr98474.c":                    {}, //TODO
		"pr98681.c":                    {}, //TODO
		"printf-2.c":                   {}, //TODO
		"return-addr.c":                {}, //TODO
		"scal-to-vec1.c":               {}, //TODO
		"scal-to-vec2.c":               {}, //TODO
		"scal-to-vec3.c":               {}, //TODO
		"simd-1.c":                     {}, //TODO
		"simd-2.c":                     {}, //TODO
		"simd-4.c":                     {}, //TODO
		"simd-5.c":                     {}, //TODO
		"simd-6.c":                     {}, //TODO
		"stdarg-3.c":                   {}, //TODO
		"stkalign.c":                   {}, //TODO
		"strct-stdarg-1.c":             {}, //TODO
		"strct-varg-1.c":               {}, //TODO
		"string-opt-18.c":              {}, //TODO
		"string-opt-5.c":               {}, //TODO
		"user-printf.c":                {}, //TODO
		"va-arg-2.c":                   {}, //TODO
		"va-arg-22.c":                  {}, //TODO
		"va-arg-pack-1.c":              {}, //TODO
		"zero-struct-2.c":              {}, //TODO
	}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "linux/386":
		// asm
		blacklist["960830-1.c"] = struct{}{}
	case "linux/arm64":
		blacklist["vfprintf-chk-1.c"] = struct{}{} //TODO
	case "linux/s390x":
		blacklist["pr58574.c"] = struct{}{}        //TODO
		blacklist["vfprintf-chk-1.c"] = struct{}{} //TODO
	case "windows/amd64":
		blacklist["941014-2.c"] = struct{}{} //TODO
		blacklist["pr36339.c"] = struct{}{}  //TODO
		blacklist["pr78622.c"] = struct{}{}  //TODO
	case "windows/386":
		blacklist["941014-2.c"] = struct{}{} //TODO
		blacklist["960830-1.c"] = struct{}{} //TODO
		blacklist["pr78622.c"] = struct{}{}  //TODO
	}
	binary := map[string]bool{}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if strings.Contains(pth, "/ieee/") || !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			base := filepath.Base(pth)
			failed[pth] = struct{}{}
			go run(pth, binary[base], true, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestGCCExecuteIEEE(t *testing.T) {
	const root = "/github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/ieee"
	g := newGolden(t, fmt.Sprintf("testdata/gcc_ieee_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	blacklist := map[string]struct{}{
		"compare-fp-1.c": {}, //TODO
		"compare-fp-4.c": {}, //TODO
		"copysign1.c":    {}, //TODO
		"copysign2.c":    {}, //TODO
		"fp-cmp-4.c":     {}, //TODO
		"fp-cmp-4f.c":    {}, //TODO
		"fp-cmp-4l.c":    {}, //TODO
		"fp-cmp-5.c":     {}, //TODO
		"fp-cmp-8.c":     {}, //TODO
		"fp-cmp-8f.c":    {}, //TODO
		"fp-cmp-8l.c":    {}, //TODO
		"inf-1.c":        {}, //TODO
		"inf-3.c":        {}, //TODO
		"mzero4.c":       {}, //TODO
		"pr36332.c":      {}, //TODO
		"pr38016.c":      {}, //TODO
		"pr50310.c":      {}, //TODO
		"pr72824-2.c":    {}, //TODO
	}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "linux/386":
		blacklist["rbug.c"] = struct{}{} //TODO Go1.18 https://github.com/golang/go/issues/48807
	case "linux/arm":
		blacklist["compare-fp-3.c"] = struct{}{} //TODO
		blacklist["rbug.c"] = struct{}{}         //TODO
	case "linux/s390x":
		blacklist["compare-fp-3.c"] = struct{}{} //TODO
	case "windows/amd64":
		blacklist["fp-cmp-1.c"] = struct{}{} //TODO
		blacklist["fp-cmp-2.c"] = struct{}{} //TODO
		blacklist["fp-cmp-3.c"] = struct{}{} //TODO
	case "windows/386":
		blacklist["fp-cmp-1.c"] = struct{}{} //TODO
		blacklist["fp-cmp-2.c"] = struct{}{} //TODO
		blacklist["fp-cmp-3.c"] = struct{}{} //TODO
		blacklist["rbug.c"] = struct{}{}     //TODO
	case "netbsd/amd64":
		blacklist["compare-fp-3.c"] = struct{}{} //TODO
		blacklist["fp-cmp-7.c"] = struct{}{}     //TODO
	}
	binary := map[string]bool{}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, binary[filepath.Base(pth)], true, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestCxgo(t *testing.T) {
	const root = "/github.com/cxgo"
	g := newGolden(t, fmt.Sprintf("testdata/cxgo_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	needFiles(t, root, []string{})
	blacklist := map[string]struct{}{
		"inet.c": {}, //TODO
		"math.c": {}, //TODO
	}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, false, false, true, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestSQLite(t *testing.T) {
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	testSQLite(t, root)
}

func testSQLite(t *testing.T, dir string) {
	const main = "main.go"
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case *oKeep:
		t.Log(temp)
	default:
		defer os.RemoveAll(temp)
	}

	if _, _, err := CopyDir(temp, dir, nil); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	ccgoArgs := []string{
		"ccgo",

		"-DHAVE_USLEEP",
		"-DLONGDOUBLE_TYPE=double",
		"-DSQLITE_DEBUG",
		"-DSQLITE_DEFAULT_MEMSTATUS=0",
		"-DSQLITE_ENABLE_DBPAGE_VTAB",
		"-DSQLITE_LIKE_DOESNT_MATCH_BLOBS",
		"-DSQLITE_MEMDEBUG",
		"-DSQLITE_THREADSAFE=0",
		"-all-errors",
		"-o", main,
		"-verify-structs",
		"shell.c",
		"sqlite3.c",
	}
	if *oDebug {
		ccgoArgs = append(ccgoArgs, "-DSQLITE_DEBUG_OS_TRACE", "-DSQLITE_FORCE_OS_TRACE")
	}
	if !func() (r bool) {
		defer func() {
			if err := recover(); err != nil {
				if *oStackTrace {
					fmt.Printf("%s\n", stack())
				}
				if *oTrace {
					fmt.Println(err)
				}
				t.Errorf("%v", err)
				r = false
			}
			if *oTraceF {
				b, _ := ioutil.ReadFile(main)
				fmt.Printf("\n----\n%s\n----\n", b)
			}
		}()

		if err := NewTask(ccgoArgs, nil, nil).Main(); err != nil {
			if *oTrace {
				fmt.Println(err)
			}
			err = cpp(*oCpp, ccgoArgs, err)
			t.Errorf("%v", err)
			return false
		}

		return true
	}() {
		return
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/sqlite"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	shell := "./shell"
	if runtime.GOOS == "windows" {
		shell = "shell.exe"
	}
	args := []string{"build"}
	if s := *oXTags; s != "" {
		args = append(args, "-tags", s)
	}
	args = append(args, "-o", shell, main)
	if out, err := exec.Command("go", args...).CombinedOutput(); err != nil {
		s := strings.TrimSpace(string(out))
		if s != "" {
			s += "\n"
		}
		t.Errorf("%s%v", s, err)
		return
	}

	var out []byte
	switch {
	case *oDebug:
		out, err = exec.Command(shell, "tmp", ".log stdout", "create table t(i); insert into t values(42); select 11*i from t;").CombinedOutput()
	default:
		out, err = exec.Command(shell, "tmp", "create table t(i); insert into t values(42); select 11*i from t;").CombinedOutput()
	}
	if err != nil {
		if *oTrace {
			fmt.Printf("%s\n%s\n", out, err)
		}
		t.Errorf("%s\n%v", out, err)
		return
	}

	if g, e := strings.TrimSpace(string(out)), "462"; g != e {
		t.Errorf("got: %s\nexp: %s", g, e)
	}
	if *oTraceO {
		fmt.Printf("%s\n", out)
	}

	if out, err = exec.Command(shell, "tmp", "select 13*i from t;").CombinedOutput(); err != nil {
		if *oTrace {
			fmt.Printf("%s\n%s\n", out, err)
		}
		t.Errorf("%v", err)
		return
	}

	if g, e := strings.TrimSpace(string(out)), "546"; g != e {
		t.Errorf("got: %s\nexp: %s", g, e)
	}
	if *oTraceO {
		fmt.Printf("%s\n", out)
	}
}

func TestBug(t *testing.T) {
	const root = "/ccgo/bug"
	g := newGolden(t, fmt.Sprintf("testdata/bug_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	mustEmptyDir(t, tempDir, keep)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	}()

	blacklist := map[string]struct{}{}
	switch fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH) {
	case "linux/s390x":
		blacklist["bitfield.c"] = struct{}{} //TODO
	}
	var rq, res, ok int
	limit := runtime.GOMAXPROCS(0)
	limiter := make(chan struct{}, limit)
	success := make([]string, 0, 0)
	results := make(chan *runResult, limit)
	failed := map[string]struct{}{}
	err = walk(root, func(pth string, fi os.FileInfo) error {
		if !strings.HasSuffix(pth, ".c") {
			return nil
		}

		switch {
		case re != nil:
			if !re.MatchString(pth) {
				return nil
			}
		default:
			if _, ok := blacklist[filepath.Base(pth)]; ok {
				return nil
			}
		}

	more:
		select {
		case r := <-results:
			res++
			<-limiter
			switch r.err.(type) {
			case nil:
				ok++
				success = append(success, filepath.Base(r.name))
				delete(failed, r.name)
			case skipErr:
				delete(failed, r.name)
				t.Logf("%v: %v\n%s", r.name, r.err, r.out)
			default:
				t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
			}
			goto more
		case limiter <- struct{}{}:
			rq++
			if *oTrace {
				fmt.Fprintf(os.Stderr, "%v: %s\n", rq, pth)
			}
			failed[pth] = struct{}{}
			go run(pth, false, false, false, results)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for res != rq {
		r := <-results
		res++
		<-limiter
		switch r.err.(type) {
		case nil:
			ok++
			success = append(success, filepath.Base(r.name))
			delete(failed, r.name)
		case skipErr:
			delete(failed, r.name)
			t.Logf("%v: %v\n%s", r.name, r.err, r.out)
		default:
			t.Errorf("%v: %v\n%s", r.name, r.err, r.out)
		}
	}
	t.Logf("files %v, ok %v, failed %v", rq, ok, len(failed))
	sort.Strings(success)
	for _, fpath := range success {
		g.w.Write([]byte(fpath))
		g.w.Write([]byte{'\n'})
	}
	if len(failed) == 0 {
		return
	}

	var a []string
	for k := range failed {
		a = append(a, k)
	}
	sort.Strings(a)
	for _, v := range a {
		t.Logf("FAIL %s", v)
	}
}

func TestCSmith(t *testing.T) {
	gcc := os.Getenv("CC")
	if gcc == "" {
		gcc = "gcc"
	}
	gcc, err := exec.LookPath(gcc)
	if err != nil {
		t.Skip(err)
		return
	}

	if testing.Short() {
		t.Skip("skipped: -short")
	}

	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Skip(err)
		return
	}
	binaryName := filepath.FromSlash("./a.out")
	mainName := filepath.FromSlash("main.go")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/csmith"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	fixedBugs := []string{
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 1906742816",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 612971101",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 3629008936",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4130344133",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3130410542",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1833258637",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3126091077",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2205128324",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3043990076",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2517344771",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 56498550",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3645367888",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 169375684",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3578720023",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1885311141",
		"--no-bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3720922579",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 241244373",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 517639208",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2205128324",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2876930815",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3365074920",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3329111231",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2648215054",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3919255949",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 890611563",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4101947480",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4058772172",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2273393378",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3100949894",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 15739796933983044010", //TODO fails on linux/s390x

		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 963985971",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3363122597",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4146870674",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 1236173074", //TODO fails on darwin/amd64
	}
	ch := time.After(*oCSmith)
	t0 := time.Now()
	var files, ok int
	var size int64
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
out:
	for i := 0; ; i++ {
		extra := ""
		var args string
		switch {
		case i < len(fixedBugs):
			if re != nil && !re.MatchString(fixedBugs[i]) {
				continue
			}

			args += fixedBugs[i]
			a := strings.Split(fixedBugs[i], " ")
			extra = strings.Join(a[len(a)-2:], " ")
			t.Log(args)
		default:
			select {
			case <-ch:
				break out
			default:
			}

			args += csmithDefaultArgs
		}
		csOut, err := exec.Command(csmith, strings.Split(args, " ")...).Output()
		if err != nil {
			t.Fatalf("%v\n%s", err, csOut)
		}

		if fn := *oBlackBox; fn != "" {
			if err := ioutil.WriteFile(fn, csOut, 0660); err != nil {
				t.Fatal(err)
			}
		}

		if err := ioutil.WriteFile("main.c", csOut, 0660); err != nil {
			t.Fatal(err)
		}

		csp := fmt.Sprintf("-I%s", filepath.FromSlash("/usr/include/csmith"))
		if s := os.Getenv("CSMITH_PATH"); s != "" {
			csp = fmt.Sprintf("-I%s", s)
		}

		ccOut, err := exec.Command(gcc, "-o", binaryName, "main.c", csp).CombinedOutput()
		if err != nil {
			t.Fatalf("%s\n%s\ncc: %v", extra, ccOut, err)
		}

		binOutA, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			return exec.CommandContext(ctx, binaryName).CombinedOutput()
		}()
		if err != nil {
			continue
		}

		size += int64(len(csOut))

		if err := os.Remove(binaryName); err != nil {
			t.Fatal(err)
		}

		files++
		var stdout, stderr bytes.Buffer
		j := NewTask([]string{
			"ccgo",

			"-o", mainName,
			"-verify-structs",
			"main.c",
			csp,
		}, &stdout, &stderr)
		j.cfg.MaxSourceLine = 1 << 20

		func() {

			defer func() {
				if err := recover(); err != nil {
					t.Errorf("%s\n%s\nccgo: %s\n%s\n%s", extra, csOut, stdout.Bytes(), stderr.Bytes(), debug.Stack())
					t.Fatal(err)
				}
			}()

			if err := j.Main(); err != nil || stdout.Len() != 0 {
				t.Errorf("%s\n%s\nccgo: %s\n%s", extra, csOut, stdout.Bytes(), stderr.Bytes())
				t.Fatal(err)
			}
		}()

		binOutB, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()

			return exec.CommandContext(ctx, "go", "run", "-tags=libc.memgrind", mainName).CombinedOutput()
		}()
		if err != nil {
			t.Errorf("%s\n%s\n%s\nccgo: %v", extra, csOut, binOutB, err)
			break
		}

		if g, e := binOutB, binOutA; !bytes.Equal(g, e) {
			t.Errorf("%s\n%s\nccgo: %v\ngot: %s\nexp: %s", extra, csOut, err, g, e)
			break
		}

		ok++
		if *oTrace {
			fmt.Fprintln(os.Stderr, time.Since(t0), files, ok)
		}

		if err := os.Remove(mainName); err != nil {
			t.Fatal(err)
		}
	}
	d := time.Since(t0)
	t.Logf("files %v, bytes %v, ok %v in %v", h(files), h(size), h(ok), d)
}

func dumpInitializer(s []*cc.Initializer) string {
	if len(s) == 0 {
		return "<empty>"
	}
	var a []string
	for _, v := range s {
		var s string
		if f := v.Field; f != nil {
			s = fmt.Sprintf("fld %q bitfield %v bitoff %2d", f.Name(), f.IsBitField(), f.BitFieldOffset())
		}
		a = append(a, fmt.Sprintf("%v: off %#04x val %v %s", v.Position(), v.Offset, v.AssignmentExpression.Operand.Value(), s))
	}
	return strings.Join(a, "\n")
}
