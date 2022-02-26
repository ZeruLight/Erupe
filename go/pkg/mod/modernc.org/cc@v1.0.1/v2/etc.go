// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

import (
	"bytes"
	"fmt"
	"go/scanner"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"modernc.org/strutil"
	"modernc.org/xc"
)

var (
	bNL    = []byte{'\n'}
	bPanic = []byte("panic")
)

// PrettyString returns pretty strings for things produced by this package.
func PrettyString(v interface{}) string {
	return strutil.PrettyString(v, "", "", printHooks)
}

func debugStack() []byte {
	b := debug.Stack()
	i := bytes.Index(b, bPanic) + 1
	if i <= 0 {
		return b
	}

	b = b[i:]
	if i = bytes.Index(b, bPanic); i <= 0 {
		return b
	}

	b = b[i:]
	if i = bytes.Index(b, bPanic); i > 0 {
		b = b[i:]
	}
	return b
}

func cppTrimSpace(toks []cppToken) []cppToken {
	for len(toks) != 0 && toks[0].Rune == ' ' {
		toks = toks[1:]
	}
	for len(toks) != 0 && toks[len(toks)-1].Rune == ' ' {
		toks = toks[:len(toks)-1]
	}
	return toks
}

func trimSpace(toks []xc.Token) []xc.Token {
	for len(toks) != 0 && toks[0].Rune == ' ' {
		toks = toks[1:]
	}
	for len(toks) != 0 && toks[len(toks)-1].Rune == ' ' {
		toks = toks[:len(toks)-1]
	}
	return toks
}

func cppTrimAllSpace(toks []cppToken) []cppToken {
	w := 0
	for _, v := range toks {
		switch v.Rune {
		case ' ', '\n':
			// nop
		default:
			toks[w] = v
			w++
		}
	}
	return toks[:w]
}

func trimAllSpace(toks []xc.Token) []xc.Token {
	w := 0
	for _, v := range toks {
		switch v.Rune {
		case ' ', '\n':
			// nop
		default:
			toks[w] = v
			w++
		}
	}
	return toks[:w]
}

func isVaList(t Type) bool { //TODO export and use
	x, ok := t.(*NamedType)
	return ok && (x.Name == idVaList || x.Name == idBuiltinVaList)
}

func cppToksDump(toks []cppToken, sep string) string {
	var a []string
	for _, t := range toks {
		if t.Rune == '\n' {
			continue
		}

		a = append(a, TokSrc(t.Token))
	}
	return strings.Join(a, sep)
}

func toksDump(toks []xc.Token, sep string) string {
	var a []string
	for _, t := range toks {
		if t.Rune == '\n' {
			continue
		}

		a = append(a, TokSrc(t))
	}
	return strings.Join(a, sep)
}

func prefer(d *Declarator) bool {
	if d.DeclarationSpecifier.IsExtern() {
		return false
	}

	t := d.Type
	for {
		switch x := UnderlyingType(t).(type) {
		case *ArrayType:
			return x.Size.Type != nil
		case *FunctionType:
			return d.FunctionDefinition != nil
		case
			*EnumType,
			*StructType,
			*UnionType:

			return true
		case *PointerType:
			t = x.Item
		case *TaggedStructType:
			return x.Type != nil
		case TypeKind:
			if x.IsScalarType() || x == Void {
				return true
			}

			panic(x)
		default:
			panic(x)
		}
	}
}

func env(key, val string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}

	return val
}

