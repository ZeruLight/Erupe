// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate rm -f scanner.go trigraphs.go
//go:generate golex -o trigraphs.go trigraphs.l
//go:generate golex -o scanner.go scanner.l

//go:generate rm -f ast.go
//go:generate yy -kind Case -o parser.y -astImport "\"modernc.org/xc\";\"go/token\";\"fmt\"" -prettyString PrettyString parser.yy

//go:generate rm -f parser.go
//go:generate goyacc -o /dev/null -xegen xegen parser.y
//go:generate goyacc -o parser.go -pool -fs -xe xegen -dlvalf "%v" -dlval "PrettyString(lval.Token)" parser.y
//go:generate rm -f xegen

//go:generate stringer -output stringer.go -type=cond,Linkage,StorageDuration enum.go
//go:generate sh -c "go test -run ^Example |fe"
//go:generate gofmt -l -s -w .

// Package cc is a C99 compiler front end. Work In Progress. API unstable.
//
// This package is no longer maintained. Please see the v3 version at
//
// 	https://modernc.org/cc/v3
package cc // import "modernc.org/cc/v2"

import (
	"bufio"
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"modernc.org/ir"
	"modernc.org/strutil"
	"modernc.org/xc"
)

const (
	// CRT0Source is the source code of the C startup code.
	CRT0Source = `int main();

__FILE_TYPE__ __stdfiles[3];
char **environ;
void *stdin = &__stdfiles[0], *stdout = &__stdfiles[1], *stderr = &__stdfiles[2];

void _start(int argc, char **argv)
{
	__register_stdfiles(stdin, stdout, stderr, &environ);
	__builtin_exit(((int (*)(int, char **))main) (argc, argv));
}	
`
	cacheSize = 500
)

var (
	_ Source = (*FileSource)(nil)
	_ Source = (*StringSource)(nil)

	_ debug.GCStats

	// YYDebug points to parser's yyDebug variable.
	YYDebug        = &yyDebug
	traceMacroDefs bool

	cache   = make(map[cacheKey][]uint32, cacheSize)
	cacheMu sync.Mutex // Guards cache, fset
	fset    = token.NewFileSet()

	packageDir     string
	headers        string
	selfImportPath string
)

func init() {
	ip, err := strutil.ImportPath()
	if err != nil {
		panic(err)
	}

	selfImportPath = ip
	if packageDir, err = findRepo(ip); err != nil {
		panic(err)
	}

	headers = filepath.Join(packageDir, "headers", fmt.Sprintf("%v_%v", env("GOOS", runtime.GOOS), env("GOARCH", runtime.GOARCH)))
}

func findRepo(s string) (string, error) {
	s = filepath.FromSlash(s)
	for _, v := range strings.Split(strutil.Gopath(), string(os.PathListSeparator)) {
		p := filepath.Join(v, "src", s)
		fi, err := os.Lstat(p)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			wd, err := os.Getwd()
			if err != nil {
				return "", err
			}

			if p, err = filepath.Rel(wd, p); err != nil {
				return "", err
			}

			if p, err = filepath.Abs(p); err != nil {
				return "", err
			}

			return p, nil
		}
	}
	return "", fmt.Errorf("%q: cannot find repository", s)
}

// ImportPath reports the import path of this package.
func ImportPath() string { return selfImportPath }

// Builtin returns the Source for built-in and predefined stuff or an error, if any.
func Builtin() (Source, error) { return NewFileSource(filepath.Join(headers, "builtin.h")) }

// Crt0 returns the Source for program initialization code.
func Crt0() (Source, error) { return NewStringSource("crt0.c", CRT0Source), nil } //TODO mv to ccgo

// MustBuiltin is like Builtin but panics on error.
func MustBuiltin() Source { return MustFileSource(filepath.Join(headers, "builtin.h")) }

// MustCrt0 is like Crt0 but panics on error.
func MustCrt0() Source { //TODO mv to ccgo
	s, err := Crt0()
	if err != nil {
		panic(err)
	}

	return s
}

