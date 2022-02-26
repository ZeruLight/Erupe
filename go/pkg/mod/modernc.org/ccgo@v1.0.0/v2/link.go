// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v2"

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"modernc.org/cc/v2"
	"modernc.org/ccgo/v2/internal/object"
	"modernc.org/sortutil"
)

/*

-------------------------------------------------------------------------------
Linker constants (const Lx = "value")
-------------------------------------------------------------------------------

LD<mangled name>	Macro value definition. Value: value.
La<mangled name>	Declarator with external linkage has alias attribute. Value: mangled alias name with external linkage.
Lb<mangled name>	Declarator with external linkage has alias attribute. Value: mangled alias name with internal linkage.
Ld<mangled name>	Definition (provides) with external linkage. Value: type.
Le<mangled name>	Declaration (requires) with external linkage. Value: type.
Lf			Translation unit boundary. Value: file name.
Lv<mangled name>		Visibility of a declarator with external linkage. Value: eg. "hidden".
Lw<mangled name>		Declarator with external linkage has weak attribute. Value: none.

-------------------------------------------------------------------------------
Linker magic names
-------------------------------------------------------------------------------

Lb + n			-> bss + off
Ld + "foo"		-> dss + off
Lw + "foo"		-> ts + off wchar_t string

*/

// ============================================================================
// Weak aliases
//
// ----------------------------------------------------------------------------
// C
//
// int __pthread_mutex_unlock(pthread_mutex_t *m)
// {
// 	// ...
// }
//
// extern typeof(__pthread_mutex_unlock) pthread_mutex_unlock __attribute__((weak, alias("__pthread_mutex_unlock")));
//
// ----------------------------------------------------------------------------
// Go
//
// // const LdX__pthread_mutex_unlock = "func(TLS, uintptr) int32" //TODO-
//
// // X__pthread_mutex_unlock is defined at src/thread/pthread_mutex_unlock.c:3:5
// func X__pthread_mutex_unlock(tls TLS, _m uintptr /* *Tpthread_mutex_t = struct{F__u struct{F int64; _ [32]byte};} */) (r int32) {
// 	// ...
// }
//
// // const LeXpthread_mutex_unlock = "func(TLS, uintptr) int32" //TODO-
//
// // const LwXpthread_mutex_unlock = "" //TODO-
//
// // const LaXpthread_mutex_unlock = "X__pthread_mutex_unlock" //TODO-

const (
	lConstPrefix = "const L"
)

var (
	traceLConsts bool
)

// NewSharedObject writes shared linker object files from in to out.
func NewSharedObject(out io.Writer, goos, goarch string, in io.Reader) (err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack2())
		}
		if e != nil && err == nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	r, w := io.Pipe()
	var e2 error

	go func() {
		defer func() {
			if err := w.Close(); err != nil && e2 == nil {
				e2 = err
			}
		}()

		_, e2 = io.Copy(w, in)
	}()

	err = object.Encode(out, goos, goarch, object.ObjVersion, object.ObjMagic, r)
	if e2 != nil && err == nil {
		err = e2
	}
	returned = true
	return err
}

// NewObjectTweaks amend NewObject behavior.
type NewObjectTweaks struct {
	DefineValues bool // --ccgo-define-values
	FreeStanding bool // -ffreestanding
	FullTLDPaths bool // --ccgo-full-paths
	StructChecks bool // --ccgo-struct-checks
	Watch        bool // --ccgo-watch
}

// NewObject writes a linker object file produced from in that comes from file
// to out.
func NewObject(out io.Writer, goos, goarch, file string, in *cc.TranslationUnit, tweaks *NewObjectTweaks) (err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack2())
		}
		if e != nil && err == nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	r, w := io.Pipe()
	g := newNGen(w, in, file, tweaks)

	go func() {
		defer func() {
			if err := recover(); err != nil && g.err == nil {
				g.err = fmt.Errorf("%v", err)
			}
			if err := w.Close(); err != nil && g.err == nil {
				g.err = err
			}
		}()

		g.gen()
	}()

	err = object.Encode(out, goos, goarch, object.ObjVersion, object.ObjMagic, r)
	if e := g.err; e != nil && err == nil {
		err = e
	}
	returned = true
	return err
}

type unit struct { // Translation unit
	provides map[string]struct{} // key: mangled declarator name with external linkage
	requires map[string]struct{} // key: mangled declarator name with external linkage
}

