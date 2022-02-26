// Copyright 2018 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//TODO remove all produced files on error/panic.

// Command ccgo is a C compiler targeting Go.
//
// Usage
//
//     $ ccgo [options] [files]
//
//       -c                          Compile and assemble, but do not link
//       -dM                         With -E: generate a list of ‘#define’ directives
//                                   for all the macros defined during the execution
//                                   of the preprocessor, including predefined macros.
//       -D<macro>[=<val>]           Define a <macro> with <val> as its value.  If
//                                   just <macro> is given, <val> is taken to be 1
//       -e ADDRESS, --entry ADDRESS Set start address (ignored)
//       -E                          Preprocess only; do not compile, assemble or link
//       -ffreestanding              Do not assume that standard C libraries and
//                                   "main" exist
//       -fomit-frame-pointer        When possible do not generate stack frames (ignored)
//       -fPIC                       Generate position-independent code if possible
//       --help                      Display this information
//       -g --gen-debug              generate debugging information (ignored)
//       -h FILENAME, -soname FILENAME
//                                   Set internal name of shared library
//       -I <dir>                    Add <dir> to the end of the main include path
//       -l LIBNAME, --library LIBNAME
//                                   Search for library LIBNAME
//       -L DIRECTORY, --library-path DIRECTORY
//                                   Add DIRECTORY to library search path
//       -m64                        Generate 64bit x86-64 code
//       -nostdlib                   Do not look for object files in standard path (ignored)
//       -o <file>                   Place the output into <file>. Use .go extension
//                                   to produce a Go source file instead of a binary.
//       -O                          Optimize output file (ignored)
//       -rpath PATH                 Set runtime shared library search path
//       -shared                     Create a shared library
//       -v                          Display the programs invoked by the compiler
//       --version                   Display compiler version information
//       --warn-go-build             Report 'go build' errors as warning
//       --warn-unresolved-libs      Report unresolved libraries as warnings
//       --warn-unresolved-symbols   Report unresolved symbols as warnings
//       -W  --no-warn               suppress warnings (ignored)
//       -Wall                       Enable most warning messages (ignored)
//       -Wl,<options>               Pass comma-separated <options> on to the linker
//       -x <language>               Specify the language of the following input files.
//                                   Permissible languages include: c.
//
//       --ccgo-define-values        Emit #defines that evaluate to a constant
//       --ccgo-full-paths           Keep full source code positions instead of
//                                   basenames
//       --ccgo-go                   Do not remove the Go source file used to link the
//                                   executable file and print its path
//       --ccgo-import <paths>       Add import comma separated paths
//       --ccgo-pkg-name             Set output Go file package name
//       --ccgo-struct-checks        Generate code to verify struct/union sizes
//                                   and field offsets.
//       --ccgo-use-import <exprs>   Add import usage comma separated expressions
//       --ccgo-watch                Enable run time watch instrumentation
//
// Installation
//
// To install or update ccgo and its accompanying tools
//
//      $ go get [-u] modernc.org/ccgo/v2/...
//
// Online documentation: [godoc.org/modernc.org/ccgo/v2/ccgo](http://godoc.org/modernc.org/ccgo/v2/ccgo)
//
// Changelog
//
// TODO
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"modernc.org/cc/v2"
	"modernc.org/ccgo/v2"
	"modernc.org/ccgo/v2/internal/object"
	"modernc.org/crt"
)

