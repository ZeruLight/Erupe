// Copyright 2019 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast2

import (
	"bytes"
	"strconv"
	"testing"

	"modernc.org/cc/v3"
)

// eqExpr implements a limited deep equality for expressions.
// Namely, it ignores types and identifier matching (only matches by the name).
func eqExpr(e1, e2 Expr) bool {
	switch e1 := e1.(type) {
	case IdentExpr:
		e2, ok := e2.(IdentExpr)
		return ok && e1.Ident.Name == e2.Ident.Name
	case *Literal:
		e2, ok := e2.(*Literal)
		return ok && e1.kind == e2.kind && e1.value == e2.value
	case *ParenExpr:
		e2, ok := e2.(*ParenExpr)
		return ok && eqExpr(e1.X, e2.X)
	case *UnaryExpr:
		e2, ok := e2.(*UnaryExpr)
		return ok && e1.Op == e2.Op && eqExpr(e1.X, e2.X)
	case *IncDecExpr:
		e2, ok := e2.(*IncDecExpr)
		return ok && e1.Op == e2.Op && eqExpr(e1.X, e2.X)
	case *BinaryExpr:
		e2, ok := e2.(*BinaryExpr)
		return ok && e1.Op == e2.Op && eqExpr(e1.X, e2.X) && eqExpr(e1.Y, e2.Y)
	case *AssignExpr:
		e2, ok := e2.(*AssignExpr)
		return ok && e1.Op == e2.Op && eqExpr(e1.Left, e2.Left) && eqExpr(e1.Right, e2.Right)
	case *IndexExpr:
		e2, ok := e2.(*IndexExpr)
		return ok && eqExpr(e1.X, e2.X) && eqExpr(e1.Ind, e2.Ind)
	case *SelectExpr:
		e2, ok := e2.(*SelectExpr)
		return ok && e1.Ptr == e2.Ptr && eqExpr(e1.X, e2.X) && eqExpr(IdentExpr{e1.Sel}, IdentExpr{e2.Sel})
	case *CondExpr:
		e2, ok := e2.(*CondExpr)
		return ok && eqExpr(e1.Cond, e2.Cond) && eqExpr(e1.Then, e2.Then) && eqExpr(e1.Else, e2.Else)
	case *CallExpr:
		e2, ok := e2.(*CallExpr)
		if !ok || len(e1.Args) != len(e2.Args) || !eqExpr(e1.Func, e2.Func) {
			return false
		}
		for i := range e1.Args {
			if !eqExpr(e1.Args[i], e2.Args[i]) {
				return false
			}
		}
		return true
	case CommaExpr:
		e2, ok := e2.(CommaExpr)
		if !ok || len(e1) != len(e2) {
			return false
		}
		for i := range e1 {
			if !eqExpr(e1[i], e2[i]) {
				return false
			}
		}
		return true
	}
	panic(e1)
}

func ident(s string) *Ident {
	return &Ident{Name: s}
}

func identExpr(s string) IdentExpr {
	return IdentExpr{ident(s)}
}

func lit(kind LiteralKind, v string) Expr {
	return &Literal{kind: kind, value: v}
}

func intLit(v int) Expr {
	return lit(LiteralInt, strconv.Itoa(v))
}

func newTestABI(t testing.TB) cc.ABI {
	abi, err := cc.NewABIFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	return abi
}