func newUnit() *unit { return &unit{} }

func (u *unit) provide(nm string) {
	if u.provides == nil {
		u.provides = map[string]struct{}{}
	}
	u.provides[nm] = struct{}{}
}

func (u *unit) require(nm string) {
	if u.requires == nil {
		u.requires = map[string]struct{}{}
	}
	u.requires[nm] = struct{}{}
}

type attr struct {
	alias string
	weak  bool
}

// Linker produces Go files from object files.
type Linker struct {
	attrs            map[string]attr // name: attr
	bss              int64
	crtPrefix        string
	declaredExterns  map[string]string // name: type
	definedExterns   map[string]string // name: type
	ds               []byte
	errs             scanner.ErrorList
	errsMu           sync.Mutex
	goarch           string
	goos             string
	helpers          map[string]int
	macroDefs        map[string]string
	num              int
	out              *bufio.Writer
	producedExterns  map[string]struct{} // name: -
	renamedHelperNum map[string]int
	renamedHelpers   map[string]int
	renamedNameNum   map[string]int
	renamedNames     map[string]int
	strings          map[int]int64
	tempFile         *os.File
	text             []int
	tld              []string
	ts               int64
	unit             *unit
	units            []*unit
	visibility       map[string]string // name: type
	visitor          visitor
	wout             *bufio.Writer

	Main             bool // Seen external definition of main.
	bool2int         bool
	ignoreDeclarator bool
}

// NewLinker returns a newly created Linker writing to out.
//
// The Linker must be eventually closed to prevent resource leaks.
func NewLinker(out io.Writer, goos, goarch string) (*Linker, error) {
	bin, ok := out.(*bufio.Writer)
	if !ok {
		bin = bufio.NewWriter(out)
	}

	tempFile, err := ioutil.TempFile("", "ccgo-linker-")
	if err != nil {
		return nil, err
	}

	r := &Linker{
		attrs:            map[string]attr{},
		crtPrefix:        crt,
		declaredExterns:  map[string]string{},
		definedExterns:   map[string]string{},
		goarch:           goarch,
		goos:             goos,
		helpers:          map[string]int{},
		macroDefs:        map[string]string{},
		out:              bin,
		producedExterns:  map[string]struct{}{},
		renamedHelperNum: map[string]int{},
		renamedHelpers:   map[string]int{},
		renamedNameNum:   map[string]int{},
		renamedNames:     map[string]int{},
		strings:          map[int]int64{},
		tempFile:         tempFile,
		visibility:       map[string]string{},
		wout:             bufio.NewWriter(tempFile),
	}
	r.visitor = visitor{r}
	return r, nil
}

func (l *Linker) w(s string, args ...interface{}) {
	if _, err := fmt.Fprintf(l.wout, s, args...); err != nil {
		todo("", err)
	}
}

func (l *Linker) err(msg string, args ...interface{}) {
	l.errsMu.Lock()
	l.errs.Add(token.Position{}, fmt.Sprintf(msg, args...))
	l.errsMu.Unlock()
}

func (l *Linker) error() error {
	l.errsMu.Lock()

	defer l.errsMu.Unlock()

	if len(l.errs) == 0 {
		return nil
	}

	var a []string
	for _, v := range l.errs {
		a = append(a, v.Error())
	}
	return fmt.Errorf("%s", strings.Join(a[:sortutil.Dedupe(sort.StringSlice(a))], "\n"))
}

// Link incerementaly links objects files.
func (l *Linker) Link(fn string, obj io.Reader) (err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack2())
		}
		if e != nil && err == nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	if err := l.link(fn, obj); err != nil {
		l.err("%v", err)
	}
	err = l.error()
	returned = true
	return err
}

func (l *Linker) link(fn string, obj io.Reader) error {
	r, w := io.Pipe()

	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				l.err(err.Error())
			}
		}()

		if err := object.Decode(w, l.goos, l.goarch, object.ObjVersion, object.ObjMagic, obj); err != nil {
			l.err("%v", err)
		}
	}()

	l.w("\n%sf = %q\n", lConstPrefix, fn)
	l.unit = newUnit()
	l.units = append(l.units, l.unit)

	sc := newLineScanner(r)
	for sc.Scan() {
		s := sc.Text()
		switch {
		case strings.HasPrefix(s, lConstPrefix):
			l.lConst(s[len(lConstPrefix):])
		default:
			l.w("%s\n", s)
		}
	}
	return sc.Err()
}