const (
	version   = "0.0.1"
	crtPrefix = "crt."

	help = `
  -c                          Compile and assemble, but do not link
  -dM                         With -E: generate a list of ‘#define’ directives
                              for all the macros defined during the execution
                              of the preprocessor, including predefined macros.
  -D<macro>[=<val>]           Define a <macro> with <val> as its value.  If
                              just <macro> is given, <val> is taken to be 1
  -e ADDRESS, --entry ADDRESS Set start address (ignored)
  -E                          Preprocess only; do not compile, assemble or link
  -ffreestanding              Do not assume that standard C libraries and
                              "main" exist
  -fomit-frame-pointer        When possible do not generate stack frames (ignored)
  -fPIC                       Generate position-independent code if possible
  --help                      Display this information
  -g --gen-debug              generate debugging information (ignored)
  -h FILENAME, -soname FILENAME
                              Set internal name of shared library
  -I <dir>                    Add <dir> to the end of the main include path
  -l LIBNAME, --library LIBNAME
                              Search for library LIBNAME
  -L DIRECTORY, --library-path DIRECTORY
                              Add DIRECTORY to library search path
  -m64                        Generate 64bit x86-64 code
  -nostdlib                   Do not look for object files in standard path
  -o <file>                   Place the output into <file>. Use .go extension
                              to produce a Go source file instead of a binary.
  -O                          Optimize output file (ignored)
  -rpath PATH                 Set runtime shared library search path
  -shared                     Create a shared library
  -v                          Display the programs invoked by the compiler
  --version                   Display compiler version information
  --warn-go-build             Report 'go build' errors as warning
  --warn-unresolved-libs      Report unresolved libraries as warnings
  --warn-unresolved-symbols   Report unresolved symbols as warnings
  -W  --no-warn               suppress warnings (ignored)
  -Wall                       Enable most warning messages (ignored)
  -Wl,<options>               Pass comma-separated <options> on to the linker
  -x <language>               Specify the language of the following input files.
                              Permissible languages include: c.

  --ccgo-define-values        Emit #defines that evaluate to a constant
  --ccgo-full-paths           Keep full source code positions instead of
                              basenames
  --ccgo-go                   Do not remove the Go source file used to link the
                              executable file and print its path
  --ccgo-import <paths>       Add import comma separated paths
  --ccgo-pkg-name             Set output Go file package name
  --ccgo-struct-checks        Generate code to verify struct/union sizes
                              and field offsets.
  --ccgo-use-import <exprs>   Add import usage comma separated expressions
  --ccgo-watch                Enable run time watch instrumentation
`

	pkgHeader = `// Code generated by '%[1]s', DO NOT EDIT.

/` + `/ +build %[4]s,%[5]s

package %[2]s

import (
	"math"
	"unsafe"%[6]s
)

const null = uintptr(0)

var _ = math.Pi
var _ = unsafe.Pointer(null)
%[7]s
`
	mainHeader = `func main() { %[3]sMain(Xmain) }

`
)

var (
	log     = func(string, ...interface{}) {}
	logging bool
)

func main() {
	r, err := main1(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, strings.TrimRight(expandError(err).Error(), "\n\t "))
	}
	os.Exit(r)
}

type config struct {
	D          []string // -D
	I          []string // -I
	L          []string // -L
	Wl         []string // -Wl
	imports    []string // --ccgo-import
	l          []string // -l
	o          string   // -o
	pkgName    string   // --ccgo-pkg-name
	useImports []string // --ccgo-use-import

	arg0         string
	args         []string
	goarch       string
	goos         string
	incPaths     []string
	linkOrder    []string
	objMap       map[string]string
	objects      []string
	osArgs       []string
	remove       []string
	sysPaths     []string
	linkerConfig *linkerConfig

	E                bool // -E
	O                bool // -O* (ignored)
	W                bool // -W (ignored)
	Wall             bool // -Wall (ignored)
	c                bool // -c
	dM               bool // -dM
	defineValues     bool // --ccgo-define-values
	fPIC             bool // -fPIC (ignored)
	ffreeStanding    bool // -ffreestanding
	fullPaths        bool // --ccgo-full-paths
	g                bool // -g --gen-debug (ignored)
	help             bool // --help
	keepGo           bool // --ccgo-go
	m64              bool // -m64
	noStdLib         bool // -nostdlib (ignored)
	omitFramePointes bool // -fomit-frame-pointer
	shared           bool // -shared
	structChecks     bool // --ccgo-struct-checks
	v                bool // -v
	version          bool // --version
	watch            bool // --ccgo-watch
}