// IncompatibleTypeDiff is a debug helper.
func IncompatibleTypeDiff(a, b Type) {
	if a == nil || b == nil {
		panic(fmt.Errorf("TODO\n\t%v\n\t%v", a, b))
	}

	ok0 := a.IsCompatible(b)
	a0 := a
	b0 := b
	a = UnderlyingType(a)
	b = UnderlyingType(b)
	if ok := a.IsCompatible(b); ok != ok0 {
		fmt.Printf(`UnderlyingType changed compatibility from %v to %v (%T %T)\n
	%v
	%v
	----
	%v
	%v
	====
	%v
	%v
	----
	%v
	%v
	----
`, ok0, ok, a0, b0, a0, b0, a, b, PrettyString(a0), PrettyString(b0), PrettyString(a), PrettyString(b))
	}
	switch x := a.(type) {
	case *FunctionType:
		y := b.(*FunctionType)
		if g, e := x.Variadic, y.Variadic; g != e {
			fmt.Printf("different .Variadic %v %v\n", g, e)
		}
		if g, e := x.Result, y.Result; g != e && !g.IsCompatible(e) {
			fmt.Printf("Incompatible result types in %v, %T(%v) and %T(%v)\n", x, g, g, e, e)
			IncompatibleTypeDiff(g, e)
		}
		if g, e := len(x.Params), len(y.Params); g != e {
			fmt.Printf("len(Params) %v %v\n", g, e)
			return
		}

		for i, v := range x.Params {
			w := y.Params[i]
			if g, e := v, w; g != e && !g.IsCompatible(e) {
				fmt.Printf("Parameter %v.Type: types not compatible %T %T\n\t%v\n\t%v\n", i, g, e, g, e)
				IncompatibleTypeDiff(g, e)
			}
		}
	case *PointerType:
		y := b.(*PointerType)
		IncompatibleTypeDiff(x.Item, y.Item)
	case *StructType:
		switch y := b.(type) {
		case *StructType:
			if g, e := x.Tag, y.Tag; g != e {
				fmt.Printf("Tags %q %q\n", dict.S(g), dict.S(e))
				return
			}

			if g, e := len(x.Fields), len(y.Fields); g != e {
				fmt.Printf("len(Fields) %v %v\n", g, e)
				return
			}

			for i, v := range x.Fields {
				w := y.Fields[i]
				if g, e := v.Name, w.Name; g != e {
					fmt.Printf("Field %v.Name: %q %q\n", i, dict.S(g), dict.S(e))
				}
				if g, e := v.Bits, w.Bits; g != e {
					fmt.Printf("Field %v.Bits: %v %v\n", i, g, e)
				}
				if g, e := v.PackedType, w.PackedType; g != e && !g.IsCompatible(e) {
					fmt.Printf("Field %v.PackedType: incompatible types\n\t%v\n\t%v\n", i, g, e)
					IncompatibleTypeDiff(g, e)
				}
				if g, e := v.Type, w.Type; g != e && !g.IsCompatible(e) {
					fmt.Printf("Field %v.Type: incompatible types\n\t%v\n\t%v\n", i, g, e)
					IncompatibleTypeDiff(g, e)
				}
			}
		case *TaggedStructType:
			b = y.Type
			if b == y {
				panic("TODO")
			}

			if b == nil {
				panic(fmt.Errorf("%T(%v).Type: %v", y, y, y.Type))
			}

			IncompatibleTypeDiff(x, b)
		default:
			panic(fmt.Errorf("%T", y))
		}
	case TypeKind:
		y := b.(TypeKind)
		if g, e := x, y; g != e && !g.IsCompatible(e) {
			fmt.Printf("TypeKinds differ: %v %v\n", g, e)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// ErrString is like error.Error() but expands scanner.ErrorList.
func ErrString(err error) string {
	var b bytes.Buffer
	printError(&b, "", err)
	return b.String()
}

func printError(w io.Writer, pref string, err error) {
	switch x := err.(type) {
	case scanner.ErrorList:
		for i, v := range x {
			fmt.Fprintf(w, "%s%v\n", pref, v)
			if i == 50 {
				fmt.Fprintln(w, "too many errors")
				break
			}
		}
	default:
		fmt.Fprintf(w, "%s%v\n", pref, err)
	}
}

func cppToks(toks []xc.Token) []cppToken {
	r := make([]cppToken, len(toks))
	for i, v := range toks {
		r[i].Token = v
	}
	return r
}

func xcToks(toks []cppToken) []xc.Token {
	r := make([]xc.Token, len(toks))
	for i, v := range toks {
		r[i] = v.Token
	}
	return r
}
