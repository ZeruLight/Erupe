// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v2"

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
)

var (
	_ io.Writer = (*opt)(nil)
	_ io.Writer = (*nopt)(nil)
)

type opt struct {
	fn           string
	fset         *token.FileSet
	needBool2int int
	out          io.Writer
	out0         bytes.Buffer
	used         map[string]struct{}

	forceBool2int bool
	noBool2int    bool
	write         bool
}

type nopt struct {
	fn    string
	fset  *token.FileSet
	out   bytes.Buffer
	used  map[string]struct{}
	write bool
}

func newOpt() *opt { return &opt{} }

func (o *opt) Write(b []byte) (int, error) {
	if o.write {
		return o.out0.Write(b)
	}

	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		o.write = true
		n, err := o.out0.Write(b[i+1:])
		return n + i, err
	}

	return len(b), nil
}

func (o *opt) pos(n ast.Node) token.Position {
	if n == nil {
		return token.Position{}
	}

	return o.fset.Position(n.Pos())
}

func (o *opt) do(out io.Writer, in io.Reader, fn string, needBool2int int) error {
	o.fn = fn
	o.needBool2int = needBool2int
	o.out = out
	o.fset = token.NewFileSet()
	ast, err := parser.ParseFile(o.fset, "", io.MultiReader(bytes.NewBufferString(fmt.Sprintf("package p // %s\n", fn)), in), parser.ParseComments)
	if err != nil {
		return err
	}

	o.file(ast)
	if err := format.Node(o, o.fset, ast); err != nil {
		return err
	}

	if o.needBool2int != 0 && !o.noBool2int || o.forceBool2int {
		_, err = o.Write([]byte(`
func bool2int(b bool) int32 {
	if b {
		return 1
	}

	return 0
}
`))
	}
	if err != nil {
		return err
	}

	b := bytes.Replace(o.out0.Bytes(), []byte("\n\n}"), []byte("\n}"), -1)
	b = bytes.Replace(b, []byte("\n\t;\n"), []byte("\n"), -1)
	b = bytes.Replace(b, []byte("{\n\n"), []byte("{\n"), -1)
	b = bytes.Replace(b, []byte(":\n\n"), []byte(":\n"), -1)
	b = bytes.Replace(b, []byte("\n\n\tdefer"), []byte("\n\tdefer"), -1)
	b = bytes.Replace(b, []byte("\n\n\t\t"), []byte("\n\t\t"), -1)
	b = bytes.Replace(b, []byte("\n\n\t)"), []byte("\n\t)"), -1)
	if traceOpt {
		os.Stderr.Write(b)
	}
	_, err = o.out.Write(b)
	return err
}

func (o *opt) file(n *ast.File) {
	for i := range n.Decls {
		o.decl(&n.Decls[i])
	}
}