func newConfig(args []string) (c *config, err error) {
	var errs []string

	defer func() {
		if len(errs) != 0 {
			if err != nil {
				errs = append([]string{err.Error()}, errs...)
			}
			err = fmt.Errorf("%s", strings.Join(errs, "\n"))
		}
		if err != nil {
			c = nil
		}
	}()

	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments to parse")
	}

	c = &config{
		arg0:     args[0],
		goarch:   env("GOARCH", runtime.GOARCH),
		goos:     env("GOOS", runtime.GOOS),
		incPaths: []string{"@"},
		objMap:   map[string]string{},
		osArgs:   args,
	}
	log("goos=%v goarch=%v", c.goos, c.goarch)
	args = args[1:]
	for len(args) != 0 {
		switch arg := args[0]; {
		case strings.HasPrefix(arg, "-D"):
			a := strings.SplitN(arg, "=", 2)
			if len(a) == 1 {
				a = append(a, "1")
			}
			c.D = append(c.D, fmt.Sprintf("%s %s", a[0][2:], a[1]))
		case arg == "-v":
			c.v = true
		case arg == "-c":
			c.c = true
		case arg == "-o":
			switch {
			case len(args) < 2:
				errs = append(errs, "-o option requires an argument")
			default:
				c.o = args[1]
				args = args[1:]
			}
		case arg == "--ccgo-go": // keep the .go file when linking a main program
			c.keepGo = true
		case arg == "-nostdlib":
			c.noStdLib = true
		case arg == "-ffreestanding":
			c.ffreeStanding = true
		case arg == "-fomit-frame-pointer":
			c.omitFramePointes = true
		case arg == "--ccgo-full-paths":
			c.fullPaths = true
		case arg == "--ccgo-watch":
			c.watch = true
		case arg == "--ccgo-import":
			switch {
			case len(args) < 2:
				errs = append(errs, "--ccgo-import option requires an argument")
			default:
				c.imports = append(c.imports, strings.Split(args[1], ",")...)
				args = args[1:]
			}
		case arg == "--ccgo-use-import":
			switch {
			case len(args) < 2:
				errs = append(errs, "--ccgo-use-import option requires an argument")
			default:
				c.useImports = append(c.useImports, strings.Split(args[1], ",")...)
				args = args[1:]
			}
		case arg == "--ccgo-struct-checks":
			c.structChecks = true
		case arg == "--ccgo-define-values":
			c.defineValues = true
		case arg == "--ccgo-pkg-name":
			switch {
			case len(args) < 2:
				errs = append(errs, "--ccgo-pkg-name option requires an argument")
			default:
				c.pkgName = args[1]
				args = args[1:]
			}
		case arg == "-dM":
			c.dM = true
		case arg == "-m64":
			c.E = true
		case arg == "--help":
			c.help = true
		case arg == "-Wall":
			c.Wall = true
		case arg == "-E":
			c.E = true
		case arg == "--version":
			c.version = true
		case arg == "-fPIC":
			c.fPIC = true
		case arg == "-W", arg == "--no-warn":
			c.W = true
		case arg == "-g", arg == "--gen-debug":
			c.g = true
		case arg == "-I":
			switch {
			case len(args) < 2:
				errs = append(errs, "-I option requires an argument")
			default:
				c.I = append(c.I, args[1])
				args = args[1:]
			}
		case strings.HasPrefix(arg, "-I"):
			c.I = append(c.I, arg[2:])
		case strings.HasPrefix(arg, "-x"):
			c.args = append(c.args, arg)
		case arg == "-L", arg == "--library-path":
			switch {
			case len(args) < 2:
				errs = append(errs, "-L option requires an argument")
			default:
				c.L = append(c.L, args[1])
				args = args[1:]
			}
		case strings.HasPrefix(arg, "-L"), strings.HasPrefix(arg, "--library-path"):
			c.L = append(c.L, arg[2:])
		case arg == "-shared":
			c.shared = true
		case strings.HasPrefix(arg, "-l"), strings.HasPrefix(arg, "--library"):
			s := arg[2:]
			c.l = append(c.l, s)
			c.linkOrder = append(c.linkOrder, arg)
		case strings.HasPrefix(arg, "-Wl,"):
			c.Wl = append(c.Wl, strings.Split(arg[4:], ",")...)
		case arg == "-":
			switch {
			case len(args) > 1:
				errs = append(errs, "no arguments allowed after -")
			default:
				c.args = append(c.args, "") // stdin
				c.linkOrder = append(c.linkOrder, "<stdin>")
			}

		case strings.HasPrefix(arg, "-O"):
			c.O = true
		case !strings.HasPrefix(arg, "-"):
			c.args = append(c.args, arg)
			c.linkOrder = append(c.linkOrder, arg)

		// Linker flags -----------------------------------------------
		case arg == "-rpath":
			switch {
			case len(args) < 2:
				errs = append(errs, "missing -rpath argument")
			default:
				c.Wl = append(c.Wl, arg, args[1])
				args = args[1:]
			}
		case arg == "-soname", arg == "-h":
			switch {
			case len(args) < 2:
				errs = append(errs, "missing -soname argument")
			default:
				c.Wl = append(c.Wl, arg, args[1])
				args = args[1:]
			}
		case
			arg == "--export-dynamic",
			arg == "--warn-go-build",
			arg == "--warn-unresolved-libs",
			arg == "--warn-unresolved-symbols":

			c.Wl = append(c.Wl, arg)
		default:
			errs = append(errs, fmt.Sprintf("%s: error: unrecognized command line option '%v'", c.arg0, arg))
		}
		args = args[1:]
	}
	if c.m64 {
		switch c.goarch {
		case "amd64":
			// ok
		default:
			errs = append(errs, fmt.Sprintf("-m64 used with invalid architecture %s", c.goarch))
		}
	}
	c.incPaths = append(c.incPaths, c.I...)
	c.sysPaths = append(c.sysPaths, c.I...)
	if c.linkerConfig, err = newLinkerConfig(c.arg0, c.Wl); err != nil {
		return nil, err
	}

	return c, nil
}