// MustFileSource is like NewFileSource but panics on error.
func MustFileSource(nm string) *FileSource { return MustFileSource2(nm, true) }

// MustFileSource2 is like NewFileSource but panics on error.
func MustFileSource2(nm string, cacheable bool) *FileSource {
	src, err := NewFileSource2(nm, cacheable)
	if err != nil {
		wd, _ := os.Getwd()
		panic(fmt.Errorf("%v: %v (wd %v)", nm, err, wd))
	}

	return src
}

// Paths returns the system header search paths, or an error, if any. If local
// is true the package-local, cached header search paths are returned.
func Paths(local bool) ([]string, error) {
	p := filepath.Join(headers, "paths")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	a := strings.Split(string(b), "\n")
	for i, v := range a {
		switch {
		case local:
			a[i] = filepath.Join(headers, strings.TrimSpace(v))
		default:
			a[i] = strings.TrimSpace(v)
		}
	}
	return a, nil
}

// HostConfig executes HostCppConfig with the cpp argument set to "cpp". For
// more info please see the documentation of HostCppConfig.
func HostConfig(opts ...string) (predefined string, includePaths, sysIncludePaths []string, err error) {
	return HostCppConfig("cpp", opts...)
}

// HostCppConfig returns the system C preprocessor configuration, or an error,
// if any.  The configuration is obtained by running the cpp command. For the
// predefined macros list the '-dM' options is added. For the include paths
// lists, the option '-v' is added and the output is parsed to extract the
// "..." include and <...> include paths. To add any other options to cpp, list
// them in opts.
//
// The function relies on a POSIX compatible C preprocessor installed.
// Execution of HostConfig is not free, so caching the results is recommended
// whenever possible.
func HostCppConfig(cpp string, opts ...string) (predefined string, includePaths, sysIncludePaths []string, err error) {
	args := append(append([]string{"-dM"}, opts...), os.DevNull)
	// cross-compile e.g. win64 -> win32
	if env("GOARCH", runtime.GOARCH) == "386" {
		args = append(args, "-m32")
	}
	pre, err := exec.Command(cpp, args...).Output()
	if err != nil {
		return "", nil, nil, err
	}

	args = append(append([]string{"-v"}, opts...), os.DevNull)
	out, err := exec.Command(cpp, args...).CombinedOutput()
	if err != nil {
		return "", nil, nil, err
	}

	sep := "\n"
	if env("GOOS", runtime.GOOS) == "windows" {
		sep = "\r\n"
	}

	a := strings.Split(string(out), sep)
	for i := 0; i < len(a); {
		switch a[i] {
		case "#include \"...\" search starts here:":
		loop:
			for i = i + 1; i < len(a); {
				switch v := a[i]; {
				case strings.HasPrefix(v, "#") || v == "End of search list.":
					break loop
				default:
					includePaths = append(includePaths, strings.TrimSpace(v))
					i++
				}
			}
		case "#include <...> search starts here:":
			for i = i + 1; i < len(a); {
				switch v := a[i]; {
				case strings.HasPrefix(v, "#") || v == "End of search list.":
					return string(pre), includePaths, sysIncludePaths, nil
				default:
					sysIncludePaths = append(sysIncludePaths, strings.TrimSpace(v))
					i++
				}
			}
		default:
			i++
		}
	}
	return "", nil, nil, fmt.Errorf("failed parsing %s -v output", cpp)
}

type cacheKey struct {
	name  string
	mtime int64
}

// FlushCache removes all items in the file cache used by instances of FileSource.
func FlushCache() { //TODO-
	cacheMu.Lock()
	cache = make(map[cacheKey][]uint32, cacheSize)
	fset = token.NewFileSet()
	cacheMu.Unlock()
}

// TranslationUnit represents a translation unit, see [0]6.9.
type TranslationUnit struct {
	ExternalDeclarationList *ExternalDeclarationList
	FileScope               *Scope
	FileSet                 *token.FileSet
	IncludePaths            []string
	Macros                  map[int]*Macro
	Model                   Model
	SysIncludePaths         []string
}

