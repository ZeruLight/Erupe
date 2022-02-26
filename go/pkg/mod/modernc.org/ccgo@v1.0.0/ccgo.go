// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ccgo translates cc ASTs to Go source code. (Work In Progress)
//
// Changelog
//
// 2018-07-01 This package is no longer maintained. Please see the v2 version at
//
//	https://modernc.org/ccgo/v2
//
// 2018-04-10: This code no longer passes tests and soon it will not even build
// due to the upcoming changes in cznic/crt. For that reason the current crt
// master branch package is now included in this repository for the improbable
// case someone wants to make the code work again.
package ccgo // import "modernc.org/ccgo"

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"modernc.org/cc"
	"modernc.org/ccir"
	"modernc.org/ir"
	"modernc.org/irgo"
	"modernc.org/virtual"
	"modernc.org/xc"
)

var (
	// Testing amends things for tests.
	Testing bool

	dict      = xc.Dict
	idVoidPtr = ir.TypeID(dict.SID("*struct{}"))
)

//TODO remove me.
func TODO(msg string, more ...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "%s:%d: %v\n", path.Base(fn), fl, fmt.Sprintf(msg, more...))
	os.Stderr.Sync()
	panic(fmt.Errorf("%s:%d: %v", path.Base(fn), fl, fmt.Sprintf(msg, more...)))
}

func typ(o ir.Object, tc ir.TypeCache, tm map[ir.TypeID]string, id ir.TypeID, nm, pkg ir.NameID) {
	if nm == 0 {
		return
	}

	s := dict.S(int(nm))
	t := tc.MustType(id)
	for {
		done := true
		switch t.Kind() {
		case ir.Array:
			t = t.(*ir.ArrayType).Item
			for s[0] != ']' {
				s = s[1:]
			}
			s = s[1:]
			done = false
		case ir.Pointer:
			if t.ID() == idVoidPtr {
				return
			}

			t = t.(*ir.PointerType).Element
			s = s[1:]
			done = false
		}
		if done {
			break
		}
	}
	id = t.ID()
	switch t.Kind() {
	case ir.Struct, ir.Union:
		if _, ok := tm[id]; !ok {
			b := s
			if pkg != 0 {
				b = nil
				b = append(b, dict.S(int(pkg))...)
				b = append(b, '.')
				b = append(b, s...)
			}
			tm[id] = string(b)
		}
	}
}

type options struct {
	ast        []*cc.TranslationUnit
	libcTypes  map[ir.TypeID]string
	library    bool
	qualifiers []string
}

// Option is a configuration/setup function that can be passed to the New
// function.
type Option func(*options) error

// Packages annotate the translation units with a package qualifier. Items
// annotated with a package qualifier are not rendered and references to
// external definitions in such translation units are prefixed with the
// respective qualifier.
func Packages(qualifiers []string) Option {
	return func(o *options) error {
		if g, e := len(qualifiers), len(o.ast); g > e {
			return fmt.Errorf("too many package qualifiers: %v > %v", g, e)
		}

		o.qualifiers = qualifiers
		return nil
	}
}

// Library selects the library linking mode, ie. the linkew will include all
// objects having external linkage.
func Library() Option {
	return func(o *options) error {
		o.library = true
		return nil
	}
}

// LibcTypes makes code refering to libc types import them from the CRT package.
func LibcTypes() Option {
	return func(o *options) error {
		o.libcTypes = typeMap
		return nil
	}
}

// New writes Go code generated from ast to out. No package or import clause is
// generated.
func New(ast []*cc.TranslationUnit, out io.Writer, opts ...Option) (err error) {
	if !Testing {
		defer func() {
			switch x := recover().(type) {
			case nil:
				// ok
			default:
				err = fmt.Errorf("ccgo.New: PANIC: %v", x)
			}
		}()
	}

	o := &options{ast: ast}
	for _, v := range opts {
		if err := v(o); err != nil {
			return err
		}
	}

	tc := ir.TypeCache{}
	tm := map[ir.TypeID]string{}
	for k, v := range o.libcTypes {
		tm[k] = v
	}
	var build [][]ir.Object
	for i, v := range ast {
		obj, err := ccir.New(v, ccir.TypeCache(tc))
		if err != nil {
			return err
		}

		var pkg, tpkg ir.NameID
		if i < len(o.qualifiers) {
			pkg = ir.NameID(dict.SID(o.qualifiers[i]))
		}
		for _, v := range obj {
			// switch x := v.(type) { //TODO-
			// case *ir.DataDefinition:
			// 	fmt.Printf("# [%v]: %T %v %v\n", i, x, x.ObjectBase, x.Value)
			// case *ir.FunctionDefinition:
			// 	fmt.Printf("# [%v]: %T %v %v\n", i, x, x.ObjectBase, x.Arguments)
			// 	for i, v := range x.Body {
			// 		fmt.Printf("%#05x\t%v\n", i, v)
			// 	}
			// }
			if err := v.Verify(); err != nil {
				return err
			}

			if b := v.Base(); !virtual.IsBuiltin(b.NameID) {
				b.Package = pkg
				tpkg = pkg
			}
			switch x := v.(type) {
			case *ir.DataDefinition:
				typ(x, tc, tm, x.TypeID, x.TypeName, tpkg)
			case *ir.FunctionDefinition:
				for _, v := range x.Body {
					switch y := v.(type) {
					case *ir.VariableDeclaration:
						typ(x, tc, tm, y.TypeID, y.TypeName, tpkg)
					}
				}
			}
		}

		build = append(build, obj)
	}

	var obj []ir.Object
	switch {
	case o.library:
		if obj, err = ir.LinkLib(build...); err != nil {
			return err
		}
	default:
		if obj, err = ir.LinkMain(build...); err != nil {
			return err
		}
	}

	for _, v := range obj {
		if err := v.Verify(); err != nil {
			return err
		}
	}

	return irgo.New(out, obj, tm, irgo.TypeCache(tc))
}