type linkerConfig struct {
	e       string   // -e (ignored)
	rpath   []string // -rpath dir		Add a directory to the runtime library search path
	soname  string
	sonames []string // -soname

	// --warn-unresolved-symbols
	//
	// If the linker is going to report an unresolved symbol (see the
	// option --unresolved-symbols) it will normally generate an error.
	// This option makes it generate a warning instead.
	warnUnresolvedSymbols bool

	exportDynamic      bool // --export-dynamic
	warnGoBuild        bool // --warn-go-build
	warnUnresolvedLibs bool // --warn-unresolved-libs
}

func newLinkerConfig(prog string, args []string) (c *linkerConfig, err error) {
	var errs []string

	defer func() {
		if len(errs) != 0 {
			if err != nil {
				errs = append([]string{err.Error()}, errs...)
			}
			err = fmt.Errorf("%s", strings.Join(errs, "\n"))
		}
		if err != nil {
			c = nil
		}
	}()

	c = &linkerConfig{}
	for ; len(args) != 0; args = args[1:] {
		switch arg := args[0]; {
		case arg == "--export-dynamic":
			c.exportDynamic = true
		case arg == "-e", arg == "--entry":
			switch {
			case len(args) < 2:
				errs = append(errs, "missing -e argument")
			default:
				c.e = args[1]
			}
			args = args[1:]
		case arg == "-soname", arg == "-h":
			switch {
			case len(args) < 2:
				errs = append(errs, "missing -soname argument")
			default:
				c.sonames = append(c.sonames, args[1])
				c.soname = args[1]
			}
			args = args[1:]
		case arg == "-rpath":
			switch {
			case len(args) < 2:
				return nil, fmt.Errorf("missing -rpath argument")
			default:
				c.rpath = append(c.rpath, args[1])
				args = args[1:]
			}
		case arg == "--warn-unresolved-symbols":
			c.warnUnresolvedSymbols = true
		case arg == "--warn-unresolved-libs":
			c.warnUnresolvedLibs = true
		case arg == "--warn-go-build":
			c.warnGoBuild = true
		default:
			errs = append(errs, fmt.Sprintf("%s: error: unrecognized linker option '%v'", prog, arg))
		}
	}
	if len(c.sonames) > 1 {
		return nil, fmt.Errorf("multiple -sonam options: %s", strings.Join(c.sonames, ""))
	}
	return c, nil
}