func (l *Linker) lConst(s string) { // x<name> = "value"
	if traceLConsts {
		l.w("\n// %s%s\n", lConstPrefix, s)
	}
	l.w("\n%s%s\n", lConstPrefix, s)
	a := strings.SplitN(s, " ", 3)
	nm := a[0]
	arg, err := strconv.Unquote(a[2])
	if err != nil {
		todo("%s: %s", s, err)
	}

	switch {
	case strings.HasPrefix(nm, "d"): // defines (provides)
		nm = nm[1:]
		if nm == "Xmain" {
			l.Main = true
		}
		l.definedExterns[nm] = arg
	case strings.HasPrefix(nm, "e"): // declares (requires)
		nm = nm[1:]
		l.declaredExterns[nm] = arg
	case strings.HasPrefix(nm, "v"):
		nm = nm[1:]
		switch arg {
		case "hidden":
			l.visibility[nm] = arg
		default:
			todo("%q", s)
		}
	case strings.HasPrefix(nm, "w"):
		nm = nm[1:]
		attr := l.attrs[nm]
		attr.weak = true
		l.attrs[nm] = attr
	case strings.HasPrefix(nm, "D"):
		nm = nm[1:]
		if _, ok := l.macroDefs[nm]; ok {
			break
		}

		l.macroDefs[nm] = arg
	case
		strings.HasPrefix(nm, "a"),
		strings.HasPrefix(nm, "b"):

		nm = nm[1:]
		attr := l.attrs[nm]
		attr.alias = arg
		l.attrs[nm] = attr
	case strings.HasPrefix(nm, "h"): // helper
		l.num++
		l.helpers[arg] = l.num
	case
		nm == "f",
		strings.HasPrefix(s, "sofile "),
		strings.HasPrefix(s, "soname "):

		// nop
	case nm == "freestanding":
		l.crtPrefix = ""
	default:
		todo("%s", s)
		panic("unreachable")
	}
}

func (l *Linker) parseID(s string) (string, int) {
	for i := len(s) - 1; i >= 0; i-- {
		if c := s[i]; c < '0' || c > '9' {
			i++
			n, err := strconv.ParseInt(s[i:], 10, 31)
			if err != nil {
				todo("", err)
			}

			return s[:i], int(n)
		}
	}
	todo("missing helper local ID")
	panic("unreachable")
}

// Close finihes the linking. The header argument is written prior to any other
// linker's own output, which does not include the package clause.
func (l *Linker) Close(header string) (err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack2())
		}
		if e != nil && err == nil {
			err = fmt.Errorf("%v", e)
		}
		if e := os.Remove(l.tempFile.Name()); e != nil && err == nil {
			err = e
		}
	}()

	if err := l.close(header); err != nil {
		l.err("%v", err)
	}
	err = l.error()
	returned = true
	return err
}