// Tweaks amend the behavior of the parser.
type Tweaks struct { //TODO- remove all options
	TrackExpand   func(string)
	TrackIncludes func(string)

	DefinesOnly                 bool // like in CC -E -dM foo.c
	EnableAnonymousStructFields bool // struct{int;}
	EnableBinaryLiterals        bool // 0b101010 == 42
	EnableEmptyStructs          bool // struct{}
	EnableImplicitBuiltins      bool // Undefined printf becomes __builtin_printf.
	EnableImplicitDeclarations  bool // eg. using exit(1) w/o #include <stdlib.h>
	EnableOmitFuncDeclSpec      bool // foo() { ... } == int foo() { ... }
	EnablePointerCompatibility  bool // All pointers are assignment compatible.
	EnableReturnExprInVoidFunc  bool // void f() { return 1; }
	EnableTrigraphs             bool
	EnableUnionCasts            bool // (union foo)0
	IgnoreUnknownPragmas        bool // #pragma
	InjectFinalNL               bool // Specs want the source to always end in a newline.
	PreprocessOnly              bool // like in CC -E foo.c
	cppExpandTest               bool // Fake includes
}

// Translate preprocesses, parses and type checks a translation unit using
// includePaths and sysIncludePaths for looking for "foo.h" and <foo.h> files.
// A special path "@" is interpretted as 'the same directory as where the file
// with the #include is'. The input consists of sources which must include any
// predefined/builtin stuff.
func Translate(tweaks *Tweaks, includePaths, sysIncludePaths []string, sources ...Source) (tu *TranslationUnit, err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			if e != nil {
				err = fmt.Errorf("%v\n%s", e, debugStack())
				return
			}

			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack())
		}
	}()

	model, err := NewModel()
	if err != nil {
		return nil, err
	}

	ctx, err := newContext(tweaks)
	if err != nil {
		return nil, err
	}

	ctx.model = model
	ctx.includePaths = append([]string(nil), includePaths...)
	ctx.sysIncludePaths = append([]string(nil), sysIncludePaths...)
	if tu, err = ctx.parse(sources); err != nil {
		return nil, err
	}

	if tweaks.PreprocessOnly {
		returned = true
		return nil, nil
	}

	if err := tu.ExternalDeclarationList.check(ctx); err != nil {
		return nil, err
	}

	if err := ctx.error(); err != nil {
		return nil, err
	}

	tu.IncludePaths = append([]string(nil), includePaths...)
	tu.SysIncludePaths = append([]string(nil), sysIncludePaths...)
	returned = true
	return tu, nil
}

// Translation unit context.
type context struct {
	errors       scanner.ErrorList
	exampleAST   interface{}
	exampleRule  int
	includePaths []string
	model        Model
	scope        *Scope
	sync.Mutex
	sysIncludePaths []string
	tweaks          *Tweaks
}

func newContext(t *Tweaks) (*context, error) {
	return &context{
		scope:  newScope(nil),
		tweaks: t,
	}, nil
}

func (c *context) err(n Node, msg string, args ...interface{}) { c.errPos(n.Pos(), msg, args...) }
func (c *context) newScope() *Scope                            { c.scope = newScope(c.scope); return c.scope }
func (c *context) newStructScope()                             { c.scope = newScope(c.scope); c.scope.structScope = true }

func (c *context) position(n Node) (r token.Position) {
	if n != nil {
		return fset.PositionFor(n.Pos(), true)
	}

	return r
}

func (c *context) errPos(pos token.Pos, msg string, args ...interface{}) {
	c.Lock()
	s := fmt.Sprintf(msg, args...)
	//s = fmt.Sprintf("%s\n====%s\n----", s, debug.Stack())
	c.errors.Add(fset.PositionFor(pos, true), s)
	c.Unlock()
}

func (c *context) error() error {
	c.Lock()

	defer c.Unlock()

	if len(c.errors) == 0 {
		return nil
	}

	c.errors.Sort()
	err := append(scanner.ErrorList(nil), c.errors...)
	return err
}