func (o *opt) decl(n *ast.Decl) {
	switch x := (*n).(type) {
	case *ast.FuncDecl:
		o.used = map[string]struct{}{}
		o.blockStmt(x.Body, true)
	case *ast.GenDecl:
		for i := range x.Specs {
			o.spec(&x.Specs[i])
		}
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *opt) spec(n *ast.Spec) {
	switch x := (*n).(type) {
	case *ast.TypeSpec:
		// nop
	case *ast.ValueSpec:
		use := x.Names[0].Name != "_"
		for i := range x.Values {
			o.expr(&x.Values[i], use)
			switch x2 := x.Values[i].(type) {
			case *ast.ParenExpr:
				x.Values[i] = x2.X
			}
		}
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *opt) blockStmt(n *ast.BlockStmt, outermost bool) {
	if n == nil {
		return
	}

	o.body(&n.List)
	if !outermost {
		return
	}

	w := 0
	for _, v := range n.List {
		switch x := v.(type) {
		case *ast.AssignStmt:
			if x2, ok := x.Lhs[0].(*ast.Ident); ok && x2.Name == "_" {
				if _, used := o.used[x.Rhs[0].(*ast.Ident).Name]; used {
					continue
				}
			}
		case *ast.DeclStmt:
			switch x2 := x.Decl.(type) {
			case *ast.GenDecl:
				w := 0
				for _, v := range x2.Specs {
					if x3, ok := v.(*ast.ValueSpec); ok && x3.Names[0].Name == "_" {
						if _, used := o.used[x3.Values[0].(*ast.Ident).Name]; used {
							continue
						}
					}

					x2.Specs[w] = v
					w++
				}
				x2.Specs = x2.Specs[:w]
			}
		}
		n.List[w] = v
		w++
	}
	n.List = n.List[:w]
}

func (o *opt) body(l0 *[]ast.Stmt) {
	l := *l0
	w := 0
	for i := range l {
		o.stmt(&l[i])
		switch x := l[i].(type) {
		case *ast.EmptyStmt:
			// nop
		default:
			l[w] = x
			w++
		}
	}
	*l0 = l[:w]
}

func (o *opt) stmt(n *ast.Stmt) {
	switch x := (*n).(type) {
	case nil:
		// nop
	case *ast.AssignStmt:
		for i := range x.Lhs {
			o.expr(&x.Lhs[i], false)
			switch x2 := x.Lhs[i].(type) {
			case *ast.ParenExpr:
				x.Lhs[i] = x2.X
			}
		}
		use := true
		if x, ok := x.Lhs[0].(*ast.Ident); ok && x.Name == "_" {
			use = false
		}
		for i := range x.Rhs {
			o.expr(&x.Rhs[i], use)
			switch x2 := x.Rhs[i].(type) {
			case *ast.ParenExpr:
				x.Rhs[i] = x2.X
			}
		}
	case *ast.BlockStmt:
		o.blockStmt(x, false)
	case *ast.BranchStmt:
		// nop
	case *ast.CaseClause:
		for i := range x.List {
			o.expr(&x.List[i], true)
		}
		o.body(&x.Body)
	case *ast.DeclStmt:
		o.decl(&x.Decl)
	case *ast.DeferStmt:
		o.call(x.Call)
	case *ast.EmptyStmt:
		// nop
	case *ast.ExprStmt:
		o.expr(&x.X, false)
		switch x2 := x.X.(type) {
		case *ast.ParenExpr:
			x.X = x2.X
		}
	case *ast.ForStmt:
		o.stmt(&x.Init)
		o.expr(&x.Cond, true)
		o.stmt(&x.Post)
		o.blockStmt(x.Body, false)
	case *ast.IfStmt:
		o.stmt(&x.Init)
		o.expr(&x.Cond, true)
		o.blockStmt(x.Body, false)
		o.stmt(&x.Else)
		if len(x.Body.List) == 0 {
			x.Cond = o.not(x.Cond)
			switch y := x.Else.(type) {
			case *ast.BlockStmt:
				x.Body.List = y.List
			case nil:
				// ok
			default:
				todo("%T", y)
			}
			x.Else = nil
		}
	case *ast.IncDecStmt:
		o.expr(&x.X, true)
	case *ast.LabeledStmt:
		o.stmt(&x.Stmt)
	case *ast.RangeStmt:
		o.expr(&x.Key, false)
		o.expr(&x.Value, false)
		o.expr(&x.X, true)
		o.blockStmt(x.Body, false)
	case *ast.ReturnStmt:
		for i := range x.Results {
			o.expr(&x.Results[i], true)
			switch y := x.Results[i].(type) {
			case *ast.ParenExpr:
				x.Results[i] = y.X
			}
		}
	case *ast.SwitchStmt:
		o.stmt(&x.Init)
		o.expr(&x.Tag, true)
		o.blockStmt(x.Body, false)
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *opt) expr(n *ast.Expr, use bool) {
	switch x := (*n).(type) {
	case *ast.ArrayType:
		o.expr(&x.Len, false)
		o.expr(&x.Elt, false)
	case *ast.BasicLit:
		// nop
	case *ast.BinaryExpr:
		o.expr(&x.X, true)
		o.expr(&x.Y, true)
		switch x.Op {
		case token.SHL, token.SHR:
			switch rhs := x.Y.(type) {
			case *ast.BasicLit:
				if rhs.Value == "0" {
					*n = x.X
					return
				}
			}
		}
		switch rhs := x.Y.(type) {
		case *ast.BasicLit:
			switch x.Op {
			case token.ADD, token.SUB:
				if rhs.Value == "0" {
					*n = x.X
					return
				}
			case token.MUL, token.QUO:
				if rhs.Value == "1" {
					*n = x.X
					return
				}
			}
		}
		switch lhs := x.X.(type) {
		case *ast.BasicLit:
			switch x.Op {
			case token.ADD, token.SUB:
				if lhs.Value == "0" {
					*n = x.Y
					return
				}
			case token.MUL, token.QUO:
				if lhs.Value == "1" {
					*n = x.Y
					return
				}
			}
		case *ast.CallExpr:
			switch x2 := lhs.Fun.(type) {
			case *ast.Ident:
				if x2.Name == "bool2int" {
					switch x.Op {
					case token.EQL:
						switch rhs := x.Y.(type) {
						case *ast.BasicLit:
							if rhs.Value == "0" {
								*n = o.not(lhs.Args[0])
								o.needBool2int--
								return
							}
						}
					case token.NEQ:
						switch rhs := x.Y.(type) {
						case *ast.BasicLit:
							if rhs.Value == "0" {
								*n = lhs.Args[0]
								o.needBool2int--
								return
							}
						}
					}
				}
			}
		}
	case *ast.CallExpr:
		o.call(x)
	case *ast.FuncLit:
		o.body(&x.Body.List)
	case *ast.Ident:
		if use && o.used != nil {
			o.used[x.Name] = struct{}{}
		}
	case *ast.IndexExpr:
		o.expr(&x.X, true)
		o.expr(&x.Index, true)
	case *ast.ParenExpr:
		o.expr(&x.X, use)
		switch x2 := x.X.(type) {
		case *ast.BasicLit:
			*n = x2
		case *ast.CallExpr:
			*n = x2
		case *ast.Ident:
			*n = x2
		case *ast.ParenExpr:
			*n = x2.X
		case *ast.SelectorExpr:
			switch x2.X.(type) {
			case *ast.Ident:
				*n = x2
			}
		case *ast.StarExpr:
			*n = x2
		case *ast.UnaryExpr:
			switch x2.Op {
			case token.AND, token.MUL:
				switch x2.X.(type) {
				case
					*ast.CallExpr,
					*ast.Ident,
					*ast.SelectorExpr:

					*n = x2
				}
			}
		}
	case *ast.SelectorExpr:
		o.expr(&x.X, true)
	case *ast.StarExpr:
		o.expr(&x.X, use)
		switch x2 := x.X.(type) {
		case *ast.ParenExpr:
			switch x3 := x2.X.(type) {
			case *ast.UnaryExpr:
				switch x3.Op {
				case token.AND, token.MUL:
					*n = x3.X
				}
			}
		case *ast.UnaryExpr:
			switch x2.Op {
			case token.AND:
				*n = x2.X
			}
		}
	case
		*ast.FuncType,
		*ast.StructType:

		// nop
	case *ast.UnaryExpr:
		o.expr(&x.X, use)
	case *ast.CompositeLit:
		for i := range x.Elts {
			o.expr(&x.Elts[i], true)
		}
	case *ast.KeyValueExpr:
		o.expr(&x.Key, true)
		switch x2 := x.Key.(type) {
		case *ast.ParenExpr:
			x.Key = x2.X
		}
		o.expr(&x.Value, true)
		switch x2 := x.Value.(type) {
		case *ast.ParenExpr:
			x.Value = x2.X
		}
	case *ast.InterfaceType:
		// nop
	case nil:
		// nop
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *opt) not(n ast.Expr) ast.Expr {
	switch x := n.(type) {
	case *ast.BinaryExpr:
		switch x.Op {
		case token.LEQ:
			x.Op = token.GTR
			return x
		case token.LSS:
			x.Op = token.GEQ
			return x
		case token.EQL:
			x.Op = token.NEQ
			return x
		case token.NEQ:
			x.Op = token.EQL
			return x
		case token.GEQ:
			x.Op = token.LSS
			return x
		case token.LAND:
			x.X = o.not(x.X)
			x.Op = token.LOR
			x.Y = o.not(x.Y)
			return x
		case token.LOR:
			x.X = o.not(x.X)
			x.Op = token.LAND
			x.Y = o.not(x.Y)
			return x
		case token.GTR:
			x.Op = token.LEQ
			return x
		default:
			todo("%v: %v", o.pos(n), x.Op)
		}
	case *ast.ParenExpr:
		return o.not(x.X)
	default:
		todo("%v: %T %s", o.pos(n), x, o.fn)
	}
	panic("unreachable")
}

func (o *opt) call(n *ast.CallExpr) {
	o.expr(&n.Fun, true)
	for i := range n.Args {
		o.expr(&n.Args[i], true)
		switch x := n.Args[i].(type) {
		case *ast.ParenExpr:
			n.Args[i] = x.X
		}
	}
}

func newNOpt() *nopt { return &nopt{} }

func (o *nopt) Write(b []byte) (int, error) {
	if o.write {
		return o.out.Write(b)
	}

	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		o.write = true
		n, err := o.out.Write(b[i+1:])
		return n + i, err
	}

	return len(b), nil
}

func (o *nopt) pos(n ast.Node) token.Position {
	if n == nil {
		return token.Position{}
	}

	return o.fset.Position(n.Pos())
}

var optTrap = []byte("uintptr(unsafe.Pointer(&")

func (o *nopt) do(out io.Writer, in io.Reader, fn string) error { //TODO reuse the nopt object
	o.fn = fn
	o.fset = token.NewFileSet()
	in = io.MultiReader(bytes.NewBufferString(fmt.Sprintf("package p // %s\n", fn)), in)
	ast, err := parser.ParseFile(o.fset, "", in, parser.ParseComments)
	if err != nil {
		return err
	}

	o.file(ast)
	if err := format.Node(o, o.fset, ast); err != nil {
		return err
	}

	b := o.out.Bytes()
	if i := bytes.Index(b, optTrap); i >= 0 {
		a := bytes.LastIndex(b[:i], []byte{'\n'})
		if a < 0 {
			a = 0
		}
		z := bytes.Index(b[i:], []byte{'\n'})
		if z < 0 {
			z = 0
		}
		todo("invalid unsafe.Pointer to uintptr conversion\n%s", bytes.TrimSpace(b[a:i+z]))
	}

	b = bytes.Replace(b, []byte("\n\n}"), []byte("\n}"), -1)
	b = bytes.Replace(b, []byte("\n\t;\n"), []byte("\n"), -1)
	b = bytes.Replace(b, []byte("{\n\n"), []byte("{\n"), -1)
	b = bytes.Replace(b, []byte(":\n\n"), []byte(":\n"), -1)
	b = bytes.Replace(b, []byte("\n\n\tdefer"), []byte("\n\tdefer"), -1)
	b = bytes.Replace(b, []byte("\n\n\t\t"), []byte("\n\t\t"), -1)
	b = bytes.Replace(b, []byte("\n\n\t)"), []byte("\n\t)"), -1)
	if traceOpt {
		os.Stderr.Write(b)
	}
	_, err = out.Write(b)
	return err
}

func (o *nopt) file(n *ast.File) {
	for i := range n.Decls {
		o.decl(&n.Decls[i])
	}
}

func (o *nopt) decl(n *ast.Decl) {
	switch x := (*n).(type) {
	case *ast.FuncDecl:
		o.used = map[string]struct{}{}
		o.blockStmt(x.Body, true)
	case *ast.GenDecl:
		for i := range x.Specs {
			o.spec(&x.Specs[i])
		}
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *nopt) spec(n *ast.Spec) {
	switch x := (*n).(type) {
	case *ast.TypeSpec:
		// nop
	case *ast.ValueSpec:
		use := x.Names[0].Name != "_"
		for i := range x.Values {
			o.expr(&x.Values[i], use)
			switch x2 := x.Values[i].(type) {
			case *ast.ParenExpr:
				x.Values[i] = x2.X
			}
		}
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *nopt) blockStmt(n *ast.BlockStmt, outermost bool) {
	if n == nil {
		return
	}

	o.body(&n.List)
	if !outermost {
		return
	}

	w := 0
	for _, v := range n.List {
		switch x := v.(type) {
		case *ast.AssignStmt:
			if x2, ok := x.Lhs[0].(*ast.Ident); ok && x2.Name == "_" {
				if _, used := o.used[x.Rhs[0].(*ast.Ident).Name]; used {
					continue
				}
			}
		case *ast.DeclStmt:
			switch x2 := x.Decl.(type) {
			case *ast.GenDecl:
				w := 0
				for _, v := range x2.Specs {
					if x3, ok := v.(*ast.ValueSpec); ok && x3.Names[0].Name == "_" {
						if _, used := o.used[x3.Values[0].(*ast.Ident).Name]; used {
							continue
						}
					}

					x2.Specs[w] = v
					w++
				}
				x2.Specs = x2.Specs[:w]
			}
		}
		n.List[w] = v
		w++
	}
	n.List = n.List[:w]
}

func (o *nopt) body(l0 *[]ast.Stmt) {
	l := *l0
	w := 0
	for i := range l {
		o.stmt(&l[i])
		switch x := l[i].(type) {
		case *ast.EmptyStmt:
			// nop
		default:
			l[w] = x
			w++
		}
	}
	*l0 = l[:w]
}

func (o *nopt) stmt(n *ast.Stmt) {
	switch x := (*n).(type) {
	case nil:
		// nop
	case *ast.AssignStmt:
		for i := range x.Lhs {
			o.expr(&x.Lhs[i], false)
			switch x2 := x.Lhs[i].(type) {
			case *ast.ParenExpr:
				x.Lhs[i] = x2.X
			}
		}
		use := true
		if x, ok := x.Lhs[0].(*ast.Ident); ok && x.Name == "_" {
			use = false
		}
		for i := range x.Rhs {
			o.expr(&x.Rhs[i], use)
			switch x2 := x.Rhs[i].(type) {
			case *ast.ParenExpr:
				x.Rhs[i] = x2.X
			}
		}
	case *ast.BlockStmt:
		o.blockStmt(x, false)
	case *ast.BranchStmt:
		// nop
	case *ast.CaseClause:
		for i := range x.List {
			o.expr(&x.List[i], true)
		}
		o.body(&x.Body)
	case *ast.DeclStmt:
		o.decl(&x.Decl)
	case *ast.DeferStmt:
		o.call(x.Call)
	case *ast.EmptyStmt:
		// nop
	case *ast.ExprStmt:
		o.expr(&x.X, false)
		switch x2 := x.X.(type) {
		case *ast.ParenExpr:
			x.X = x2.X
		}
	case *ast.ForStmt:
		o.stmt(&x.Init)
		o.expr(&x.Cond, true)
		o.stmt(&x.Post)
		o.blockStmt(x.Body, false)
	case *ast.IfStmt:
		o.stmt(&x.Init)
		o.expr(&x.Cond, true)
		o.blockStmt(x.Body, false)
		o.stmt(&x.Else)
		if len(x.Body.List) == 0 {
			// Turn
			//	if cond {} else { stmtList }
			// into
			//	if !cond { stmtList }
			x.Cond = o.not(x.Cond)
			switch y := x.Else.(type) {
			case *ast.BlockStmt:
				x.Body.List = y.List
			case nil:
				// ok
			default:
				todo("%T", y)
			}
			x.Else = nil
		}
	case *ast.IncDecStmt:
		o.expr(&x.X, true)
	case *ast.LabeledStmt:
		o.stmt(&x.Stmt)
	case *ast.RangeStmt:
		o.expr(&x.Key, false)
		o.expr(&x.Value, false)
		o.expr(&x.X, true)
		o.blockStmt(x.Body, false)
	case *ast.ReturnStmt:
		for i := range x.Results {
			o.expr(&x.Results[i], true)
			switch y := x.Results[i].(type) {
			case *ast.ParenExpr:
				x.Results[i] = y.X
			}
		}
	case *ast.SwitchStmt:
		o.stmt(&x.Init)
		o.expr(&x.Tag, true)
		o.blockStmt(x.Body, false)
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *nopt) expr(n *ast.Expr, use bool) {
	switch x := (*n).(type) {
	case *ast.ArrayType:
		o.expr(&x.Len, false)
		o.expr(&x.Elt, false)
	case *ast.BasicLit:
		// nop
	case *ast.BinaryExpr:
		o.expr(&x.X, true)
		o.expr(&x.Y, true)
		switch x.Op {
		case token.SHL, token.SHR:
			switch rhs := x.Y.(type) {
			case *ast.BasicLit:
				if rhs.Value == "0" {
					*n = x.X
					return
				}
			}
		}
		switch rhs := x.Y.(type) {
		case *ast.BasicLit:
			switch x.Op {
			case token.ADD, token.SUB:
				if rhs.Value == "0" {
					*n = x.X
					return
				}
			case token.MUL, token.QUO:
				if rhs.Value == "1" {
					*n = x.X
					return
				}
			}
		}
		switch lhs := x.X.(type) {
		case *ast.BasicLit:
			switch x.Op {
			case token.ADD, token.SUB:
				if lhs.Value == "0" {
					*n = x.Y
					return
				}
			case token.MUL, token.QUO:
				if lhs.Value == "1" {
					*n = x.Y
					return
				}
			}
		case *ast.CallExpr:
			switch x2 := lhs.Fun.(type) {
			case *ast.Ident:
				if x2.Name == "bool2int" {
					switch x.Op {
					case token.EQL:
						switch rhs := x.Y.(type) {
						case *ast.BasicLit:
							if rhs.Value == "0" {
								*n = o.not(lhs.Args[0])
								return
							}
						}
					case token.NEQ:
						switch rhs := x.Y.(type) {
						case *ast.BasicLit:
							if rhs.Value == "0" {
								*n = lhs.Args[0]
								return
							}
						}
					}
				}
			}
		}
	case *ast.CallExpr:
		o.call(x)
	case *ast.FuncLit:
		o.body(&x.Body.List)
	case *ast.Ident:
		if use && o.used != nil {
			o.used[x.Name] = struct{}{}
		}
	case *ast.IndexExpr:
		o.expr(&x.X, true)
		o.expr(&x.Index, true)
	case *ast.ParenExpr:
		o.expr(&x.X, use)
		switch x2 := x.X.(type) {
		case *ast.BasicLit:
			*n = x2
		case *ast.CallExpr:
			*n = x2
		case *ast.Ident:
			*n = x2
		case *ast.ParenExpr:
			*n = x2.X
		case *ast.SelectorExpr:
			switch x2.X.(type) {
			case *ast.Ident:
				*n = x2
			}
		case *ast.StarExpr:
			*n = x2
		case *ast.UnaryExpr:
			switch x2.Op {
			case token.AND:
				switch x2.X.(type) {
				case
					*ast.Ident,
					*ast.SelectorExpr:

					*n = x2
				}
			}
		}
	case *ast.SelectorExpr:
		o.expr(&x.X, true)
	case *ast.StarExpr:
		o.expr(&x.X, use)
		switch x2 := x.X.(type) {
		case *ast.ParenExpr:
			switch x3 := x2.X.(type) {
			case *ast.UnaryExpr:
				switch x3.Op {
				case token.AND:
					*n = x3.X
				}
			}
		case *ast.UnaryExpr:
			switch x2.Op {
			case token.AND:
				*n = x2.X
			}
		}
	case
		*ast.FuncType,
		*ast.StructType:

		// nop
	case *ast.UnaryExpr:
		o.expr(&x.X, use)
		switch x.Op {
		case token.AND:
			switch x2 := x.X.(type) {
			case *ast.StarExpr:
				*n = x2.X
			}
		}
	case *ast.CompositeLit:
		for i := range x.Elts {
			o.expr(&x.Elts[i], true)
		}
	case *ast.KeyValueExpr:
		o.expr(&x.Key, true)
		switch x2 := x.Key.(type) {
		case *ast.ParenExpr:
			x.Key = x2.X
		}
		o.expr(&x.Value, true)
		switch x2 := x.Value.(type) {
		case *ast.ParenExpr:
			x.Value = x2.X
		}
	case *ast.InterfaceType:
		// nop
	case *ast.SliceExpr:
		o.expr(&x.X, use)
		o.expr(&x.Low, use)
		o.expr(&x.High, use)
		o.expr(&x.Max, use)
	case *ast.TypeAssertExpr:
		o.expr(&x.X, use)
	case nil:
		// nop
	default:
		todo("%v: %T", o.pos(x), x)
	}
}

func (o *nopt) not(n ast.Expr) ast.Expr {
	switch x := n.(type) {
	case *ast.BinaryExpr:
		switch x.Op {
		case
			token.LEQ,
			token.LSS,
			token.EQL,
			token.NEQ,
			token.GEQ,
			token.LAND,
			token.LOR,
			token.GTR:
			return &ast.UnaryExpr{Op: token.NOT, X: &ast.ParenExpr{X: x}}
		default:
			todo("%v: %v", o.pos(n), x.Op)
		}
	case *ast.ParenExpr:
		return &ast.UnaryExpr{Op: token.NOT, X: x.X}
	case *ast.UnaryExpr:
		switch x.Op {
		case token.NOT:
			return x.X
		default:
			todo("%v: %T %s", o.pos(n), x, o.fn)
		}
	default:
		todo("%v: %T %s", o.pos(n), x, o.fn)
	}
	panic("unreachable")
}

func (o *nopt) call(n *ast.CallExpr) {
	o.expr(&n.Fun, true)
	for i := range n.Args {
		o.expr(&n.Args[i], true)
		switch x := n.Args[i].(type) {
		case *ast.ParenExpr:
			n.Args[i] = x.X
		}
	}
}