func TestExpr(t *testing.T) {
	joinE := [2]string{`int func() { `, `; }`}
	joinX := [2]string{`int func(int x) { `, `; }`}
	joinXp := [2]string{`int func(int* x) { `, `; }`}
	joinXY := [2]string{`int func(int x, int y) { `, `; }`}
	cfg := &cc.Config{ABI: newTestABI(t)}
	for _, v := range []struct {
		name string
		src  string
		join [2]string
		exp  Expr
		cstr string
	}{
		{
			name: "int lit",
			src:  `42`,
			join: joinE,
			exp:  lit(LiteralInt, `42`),
		},
		{
			name: "int lit hex",
			src:  `0x42`,
			join: joinE,
			exp:  lit(LiteralInt, `0x42`),
		},
		{
			name: "int lit suffix",
			src:  `42u`,
			join: joinE,
			exp:  lit(LiteralInt, `42u`),
		},
		{
			name: "int lit pos",
			src:  `+42`,
			join: joinE,
			exp:  &UnaryExpr{Op: UnaryPlus, X: intLit(42)},
		},
		{
			name: "int lit neg",
			src:  `-42`,
			join: joinE,
			exp:  &UnaryExpr{Op: UnaryMinus, X: intLit(42)},
		},
		{
			name: "float lit",
			src:  `1.0`,
			join: joinE,
			exp:  lit(LiteralFloat, `1.0`),
		},
		{
			name: "float lit exp",
			src:  `1e5`,
			join: joinE,
			exp:  lit(LiteralFloat, `1e5`),
		},
		{
			name: "float lit pos",
			src:  `+1.0`,
			join: joinE,
			exp:  &UnaryExpr{Op: UnaryPlus, X: lit(LiteralFloat, `1.0`)},
		},
		{
			name: "float lit neg",
			src:  `-1.0`,
			join: joinE,
			exp:  &UnaryExpr{Op: UnaryMinus, X: lit(LiteralFloat, `1.0`)},
		},
		{
			name: "char lit",
			src:  `'a'`,
			join: joinE,
			exp:  lit(LiteralChar, `a`),
		},
		{
			name: "wide char lit",
			src:  `L'a'`,
			join: joinE,
			exp:  lit(LiteralWChar, `a`),
		},
		{
			name: "string lit",
			src:  `"a"`,
			join: joinE,
			exp:  lit(LiteralString, `a`),
		},
		{
			name: "wide string lit",
			src:  `L"a"`,
			join: joinE,
			exp:  lit(LiteralWString, `a`),
		},
		{
			name: "parentheses",
			src:  `(1)`,
			join: joinE,
			exp:  &ParenExpr{X: intLit(1)},
		},
		{
			name: "comma",
			src:  `1, 2, 3`,
			join: joinE,
			exp:  CommaExpr{intLit(1), intLit(2), intLit(3)},
		},
		{
			name: "ident",
			src:  `x`,
			join: joinX,
			exp:  identExpr("x"),
		},
		{
			name: "unary plus",
			src:  `+x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryPlus, X: identExpr("x")},
		},
		{
			name: "unary minus",
			src:  `-x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryMinus, X: identExpr("x")},
		},
		{
			name: "unary addr",
			src:  `&x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryAddr, X: identExpr("x")},
		},
		{
			name: "unary deref",
			src:  `*x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryDeref, X: identExpr("x")},
		},
		{
			name: "unary not",
			src:  `!x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryNot, X: identExpr("x")},
		},
		{
			name: "unary invert",
			src:  `~x`,
			join: joinX,
			exp:  &UnaryExpr{Op: UnaryInvert, X: identExpr("x")},
		},
		{
			name: "binary add",
			src:  `x + y`,
			join: joinXY,
			exp:  &BinaryExpr{X: identExpr("x"), Op: BinaryAdd, Y: identExpr("y")},
		},
		{
			name: "binary order 1",
			src:  `x * 2 + y`,
			cstr: `x*2 + y`,
			join: joinXY,
			exp:  &BinaryExpr{X: &BinaryExpr{X: identExpr("x"), Op: BinaryMul, Y: intLit(2)}, Op: BinaryAdd, Y: identExpr("y")},
		},
		{
			name: "binary order 2",
			src:  `x + y * 2`,
			cstr: `x + y*2`,
			join: joinXY,
			exp:  &BinaryExpr{X: identExpr("x"), Op: BinaryAdd, Y: &BinaryExpr{X: identExpr("y"), Op: BinaryMul, Y: intLit(2)}},
		},
		{
			name: "assign",
			src:  `x = 1`,
			join: joinX,
			exp:  &AssignExpr{Left: identExpr("x"), Op: BinaryNone, Right: intLit(1)},
		},
		{
			name: "assign op",
			src:  `x += 1`,
			join: joinX,
			exp:  &AssignExpr{Left: identExpr("x"), Op: BinaryAdd, Right: intLit(1)},
		},
		{
			name: "assign chain",
			src:  `y = x = 1`,
			join: joinXY,
			exp:  &AssignExpr{Left: identExpr("y"), Right: &AssignExpr{Left: identExpr("x"), Right: intLit(1)}},
		},
		{
			name: "inc post",
			src:  `x++`,
			join: joinX,
			exp:  &IncDecExpr{X: identExpr("x"), Op: IncPost},
		},
		{
			name: "inc pre",
			src:  `++x`,
			join: joinX,
			exp:  &IncDecExpr{X: identExpr("x"), Op: IncPre},
		},
		{
			name: "inc add 1",
			src:  `x+++y`,
			cstr: `x++ + y`,
			join: joinXY,
			exp:  &BinaryExpr{X: &IncDecExpr{X: identExpr("x"), Op: IncPost}, Op: BinaryAdd, Y: identExpr("y")},
		},
		{
			name: "inc add 2",
			src:  `++x+y`,
			cstr: `++x + y`,
			join: joinXY,
			exp:  &BinaryExpr{X: &IncDecExpr{X: identExpr("x"), Op: IncPre}, Op: BinaryAdd, Y: identExpr("y")},
		},
		{
			name: "inc add 3",
			src:  `x+y++`,
			cstr: `x + y++`,
			join: joinXY,
			exp:  &BinaryExpr{X: identExpr("x"), Op: BinaryAdd, Y: &IncDecExpr{X: identExpr("y"), Op: IncPost}},
		},
		{
			name: "index",
			src:  `x[1]`,
			join: joinXp,
			exp:  &IndexExpr{X: identExpr("x"), Ind: intLit(1)},
		},
		{
			name: "index chain",
			src:  `x[1][2]`,
			join: joinXp,
			exp:  &IndexExpr{X: &IndexExpr{X: identExpr("x"), Ind: intLit(1)}, Ind: intLit(2)},
		},
		{
			name: "select",
			src:  `x.y`,
			join: [2]string{`struct s { int y; }; int f(struct s x) { `, `; }`},
			exp:  &SelectExpr{X: identExpr("x"), Sel: ident("y"), Ptr: false},
		},
		{
			name: "select ptr",
			src:  `x->y`,
			join: [2]string{`struct s { int y; }; int f(struct s* x) { `, `; }`},
			exp:  &SelectExpr{X: identExpr("x"), Sel: ident("y"), Ptr: true},
		},
		{
			name: "select chain 1",
			src:  `x.y.z`,
			join: [2]string{`struct s1 { int z; }; struct s2 { struct s1 y; }; int f(struct s2 x) { `, `; }`},
			exp:  &SelectExpr{X: &SelectExpr{X: identExpr("x"), Sel: ident("y")}, Sel: ident("z")},
		},
		{
			name: "select chain 2",
			src:  `x->y->z`,
			join: [2]string{`struct s1 { int z; }; struct s2 { struct s1* y; }; int f(struct s2* x) { `, `; }`},
			exp:  &SelectExpr{X: &SelectExpr{X: identExpr("x"), Sel: ident("y"), Ptr: true}, Sel: ident("z"), Ptr: true},
		},
		{
			name: "select chain 3",
			src:  `x.y->z`,
			join: [2]string{`struct s1 { int z; }; struct s2 { struct s1 y; }; int f(struct s2* x) { `, `; }`},
			exp:  &SelectExpr{X: &SelectExpr{X: identExpr("x"), Sel: ident("y")}, Sel: ident("z"), Ptr: true},
		},
		{
			name: "select chain 4",
			src:  `x->y.z`,
			join: [2]string{`struct s1 { int z; }; struct s2 { struct s1* y; }; int f(struct s2 x) { `, `; }`},
			exp:  &SelectExpr{X: &SelectExpr{X: identExpr("x"), Sel: ident("y"), Ptr: true}, Sel: ident("z")},
		},
		{
			name: "call",
			src:  `x(1, 2)`,
			join: [2]string{`void x(int a1, int a2); int f() { `, `; }`},
			exp:  &CallExpr{Func: identExpr("x"), Args: []Expr{intLit(1), intLit(2)}},
		},
		{
			name: "cond",
			src:  `x ? 1 : 2`,
			join: joinX,
			exp:  &CondExpr{Cond: identExpr("x"), Then: intLit(1), Else: intLit(2)},
		},
		{
			name: "cond chain",
			src:  `x ? 1 : y ? 2 : 3`,
			join: joinXY,
			exp:  &CondExpr{Cond: identExpr("x"), Then: intLit(1), Else: &CondExpr{Cond: identExpr("y"), Then: intLit(2), Else: intLit(3)}},
		},
	} {
		t.Run(v.name, func(t *testing.T) {
			ast, err := cc.Parse(cfg, nil, nil, []cc.Source{
				{Name: "test", Value: v.join[0] + v.src + v.join[1]},
			})
			if err != nil {
				t.Fatal(err)
				return
			}
			tu := ast.TranslationUnit
			for ; tu.TranslationUnit != nil; tu = tu.TranslationUnit {
			}
			cce := tu.
				ExternalDeclaration.
				FunctionDefinition.
				CompoundStatement.
				BlockItemList.
				BlockItemList.
				BlockItem.
				Statement.
				ExpressionStatement.
				Expression
			e := NewExprFrom(cce)
			if !eqExpr(v.exp, e) {
				t.Fatalf("unexpected expression: %#v", e)
			}
			buf := bytes.NewBuffer(nil)
			if err := PrintExpr(buf, e); err != nil {
				t.Fatal(err)
			}
			exp := v.cstr
			if exp == "" {
				exp = v.src
			}
			if s := buf.String(); s != exp {
				t.Fatalf("unexpected C code printed:\nexp: %q\nvs\ngot: %q\n", exp, s)
			}
		})
	}
}