func (c *context) parse(in []Source) (_ *TranslationUnit, err error) {
	defer func() { c.scope.typedefs = nil }()

	cpp := newCPP(c)
	r, err := cpp.parse(in...)
	if err != nil {
		return nil, err
	}

	lx, err := newLexer(c, "", 0, nil)
	if err != nil {
		return nil, err
	}

	p := newTokenPipe(1024)
	if c.tweaks.PreprocessOnly {
		p.emitWhiteSpace = true
	}
	lx.tc = p

	var cppErr error
	ch := make(chan struct{})
	go func() {
		returned := false

		defer func() {
			p.close()
			e := recover()
			if !returned && cppErr == nil {
				cppErr = fmt.Errorf("PANIC: %v\n%s", e, debugStack())
				c.err(nopos, "%v", cppErr)
			}
			ch <- struct{}{}
		}()

		if cppErr = cpp.eval(r, p); cppErr != nil {
			c.err(nopos, "%v", cppErr)
		}
		returned = true
	}()

	if c.tweaks.PreprocessOnly {
		for {
			t := p.read()
			if t.Rune == ccEOF {
				break
			}

			if f := c.tweaks.TrackExpand; f != nil {
				if p := c.position(t); filepath.Base(p.Filename) != "builtin.h" {
					f(TokSrc(t.Token))
				}
			}
		}
		if err := c.error(); err != nil {
			return nil, err
		}

		return nil, cppErr
	}

	ok := lx.parse(TRANSLATION_UNIT)
	if err := c.error(); err != nil || !ok {
		go func() { // drain
			for range p.ch {
			}
		}()
		return nil, err
	}

	if c.scope.Parent != nil {
		panic("internal error")
	}

	<-ch
	if cppErr != nil {
		return nil, cppErr
	}

	tu := lx.ast.(*TranslationUnit)
	tu.Macros = cpp.macros
	return tu, nil
}

func (c *context) popScope() (old, new *Scope) {
	old = c.scope
	c.scope = c.scope.Parent
	return old, c.scope
}

func (c *context) ptrDiff() Type {
	d, ok := c.scope.LookupIdent(idPtrdiffT).(*Declarator)
	if !ok {
		psz := c.model[Ptr].Size
		for _, v := range []TypeKind{Int, Long, LongLong} {
			if c.model[v].Size >= psz {
				return v
			}
		}
		panic("internal error")
	}

	if !d.DeclarationSpecifier.IsTypedef() {
		panic(d.Type)
	}

	return d.Type
}

func (c *context) wideChar() Type {
	d, ok := c.scope.LookupIdent(idWcharT).(*Declarator)
	if !ok {
		var sz int
		switch goos := env("GOOS", ""); goos {
		case "windows":
			sz = 2
		case "linux":
			sz = 4
		default:
			panic(goos)
		}
		for _, v := range []TypeKind{SChar, Short, Int, Long, LongLong} {
			if c.model[v].Size >= sz {
				return v
			}
		}
		panic("internal error")
	}

	if !d.DeclarationSpecifier.IsTypedef() {
		panic(d.Type)
	}

	return d.Type
}

func (c *context) charConst(t xc.Token) Operand {
	switch t.Rune {
	case CHARCONST:
		s := string(t.S())
		s = s[1 : len(s)-1] // Remove outer 's.
		if len(s) == 1 {
			return Operand{Type: Int, Value: &ir.Int64Value{Value: int64(s[0])}}
		}

		runes := []rune(s)
		var r rune
		switch runes[0] {
		case '\\':
			r, _ = decodeEscapeSequence(runes)
			if r < 0 {
				r = -r
			}
		default:
			r = runes[0]
		}
		return Operand{Type: Int, Value: &ir.Int64Value{Value: int64(r)}}
	case LONGCHARCONST:
		s := t.S()
		var buf bytes.Buffer
		s = s[2 : len(s)-1]
		runes := []rune(string(s))
		for i := 0; i < len(runes); {
			switch r := runes[i]; {
			case r == '\\':
				r, n := decodeEscapeSequence(runes[i:])
				switch {
				case r < 0:
					buf.WriteByte(byte(-r))
				default:
					buf.WriteRune(r)
				}
				i += n
			default:
				buf.WriteByte(byte(r))
				i++
			}
		}
		s = buf.Bytes()
		runes = []rune(string(s))
		if len(runes) != 1 {
			panic("TODO")
		}

		return Operand{Type: Long, Value: &ir.Int64Value{Value: int64(runes[0])}}
	default:
		panic("internal error")
	}
}