func main1(args []string) (r int, err error) {
	if fn := os.Getenv("CCGOLOG"); fn != "" {
		logging = true
		var f *os.File
		if f, err = os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644); err != nil {
			return 1, err
		}

		pid := fmt.Sprintf("[pid %v] ", os.Getpid())

		log = func(s string, args ...interface{}) {
			if s == "" {
				s = strings.Repeat("%v ", len(args))
			}
			_, fn, fl, _ := runtime.Caller(1)
			s = fmt.Sprintf(pid+"%s:%d: "+s, append([]interface{}{filepath.Base(fn), fl}, args...)...)
			switch {
			case len(s) != 0 && s[len(s)-1] == '\n':
				fmt.Fprint(f, s)
			default:
				fmt.Fprintln(f, s)
			}
		}

		defer func() {
			log("---- exit status %v, err %v", r, err)
			f.Close()
		}()

		log("==== %v", args)
	}

	returned := false

	defer func() {
		e := recover()
		if !returned && e != nil {
			err = errs(err, fmt.Errorf("PANIC: %v #%s", e, debugStack2()))
		}
	}()

	c, err := newConfig(args)
	if err != nil {
		return 2, err
	}

	if c.version {
		return 0, fmt.Errorf("%s %s", c.arg0, version)
	}

	if c.help {
		return 2, fmt.Errorf("%s", help[1:])
	}

	nin := 0
	for _, v := range c.args {
		if !strings.HasPrefix(v, "-") {
			nin++
		}
	}
	if c.c && c.o != "" && nin > 1 {
		return 2, fmt.Errorf("-o cannot be used with -c and multiple input files")
	}

	if len(c.args) == 0 {
		if c.v {
			return 0, fmt.Errorf("%s %s", c.arg0, version)
		}

		return 2, fmt.Errorf(`
%s: fatal error: no input files
compilation terminated`, c.arg0)
	}

	libc := filepath.Join(crt.RepositoryPath, "libc")
	var libcArch string
	switch c.goarch {
	case "386":
		libcArch = "i386"
	case "amd64":
		libcArch = "x86_64"
	default:
		return 1, fmt.Errorf("unknown/unsupported GOARCH: %s", c.goarch)
	}

	var sysPaths []string
	if !c.ffreeStanding {
		sysPaths = []string{
			filepath.Join(libc, "arch", libcArch),
			filepath.Join(libc, "arch", "generic"),
			//TODO filepath.Join(libc, "obj", "src", "internal"),
			//TODO filepath.Join(libc, "src", "internal"),
			filepath.Join(libc, "obj", "include"),
			filepath.Join(libc, "include"),
		}
	}

	c.incPaths = append(c.incPaths, sysPaths...)
	c.sysPaths = append(c.sysPaths, sysPaths...)
	var forceExt string
	for _, in := range c.args {
		if strings.HasPrefix(in, "-xc") {
			forceExt = ".c"
			continue
		}

		ext := filepath.Ext(in)
		if forceExt != "" {
			ext = forceExt
		}
		switch ext {
		case ".c":
			if in == "" {
				arg := "<stdin>"
				b, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
				if err != nil {
					return 1, err
				}

				log("stdin\n%s\n----", b)
				if err = c.compileSource("stdin.o", arg, cc.NewStringSource("stdin", string(b))); err != nil {
					return 1, err
				}

				continue
			}

			fallthrough
		case ".s":
			if _, err = c.compile(in); err != nil {
				return 1, err
			}
		case ".a", ".o", ".so", ".lo":
			c.objects = append(c.objects, in)
			c.objMap[in] = in
		default:
			return 1, fmt.Errorf("%s: file not recognized", in)
		}
	}
	if c.c || c.E {
		returned = true
		return 0, nil
	}

	if c.shared {
		if err = c.linkShared(); err != nil {
			returned = true
			return 1, err
		}

		returned = true
		return 0, nil
	}

	defer func() {
		for _, v := range c.remove {
			os.Remove(v)
		}
	}()

	if err := c.linkExecutable(); err != nil {
		returned = true
		return 1, err
	}

	returned = true
	return 0, nil
}