func (l *Linker) close(header string) (err error) {
	if err = l.wout.Flush(); err != nil {
		return err
	}

	if _, err = l.tempFile.Seek(0, os.SEEK_SET); err != nil {
		return err
	}

	l.wout = l.out
	l.w("%s\n", strings.TrimSpace(header))

	defer func() {
		if e := l.wout.Flush(); e != nil && err == nil {
			err = e
		}
	}()

	const (
		skipBlank = iota
		collectComments
		copy
		copyFunc
		copyParen
	)

	sc := newLineScanner(l.tempFile)
	state := skipBlank
	for l.scan(sc) {
		s := sc.Text()
		switch state {
		case skipBlank:
			if len(s) == 0 {
				break
			}

			l.tld = l.tld[:0]
			state = collectComments
			fallthrough
		case collectComments:
			if s == "" {
				l.emit()
				state = skipBlank
				break
			}

			l.tld = append(l.tld, s)
			if strings.HasPrefix(s, "//") {
				break
			}

			switch {
			case strings.HasPrefix(s, "const ("):
				state = copyParen
			case strings.HasPrefix(s, "var"):
				state = copy
			case strings.HasPrefix(s, "func"):
				if strings.HasSuffix(s, "}") {
					l.emit()
					state = skipBlank
					break
				}

				state = copyFunc
			case strings.HasPrefix(s, "type"):
				if !strings.HasSuffix(s, "{") {
					l.emit()
					state = skipBlank
					break
				}

				state = copyFunc
			default:
				todo("%q", s)
			}
		case copy:
			l.tld = append(l.tld, s)
			if len(s) == 0 {
				l.emit()
				state = skipBlank
				break
			}
		case copyFunc:
			l.tld = append(l.tld, s)
			if s == "}" {
				l.emit()
				state = skipBlank
			}
		case copyParen:
			l.tld = append(l.tld, s)
			if s == ")" {
				l.emit()
				state = skipBlank
			}
		default:
			todo("", state)
		}
	}

	if err = sc.Err(); err != nil {
		return err
	}

	if len(l.tld) != 0 {
		l.emit()
	}

	l.genDefs()
	l.genWeak()
	l.genHelpers()

	if l.bss != 0 {
		l.w(`
var bss = %sBSS(&bssInit[0])

var bssInit [%d]byte
`, l.crtPrefix, l.bss)
	}
	if n := len(l.ds); n != 0 {
		if n < 16 {
			l.ds = append(l.ds, make([]byte, 16-n)...)
		}
		l.w("\nvar ds = %sDS(dsInit)\n", l.crtPrefix)
		l.w("\nvar dsInit = []byte{")
		if isTesting {
			l.w("\n\t\t")
		}
		for i, v := range l.ds {
			l.w("%#02x, ", v)
			if isTesting && i&15 == 15 {
				l.w("// %#x\n\t", i&^15)
			}
		}
		if isTesting && len(l.ds)&15 != 0 {
			l.w("// %#x\n", len(l.ds)&^15)
		}
		l.w("}\n")
	}
	if l.ts != 0 {
		l.w("\nvar ts = %sTS(\"", l.crtPrefix)
		for _, v := range l.text {
			s := fmt.Sprintf("%q", dict.S(v))
			l.w("%s\\x00", s[1:len(s)-1])
			for n := len(dict.S(v)) + 1; n%4 != 0; n++ {
				l.w("\\x00")
			}
		}
		l.w("\")\n")
	}
	return nil
}

func (l *Linker) genDefs() {
	var a []string
	for k := range l.macroDefs {
		a = append(a, k)
	}
	if len(a) == 0 {
		return
	}
	sort.Strings(a)
	l.w("\nconst (")
	for _, k := range a {
		l.w("\n\t%s = %s", k, l.macroDefs[k])
	}
	l.w("\n)\n")
}

func (l *Linker) genWeak() {
	var a []string
	for k, v := range l.attrs {
		if v.weak {
			if _, ok := l.definedExterns[k]; !ok {
				a = append(a, k)
			}
		}
	}
	sort.Strings(a)
	for _, k := range a {
		attr := l.attrs[k]
		if attr.alias == "" {
			continue
		}

		l.w("\nvar %s = %s\n", k, attr.alias)
	}
}

func (l *Linker) isFunc(nm string) bool {
	t := l.declaredExterns[nm]
	if t == "" {
		t = l.definedExterns[nm]
	}
	return strings.HasPrefix(t, "func")
}

func (l *Linker) genHelpers() {
	if l.bool2int {
		l.w(`
func bool2int(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
`)
	}
}