func (c *context) strConst(t xc.Token) Operand {
	s := t.S()
	switch t.Rune {
	case LONGSTRINGLITERAL:
		s = s[1:] // Remove leading 'L'.
		fallthrough
	case STRINGLITERAL:
		var buf bytes.Buffer
		s = s[1 : len(s)-1] // Remove outer "s.
		runes := []rune(string(s))
		for i := 0; i < len(runes); {
			switch r := runes[i]; {
			case r == '\\':
				r, n := decodeEscapeSequence(runes[i:])
				switch {
				case r < 0:
					buf.WriteByte(byte(-r))
				default:
					buf.WriteRune(r)
				}
				i += n
			default:
				buf.WriteByte(byte(r))
				i++
			}
		}
		switch t.Rune {
		case LONGSTRINGLITERAL:
			runes := []rune(buf.String())
			typ := c.wideChar()
			return Operand{
				Type:  &ArrayType{Item: typ, Size: Operand{Type: Int, Value: &ir.Int64Value{Value: c.model.Sizeof(typ) * int64(len(runes)+1)}}},
				Value: &ir.WideStringValue{Value: runes},
			}
		case STRINGLITERAL:
			return Operand{
				Type:  &ArrayType{Item: Char, Size: Operand{Type: Int, Value: &ir.Int64Value{Value: int64(len(buf.Bytes()) + 1)}}},
				Value: &ir.StringValue{StringID: ir.StringID(dict.ID(buf.Bytes()))},
			}
		}
	}
	panic("internal error")
}

func (c *context) sizeof(t Type) Operand {
	sz := c.model.Sizeof(t)
	d, ok := c.scope.LookupIdent(idSizeT).(*Declarator)
	if !ok {
		psz := c.model[Ptr].Size
		for _, v := range []TypeKind{UInt, ULong, ULongLong} {
			if c.model[v].Size >= psz {
				return newIntConst(c, nopos, uint64(sz), v)
			}
		}
		panic("internal error")
	}

	if !d.DeclarationSpecifier.IsTypedef() {
		panic(d.Type)
	}

	r := Operand{Type: d.Type, Value: &ir.Int64Value{Value: sz}}
	if x, ok := underlyingType(t, false).(*ArrayType); ok && x.Size.Value == nil {
		r.Value = nil
	}
	return r
}

func (c *context) toC(ch rune, val int) rune {
	if ch != IDENTIFIER {
		return ch
	}

	if x, ok := keywords[val]; ok {
		return x
	}

	return ch
}

// Source represents parser's input.
type Source interface {
	Cache([]uint32)                     // Optionally cache the encoded source. Can be a no-operation.
	Cached() []uint32                   // Return nil or the optionally encoded source cached by earlier call to Cache.
	Name() string                       // Result will be used in reporting source code positions.
	ReadCloser() (io.ReadCloser, error) // Where to read the source from
	Size() (int64, error)               // Report the size of the source in bytes.
	String() string
}

// FileSource is a Source reading from a named file.
type FileSource struct {
	*bufio.Reader
	f    *os.File
	path string
	key  cacheKey
	size int64

	cacheable bool
}

// NewFileSource returns a newly created *FileSource reading from name.
func NewFileSource(name string) (*FileSource, error) { return NewFileSource2(name, true) }

// NewFileSource2 returns a newly created *FileSource reading from name.
func NewFileSource2(name string, cacheable bool) (*FileSource, error) { //TODO-?
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	r := &FileSource{f: f, path: name, cacheable: cacheable}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	r.size = fi.Size()
	r.key = cacheKey{name, fi.ModTime().UnixNano()}
	return r, nil
}