func (c *config) linkShared() (err error) {
	lc := c.linkerConfig

	var fn string
	if c.o != "" {
		fn = c.o
	}
	if fn == "" && lc.soname != "" {
		fn = lc.soname
	}
	if fn == "" {
		fn = "a.so"
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}

	defer func() { err = errs(err, f.Close()) }()

	b := bufio.NewWriter(f)

	defer func() { err = errs(err, b.Flush()) }()

	r, w := io.Pipe()
	var e2 error

	go func() {
		defer func() {
			if err := recover(); err != nil && e2 == nil {
				e2 = fmt.Errorf("%v", err)
			}
			if err := w.Close(); err != nil && e2 == nil {
				e2 = err
			}
		}()

		if lc.soname != "" {
			if _, e2 = fmt.Fprintf(w, "const Lsoname = %q\n\n", lc.soname); e2 != nil {
				return
			}
		}

		for _, v := range c.linkOrder {
			switch {
			case strings.HasPrefix(v, "-"):
				//TODO
			default:
				fn := c.objMap[v]
				if fn == "" {
					e2 = fmt.Errorf("internal error: missing object for %q", v)
					return
				}

				if _, e2 = fmt.Fprintf(w, "\n\nconst Lsofile = %q\n\n", fn); e2 != nil {
					return
				}

				f, err := os.Open(fn)
				if err != nil {
					e2 = err
					return
				}

				if e2 = object.Decode(w, c.goos, c.goarch, object.ObjVersion, object.ObjMagic, bufio.NewReader(f)); e2 != nil {
					return
				}
			}
		}
	}()

	err = ccgo.NewSharedObject(b, c.goos, c.goarch, r)
	if err == nil {
		err = e2
	}
	return err
}

func (c *config) linkExecutable() (err error) {
	fn := "a.out"
	if c.goos == "windows" {
		fn = "a.exe"
	}
	if c.o != "" {
		fn = c.o
	}

	if filepath.Ext(fn) == ".go" {
		return c.linkGo(fn)
	}

	dir, err := ioutil.TempDir("", "ccgo-linker-")
	if err != nil {
		return err
	}

	src := filepath.Join(dir, "main.go")

	defer func() {
		if c.keepGo {
			fmt.Fprintf(os.Stderr, "%s\n", src)
			return
		}

		err = errs(err, os.RemoveAll(dir))
	}()

	if err := c.linkGo(src); err != nil {
		return err
	}

	if err := c.buildExecutable(fn, src); err != nil {
		if c.linkerConfig.warnGoBuild {
			msg := err.Error()
			if !isArgumentMismatchError(msg) {
				return err
			}

			log("faking a --warn-go-build binary\n%s", msg)
			src = filepath.Join(dir, "error.go")
			f, err := os.Create(src)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintf(f, `package main

import (
	"fmt"
	"os"
)

func main() {
	if fn := os.Getenv("CCGOLOG"); fn != "" {
		f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(f, "[pid %%v] EXEC %%v\n", os.Getpid(), os.Args)
		//TODO want exit status
		f.Close()
	}
	fmt.Fprintln(os.Stderr, %q)
	os.Exit(1)
}
`, msg); err != nil {
				return err
			}

			return c.buildExecutable(fn, src)
		}

		return err
	}
	return nil
}

func (c *config) buildExecutable(bin, src string) error {
	a := []string{"go", "build", "-gcflags=-e", "-o", bin, src}
	cmd := exec.Command(a[0], a[1:]...)
	for _, v := range os.Environ() {
		if v != "CC=ccgo" {
			cmd.Env = append(cmd.Env, v)
		}
	}
	if c.v {
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(a, " "))
	}
	if co, err := cmd.CombinedOutput(); err != nil {
		if c.linkerConfig.warnGoBuild {
			fmt.Printf("warning: go build %s\n%s\n%v\n", bin, co, err)
		}
		return fmt.Errorf("%s\n%v", co, err)
	}
	return nil
}