func (l *Linker) emit() (err error) {
	// fmt.Printf("==== emit\n%s\n----\n", strings.Join(l.tld, "\n")) //TODO- DBG
	defer func() { l.tld = l.tld[:0] }()

	if l.ignoreDeclarator {
		l.ignoreDeclarator = false
		return
	}

	s := strings.Join(l.tld, "\n")
	fset := token.NewFileSet()
	in := io.MultiReader(bytes.NewBufferString("package p\n"), bytes.NewBufferString(s))
	file, err := parser.ParseFile(fset, "", in, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Walk(&l.visitor, file)
	e := emitor{out: l.wout}
	format.Node(&e, fset, file)
	return nil
}

type visitor struct {
	*Linker
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
out:
	switch x := node.(type) {
	case *ast.SelectorExpr:
		ast.Walk(v, x.X)
		return nil
	case *ast.BinaryExpr:
		switch x2 := x.X.(type) {
		case *ast.Ident:
			switch {
			case x2.Name == "Lb":
				rhs := x.Y.(*ast.BasicLit)
				n, err := strconv.ParseInt(rhs.Value, 10, 63)
				if err != nil {
					todo("", err)
				}

				if n == 0 {
					n++
				}
				x2.Name = "bss"
				rhs.Value = fmt.Sprint(v.bss)
				v.bss += roundup(n, 8) // keep alignment
				return nil
			case x2.Name == "Ld":
				rhs := x.Y.(*ast.BasicLit)
				s, err := strconv.Unquote(rhs.Value)
				if err != nil {
					todo("", err)
				}

				x2.Name = "ds"
				rhs.Value = fmt.Sprint(v.allocDS(s))
				return nil
			case x2.Name == "Lw":
				rhs := x.Y.(*ast.BasicLit)
				s, err := strconv.Unquote(rhs.Value)
				if err != nil {
					todo("", err)
				}

				x2.Name = "ts"
				rhs.Value = fmt.Sprint(v.allocString(dict.SID(s)))
				return nil
			}
		}
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			if x.Value[0] == '`' {
				break
			}

			s, err := strconv.Unquote(x.Value)
			if err != nil {
				todo("", err)
			}

			x.Value = fmt.Sprintf("ts+%d %s", v.allocString(dict.SID(s)), strComment2([]byte(s)))
		}
	case *ast.Ident:
		nm := x.Name
		switch {
		case nm == "Lb":
			x.Name = "bss+0"
			v.bss += roundup(1, 8) // keep alignment
			break out
		case nm == "Ld":
			x.Name = "ds+0"
			v.allocDS("\x00")
			break out
		}

		if _, ok := v.renamedHelpers[nm]; ok {
			x.Name = v.renameHelper(nm)
			break
		}

		switch {
		case strings.HasPrefix(nm, "X"):
			if _, ok := v.definedExterns[nm]; !ok {
				x.Name = fmt.Sprintf("%s%s", v.crtPrefix, nm)
			}
		case strings.HasPrefix(nm, "x") && len(nm) > 1: // Static linkage
			x.Name = v.rename("x", "x", nm[1:])
		case strings.HasPrefix(nm, "C") && nm != "Copy": // Enum constant
			x.Name = v.rename("C", "c", nm[1:])
		case strings.HasPrefix(nm, "E"): // Tagged enum type
			x.Name = v.rename("E", "e", nm[1:])
		case strings.HasPrefix(nm, "N") && !strings.HasPrefix(nm, "Nz"): // Named type
			x.Name = v.rename("N", "n", nm[1:])
		case strings.HasPrefix(nm, "S"): // Tagged struct type
			x.Name = v.rename("S", "s", nm[1:])
		case strings.HasPrefix(nm, "T") && nm != "TLS": // Named type
			x.Name = v.rename("T", "t", nm[1:])
		case strings.HasPrefix(nm, "U"): // Tagged union type
			x.Name = v.rename("U", "u", nm[1:])
		case nm == "bool2int":
			v.bool2int = true
		}
	}
	return v
}

func (l *Linker) allocDS(s string) int64 {
	up := roundup(int64(len(l.ds)), 8) // keep alignment
	if n := up - int64(len(l.ds)); n != 0 {
		l.ds = append(l.ds, make([]byte, n)...)
	}
	r := len(l.ds)
	l.ds = append(l.ds, s...)
	return int64(r)
}

func (l *Linker) renameHelper(nm string) string {
	n := l.renamedHelpers[nm]
	if n == 0 {
		l.num++
		n = l.num
		l.renamedHelpers[nm] = n
	}
	for i := 0; i < len(nm); i++ {
		if c := nm[i]; c < '0' || c > '9' {
			continue
		}

		var j int
		for j = i + 1; j < len(nm); j++ {
			if c := nm[j]; c < '0' || c > '9' {
				break
			}
		}
		// "abc123def" i: 3, j:6
		return fmt.Sprintf("%s%d%s", nm[:i], n, nm[j:])
	}
	todo("%q", nm)
	panic("unreachable")
}