func (s *FileSource) String() string { return s.Name() }

// Cache implements Source.
func (s *FileSource) Cache(a []uint32) {
	if !s.cacheable {
		return
	}

	cacheMu.Lock()
	if len(cache) > cacheSize {
		for k := range cache {
			delete(cache, k)
			break
		}
	}
	cache[s.key] = a
	cacheMu.Unlock()
}

// Cached implements Source.
func (s *FileSource) Cached() (r []uint32) {
	if !s.cacheable {
		return nil
	}

	cacheMu.Lock()
	var ok bool
	r, ok = cache[s.key]
	cacheMu.Unlock()
	if ok {
		s.Close()
	}
	return r
}

// Close implements io.ReadCloser.
func (s *FileSource) Close() error {
	if f := s.f; f != nil {
		s.f = nil
		return f.Close()
	}

	return nil
}

// Name implements Source.
func (s *FileSource) Name() string { return s.path }

// ReadCloser implements Source.
func (s *FileSource) ReadCloser() (io.ReadCloser, error) {
	s.Reader = bufio.NewReader(s.f)
	return s, nil
}

// Size implements Source.
func (s *FileSource) Size() (int64, error) { return s.size, nil }

// StringSource is a Source reading from a string.
type StringSource struct {
	*strings.Reader
	name string
	src  string
}

// NewStringSource returns a newly created *StringSource reading from src and
// having the presumed name.
func NewStringSource(name, src string) *StringSource { return &StringSource{name: name, src: src} }

func (s *StringSource) String() string { return s.Name() }

// Cache implements Source.
func (s *StringSource) Cache(a []uint32) {
	cacheMu.Lock()
	if len(cache) > cacheSize {
		for k := range cache {
			delete(cache, k)
			break
		}
	}
	cache[cacheKey{mtime: -1, name: s.src}] = a
	cacheMu.Unlock()
}

// Cached implements Source.
func (s *StringSource) Cached() (r []uint32) {
	cacheMu.Lock()
	r = cache[cacheKey{mtime: -1, name: s.src}]
	cacheMu.Unlock()
	return r
}

// Close implements io.ReadCloser.
func (s *StringSource) Close() error { return nil }

// Name implements Source.
func (s *StringSource) Name() string { return s.name }

// Size implements Source.
func (s *StringSource) Size() (int64, error) { return int64(len(s.src)), nil }

// ReadCloser implements Source.
func (s *StringSource) ReadCloser() (io.ReadCloser, error) {
	s.Reader = strings.NewReader(s.src)
	return s, nil
}

// Scope binds names to declarations.
type Scope struct {
	EnumTags   map[int]*EnumSpecifier // name ID: *EnumSpecifier
	Idents     map[int]Node           // name ID: Node in {*Declarator, EnumerationConstant}
	Labels     map[int]*LabeledStmt   // name ID: label
	Parent     *Scope
	StructTags map[int]*StructOrUnionSpecifier // name ID: *StructOrUnionSpecifier

	// parser support
	typedefs map[int]bool // name: isTypedef
	fixDecl  int

	forStmtEndScope *Scope
	structScope     bool
	typedef         bool
}

func newScope(parent *Scope) *Scope { return &Scope{Parent: parent} }

func (s *Scope) insertLabel(ctx *context, st *LabeledStmt) {
	for s.Parent != nil && s.Parent.Parent != nil {
		s = s.Parent
	}
	if s.Labels == nil {
		s.Labels = map[int]*LabeledStmt{}
	}
	if ex := s.Labels[st.Token.Val]; ex != nil {
		panic("TODO")
	}

	s.Labels[st.Token.Val] = st
}

func (s *Scope) insertEnumTag(ctx *context, nm int, es *EnumSpecifier) {
	for s.structScope {
		s = s.Parent
	}
	if s.EnumTags == nil {
		s.EnumTags = map[int]*EnumSpecifier{}
	}
	if ex := s.EnumTags[nm]; ex != nil {
		if ex == es || ex.isCompatible(es) {
			return
		}

		panic(fmt.Errorf("%s\n----\n%s", ex, es))
	}

	s.EnumTags[nm] = es
}