func (c *config) linkGo(fn string) (err error) {
	lc := c.linkerConfig
	pkgName := toExt(filepath.Base(fn), "")
	if c.pkgName != "" {
		pkgName = c.pkgName
	}

	f, err := os.Create(fn)
	if err != nil {
		return err
	}

	out := bufio.NewWriter(f)
	defer func() { err = errs(err, out.Flush()) }()

	l, err := ccgo.NewLinker(out, c.goos, c.goarch)
	if err != nil {
		return err
	}

	crtPrefix := crtPrefix
	imports := "\n\n\t\"modernc.org/crt\""
	if c.ffreeStanding {
		imports = ""
		crtPrefix = ""
	}

	var a []string
	for _, v := range c.imports {
		a = append(a, fmt.Sprintf("%q", v))
	}
	if len(a) != 0 {
		imports = fmt.Sprintf("\n\t%s%s", strings.Join(a, "\n\t"), imports)
	}

	var useImportsA []string
	for _, v := range c.useImports {
		useImportsA = append(useImportsA, fmt.Sprintf("var _ = %s", v))
	}
	if len(useImportsA) != 0 {
		useImportsA = append(useImportsA, "")
	}
	useImports := strings.Join(useImportsA, "\n")

	header := fmt.Sprintf(pkgHeader, strings.Join(c.osArgs, " "), pkgName, crtPrefix, c.goos, c.goarch, imports, useImports)

	defer func() { err = errs(err, l.Close(header)) }()

	for _, v := range c.linkOrder {
		switch {
		case strings.HasPrefix(v, "-"):
			switch {
			case strings.HasPrefix(v, "-l"):
				fn := c.findLib(v[2:])
				if fn == "" {
					switch {
					case lc.warnUnresolvedLibs:
						fmt.Printf("warning: cannot find %s\n", v)
						continue
					default:
						return fmt.Errorf("cannot find %s", v)
					}
				}

				if err = c.linkFile(l, fn); err != nil {
					return err
				}
			default:
				panic(fmt.Errorf("TODO %q", v))
			}
		default:
			fn := c.objMap[v]
			if fn == "" {
				return fmt.Errorf("internal error: missing object for %q", v)
			}

			if err = c.linkFile(l, fn); err != nil {
				return err
			}
		}
	}
	if l.Main {
		header = fmt.Sprintf(pkgHeader+mainHeader, strings.Join(c.osArgs, " "), "main", crtPrefix, c.goos, c.goarch, imports, useImports)
	}
	return nil
}

func (c *config) findLib(nm string) string {
	list := append([]string{""}, c.L...)
	for _, v := range list {
		pat := filepath.Join(v, fmt.Sprintf("lib%s.so", nm))
		m, err := filepath.Glob(pat)
		if err != nil || len(m) == 0 {
			continue
		}

		return m[0]
	}
	for _, v := range list {
		pat := filepath.Join(v, fmt.Sprintf("lib%s.a", nm))
		m, err := filepath.Glob(pat)
		if err != nil || len(m) == 0 {
			continue
		}

		return m[0]
	}
	return ""
}

func (c *config) linkFile(l *ccgo.Linker, fn string) (err error) {
	var f *os.File
	if f, err = os.Open(fn); err != nil {
		return err
	}

	defer func() { err = errs(err, f.Close()) }()

	switch ext := filepath.Ext(fn); ext {
	case ".a":
		r, err := newArReader(f)
		if err != nil {
			return err
		}

		for r.Next() {
			if err := l.Link(r.fn, r); err != nil {
				return fmt.Errorf("%s: %v", fn, err)
			}
		}
		return r.err
	case ".o", ".so":
		if err := l.Link(fn, bufio.NewReader(f)); err != nil {
			return fmt.Errorf("%s: %v", fn, err)
		}
	default:
		return fmt.Errorf("unknown linker object type: %s", fn)
	}

	return nil
}