func (l *Linker) rename(prefix0, prefix, nm string) string {
	switch c := nm[0]; {
	case c >= '0' && c <= '9':
		n := l.renamedNames[nm]
		if n == 0 {
			l.num++
			n = l.num
			l.renamedNames[nm] = n
		}
		for {
			if c := nm[0]; c < '0' || c > '9' {
				break
			}

			nm = nm[1:]
		}
		return fmt.Sprintf("%s%d%s", prefix, n, nm)
	default:
		n, ok := l.renamedNames[nm]
		if !ok {
			n = l.renamedNameNum[nm]
			l.renamedNames[nm] = n
			l.renamedNameNum[nm]++
		}
		for {
			if c := nm[0]; c < '0' || c > '9' {
				break
			}

			nm = nm[1:]
		}
		if n == 0 {
			return fmt.Sprintf("%s%s", prefix0, nm)
		}

		return fmt.Sprintf("%s%d%s", prefix, n, nm)
	}
}

func (l *Linker) allocString(s int) int64 {
	if n, ok := l.strings[s]; ok {
		return n
	}

	r := l.ts
	l.strings[s] = r
	l.ts += int64(len(dict.S(s))) + 1
	for l.ts%4 != 0 {
		l.ts++
	}
	l.text = append(l.text, s)
	return r
}

type emitor struct {
	out  io.Writer
	gate bool
}

func (e *emitor) Write(b []byte) (int, error) {
	if e.gate {
		return e.out.Write(b)
	}

	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		e.gate = true
		n, err := e.out.Write(b[i+1:])
		return n + i, err
	}

	return len(b), nil
}

func (l *Linker) scan(sc *lineScanner) bool {
	for {
		if !sc.Scan() {
			return false
		}

		if s := sc.Text(); strings.HasPrefix(s, lConstPrefix) {
			l.lConst2(s[len(lConstPrefix):])
			continue
		}

		return true
	}
}

func (l *Linker) lConst2(s string) { // x<name> = "value"
	if traceLConsts {
		l.w("\n// %s%s\n", lConstPrefix, s)
	}
	//l.w("\n// %s%s //TODO- \n", lConstPrefix, s)
	a := strings.SplitN(s, " ", 3)
	nm := a[0]
	arg, err := strconv.Unquote(a[2])
	if err != nil {
		todo("", err)
	}
	switch {
	case
		strings.HasPrefix(nm, "D"),
		strings.HasPrefix(nm, "a"),
		strings.HasPrefix(nm, "e"),
		strings.HasPrefix(nm, "freestanding"),
		strings.HasPrefix(nm, "v"),
		strings.HasPrefix(nm, "w"):

		// nop
	case strings.HasPrefix(nm, "b"):
		nm = nm[1:]
		attr := l.attrs[nm]
		attr.alias = l.rename("x", "x", arg[1:])
		l.attrs[nm] = attr
	case strings.HasPrefix(nm, "d"):
		nm = nm[1:]
		if _, ok := l.producedExterns[nm]; ok {
			l.ignoreDeclarator = true
			break
		}

		l.producedExterns[nm] = struct{}{}
	case
		nm == "f",
		strings.HasPrefix(s, "sofile "),
		strings.HasPrefix(s, "soname "):

		l.w("\n// linking %s\n", arg)
		for k := range l.renamedNames {
			delete(l.renamedNames, k)
		}
		for k := range l.renamedHelpers {
			delete(l.renamedHelpers, k)
		}
	case strings.HasPrefix(nm, "h"): // helper
		l.genHelper(l.renameHelper(nm[1:]), strings.Split(arg, "$"))
	default:
		todo("%s", s)
	}
}