func (s *Scope) insertDeclarator(ctx *context, d *Declarator) {
	if s.Idents == nil {
		s.Idents = map[int]Node{}
	}
	nm := d.Name()
	if ex := s.Idents[nm]; ex != nil {
		panic("internal error 8")
	}

	s.Idents[nm] = d
}

func (s *Scope) insertEnumerationConstant(ctx *context, c *EnumerationConstant) {
	for s.structScope {
		s = s.Parent
	}
	if s.Idents == nil {
		s.Idents = map[int]Node{}
	}
	nm := c.Token.Val
	if ex := s.Idents[nm]; ex != nil {
		if ex == c {
			return
		}

		if x, ok := ex.(*EnumerationConstant); ok && x.equal(c) {
			return
		}

		panic(fmt.Errorf("%v: %v, %v", ctx.position(c), ex, c))
	}

	s.Idents[nm] = c
}

func (s *Scope) insertStructTag(ctx *context, ss *StructOrUnionSpecifier) {
	for s.structScope {
		s = s.Parent
	}
	if s.StructTags == nil {
		s.StructTags = map[int]*StructOrUnionSpecifier{}
	}
	nm := ss.IdentifierOpt.Token.Val
	if ex := s.StructTags[nm]; ex != nil && !ex.typ.IsCompatible(ss.typ) {
		panic(fmt.Errorf("%v: %v, %v", ctx.position(ss), ex.typ, ss.typ))
	}

	s.StructTags[nm] = ss
}

func (s *Scope) insertTypedef(ctx *context, nm int, isTypedef bool) {
	//dbg("%p(parent %p).insertTypedef(%q, %v)", s, s.Parent, dict.S(nm), isTypedef)
	if s.typedefs == nil {
		s.typedefs = map[int]bool{}
	}
	// Redefinitions, if any, are ignored during parsing, but checked later in insertDeclarator.
	s.typedefs[nm] = isTypedef
}

func (s *Scope) isTypedef(nm int) bool {
	//dbg("==== %p(parent %p).isTypedef(%q)", s, s.Parent, dict.S(nm))
	for s != nil {
		//dbg("%p", s)
		if v, ok := s.typedefs[nm]; ok {
			if s.structScope && !v {
				s = s.Parent
				continue
			}

			//dbg("%p -> %v %v", s, v, ok)
			return v
		}

		s = s.Parent
	}
	return false
}

// LookupIdent will return the Node associated with name ID nm.
func (s *Scope) LookupIdent(nm int) Node {
	for s != nil {
		if n := s.Idents[nm]; n != nil {
			return n
		}

		s = s.Parent
	}
	return nil
}

// LookupLabel will return the Node associated with label ID nm.
func (s *Scope) LookupLabel(nm int) Node {
	for s != nil {
		if n := s.Labels[nm]; n != nil {
			if s.Parent == nil && s.Parent.Parent != nil {
				panic("internal error")
			}

			return n
		}

		s = s.Parent
	}
	return nil
}

func (s *Scope) lookupEnumTag(nm int) *EnumSpecifier {
	for s != nil {
		if n := s.EnumTags[nm]; n != nil {
			return n
		}

		s = s.Parent
	}
	return nil
}

func (s *Scope) lookupStructTag(nm int) *StructOrUnionSpecifier {
	for s != nil {
		if n := s.StructTags[nm]; n != nil {
			return n
		}

		s = s.Parent
	}
	return nil
}

func (s *Scope) String() string {
	var a []string
	for _, v := range s.Idents {
		switch x := v.(type) {
		case *Declarator:
			a = append(a, string(dict.S(x.Name())))
		default:
			panic(fmt.Errorf("%T", x))
		}
	}
	sort.Strings(a)
	return "{" + strings.Join(a, ", ") + "}"
}