func (c *config) compile(in string) (out string, err error) {
	out = filepath.Base(toExt(in, ".o"))
	if c.c && c.o != "" {
		out = c.o
	}
	if logging {
		b, err := ioutil.ReadFile(in)
		if err != nil {
			return "", err
		}

		log("file %s\n%s\n----", in, b)
	}
	src, err := cc.NewFileSource2(in, true)
	if err != nil {
		return "", err
	}

	if filepath.Ext(in) != ".s" {
		return out, c.compileSource(out, in, src)
	}

	f, err := os.Open(in)
	if err != nil {
		return "", err
	}

	sc := bufio.NewScanner(f)

	g, err := os.Create(out)
	if err != nil {
		return "", err
	}

	defer func() { err = errs(err, g.Close()) }()

	gb := bufio.NewWriter(g)

	defer func() { err = errs(err, gb.Flush()) }()

	r, w := io.Pipe()

	var e2 error
	go func() {
		defer func() {
			e2 = errs(e2, w.Close())
		}()

		if _, e2 = fmt.Fprintf(w, `
const Lf = %q

`, in); e2 != nil {
			return
		}
		for sc.Scan() {
			if _, e2 = fmt.Fprintf(w, "// %s\n", sc.Text()); e2 != nil {
				return
			}
		}

		_, e2 = fmt.Fprintf(w, "\n")
	}()

	err = ccgo.NewSharedObject(gb, c.goos, c.goarch, r)
	if err = errs(err, e2); err != nil {
		out = ""
	}
	return out, err
}

func (c *config) compileSource(out, in string, src cc.Source) (err error) {
	c.objects = append(c.objects, out)
	c.objMap[in] = out
	if !c.c {
		c.remove = append(c.remove, out)
	}
	defs := []string{`
#define _DEFAULT_SOURCE 1
#define _POSIX_C_SOURCE 200809
#define __FUNCTION__ __func__ // gcc compatibility
#define __ccgo__ 1
`}
	for _, v := range c.D {
		defs = append(defs, fmt.Sprintf("#define %s", v))
	}

	tweaks := &cc.Tweaks{
		// TrackExpand:   func(s string) { fmt.Print(s) },
		// TrackIncludes: func(s string) { fmt.Printf("[#include %s]\n", s) },
		EnableAnonymousStructFields: true,
		EnableBinaryLiterals:        true,
		EnableEmptyStructs:          true,
		EnableImplicitBuiltins:      true,
		EnableOmitFuncDeclSpec:      true,
		EnableReturnExprInVoidFunc:  true,
		EnableUnionCasts:            true,
		IgnoreUnknownPragmas:        true,
		InjectFinalNL:               true,
	}

	sources := []cc.Source{cc.NewStringSource("<defines>", strings.Join(defs, "\n"))}
	builtin, err := cc.Builtin()
	if err != nil {
		return err
	}

	sources = append(sources, builtin)

	if c.E {
		tweaks.PreprocessOnly = true
		switch {
		case c.dM:
			prev := "\n"
			last := "\n"
			tweaks.DefinesOnly = true
			tweaks.TrackExpand = func(s string) {
				ts := strings.TrimSpace(s)
				if !strings.HasPrefix(ts, "#") {
					return
				}

				ts = strings.TrimSpace(ts[1:])
				if !strings.HasPrefix(ts, "define") {
					return
				}
				s += "\n"
				if s == "\n" && last == "\n" && prev == "\n" {
					return
				}

				fmt.Print(s)
				prev = last
				last = s
			}
		default:
			prev := "\n"
			last := "\n"
			tweaks.TrackExpand = func(s string) {
				ts := strings.TrimSpace(s)
				if strings.HasPrefix(ts, "#") {
					return
				}

				if s == "\n" && last == "\n" && prev == "\n" {
					return
				}

				fmt.Print(s)
				prev = last
				last = s

			}
		}
	}
	sources = append(sources, src)
	tu, err := cc.Translate(tweaks, c.incPaths, c.sysPaths, sources...)
	if err != nil {
		return err
	}

	if c.E {
		return nil
	}

	f, err := os.Create(out)
	if err != nil {
		return err
	}

	defer func() { err = errs(err, f.Close()) }()

	b := bufio.NewWriter(f)

	defer func() { err = errs(err, b.Flush()) }()

	objTweaks := &ccgo.NewObjectTweaks{
		DefineValues: c.defineValues,
		FreeStanding: c.ffreeStanding,
		FullTLDPaths: c.fullPaths,
		StructChecks: c.structChecks,
		Watch:        c.watch,
	}
	return ccgo.NewObject(b, c.goos, c.goarch, src.Name(), tu, objTweaks)
}