func (l *Linker) genHelper(nm string, a []string) {
	l.w("\nfunc %s", nm)
	switch a[0] {
	case "add%d", "and%d", "div%d", "mod%d", "mul%d", "or%d", "sub%d", "xor%d":
		// eg.: [0: "add%d" 1: op "+" 2: lhs type "uint16" 3: rhs type "uint8" 4: promotion type "int32"]
		l.w("(p *%[2]s, v %[3]s) (r %[2]s) { r = %[2]s(%[4]s(*p) %[1]s %[4]s(v)); *p = r; return r }", a[1], a[2], a[3], a[4])
	case "and%db", "or%db", "xor%db":
		// eg.: [0: "or%db" 1: op "|" 2: lhs type "uint16" 3: rhs type "uint8" 4: promotion type "int32" 5: packed type "uint32" 6: bitoff 7: promotion type bits 8: bits 9: lhs type bits]
		l.w(`(p *%[5]s, v %[3]s) (r %[2]s) {
	r = %[2]s((%[4]s(%[2]s(*p>>%[6]s))<<(%[7]s-%[8]s)>>(%[7]s-%[8]s)) %[1]s %[4]s(v))
	*p = (*p &^ ((1<<%[8]s - 1) << %[6]s)) | (%[5]s(r) << %[6]s & ((1<<%[8]s - 1) << %[6]s))
	return r<<(%[9]s-%[8]s)>>(%[9]s-%[8]s)
}`, a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9])
	case "set%d": // eg.: [0: "set%d" 1: op "" 2: operand type "uint32"]
		l.w("(p *%[2]s, v %[2]s) %[2]s { *p = v; return v }", a[1], a[2])
	case "setb%d":
		// eg.: [0: "set%db" 1: packed type "uint32" 2: lhs type "int16" 3: rhs type "char" 4: bitoff 5: bits 6: lhs type bits]
		l.w(`(p *%[1]s, v %[3]s) %[2]s { 
	w := %[1]s(v) & (1<<%[5]s-1)
	*p = (*p &^ ((1<<%[5]s - 1) << %[4]s)) | (w << %[4]s)
	return %[2]s(w)<<(%[6]s-%[5]s)>>(%[6]s-%[5]s)
}`, a[1], a[2], a[3], a[4], a[5], a[6])

	case "rsh%d":
		// eg.: [0: "rsh%d" 1: op ">>" 2: lhs type "uint32" 3: promotion type]
		l.w("(p *%[2]s, v uint) (r %[2]s) { r = %[2]s(%[3]s(*p) >> v); *p = r; return r }", a[1], a[2], a[3])
	case "fn%d":
		// eg.: [0: "fn%d" 1: type "unc()"]
		l.w("(p uintptr) %[1]s { return *(*%[1]s)(unsafe.Pointer(&p)) }", a[1])
	case "fp%d":
		l.w("(f %[1]s) uintptr { return *(*uintptr)(unsafe.Pointer(&f)) }", a[1])
	case "postinc%d":
		// eg.: [0: "postinc%d" 1: operand type "int32" 2: delta "1"]
		l.w("(p *%[1]s) %[1]s { r := *p; *p += %[2]s; return r }", a[1], a[2])
	case "preinc%d":
		// eg.: [0: "preinc%d" 1: operand type "int32" 2: delta "1"]
		l.w("(p *%[1]s) %[1]s { *p += %[2]s; return *p }", a[1], a[2])
	case "postinc%db":
		//TODO op.type(fp.type(*p>>fp.bitoff)<<x>>x)
		// eg.: [0: "postinc%db" 1: delta "1" 2: lhs type "int32" 3: pack type "uint8" 4: lhs type bits "32" 5: bits "3" 6: bitoff "2"]

		l.w(`(p *%[3]s) %[2]s {
	r := %[2]s(*p>>%[6]s)<<(%[4]s-%[5]s)>>(%[4]s-%[5]s)
	*p = (*p &^ ((1<<%[5]s - 1) << %[6]s)) | (%[3]s(r+%[1]s) << %[6]s & ((1<<%[5]s - 1) << %[6]s))
	return r
}`, a[1], a[2], a[3], a[4], a[5], a[6])
	case "preinc%db":
		//TODO op.type(fp.type(*p>>fp.bitoff)<<x>>x)
		// eg.: [0: "preinc%db" 1: delta "1" 2: lhs type "int32" 3: pack type "uint8" 4: lhs type bits "32" 5: bits "3" 6: bitoff "2"]
		l.w(`(p *%[3]s) %[2]s {
	r := (%[2]s(*p>>%[6]s+%[1]s)<<(%[4]s-%[5]s)>>(%[4]s-%[5]s))
	*p = (*p &^ ((1<<%[5]s - 1) << %[6]s)) | (%[3]s(r) << %[6]s & ((1<<%[5]s - 1) << %[6]s))
	return r
}`, a[1], a[2], a[3], a[4], a[5], a[6])

	case "float2int%d":
		// eg.: [0: "float2int%d" 1: type "uint64" 2: max "18446744073709551615"]
		l.w("(f float32) %[1]s { if f > %[2]s { return 0 }; return %[1]s(f) }", a[1], a[2])
	default:
		todo("%q", a)
	}
	l.w("\n")
}
