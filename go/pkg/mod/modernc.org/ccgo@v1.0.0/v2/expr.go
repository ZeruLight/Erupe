// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v2"

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"modernc.org/cc/v2"
	"modernc.org/ir"
	"modernc.org/mathutil"
)

func (g *gen) exprListOpt(n *cc.ExprListOpt, void bool) {
	if n == nil {
		return
	}

	g.exprList(n.ExprList, void)
}

func (g *ngen) exprListOpt(n *cc.ExprListOpt, void bool) {
	if n == nil {
		return
	}

	g.exprList(n.ExprList, void)
}

func (g *gen) exprList(n *cc.ExprList, void bool) {
	switch l := g.pexprList(n); {
	case void:
		for _, v := range l {
			g.void(v)
			g.w(";")
		}
	default:
		switch {
		case len(l) == 1:
			g.value(l[0], false)
		default:
			g.w("func() %v {", g.typ(n.Operand.Type))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.value(l[len(l)-1], false)
			g.w("}()")
		}
	}
}

func (g *ngen) exprList(n *cc.ExprList, void bool) {
	switch l := g.pexprList(n); {
	case void:
		for _, v := range l {
			g.void(v)
			g.w(";")
		}
	default:
		switch {
		case len(l) == 1:
			g.value(l[0], false)
		default:
			g.w("func() %v {", g.typ(n.Operand.Type))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.value(l[len(l)-1], false)
			g.w("}()")
		}
	}
}

func (g *gen) exprList2(n *cc.ExprList, t cc.Type) {
	switch l := g.pexprList(n); {
	case len(l) == 1:
		g.convert(l[0], t)
	default:
		g.w("func() %v {", g.typ(t))
		for _, v := range l[:len(l)-1] {
			g.void(v)
			g.w(";")
		}
		g.w("return ")
		g.convert(l[len(l)-1], t)
		g.w("}()")
	}
}

func (g *ngen) exprList2(n *cc.ExprList, t cc.Type) {
	switch l := g.pexprList(n); {
	case len(l) == 1:
		g.convert(l[0], t)
	default:
		g.w("func() %v {", g.typ(t))
		for _, v := range l[:len(l)-1] {
			g.void(v)
			g.w(";")
		}
		g.w("return ")
		g.convert(l[len(l)-1], t)
		g.w("}()")
	}
}

func (g *gen) void(n *cc.Expr) {
	if n.Case == cc.ExprCast && n.Expr.Case == cc.ExprIdent && !isVaList(n.Expr.Operand.Type) {
		g.enqueue(n.Expr.Declarator)
		return
	}

	if g.voidCanIgnore(n) {
		return
	}

	switch n.Case {
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		g.value(n, false)
	case cc.ExprAssign: // Expr '=' Expr
		if n.Expr.Equals(n.Expr2) {
			return
		}

		op := n.Expr.Operand
		if op.Bits() != 0 {
			g.assignmentValue(n)
			return
		}

		if isVaList(n.Expr.Operand.Type) {
			panic("TODO")
			// switch rhs := n.Expr2.Operand; {
			// case isVaList(rhs.Type): // va_copy
			// 	g.w("{ x := *")
			// 	g.value(n.Expr2, false)
			// 	g.w("; ")
			// 	g.w("*")
			// 	g.lvalue(n.Expr)
			// 	g.w(" = &x }")
			// 	return
			// case n.Expr2.Declarator != nil && n.Expr2.Declarator.Name() == idVaStart:
			// 	g.w("{ x := ap; *")
			// 	g.lvalue(n.Expr)
			// 	g.w(" = &x }")
			// 	return
			// case n.Expr2.Declarator != nil && n.Expr2.Declarator.Name() == idVaEnd:
			// 	g.w("*")
			// 	g.lvalue(n.Expr)
			// 	g.w(" = nil")
			// 	return
			// }
			// panic(fmt.Errorf("%v: %v = %v", g.position0(n), n.Expr.Operand, n.Expr2.Operand))
		}

		g.w("*")
		g.lvalue(n.Expr)
		g.w(" = ")
		if isVaList(n.Expr.Operand.Type) && n.Expr2.Case == cc.ExprCast {
			if ec := n.Expr2; g.voidCanIgnore(ec) {
				switch op := ec.Expr; {
				case op.IsNonZero() && g.voidCanIgnore(op):
					g.w("&%s", ap)
					return
				case op.IsZero() && g.voidCanIgnore(op):
					g.w("nil")
					return
				}
			}
			panic(g.position0(n))
		}

		g.convert(n.Expr2, n.Expr.Operand.Type)
	case
		cc.ExprPostInc, // Expr "++"
		cc.ExprPreInc:  // "++" Expr

		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			switch sz := g.model.Sizeof(x.Item); {
			case sz == 1:
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")++")
			default:
				g.value(n.Expr, false)
				g.w(" += %d", sz)
			}
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")++")
				return
			}
			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case
		cc.ExprPostDec, // Expr "--"
		cc.ExprPreDec:  // "--" Expr

		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			switch sz := g.model.Sizeof(x.Item); {
			case sz == 1:
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")--")
			default:
				g.value(n.Expr, false)
				g.w(" -= %d", sz)
			}
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")--")
				return
			}
			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprAddAssign: // Expr "+=" Expr
		switch {
		case cc.UnderlyingType(n.Expr.Operand.Type).Kind() == cc.Ptr:
			g.w(" *(")
			g.lvalue(n.Expr)
			g.w(") += %d*uintptr(", g.model.Sizeof(n.Expr.Operand.Type.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		default:
			g.voidArithmeticAsop(n)
		}
	case cc.ExprSubAssign: // Expr "-=" Expr
		switch {
		case n.Expr.Operand.Type.Kind() == cc.Ptr:
			g.w(" *(")
			g.lvalue(n.Expr)
			g.w(") -= %d*uintptr(", g.model.Sizeof(n.Expr.Operand.Type.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		default:
			g.voidArithmeticAsop(n)
		}
	case
		cc.ExprAndAssign, // Expr "&=" Expr
		cc.ExprDivAssign, // Expr "/=" Expr
		cc.ExprLshAssign, // Expr "<<=" Expr
		cc.ExprModAssign, // Expr "%=" Expr
		cc.ExprMulAssign, // Expr "*=" Expr
		cc.ExprOrAssign,  // Expr "|=" Expr
		cc.ExprRshAssign, // Expr ">>=" Expr
		cc.ExprXorAssign: // Expr "^=" Expr

		g.voidArithmeticAsop(n)
	case cc.ExprPExprList: // '(' ExprList ')'
		for l := n.ExprList; l != nil; l = l.ExprList {
			g.void(l.Expr)
			g.w(";")
		}
	case cc.ExprCast: // '(' TypeName ')' Expr
		if isVaList(n.Expr.Operand.Type) { //TODO- ?
			g.w("%sVA%s(", crt, g.typ(cc.UnderlyingType(n.TypeName.Type)))
			g.value(n.Expr, false)
			g.w(")")
			return
		}

		g.void(n.Expr)
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		switch {
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			g.void(n.Expr2)
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			g.exprList(n.ExprList, true)
		default:
			// if expr != 0 {
			//	exprList
			// } else {
			//	expr2
			// }
			g.w("if ")
			g.value(n.Expr, false)
			g.w(" != 0 {")
			g.exprList(n.ExprList, true)
			g.w("} else {")
			g.void(n.Expr2)
			g.w("}")
		}
	case cc.ExprLAnd: // Expr "&&" Expr
		if n.Expr.IsZero() && g.voidCanIgnore(n.Expr) {
			return
		}

		g.w("if ")
		g.value(n.Expr, false)
		g.w(" != 0 {")
		g.void(n.Expr2)
		g.w("}")
	case cc.ExprLOr: // Expr "||" Expr
		if n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr) {
			return
		}

		g.w("if ")
		g.value(n.Expr, false)
		g.w(" == 0 {")
		g.void(n.Expr2)
		g.w("}")
	case cc.ExprIndex: // Expr '[' ExprList ']'
		g.void(n.Expr)
		if !g.voidCanIgnoreExprList(n.ExprList) {
			g.w("\n")
		}
		g.exprList(n.ExprList, true)
	case // Unary
		cc.ExprAddrof,     // '&' Expr
		cc.ExprCpl,        // '~' Expr
		cc.ExprDeref,      // '*' Expr
		cc.ExprNot,        // '!' Expr
		cc.ExprUnaryMinus, // '-' Expr
		cc.ExprUnaryPlus:  // '+' Expr

		g.void(n.Expr)
	case // Binary
		cc.ExprAdd, // Expr '+' Expr
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprEq,  // Expr "==" Expr
		cc.ExprGe,  // Expr ">=" Expr
		cc.ExprGt,  // Expr ">" Expr
		cc.ExprLe,  // Expr "<=" Expr
		cc.ExprLsh, // Expr "<<" Expr
		cc.ExprLt,  // Expr '<' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprNe,  // Expr "!=" Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprRsh, // Expr ">>" Expr
		cc.ExprSub, // Expr '-' Expr
		cc.ExprXor: // Expr '^' Expr

		g.void(n.Expr)
		if !g.voidCanIgnore(n.Expr2) {
			g.w(";")
		}
		g.void(n.Expr2)
	default:
		todo("", g.position0(n), n.Case, n.Operand) // void
	}
}

func (g *ngen) void(n *cc.Expr) {
	if n.Case == cc.ExprCast && n.Expr.Case == cc.ExprIdent && !isVaList(n.Expr.Operand.Type) {
		g.enqueue(n.Expr.Declarator)
		return
	}

	if g.voidCanIgnore(n) {
		return
	}

	switch n.Case {
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		if e := n.Expr; e.Case == cc.ExprIdent && (e.Token.Val == idGo || e.Token.Val == idGo2) {
			g.w("%s", dict.S(int(n.ArgumentExprListOpt.ArgumentExprList.Expr.Operand.Value.(*ir.StringValue).StringID)))
			return
		}

		var t0 cc.Type
		if !isFnPtr(n.Expr.Operand.Type, &t0) {
			todo("%v: %v", g.position(n), n.Expr.Operand.Type)
		}

		t := cc.UnderlyingType(t0).(*cc.FunctionType)
		var args []*cc.Expr
		if o := n.ArgumentExprListOpt; o != nil {
			for l := o.ArgumentExprList; l != nil; l = l.ArgumentExprList {
				args = append(args, l.Expr)
			}
		}
		params := t.Params
		var voidParams bool
		if voidParams = len(params) == 1 && params[0].Kind() == cc.Void; voidParams {
			params = nil
		}
		switch {
		case voidParams && len(args) != 0:
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		case len(args) < len(params):
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		case len(args) == len(params):
			g.value(n, false)

		// len(args) > len(params)
		case t.Variadic:
			g.value(n, false)
		case len(params) == 0:
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		default:
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		}
	case cc.ExprAssign: // Expr '=' Expr
		if n.Expr.Equals(n.Expr2) {
			return
		}

		op := n.Expr.Operand
		if op.Bits() != 0 {
			g.assignmentValue(n)
			return
		}

		g.w("*")
		g.lvalue(n.Expr)
		g.w(" = ")
		g.convert(n.Expr2, n.Expr.Operand.Type)
	case
		cc.ExprPostInc, // Expr "++"
		cc.ExprPreInc:  // "++" Expr

		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			switch sz := g.model.Sizeof(x.Item); {
			case sz == 1:
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")++")
			default:
				g.value(n.Expr, false)
				g.w(" += %d", sz)
			}
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")++")
				return
			}
			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case
		cc.ExprPostDec, // Expr "--"
		cc.ExprPreDec:  // "--" Expr

		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			switch sz := g.model.Sizeof(x.Item); {
			case sz == 1:
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")--")
			default:
				g.value(n.Expr, false)
				g.w(" -= %d", sz)
			}
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w(" *(")
				g.lvalue(n.Expr)
				g.w(")--")
				return
			}
			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprAddAssign: // Expr "+=" Expr
		switch {
		case cc.UnderlyingType(n.Expr.Operand.Type).Kind() == cc.Ptr:
			g.w(" *(")
			g.lvalue(n.Expr)
			g.w(") += %d*uintptr(", g.model.Sizeof(n.Expr.Operand.Type.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		default:
			g.voidArithmeticAsop(n)
		}
	case cc.ExprSubAssign: // Expr "-=" Expr
		switch {
		case n.Expr.Operand.Type.Kind() == cc.Ptr:
			g.w(" *(")
			g.lvalue(n.Expr)
			g.w(") -= %d*uintptr(", g.model.Sizeof(n.Expr.Operand.Type.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		default:
			g.voidArithmeticAsop(n)
		}
	case
		cc.ExprAndAssign, // Expr "&=" Expr
		cc.ExprDivAssign, // Expr "/=" Expr
		cc.ExprLshAssign, // Expr "<<=" Expr
		cc.ExprModAssign, // Expr "%=" Expr
		cc.ExprMulAssign, // Expr "*=" Expr
		cc.ExprOrAssign,  // Expr "|=" Expr
		cc.ExprRshAssign, // Expr ">>=" Expr
		cc.ExprXorAssign: // Expr "^=" Expr

		g.voidArithmeticAsop(n)
	case cc.ExprPExprList: // '(' ExprList ')'
		for l := n.ExprList; l != nil; l = l.ExprList {
			g.void(l.Expr)
			g.w(";")
		}
	case cc.ExprCast: // '(' TypeName ')' Expr
		if isVaList(n.Expr.Operand.Type) { //TODO- ?
			g.w("%sVA%s(", g.crtPrefix, g.typ(cc.UnderlyingType(n.TypeName.Type)))
			g.value(n.Expr, false)
			g.w(")")
			return
		}

		g.void(n.Expr)
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		switch {
		case n.Expr.IsZero():
			if !g.voidCanIgnore(n.Expr) {
				g.void(n.Expr)
			}
			g.void(n.Expr2)
		case n.Expr.IsNonZero():
			if !g.voidCanIgnore(n.Expr) {
				g.void(n.Expr)
			}
			g.exprList(n.ExprList, true)
		default:
			// if expr != 0 {
			//	exprList
			// } else {
			//	expr2
			// }
			g.w("if ")
			g.value(n.Expr, false)
			g.w(" != 0 {")
			g.exprList(n.ExprList, true)
			g.w("} else {")
			g.void(n.Expr2)
			g.w("}")
		}
	case cc.ExprLAnd: // Expr "&&" Expr
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			// nop
		case n.Expr.IsZero():
			if !g.voidCanIgnore(n.Expr) {
				g.void(n.Expr)
			}
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			g.void(n.Expr2)
		case g.voidCanIgnore(n.Expr2):
			g.void(n.Expr)
		default:
			g.w("if ")
			g.value(n.Expr, false)
			g.w(" != 0 {")
			g.void(n.Expr2)
			g.w("}")
		}
	case cc.ExprLOr: // Expr "||" Expr
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			// nop
		case n.Expr.IsNonZero():
			if !g.voidCanIgnore(n.Expr) {
				g.void(n.Expr)
			}
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			g.void(n.Expr2)
		case g.voidCanIgnore(n.Expr2):
			g.void(n.Expr)
		default:
			g.w("if ")
			g.value(n.Expr, false)
			g.w(" == 0 {")
			g.void(n.Expr2)
			g.w("}")
		}
	case cc.ExprIndex: // Expr '[' ExprList ']'
		g.void(n.Expr)
		if !g.voidCanIgnoreExprList(n.ExprList) {
			g.w("\n")
		}
		g.exprList(n.ExprList, true)
	case // Unary
		cc.ExprAddrof,     // '&' Expr
		cc.ExprCpl,        // '~' Expr
		cc.ExprDeref,      // '*' Expr
		cc.ExprNot,        // '!' Expr
		cc.ExprUnaryMinus, // '-' Expr
		cc.ExprUnaryPlus:  // '+' Expr

		g.void(n.Expr)
	case // Binary
		cc.ExprAdd, // Expr '+' Expr
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprEq,  // Expr "==" Expr
		cc.ExprGe,  // Expr ">=" Expr
		cc.ExprGt,  // Expr ">" Expr
		cc.ExprLe,  // Expr "<=" Expr
		cc.ExprLsh, // Expr "<<" Expr
		cc.ExprLt,  // Expr '<' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprNe,  // Expr "!=" Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprRsh, // Expr ">>" Expr
		cc.ExprSub, // Expr '-' Expr
		cc.ExprXor: // Expr '^' Expr

		g.void(n.Expr)
		if !g.voidCanIgnore(n.Expr2) {
			g.w(";")
		}
		g.void(n.Expr2)
	case cc.ExprStatement: // '(' CompoundStmt ')'
		g.compoundStmt(n.CompoundStmt, nil, nil, false, nil, nil, nil, nil, false, false)
	default:
		todo("", g.position(n), n.Case)
	}
} // void

func (g *gen) lvalue(n *cc.Expr) {
	g.w("&")
	g.value(n, false)
}

func (g *ngen) lvalue(n *cc.Expr) {
	g.w("&")
	g.value(n, false)
}

func (g *gen) value(n *cc.Expr, packedField bool) {
	g.w("(")

	defer g.w(")")

	if n.Operand.Value != nil && g.voidCanIgnore(n) {
		g.constant(n)
		return
	}

	switch n.Case {
	case cc.ExprIdent: // IDENTIFIER
		d := g.normalizeDeclarator(n.Declarator)
		switch {
		case d == nil:
			if n.Operand.Type == nil || n.Operand.Value == nil {
				todo("", g.position0(n), n.Operand)
			}

			// Enum const
			g.w("%s(", g.typ(n.Operand.Type))
			g.constant(n)
			g.w(")")
		default:
			g.enqueue(d)
			switch {
			case d.Type.Kind() == cc.Function:
				g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
			case g.escaped(d) && cc.UnderlyingType(d.Type).Kind() != cc.Array:
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
			default:
				g.w("%s", g.mangleDeclarator(d))
			}
		}
	case cc.ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}
		switch d := n.Declarator; {
		case g.escaped(d):
			todo("", g.position(d))
		default:
			g.w("func() %s { %s = ", g.typ(d.Type), g.mangleDeclarator(d))
			g.literal(d.Type, d.Initializer)
			g.w("; return %s }()", g.mangleDeclarator(d))
		}
	case
		cc.ExprEq, // Expr "==" Expr
		cc.ExprGe, // Expr ">=" Expr
		cc.ExprGt, // Expr ">" Expr
		cc.ExprLe, // Expr "<=" Expr
		cc.ExprLt, // Expr '<' Expr
		cc.ExprNe: // Expr "!=" Expr

		g.relop(n)
	case
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprXor: // Expr '^' Expr

		g.binop(n)
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		if d := n.Expr.Declarator; d != nil && d.Name() == idBuiltinAlloca {
			g.w("alloca(&allocs, int(")
			if n.ArgumentExprListOpt.ArgumentExprList.ArgumentExprList != nil {
				todo("", g.position0(n))
			}
			g.value(n.ArgumentExprListOpt.ArgumentExprList.Expr, false)
			g.w("))")
			return
		}

		if n.Expr.Case == cc.ExprIdent && n.Expr.Declarator == nil {
			switch x := n.Expr.Scope.LookupIdent(n.Expr.Token.Val).(type) {
			case *cc.Declarator:
				n.Expr.Declarator = x
				n.Expr.Operand.Type = &cc.PointerType{Item: x.Type}
			default:
				todo("%v: %T undefined: %q", g.position0(n), x, dict.S(n.Expr.Token.Val))
			}
		}
		var ft0 cc.Type
		if !isFnPtr(n.Expr.Operand.Type, &ft0) {
			todo("%v: %v", g.position0(n), n.Expr.Operand.Type)
		}
		ft := cc.UnderlyingType(ft0).(*cc.FunctionType)
		var args []*cc.Expr
		if o := n.ArgumentExprListOpt; o != nil {
			for l := o.ArgumentExprList; l != nil; l = l.ArgumentExprList {
				args = append(args, l.Expr)
			}
		}
		params := ft.Params
		if len(params) == 1 && params[0].Kind() == cc.Void {
			params = nil
		}
		switch np := len(params); {
		case len(args) > np && !ft.Variadic:
			for _, v := range args[np:] {
				if !g.voidCanIgnore(v) {
					todo("", g.position0(v))
				}
			}
			args = args[:np]
			fallthrough
		case
			len(args) > np && ft.Variadic,
			len(args) == np:

			g.convert(n.Expr, ft)
			g.w("(tls")
			for i, v := range args {
				g.w(", ")
				switch t := n.CallArgs[i].Type; {
				case t == nil:
					g.value(v, false)
				default:
					g.convert(v, t)
				}
			}
			g.w(")")
		default:
			todo("", g.position0(n), np, len(args), ft.Variadic)
		}
	case cc.ExprAddrof: // '&' Expr
		g.uintptr(n.Expr, false)
	case cc.ExprSelect: // Expr '.' IDENTIFIER
		fp := n.Operand.FieldProperties
		switch {
		case fp.Declarator.Type.Kind() == cc.Array:
			g.uintptr(n.Expr, false)
			g.w("+%d", fp.Offset)
		default:
			switch {
			case fp.Bits != 0 && !packedField:
				g.bitField(n)
			default:
				if n.Expr.Case == cc.ExprCall {
					g.value(n.Expr, false)
					g.w(".F%s", dict.S(n.Token2.Val))
					return
				}

				t := n.Operand.Type
				if fp.Bits != 0 {
					t = fp.PackedType
				}
				g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
				g.uintptr(n.Expr, false)
				g.w("+%d", fp.Offset)
				g.w("))")
			}
		}
	case cc.ExprPSelect: // Expr "->" IDENTIFIER
		fp := n.Operand.FieldProperties
		switch {
		case fp.Declarator.Type.Kind() == cc.Array:
			g.value(n.Expr, false)
			g.w("+%d", fp.Offset)
		default:
			switch {
			case fp.Bits != 0 && !packedField:
				g.bitField(n)
			default:
				t := n.Operand.Type
				if fp.Bits != 0 {
					t = fp.PackedType
				}
				g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
				g.value(n.Expr, false)
				g.w("+%d))", fp.Offset)
			}
		}
	case cc.ExprIndex: // Expr '[' ExprList ']'
		var it cc.Type
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case *cc.ArrayType:
			it = x.Item
		case *cc.PointerType:
			it = x.Item
		default:
			todo("%v: %T", g.position0(n), x)
		}
		switch {
		case it.Kind() == cc.Array:
			g.value(n.Expr, false)
			g.indexOff(n.ExprList, it)
		default:
			g.w("*(*%s)(unsafe.Pointer(", g.typ(n.Operand.Type))
			g.value(n.Expr, false)
			g.indexOff(n.ExprList, it)
			g.w("))")
		}
	case cc.ExprAdd: // Expr '+' Expr
		switch t, u := cc.UnderlyingType(n.Expr.Operand.Type), cc.UnderlyingType(n.Expr2.Operand.Type); {
		case t.Kind() == cc.Ptr:
			g.value(n.Expr, false)
			g.w(" + %d*uintptr(", g.model.Sizeof(t.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		case u.Kind() == cc.Ptr:
			g.w("%d*uintptr(", g.model.Sizeof(u.(*cc.PointerType).Item))
			g.value(n.Expr, false)
			g.w(") + ")
			g.value(n.Expr2, false)
		default:
			g.binop(n)
		}
	case cc.ExprSub: // Expr '-' Expr
		switch t, u := cc.UnderlyingType(n.Expr.Operand.Type), cc.UnderlyingType(n.Expr2.Operand.Type); {
		case t.Kind() == cc.Ptr && u.Kind() == cc.Ptr:
			g.w("%s((", g.typ(n.Operand.Type))
			g.value(n.Expr, false)
			g.w(" - ")
			g.value(n.Expr2, false)
			g.w(")/%d)", g.model.Sizeof(t.(*cc.PointerType).Item))
		case t.Kind() == cc.Ptr:
			g.value(n.Expr, false)
			g.w(" - %d*uintptr(", g.model.Sizeof(t.(*cc.PointerType).Item))
			g.value(n.Expr2, false)
			g.w(")")
		default:
			g.binop(n)
		}
	case cc.ExprDeref: // '*' Expr
		it := cc.UnderlyingType(n.Expr.Operand.Type).(*cc.PointerType).Item
		switch it.Kind() {
		case
			cc.Array,
			cc.Function:

			g.value(n.Expr, false)
		default:
			i := 1
			for n.Expr.Case == cc.ExprDeref {
				i++
				n = n.Expr
			}
			g.w("%[1]s(%[1]s%[2]s)(unsafe.Pointer(", strings.Repeat("*", i), g.typ(it))
			g.value(n.Expr, false)
			g.w("))")
		}
	case cc.ExprAssign: // Expr '=' Expr
		g.assignmentValue(n)
	case cc.ExprLAnd: // Expr "&&" Expr
		if n.Operand.Value != nil && g.voidCanIgnore(n) {
			g.constant(n)
			break
		}

		g.needBool2int++
		g.w(" bool2int((")
		g.value(n.Expr, false)
		g.w(" != 0) && (")
		g.value(n.Expr2, false)
		g.w(" != 0))")
	case cc.ExprLOr: // Expr "||" Expr
		if n.Operand.Value != nil && g.voidCanIgnore(n) {
			g.constant(n)
			break
		}

		g.needBool2int++
		g.w(" bool2int((")
		g.value(n.Expr, false)
		g.w(" != 0) || (")
		g.value(n.Expr2, false)
		g.w(" != 0))")
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		t := n.Operand.Type
		switch {
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			g.value(n.Expr2, false)
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			g.exprList(n.ExprList, false)
		default:
			g.w(" func() %s { if ", g.typ(t))
			g.value(n.Expr, false)
			g.w(" != 0 { return ")
			g.exprList2(n.ExprList, t)
			g.w(" }\n\nreturn ")
			g.convert(n.Expr2, t)
			g.w(" }()")
		}
	case cc.ExprCast: // '(' TypeName ')' Expr
		t := n.TypeName.Type
		op := n.Expr.Operand
		if isVaList(op.Type) {
			g.w("%sVA%s(", crt, g.typ(cc.UnderlyingType(t)))
			g.value(n.Expr, false)
			g.w(")")
			return
		}

		switch x := cc.UnderlyingType(t).(type) {
		case *cc.PointerType:
			if d := n.Expr.Declarator; x.Item.Kind() == cc.Function && d != nil && g.normalizeDeclarator(d).Type.Equal(x.Item) {
				g.value(n.Expr, false)
				return
			}
		}

		g.convert(n.Expr, t)
	case cc.ExprPreInc: // "++" Expr
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.model.Sizeof(x.Item)))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("preinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("preinc%d", g.typ(x), 1))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}

			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprPostInc: // Expr "++"
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.model.Sizeof(x.Item)))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("postinc%d", g.typ(x), 1))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}

			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprPreDec: // "--" Expr
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.int64ToUintptr(-g.model.Sizeof(x.Item))))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("preinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.convertInt64(-1, x)))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}
			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprPostDec: // Expr "--"
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.int64ToUintptr(-g.model.Sizeof(x.Item))))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value(n.Expr, true)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.convertInt64(-1, x)))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}
			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprNot: // '!' Expr
		g.needBool2int++
		g.w(" bool2int(")
		g.value(n.Expr, false)
		g.w(" == 0)")
	case cc.ExprLsh: // Expr "<<" Expr
		g.convert(n.Expr, n.Operand.Type)
		g.w(" << (uint(")
		g.value(n.Expr2, false)
		g.w(") %% %d)", g.shiftMod(cc.UnderlyingType(n.Operand.Type)))
	case cc.ExprRsh: // Expr ">>" Expr
		g.convert(n.Expr, n.Operand.Type)
		g.w(" >> (uint(")
		g.value(n.Expr2, false)
		g.w(") %% %d)", g.shiftMod(cc.UnderlyingType(n.Operand.Type)))
	case cc.ExprUnaryMinus: // '-' Expr
		g.w("- ")
		g.convert(n.Expr, n.Operand.Type)
	case cc.ExprCpl: // '~' Expr
		g.w("^(")
		g.convert(n.Expr, n.Operand.Type)
		g.w(")")
	case cc.ExprAddAssign: // Expr "+=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case *cc.PointerType:
			g.needPreInc = true
			g.w("preinc(")
			g.lvalue(n.Expr)
			g.w(", %d*uintptr(", g.model.Sizeof(x.Item))
			g.value(n.Expr2, false)
			g.w("))")
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("add%d", "+", g.typ(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprSubAssign: // Expr "-=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("sub%d", "-", g.typ(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprOrAssign: // Expr "|=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					g.w("%s(&", g.registerHelper("or%db", "|", g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
					g.value(n.Expr, true)
					g.w(", ")
					g.convert(n.Expr2, n.Operand.Type)
					g.w(")")
				default:
					g.w("%s(", g.registerHelper("or%d", "|", g.typ(x)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.convert(n.Expr2, x)
					g.w(")")
				}
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprAndAssign: // Expr "&=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					g.w("%s(&", g.registerHelper("and%db", "&", g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
					g.value(n.Expr, true)
					g.w(", ")
					g.convert(n.Expr2, n.Operand.Type)
					g.w(")")
				default:
					g.w("%s(", g.registerHelper("and%d", "&", g.typ(x)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.convert(n.Expr2, x)
					g.w(")")
				}
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprXorAssign: // Expr "^=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					g.w("%s(&", g.registerHelper("xor%db", "^", g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
					g.value(n.Expr, true)
					g.w(", ")
					g.convert(n.Expr2, n.Operand.Type)
					g.w(")")
				default:
					g.w("%s(", g.registerHelper("xor%d", "^", g.typ(x)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.convert(n.Expr2, x)
					g.w(")")
				}
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprPExprList: // '(' ExprList ')'
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.value(l[0], false)
		default:
			g.w("func() %v {", g.typ(n.Operand.Type))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], n.Operand.Type)
			g.w("}()")
		}
	case cc.ExprMulAssign: // Expr "*=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("mul%d", "*", g.typ(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprDivAssign: // Expr "/=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("div%d", "/", g.typ(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprModAssign: // Expr "%=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("mod%d", "%", g.typ(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprRshAssign: // Expr ">>=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("rsh%d", ">>", g.typ(x), g.shiftMod(x)))
				g.lvalue(n.Expr)
				g.w(", ")
				g.convert(n.Expr2, x)
				g.w(")")
				return
			}
			todo("", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprUnaryPlus: // '+' Expr
		g.convert(n.Expr, n.Operand.Type)
	case
		cc.ExprInt,        // INTCONST
		cc.ExprSizeofExpr, // "sizeof" Expr
		cc.ExprSizeofType, // "sizeof" '(' TypeName ')'
		cc.ExprString:     // STRINGLITERAL

		g.constant(n)
	default:
		todo("", g.position0(n), n.Case, n.Operand) // value
	}
}

func (g *ngen) value(n *cc.Expr, packedField bool) { g.value0(n, packedField, false) }

func (g *ngen) value0(n *cc.Expr, packedField bool, exprCall bool) {
	g.w("(")

	defer g.w(")")

	if g.escaped(n.Declarator) {
		g.value0Escaped(n, packedField, exprCall)
		return
	}

	if n.Operand.Value != nil && g.voidCanIgnore(n) {
		g.constant(n)
		return
	}

	switch n.Case {
	case cc.ExprIdent: // IDENTIFIER
		d := g.normalizeDeclarator(n.Declarator)
		switch {
		case d == nil:
			if n.Operand.Type == nil || n.Operand.Value == nil {
				todo("%v: %s, %v, %p", g.position(n), string(n.Token.S()), n.Operand, n.Declarator)
			}

			// Enum const
			g.w("%s(", g.typ(n.Operand.Type))
			g.constant(n)
			g.w(")")
		default:
			g.enqueue(d)
			arr, _ := cc.UnderlyingType(d.Type).(*cc.ArrayType)
			switch {
			case d.Type.Kind() == cc.Function:
				if exprCall {
					g.w("%s", g.mangleDeclarator(d))
					break
				}

				g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
			case g.escaped(d) && arr == nil:
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
			default:
				g.w("%s", g.mangleDeclarator(d))
			}
		}
	case cc.ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}
		if d := n.Declarator; d != nil {
			switch {
			case g.escaped(d):
				todo("", g.position(d))
			default:
				//TODO- g.w("func(/*TODO1326*/) %s { %s = ", g.typ(d.Type), g.mangleDeclarator(d))
				//TODO- g.literal(d.Type, d.Initializer)
				//TODO- g.w("; return %s }()", g.mangleDeclarator(d))
				g.literal(d.Type, d.Initializer)
			}
			break
		}

		g.literal(
			n.TypeName.Type,
			&cc.Initializer{
				Case:            cc.InitializerCompLit,
				InitializerList: n.InitializerList,
			},
		)
	case
		cc.ExprEq, // Expr "==" Expr
		cc.ExprGe, // Expr ">=" Expr
		cc.ExprGt, // Expr ">" Expr
		cc.ExprLe, // Expr "<=" Expr
		cc.ExprLt, // Expr '<' Expr
		cc.ExprNe: // Expr "!=" Expr

		g.relop(n)
	case
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprXor: // Expr '^' Expr

		g.binop(n)
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		if e := n.Expr; e.Case == cc.ExprIdent && e.Token.Val == idGo2 {
			g.w("%s", dict.S(int(n.ArgumentExprListOpt.ArgumentExprList.Expr.Operand.Value.(*ir.StringValue).StringID)))
			return
		}

		if d := n.Expr.Declarator; d != nil && d.Name() == idBuiltinAlloca {
			g.w("%sAlloca(&allocs, int(", g.crtPrefix)
			if n.ArgumentExprListOpt.ArgumentExprList.ArgumentExprList != nil {
				todo("", g.position(n))
			}
			g.value(n.ArgumentExprListOpt.ArgumentExprList.Expr, false)
			g.w("))")
			return
		}

		if n.Expr.Case == cc.ExprIdent && n.Expr.Declarator == nil {
			switch x := n.Expr.Scope.LookupIdent(n.Expr.Token.Val).(type) {
			case *cc.Declarator:
				n.Expr.Declarator = x
				n.Expr.Operand.Type = &cc.PointerType{Item: x.Type}
			default:
				todo("%v: %T undefined: %q", g.position(n), x, dict.S(n.Expr.Token.Val))
			}
		}

		var t0 cc.Type
		if !isFnPtr(n.Expr.Operand.Type, &t0) {
			todo("%v: %v", g.position(n), n.Expr.Operand.Type)
		}

		t := cc.UnderlyingType(t0).(*cc.FunctionType)
		var d *cc.Declarator
		if d0 := n.Expr.Declarator; d0 != nil && d0.FunctionDefinition != nil {
			d = d0.FunctionDefinition.Declarator
			if !isFnPtr(cc.UnderlyingType(d.Type), &t0) {
				todo("%v: %v", g.position(n), d.Type)
			}
			t = cc.UnderlyingType(t0).(*cc.FunctionType)
		}
		var args []*cc.Expr
		if o := n.ArgumentExprListOpt; o != nil {
			for l := o.ArgumentExprList; l != nil; l = l.ArgumentExprList {
				args = append(args, l.Expr)
			}
		}
		params := t.Params
		var voidParams bool
		if voidParams = len(params) == 1 && params[0].Kind() == cc.Void; voidParams {
			params = nil
		}
		switch {
		case voidParams && len(args) != 0:
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		case len(args) < len(params):
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		case len(args) == len(params):
			//ok

		// len(args) > len(params)
		case t.Variadic:
			// ok
		case len(params) == 0:
			switch {
			case d == nil:
				// ok
			default:
				for _, v := range args {
					if !g.voidCanIgnore(v) {
						todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
					}
				}
				args = nil
			}
		default:
			todo("%v: %v args %v params %v variadic %v voidParams %v", g.position(n), n.Case, len(args), len(params), t.Variadic, voidParams)
		}
		g.convert(n.Expr, t)
		g.w("(tls")
		for i, v := range args {
			g.w(", ")
			switch t := n.CallArgs[i].Type; {
			case t == nil:
				g.value(v, false)
			default:
				g.convert(v, t)
			}
		}
		g.w(")")
	case cc.ExprAddrof: // '&' Expr
		g.uintptr(n.Expr, false)
	case cc.ExprSelect: // Expr '.' IDENTIFIER
		fp := n.Operand.FieldProperties
		switch {
		case fp.Declarator.Type.Kind() == cc.Array:
			g.uintptr(n.Expr, false)
			g.w("+%d", fp.Offset)
		default:
			switch {
			case fp.Bits != 0 && !packedField:
				g.bitField(n)
			default:
				if n.Expr.Case == cc.ExprCall {
					g.value0(n.Expr, false, exprCall)
					g.w(".F%s", dict.S(n.Token2.Val))
					return
				}

				t := n.Operand.Type
				if fp.Bits != 0 {
					t = fp.PackedType
				}
				g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
				g.uintptr(n.Expr, false)
				g.w("+%d", fp.Offset)
				g.w("))")
			}
		}
	case cc.ExprPSelect: // Expr "->" IDENTIFIER
		fp := n.Operand.FieldProperties
		switch {
		case fp.Declarator.Type.Kind() == cc.Array:
			g.value0(n.Expr, false, exprCall)
			g.w("+%d", fp.Offset)
		default:
			switch {
			case fp.Bits != 0 && !packedField:
				g.bitField(n)
			default:
				t := n.Operand.Type
				if fp.Bits != 0 {
					t = fp.PackedType
				}
				g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
				g.value0(n.Expr, false, exprCall)
				g.w("+%d))", fp.Offset)
			}
		}
	case cc.ExprIndex: // Expr '[' ExprList ']'
		var it cc.Type
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case *cc.ArrayType:
			it = x.Item
		case *cc.PointerType:
			it = x.Item
		default:
			if !x.IsIntegerType() {
				todo("%v: %T", g.position(n), x)
			}

			switch y := cc.UnderlyingType(n.ExprList.Operand.Type).(type) {
			case *cc.ArrayType:
				it = y.Item
			case *cc.PointerType:
				it = y.Item
			default:
				todo("%v: %T", g.position(n), y)
			}

			// 42[p]
			switch {
			case it.Kind() == cc.Array:
				todo("%v: %v", g.position(n), it)
			default:
				g.w("*(*%s)(unsafe.Pointer(", g.typ(n.Operand.Type))
				g.exprList(n.ExprList, false)
				g.indexOff2(n.Expr, it)
				g.w("))")
			}
			return
		}

		// p[42]
		switch {
		case it.Kind() == cc.Array:
			g.value0(n.Expr, false, exprCall)
			g.indexOff(n.ExprList, it)
		default:
			g.w("*(*%s)(unsafe.Pointer(", g.typ(n.Operand.Type))
			g.value0(n.Expr, false, exprCall)
			g.indexOff(n.ExprList, it)
			g.w("))")
		}
	case cc.ExprAdd: // Expr '+' Expr
		t, u := cc.UnderlyingType(n.Expr.Operand.Type), cc.UnderlyingType(n.Expr2.Operand.Type)
		switch {
		case t.Kind() == cc.Ptr:
			g.value0(n.Expr, false, exprCall)
			g.w(" + %d*uintptr(", g.model.Sizeof(t.(*cc.PointerType).Item))
			g.value0(n.Expr2, false, exprCall)
			g.w(")")
		case u.Kind() == cc.Ptr:
			g.w("%d*uintptr(", g.model.Sizeof(u.(*cc.PointerType).Item))
			g.value0(n.Expr, false, exprCall)
			g.w(") + ")
			g.value0(n.Expr2, false, exprCall)
		default:
			g.binop(n)
		}
	case cc.ExprSub: // Expr '-' Expr
		switch t, u := cc.UnderlyingType(n.Expr.Operand.Type), cc.UnderlyingType(n.Expr2.Operand.Type); {
		case t.Kind() == cc.Ptr && u.Kind() == cc.Ptr:
			g.w("%s((", g.typ(n.Operand.Type))
			g.value0(n.Expr, false, exprCall)
			g.w(" - ")
			g.value0(n.Expr2, false, exprCall)
			g.w(")/%d)", g.model.Sizeof(t.(*cc.PointerType).Item))
		case t.Kind() == cc.Ptr:
			g.value0(n.Expr, false, exprCall)
			g.w(" - %d*uintptr(", g.model.Sizeof(t.(*cc.PointerType).Item))
			g.value0(n.Expr2, false, exprCall)
			g.w(")")
		default:
			g.binop(n)
		}
	case cc.ExprDeref: // '*' Expr
		it := cc.UnderlyingType(n.Expr.Operand.Type).(*cc.PointerType).Item
		switch it.Kind() {
		case
			cc.Array,
			cc.Function:

			g.value0(n.Expr, false, exprCall)
		default:
			i := 1
			for n.Expr.Case == cc.ExprDeref {
				i++
				n = n.Expr
			}
			g.w("%[1]s(%[1]s%[2]s)(unsafe.Pointer(", strings.Repeat("*", i), g.typ(it))
			g.value0(n.Expr, false, exprCall)
			g.w("))")
		}
	case cc.ExprAssign: // Expr '=' Expr
		g.assignmentValue(n)
	case cc.ExprLAnd: // Expr "&&" Expr
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			g.constant(n)
		case n.Expr.IsZero():
			if g.voidCanIgnore(n.Expr) {
				g.w(" 0")
				break
			}

			g.w(" bool2int(")
			g.value0(n.Expr, false, exprCall)
			g.w(" != 0)")
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			g.w(" bool2int(")
			g.value0(n.Expr2, false, exprCall)
			g.w(" != 0)")
		default:
			g.w(" bool2int((")
			g.value0(n.Expr, false, exprCall)
			g.w(" != 0) && (")
			g.value0(n.Expr2, false, exprCall)
			g.w(" != 0))")
		}
	case cc.ExprLOr: // Expr "||" Expr
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			g.constant(n)
		case n.Expr.IsNonZero():
			if g.voidCanIgnore(n.Expr) {
				g.w(" 1")
				break
			}

			g.w(" bool2int(")
			g.value0(n.Expr, false, exprCall)
			g.w(" != 0)")
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			g.w(" bool2int(")
			g.value0(n.Expr2, false, exprCall)
			g.w(" != 0)")
		default:
			g.w(" bool2int((")
			g.value0(n.Expr, false, exprCall)
			g.w(" != 0) || (")
			g.value0(n.Expr2, false, exprCall)
			g.w(" != 0))")
		}
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		t := n.Operand.Type
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			g.constant(n)
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			g.value0(n.Expr2, false, exprCall)
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			g.exprList(n.ExprList, false)
		default:
			g.w(" func() %s { if ", g.typ(t))
			g.value0(n.Expr, false, exprCall)
			g.w(" != 0 { return ")
			g.exprList2(n.ExprList, t)
			g.w(" }\n\nreturn ")
			g.convert(n.Expr2, t)
			g.w(" }()")
		}
	case cc.ExprCast: // '(' TypeName ')' Expr
		t := n.TypeName.Type
		op := n.Expr.Operand
		if isVaList(op.Type) {
			switch cc.UnderlyingType(t).(type) {
			case *cc.StructType:
				g.w("%sVAother(", g.crtPrefix)
				g.value0(n.Expr, false, exprCall)
				g.w(").(%s)", g.typ(t))
			default:
				g.w("%sVA%s(", g.crtPrefix, g.typ(cc.UnderlyingType(t)))
				g.value0(n.Expr, false, exprCall)
				g.w(")")
			}
			return
		}

		switch x := cc.UnderlyingType(t).(type) {
		case *cc.PointerType:
			//TODO- if d := n.Expr.Declarator; x.Item.Kind() == cc.Function && d != nil && g.normalizeDeclarator(d).Type.Equal(x.Item) {
			if d := n.Expr.Declarator; x.Item.Kind() == cc.Function && d != nil {
				g.value0(n.Expr, false, exprCall)
				return
			}
		}

		g.convert(n.Expr, t)
	case cc.ExprPreInc: // "++" Expr
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.model.Sizeof(x.Item)))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("preinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value0(n.Expr, true, exprCall)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("preinc%d", g.typ(x), 1))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}

			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprPostInc: // Expr "++"
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.model.Sizeof(x.Item)))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", 1, g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value0(n.Expr, true, exprCall)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("postinc%d", g.typ(x), 1))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}

			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprPreDec: // "--" Expr
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.int64ToUintptr(-g.model.Sizeof(x.Item))))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("preinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value0(n.Expr, true, exprCall)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("preinc%d", g.typ(x), g.convertInt64(-1, x)))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}
			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprPostDec: // Expr "--"
		switch x := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.int64ToUintptr(-g.model.Sizeof(x.Item))))
			g.lvalue(n.Expr)
			g.w(")")
		case cc.TypeKind:
			if op := n.Expr.Operand; op.Bits() != 0 {
				fp := op.FieldProperties
				g.w("%s(&", g.registerHelper("postinc%db", g.convertInt64(-1, x), g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
				g.value0(n.Expr, true, exprCall)
				g.w(")")
				return
			}

			if x.IsArithmeticType() {
				g.w("%s(", g.registerHelper("postinc%d", g.typ(x), g.convertInt64(-1, x)))
				g.lvalue(n.Expr)
				g.w(")")
				return
			}
			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprNot: // '!' Expr
		g.w(" bool2int(")
		g.value0(n.Expr, false, exprCall)
		g.w(" == 0)")
	case cc.ExprLsh: // Expr "<<" Expr
		g.convert(n.Expr, n.Operand.Type)
		g.w(" << (uint(")
		g.value0(n.Expr2, false, exprCall)
		g.w(") %% %d)", g.shiftMod(cc.UnderlyingType(n.Operand.Type)))
	case cc.ExprRsh: // Expr ">>" Expr
		g.convert(n.Expr, n.Operand.Type)
		g.w(" >> (uint(")
		g.value0(n.Expr2, false, exprCall)
		g.w(") %% %d)", g.shiftMod(cc.UnderlyingType(n.Operand.Type)))
	case cc.ExprUnaryMinus: // '-' Expr
		g.w("- ")
		g.convert(n.Expr, n.Operand.Type)
	case cc.ExprCpl: // '~' Expr
		g.w("^(")
		g.convert(n.Expr, n.Operand.Type)
		g.w(")")
	case cc.ExprAddAssign: // Expr "+=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case *cc.PointerType:
			g.w("%sPreinc(", g.crtPrefix)
			g.lvalue(n.Expr)
			g.w(", %d*uintptr(", g.model.Sizeof(x.Item))
			g.value0(n.Expr2, false, exprCall)
			g.w("))")
		case cc.TypeKind:
			if x.IsArithmeticType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("add%d", "+", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprSubAssign: // Expr "-=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("sub%d", "-", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprOrAssign: // Expr "|=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(&", g.registerHelper("or%db", "|", g.typ(n.Expr.Operand.Type), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type), g.typ(fp.PackedType), fp.Bitoff, g.model.Sizeof(pro.Type)*8, fp.Bits, g.model.Sizeof(op.Type)*8))
					g.value(n.Expr, true)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("or%d", "|", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprAndAssign: // Expr "&=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(&", g.registerHelper("and%db", "&", g.typ(n.Expr.Operand.Type), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type), g.typ(fp.PackedType), fp.Bitoff, g.model.Sizeof(pro.Type)*8, fp.Bits, g.model.Sizeof(op.Type)*8))
					g.value(n.Expr, true)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("and%d", "&", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprXorAssign: // Expr "^=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					fp := op.FieldProperties
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(&", g.registerHelper("xor%db", "^", g.typ(n.Expr.Operand.Type), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type), g.typ(fp.PackedType), fp.Bitoff, g.model.Sizeof(pro.Type)*8, fp.Bits, g.model.Sizeof(op.Type)*8))
					g.value(n.Expr, true)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("xor%d", "^", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprPExprList: // '(' ExprList ')'
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.value0(l[0], false, exprCall)
		default:
			g.w("func() %v {", g.typ(n.Operand.Type))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], n.Operand.Type)
			g.w("}()")
		}
	case cc.ExprMulAssign: // Expr "*=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("mul%d", "*", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprDivAssign: // Expr "/=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsArithmeticType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("div%d", "/", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprModAssign: // Expr "%=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					pro, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
					g.w("%s(", g.registerHelper("mod%d", "%", g.typ(x), g.typ(n.Expr2.Operand.Type), g.typ(pro.Type)))
					g.lvalue(n.Expr)
					g.w(", ")
					g.value(n.Expr2, false)
					g.w(")")
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprRshAssign: // Expr ">>=" Expr
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case cc.TypeKind:
			if x.IsIntegerType() {
				switch op := n.Expr.Operand; {
				case op.Bits() != 0:
					todo("", g.position(n))
				default:
					g.w("%s(", g.registerHelper("rsh%d", ">>", g.typ(n.Expr.Operand.Type), g.typ(x)))
					g.lvalue(n.Expr)
					g.w(", uint(")
					g.value(n.Expr2, false)
					g.w(")%%%d)", g.shiftMod(x))
				}
				return
			}
			todo("", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprUnaryPlus: // '+' Expr
		g.convert(n.Expr, n.Operand.Type)
	case
		cc.ExprInt,        // INTCONST
		cc.ExprSizeofExpr, // "sizeof" Expr
		cc.ExprSizeofType, // "sizeof" '(' TypeName ')'
		cc.ExprString:     // STRINGLITERAL

		g.constant(n)
	case cc.ExprStatement: // '(' CompoundStmt ')'
		g.w("func() %s {", g.typ(n.Operand.Type))
		g.compoundStmt(n.CompoundStmt, nil, nil, false, nil, nil, nil, nil, false, true)
		g.w("}()")
	default:
		//println(n.Case.String())
		todo("", g.position(n), n.Case)
	}
} // value0

func (g *ngen) value0Escaped(n *cc.Expr, packedField bool, exprCall bool) {
	d := n.Declarator
	g.enqueue(d)
	u := cc.UnderlyingType(d.Type)
	switch n.Case {
	case cc.ExprIdent: // IDENTIFIER
		if u.Kind() == cc.Array {
			g.w("%s", g.mangleDeclarator(d))
			return
		}

		if u.Kind() == cc.Function {
			g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
			return
		}

		g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
	case cc.ExprPExprList: // '(' ExprList ')'
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.value0(l[0], packedField, exprCall)
		default:
			g.w("func() %v {", g.typ(n.Operand.Type))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], n.Operand.Type)
			g.w("}()")
		}
	default:
		todo("", g.position(n), n.Case) // value0Escaped
	}
}

func (g *gen) pexprList(n *cc.ExprList) (r []*cc.Expr) { //TODO use
	for l := n; l != nil; l = l.ExprList {
		if e := l.Expr; l.ExprList == nil || !g.voidCanIgnore(e) {
			r = append(r, e)
		}
	}
	return r
}

func (g *ngen) pexprList(n *cc.ExprList) (r []*cc.Expr) { //TODO use
	for l := n; l != nil; l = l.ExprList {
		if e := l.Expr; l.ExprList == nil || !g.voidCanIgnore(e) {
			r = append(r, e)
		}
	}
	return r
}

func (g *gen) bitField(n *cc.Expr) {
	op := n.Operand
	fp := op.FieldProperties
	g.w("%s(", g.typ(op.Type))
	g.value(n, true)
	bits := int(g.model.Sizeof(op.Type) * 8)
	g.w(">>%d)<<%d>>%d", fp.Bitoff, bits-fp.Bits, bits-fp.Bits)
}

func (g *ngen) bitField(n *cc.Expr) {
	op := n.Operand
	fp := op.FieldProperties
	convType := op.Type
	switch x := underlyingType(fp.Type, true).(type) {
	case *cc.EnumType:
		if x.Min < 0 {
			break
		}

		bits := mathutil.BitLenUint64(x.Max)
		if fp.Bits == bits {
			convType = cc.Unsigned(convType)
		}
	}
	g.w("%s(", g.typ(convType))
	g.value(n, true)
	bits := int(g.model.Sizeof(op.Type) * 8)
	g.w(">>%d)<<%d>>%d", fp.Bitoff, bits-fp.Bits, bits-fp.Bits)
}

func (g *gen) indexOff(n *cc.ExprList, it cc.Type) {
	switch {
	case n.Operand.Value != nil && g.voidCanIgnoreExprList(n):
		g.w("%+d", g.model.Sizeof(it)*n.Operand.Value.(*ir.Int64Value).Value)
	default:
		g.w(" + %d*uintptr(", g.model.Sizeof(it))
		g.exprList(n, false)
		g.w(")")
	}
}

func (g *ngen) indexOff(n *cc.ExprList, it cc.Type) { // p[42]
	switch {
	case n.Operand.Value != nil && g.voidCanIgnoreExprList(n):
		g.w("%+d", g.model.Sizeof(it)*n.Operand.Value.(*ir.Int64Value).Value)
	default:
		//fmt.Printf("%v:\n", g.position(n)) //TODO- DBG
		g.w(" + %d*uintptr(", g.model.Sizeof(it))
		g.exprList(n, false)
		g.w(")")
	}
}

func (g *ngen) indexOff2(n *cc.Expr, it cc.Type) { // 42[p]
	switch {
	case n.Operand.Value != nil:
		g.w("%+d", g.model.Sizeof(it)*n.Operand.Value.(*ir.Int64Value).Value)
	default:
		g.w(" + %d*uintptr(", g.model.Sizeof(it))
		g.value(n, false)
		g.w(")")
	}
}

func (g *gen) uintptr(n *cc.Expr, packedField bool) {
	g.w("(")

	defer g.w(")")

	switch n.Case {
	case cc.ExprPExprList: // '(' ExprList ')'
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.uintptr(l[0], false)
		default:
			g.w("func() uintptr {")
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.uintptr(l[len(l)-1], packedField)
			g.w("}()")
		}
	case cc.ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}
		switch d := n.Declarator; {
		case g.escaped(d):
			g.w("func() uintptr { *(*%s)(unsafe.Pointer(%s)) = ", g.typ(d.Type), g.mangleDeclarator(d))
			g.literal(d.Type, d.Initializer)
			g.w("; return %s }()", g.mangleDeclarator(d))
		default:
			g.w("func() uintptr { %s = ", g.mangleDeclarator(d))
			g.literal(d.Type, d.Initializer)
			g.w("; return uintptr(unsafe.Pointer(&%s)) }()", g.mangleDeclarator(d))
		}
	case cc.ExprIdent: // IDENTIFIER
		d := g.normalizeDeclarator(n.Declarator)
		g.enqueue(d)
		arr := cc.UnderlyingType(d.Type).Kind() == cc.Array
		switch {
		case d.Type.Kind() == cc.Function:
			g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
		case arr:
			g.w("%s", g.mangleDeclarator(d))
		case g.escaped(d):
			g.w("%s", g.mangleDeclarator(d))
		default:
			g.w("uintptr(unsafe.Pointer(&%s))", g.mangleDeclarator(d))
		}
	case cc.ExprIndex: // Expr '[' ExprList ']'
		switch x := cc.UnderlyingType(n.Expr.Operand.Type).(type) {
		case *cc.ArrayType:
			g.uintptr(n.Expr, false)
			g.indexOff(n.ExprList, x.Item)
		case *cc.PointerType:
			g.value(n.Expr, false)
			g.indexOff(n.ExprList, x.Item)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.ExprSelect: // Expr '.' IDENTIFIER
		fp := n.Operand.FieldProperties
		if bits := fp.Bits; bits != 0 && !packedField {
			todo("", g.position0(n), n.Operand)
		}
		g.uintptr(n.Expr, packedField)
		g.w("+%d", fp.Offset)
	case cc.ExprPSelect: // Expr "->" IDENTIFIER
		fp := n.Operand.FieldProperties
		if bits := fp.Bits; bits != 0 && !packedField {
			todo("", g.position0(n), n.Operand)
		}
		g.value(n.Expr, false)
		g.w("+%d", fp.Offset)
	case cc.ExprDeref: // '*' Expr
		switch cc.UnderlyingType(cc.UnderlyingType(n.Expr.Operand.Type).(*cc.PointerType).Item).(type) {
		case *cc.ArrayType:
			g.value(n.Expr, false)
		default:
			g.value(n.Expr, false)
		}
	case cc.ExprString: // STRINGLITERAL
		g.constant(n)
	default:
		todo("", g.position0(n), n.Case, n.Operand) // uintptr
	}
}

func (g *ngen) uintptr(n *cc.Expr, packedField bool) {
	g.w("(")

	defer g.w(")")

	if g.escaped(n.Declarator) {
		g.uintptrEscaped(n)
		return
	}

	switch n.Case {
	case cc.ExprPExprList: // '(' ExprList ')'
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.uintptr(l[0], false)
		default:
			g.w("func() uintptr {")
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.uintptr(l[len(l)-1], packedField)
			g.w("}()")
		}
	case cc.ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}
		if d := n.Declarator; d != nil {
			todo("%v: %v TODO (*ngen).uintptr", g.position(n), n.Case)
		}

		t := n.TypeName.Type
		ini := &cc.Initializer{
			Case:            cc.InitializerCompLit,
			InitializerList: n.InitializerList,
		}
		if g.isConstInitializer(t, ini) {
			g.w("Ld + %q", g.allocDS(n.TypeName.Type, ini))
			break
		}

		g.w("func() uintptr { x := Lb+%d; *(*%s)(unsafe.Pointer(x)) = ", g.model.Sizeof(t), g.typ(t))
		g.literal(t, ini)
		g.w("; return x }()")
	case cc.ExprIdent: // IDENTIFIER
		d := n.Declarator
		fixMain(d)
		g.enqueue(d)
		arr := cc.UnderlyingType(d.Type).Kind() == cc.Array
		switch {
		case d.Type.Kind() == cc.Function:
			g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
		case arr:
			g.w("%s ", g.mangleDeclarator(d))
		default:
			// 		g.w("uintptr(unsafe.Pointer(&%s))", g.mangleDeclarator(d))
			todo("%v: %v TODO (*ngen).uintptr", g.position(n), n.Case)
		}
	case cc.ExprIndex: // Expr '[' ExprList ']'
		t := n.Expr.Operand.Type
		if d := n.Expr.Declarator; d != nil {
			t = d.Type
		}
		switch x := cc.UnderlyingType(t).(type) {
		case *cc.ArrayType:
			g.uintptr(n.Expr, false)
			g.indexOff(n.ExprList, x.Item)
		case *cc.PointerType:
			g.value(n.Expr, false)
			g.indexOff(n.ExprList, x.Item)
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.ExprSelect: // Expr '.' IDENTIFIER
		fp := n.Operand.FieldProperties
		if bits := fp.Bits; bits != 0 && !packedField {
			todo("", g.position(n), n.Operand)
		}
		g.uintptr(n.Expr, packedField)
		g.w("+%d", fp.Offset)
	case cc.ExprPSelect: // Expr "->" IDENTIFIER
		fp := n.Operand.FieldProperties
		if bits := fp.Bits; bits != 0 && !packedField {
			todo("", g.position(n), n.Operand)
		}
		g.value(n.Expr, false)
		g.w("+%d", fp.Offset)
	case cc.ExprDeref: // '*' Expr
		switch cc.UnderlyingType(cc.UnderlyingType(n.Expr.Operand.Type).(*cc.PointerType).Item).(type) {
		case *cc.ArrayType:
			g.value(n.Expr, false)
		default:
			g.value(n.Expr, false)
		}
	case cc.ExprString: // STRINGLITERAL
		g.constant(n)
	default:
		todo("%v: %v TODO (*ngen).uintptr", g.position(n), n.Case)
	}
} // uintptr

func (g *ngen) uintptrEscaped(n *cc.Expr) {
	d := n.Declarator
	g.enqueue(d)
	switch n.Case {
	case cc.ExprIdent: // IDENTIFIER
		switch {
		case d.Type.Kind() == cc.Function:
			fixMain(d)
			g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
		default:
			g.w("%s ", g.mangleDeclarator(d))
		}
	case cc.ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}
		g.w("func() uintptr { *(*%s)(unsafe.Pointer(%s)) = ", g.typ(d.Type), g.mangleDeclarator(d))
		g.literal(d.Type, d.Initializer)
		g.w("; return %s }()", g.mangleDeclarator(d))
	case cc.ExprPExprList:
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.uintptrEscaped(l[0])
		default:
			todo("", g.position(n))
		}
	default:
		todo("", g.position(n), n.Case)
	}
} // uintptrEscaped

func (g *gen) voidCanIgnore(n *cc.Expr) bool {
	switch n.Case {
	case
		cc.ExprAlignofExpr, // "__alignof__" Expr
		cc.ExprAlignofType, // "__alignof__" '(' TypeName ')'
		cc.ExprChar,        // CHARCONST
		cc.ExprFloat,       // FLOATCONST
		cc.ExprIdent,       // IDENTIFIER
		cc.ExprInt,         // INTCONST
		cc.ExprSizeofExpr,  // "sizeof" Expr
		cc.ExprSizeofType,  // "sizeof" '(' TypeName ')'
		cc.ExprString:      // STRINGLITERAL

		return true
	case cc.ExprPExprList: // '(' ExprList ')'
		return g.voidCanIgnoreExprList(n.ExprList)
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		switch n.Expr.Case {
		case cc.ExprIdent:
			switch n.Expr.Token.Val {
			case idBuiltinTypesCompatible:
				return true
			}
		}
		return false
	case
		cc.ExprAddAssign, // Expr "+=" Expr
		cc.ExprAndAssign, // Expr "&=" Expr
		cc.ExprAssign,    // Expr '=' Expr
		cc.ExprDivAssign, // Expr "/=" Expr
		cc.ExprLshAssign, // Expr "<<=" Expr
		cc.ExprModAssign, // Expr "%=" Expr
		cc.ExprMulAssign, // Expr "*=" Expr
		cc.ExprOrAssign,  // Expr "|=" Expr
		cc.ExprPostDec,   // Expr "--"
		cc.ExprPostInc,   // Expr "++"
		cc.ExprPreDec,    // "--" Expr
		cc.ExprPreInc,    // "++" Expr
		cc.ExprRshAssign, // Expr ">>=" Expr
		cc.ExprSubAssign, // Expr "-=" Expr
		cc.ExprXorAssign: // Expr "^=" Expr

		return false
	case cc.ExprCast: // '(' TypeName ')' Expr
		return !isVaList(n.Expr.Operand.Type) && g.voidCanIgnore(n.Expr)
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		if !g.voidCanIgnore(n.Expr) {
			return false
		}

		switch {
		case n.Expr.IsNonZero():
			return g.voidCanIgnoreExprList(n.ExprList)
		case n.Expr.IsZero():
			return g.voidCanIgnore(n.Expr2)
		}
		return false
	case
		cc.ExprAdd, // Expr '+' Expr
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprEq,  // Expr "==" Expr
		cc.ExprGe,  // Expr ">=" Expr
		cc.ExprGt,  // Expr ">" Expr
		cc.ExprLe,  // Expr "<=" Expr
		cc.ExprLsh, // Expr "<<" Expr
		cc.ExprLt,  // Expr '<' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprNe,  // Expr "!=" Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprRsh, // Expr ">>" Expr
		cc.ExprSub, // Expr '-' Expr
		cc.ExprXor: // Expr '^' Expr

		return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
	case cc.ExprLAnd: // Expr "&&" Expr
		return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
	case cc.ExprLOr: // Expr "||" Expr
		return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
	case
		cc.ExprAddrof,     // '&' Expr
		cc.ExprCpl,        // '~' Expr
		cc.ExprDeref,      // '*' Expr
		cc.ExprNot,        // '!' Expr
		cc.ExprPSelect,    // Expr "->" IDENTIFIER
		cc.ExprSelect,     // Expr '.' IDENTIFIER
		cc.ExprUnaryMinus, // '-' Expr
		cc.ExprUnaryPlus:  // '+' Expr

		return g.voidCanIgnore(n.Expr)
	case cc.ExprIndex: // Expr '[' ExprList ']'
		return g.voidCanIgnore(n.Expr) && g.voidCanIgnoreExprList(n.ExprList)
	default:
		todo("", g.position0(n), n.Case, n.Operand) // voidCanIgnore
	}
	panic("unreachable")
}

func (g *ngen) voidCanIgnore(n *cc.Expr) bool {
	switch n.Case {
	case
		cc.ExprAlignofExpr, // "__alignof__" Expr
		cc.ExprAlignofType, // "__alignof__" '(' TypeName ')'
		cc.ExprChar,        // CHARCONST
		cc.ExprFloat,       // FLOATCONST
		cc.ExprIdent,       // IDENTIFIER
		cc.ExprInt,         // INTCONST
		cc.ExprLChar,       // LONGCHARCONST
		cc.ExprLString,     // LONGSTRINGLITERAL
		cc.ExprSizeofExpr,  // "sizeof" Expr
		cc.ExprSizeofType,  // "sizeof" '(' TypeName ')'
		cc.ExprString:      // STRINGLITERAL

		return true
	case cc.ExprPExprList: // '(' ExprList ')'
		return g.voidCanIgnoreExprList(n.ExprList)
	case cc.ExprCall: // Expr '(' ArgumentExprListOpt ')'
		switch n.Expr.Case {
		case cc.ExprIdent:
			switch n.Expr.Token.Val {
			case idBuiltinTypesCompatible:
				return true
			}
		}
		return false
	case
		cc.ExprAddAssign, // Expr "+=" Expr
		cc.ExprAndAssign, // Expr "&=" Expr
		cc.ExprAssign,    // Expr '=' Expr
		cc.ExprDivAssign, // Expr "/=" Expr
		cc.ExprLshAssign, // Expr "<<=" Expr
		cc.ExprModAssign, // Expr "%=" Expr
		cc.ExprMulAssign, // Expr "*=" Expr
		cc.ExprOrAssign,  // Expr "|=" Expr
		cc.ExprPostDec,   // Expr "--"
		cc.ExprPostInc,   // Expr "++"
		cc.ExprPreDec,    // "--" Expr
		cc.ExprPreInc,    // "++" Expr
		cc.ExprRshAssign, // Expr ">>=" Expr
		cc.ExprStatement, // '(' CompoundStmt ')' //TODO we can do better
		cc.ExprSubAssign, // Expr "-=" Expr
		cc.ExprXorAssign: // Expr "^=" Expr

		return false
	case cc.ExprCast: // '(' TypeName ')' Expr
		return !isVaList(n.Expr.Operand.Type) && g.voidCanIgnore(n.Expr)
	case
		cc.ExprAdd, // Expr '+' Expr
		cc.ExprAnd, // Expr '&' Expr
		cc.ExprDiv, // Expr '/' Expr
		cc.ExprEq,  // Expr "==" Expr
		cc.ExprGe,  // Expr ">=" Expr
		cc.ExprGt,  // Expr ">" Expr
		cc.ExprLe,  // Expr "<=" Expr
		cc.ExprLsh, // Expr "<<" Expr
		cc.ExprLt,  // Expr '<' Expr
		cc.ExprMod, // Expr '%' Expr
		cc.ExprMul, // Expr '*' Expr
		cc.ExprNe,  // Expr "!=" Expr
		cc.ExprOr,  // Expr '|' Expr
		cc.ExprRsh, // Expr ">>" Expr
		cc.ExprSub, // Expr '-' Expr
		cc.ExprXor: // Expr '^' Expr

		return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
	case cc.ExprCond: // Expr '?' ExprList ':' Expr
		switch {
		case n.Expr.IsZero():
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
		case n.Expr.IsNonZero():
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnoreExprList(n.ExprList)
		default:
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnoreExprList(n.ExprList) && g.voidCanIgnore(n.Expr2)
		}
	case cc.ExprLAnd: // Expr "&&" Expr
		switch {
		case n.Expr.IsZero():
			return g.voidCanIgnore(n.Expr)
		case n.Expr.IsNonZero():
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
		default:
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
		}
	case cc.ExprLOr: // Expr "||" Expr
		switch {
		case n.Expr.IsNonZero():
			return g.voidCanIgnore(n.Expr)
		case n.Expr.IsZero():
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
		default:
			return g.voidCanIgnore(n.Expr) && g.voidCanIgnore(n.Expr2)
		}
	case
		cc.ExprAddrof,     // '&' Expr
		cc.ExprCpl,        // '~' Expr
		cc.ExprDeref,      // '*' Expr
		cc.ExprNot,        // '!' Expr
		cc.ExprPSelect,    // Expr "->" IDENTIFIER
		cc.ExprSelect,     // Expr '.' IDENTIFIER
		cc.ExprUnaryMinus, // '-' Expr
		cc.ExprUnaryPlus:  // '+' Expr

		return g.voidCanIgnore(n.Expr)
	case cc.ExprIndex: // Expr '[' ExprList ']'
		return g.voidCanIgnore(n.Expr) && g.voidCanIgnoreExprList(n.ExprList)
	default:
		todo("", g.position(n), n.Case, n.Operand)
	}
	panic("unreachable")
} // voidCanIgnore

func (g *gen) voidCanIgnoreExprList(n *cc.ExprList) bool {
	if n.ExprList == nil {
		return g.voidCanIgnore(n.Expr)
	}

	for l := n; l != nil; l = l.ExprList {
		if !g.voidCanIgnore(l.Expr) {
			return false
		}
	}

	return true
}

func (g *ngen) voidCanIgnoreExprList(n *cc.ExprList) bool {
	if n == nil {
		return true
	}

	if n.ExprList == nil {
		return g.voidCanIgnore(n.Expr)
	}

	for l := n; l != nil; l = l.ExprList {
		if !g.voidCanIgnore(l.Expr) {
			return false
		}
	}

	return true
}

func (g *gen) constant(n *cc.Expr) {
	switch x := n.Operand.Value.(type) {
	case *ir.Float32Value:
		switch {
		case math.IsInf(float64(x.Value), 1):
			g.w("math.Inf(1)")
			return
		case math.IsInf(float64(x.Value), -1):
			g.w("math.Inf(-1)")
			return
		}
		switch u := cc.UnderlyingType(n.Operand.Type).(type) {
		case cc.TypeKind:
			switch u {
			case cc.Double:
				switch {
				case x.Value == 0 && math.Copysign(1, float64(x.Value)) == -1:
					g.w(" nz64")
					g.needNZ64 = true
				default:
					g.w(" %v", float64(x.Value))
				}
				return
			case cc.Float:
				switch {
				case x.Value == 0 && math.Copysign(1, float64(x.Value)) == -1:
					g.w(" nz32")
					g.needNZ32 = true
				default:
					g.w(" %v", x.Value)
				}
				return
			default:
				todo("", g.position0(n), u)
			}
		default:
			todo("%v: %T", g.position0(n), u)
		}
	case *ir.Float64Value:
		switch {
		case math.IsInf(x.Value, 1):
			g.w("math.Inf(1)")
			return
		case math.IsInf(x.Value, -1):
			g.w("math.Inf(-1)")
			return
		}

		switch u := cc.UnderlyingType(n.Operand.Type).(type) {
		case cc.TypeKind:
			if u.IsIntegerType() {
				g.w(" %v", cc.ConvertFloat64(x.Value, u, g.model))
				return
			}

			switch u {
			case
				cc.Double,
				cc.LongDouble:

				switch {
				case x.Value == 0 && math.Copysign(1, x.Value) == -1:
					g.w(" nz64")
					g.needNZ64 = true
				default:
					g.w(" %v", x.Value)
				}
				return
			case cc.Float:
				switch {
				case x.Value == 0 && math.Copysign(1, x.Value) == -1:
					g.w(" nz32")
					g.needNZ32 = true
				default:
					g.w(" %v", float32(x.Value))
				}
				return
			default:
				todo("", g.position0(n), u)
			}
		default:
			todo("%v: %T", g.position0(n), u)
		}
	case *ir.Int64Value:
		if n.Case == cc.ExprChar {
			g.w(" %s", strconv.QuoteRuneToASCII(rune(x.Value)))
			return
		}

		f := " %d"
		m := n
		s := ""
		for done := false; !done; { //TODO-
			switch m.Case {
			case cc.ExprInt: // INTCONST
				s = string(m.Token.S())
				done = true
			case
				cc.ExprCast,       // '(' TypeName ')' Expr
				cc.ExprUnaryMinus: // '-' Expr

				m = m.Expr
			default:
				done = true
			}
		}
		s = strings.ToLower(s)
		switch {
		case strings.HasPrefix(s, "0x"):
			f = "%#x"
		case strings.HasPrefix(s, "0"):
			f = "%#o"
		}

		switch y := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			if n.IsZero() {
				g.w("%s", null)
				return
			}

			switch {
			case y.Item.Kind() == cc.Function:
				g.w("uintptr(%v)", uintptr(x.Value))
			default:
				g.w("uintptr("+f+")", uintptr(x.Value))
			}
			return
		}

		switch {
		case n.Operand.Type.IsUnsigned():
			g.w(f, uint64(cc.ConvertInt64(x.Value, n.Operand.Type, g.model)))
		default:
			g.w(f, cc.ConvertInt64(x.Value, n.Operand.Type, g.model))
		}
	case *ir.StringValue:
		g.w(" ts+%d %s", g.allocString(int(x.StringID)), strComment(x))
	case *ir.AddressValue:
		if x == cc.Null {
			g.w("%s", null)
			return
		}

		todo("", g.position0(n))
	default:
		todo("%v: %v %T(%v)", g.position0(n), n.Operand, x, x)
	}
} // constant

func (g *ngen) constant(n *cc.Expr) {
	switch x := n.Operand.Value.(type) {
	case *ir.Float32Value:
		switch {
		case math.IsInf(float64(x.Value), 1):
			g.w("math.Inf(1)")
			return
		case math.IsInf(float64(x.Value), -1):
			g.w("math.Inf(-1)")
			return
		case math.IsNaN(float64(x.Value)):
			g.w("math.NaN()")
			return
		}
		switch u := cc.UnderlyingType(n.Operand.Type).(type) {
		case cc.TypeKind:
			switch u {
			case
				cc.Double,
				cc.LongDouble:

				switch {
				case x.Value == 0 && math.Copysign(1, float64(x.Value)) == -1:
					g.w(" %sNz64", g.crtPrefix)
					g.needNZ64 = true
				default:
					g.w(" %v", float64(x.Value))
				}
				return
			case cc.Float:
				switch {
				case x.Value == 0 && math.Copysign(1, float64(x.Value)) == -1:
					g.w(" %sNz32", g.crtPrefix)
					g.needNZ32 = true
				default:
					g.w(" %v", x.Value)
				}
				return
			default:
				todo("", g.position(n), u)
			}
		default:
			todo("%v: %T", g.position(n), u)
		}
	case *ir.Float64Value:
		switch {
		case math.IsInf(x.Value, 1):
			g.w("math.Inf(1)")
			return
		case math.IsInf(x.Value, -1):
			g.w("math.Inf(-1)")
			return
		case math.IsNaN(x.Value):
			g.w("math.NaN()")
			return
		}

		switch u := cc.UnderlyingType(n.Operand.Type).(type) {
		case cc.TypeKind:
			if u.IsIntegerType() {
				g.w(" %v", cc.ConvertFloat64(x.Value, u, g.model))
				return
			}

			switch u {
			case
				cc.Double,
				cc.LongDouble:

				switch {
				case x.Value == 0 && math.Copysign(1, x.Value) == -1:
					g.w(" %sNz64", g.crtPrefix)
					g.needNZ64 = true
				default:
					g.w(" %v", x.Value)
				}
				return
			case cc.Float:
				switch {
				case x.Value == 0 && math.Copysign(1, x.Value) == -1:
					g.w(" %sNz32", g.crtPrefix)
					g.needNZ32 = true
				default:
					g.w(" %v", float32(x.Value))
				}
				return
			default:
				todo("", g.position(n), u)
			}
		default:
			todo("%v: %T", g.position(n), u)
		}
	case *ir.Int64Value:
		if n.Case == cc.ExprChar {
			g.w(" %s", strconv.QuoteRuneToASCII(rune(x.Value)))
			return
		}

		f := " %d"
		m := n
		s := ""
		for done := false; !done; { //TODO-
			switch m.Case {
			case cc.ExprInt: // INTCONST
				s = string(m.Token.S())
				done = true
			case
				cc.ExprCast,       // '(' TypeName ')' Expr
				cc.ExprUnaryMinus: // '-' Expr

				m = m.Expr
			default:
				done = true
			}
		}
		s = strings.ToLower(s)
		switch {
		case strings.HasPrefix(s, "0x"):
			f = "%#x"
		case strings.HasPrefix(s, "0"):
			f = "%#o"
		}

		switch y := cc.UnderlyingType(n.Operand.Type).(type) {
		case *cc.PointerType:
			if n.IsZero() && g.voidCanIgnore(n) {
				g.w("%s", null)
				return
			}

			switch {
			case y.Item.Kind() == cc.Function:
				g.w("uintptr(%v)", uintptr(x.Value))
			default:
				g.w("uintptr("+f+")", uintptr(x.Value))
			}
			return
		}

		switch k := n.Operand.Type; {
		case k == cc.DoubleComplex:
			g.w("complex(float64(%v), 0)", x.Value)
		case n.Operand.Type.IsUnsigned():
			g.w(f, uint64(cc.ConvertInt64(x.Value, n.Operand.Type, g.model)))
		default:
			g.w(f, cc.ConvertInt64(x.Value, n.Operand.Type, g.model))
		}
	case *ir.StringValue:
		g.w(" %q", dict.S(int(x.StringID)))
	case *ir.WideStringValue:
		wsz := int(g.model.Sizeof(n.Operand.Type.(*cc.PointerType).Item))
		b := make([]byte, len(x.Value)*wsz)
		for i, v := range x.Value {
			switch wsz {
			case 2:
				*(*int16)(unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(i*wsz))) = int16(v)
			case 4:
				*(*rune)(unsafe.Pointer(uintptr(unsafe.Pointer(&b[0])) + uintptr(i*wsz))) = v
			default:
				todo("", g.position(n))
			}
		}
		g.w(" Lw +%q", b)
	case *ir.AddressValue:
		if x == cc.Null {
			g.w("%s", null)
			return
		}

		todo("", g.position(n))
	default:
		todo("%v: %v %T(%v)", g.position(n), n.Operand, x, x)
	}
} // constant

func (g *gen) voidArithmeticAsop(n *cc.Expr) {
	var mask uint64
	op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
	lhs := n.Expr.Operand
	switch {
	case lhs.Bits() != 0:
		fp := lhs.FieldProperties
		mask = (uint64(1)<<uint(fp.Bits) - 1) << uint(fp.Bitoff)
		g.w("{ p := &")
		g.value(n.Expr, true)
		sz := int(g.model.Sizeof(lhs.Type) * 8)
		g.w("; *p = (*p &^ %#x) | (%s((%s(*p>>%d)<<%d>>%[5]d) ", mask, g.typ(fp.PackedType), g.typ(op.Type), fp.Bitoff, sz-fp.Bits)
	case n.Expr.Declarator != nil:
		g.w(" *(")
		g.lvalue(n.Expr)
		g.w(") = %s(", g.typ(n.Expr.Operand.Type))
		g.convert(n.Expr, op.Type)
	default:
		g.w("{ p := ")
		g.lvalue(n.Expr)
		g.w("; *p = %s(%s(*p)", g.typ(n.Expr.Operand.Type), g.typ(op.Type))
	}
	switch n.Token.Rune {
	case cc.ANDASSIGN:
		g.w("&")
	case cc.ADDASSIGN:
		g.w("+")
	case cc.SUBASSIGN:
		g.w("-")
	case cc.MULASSIGN:
		g.w("*")
	case cc.DIVASSIGN:
		g.w("/")
	case cc.ORASSIGN:
		g.w("|")
	case cc.RSHASSIGN:
		g.w(">>")
		op.Type = cc.UInt
	case cc.XORASSIGN:
		g.w("^")
	case cc.MODASSIGN:
		g.w("%%")
	case cc.LSHASSIGN:
		g.w("<<")
		op.Type = cc.UInt
	default:
		todo("", g.position0(n), cc.TokSrc(n.Token))
	}
	if n.Expr.Operand.Bits() != 0 {
		g.w("(")
	}
	g.convert(n.Expr2, op.Type)
	switch {
	case lhs.Bits() != 0:
		g.w("))<<%d&%#x) }", lhs.FieldProperties.Bitoff, mask)
	case n.Expr.Declarator != nil:
		g.w(")")
	default:
		g.w(")}")
	}
}

func (g *ngen) voidArithmeticAsop(n *cc.Expr) {
	var mask uint64
	op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
	lhs := n.Expr.Operand
	switch {
	case lhs.Bits() != 0:
		fp := lhs.FieldProperties
		mask = (uint64(1)<<uint(fp.Bits) - 1) << uint(fp.Bitoff)
		g.w("{ p := &")
		g.value(n.Expr, true)
		bits := int(g.model.Sizeof(fp.Type) * 8)
		g.w("; *p = (*p &^ %#x) | (%s((%s(%s(*p>>%d)<<%d>>%[6]d)) ", mask, g.typ(fp.PackedType), g.typ(op.Type), g.typ(fp.Type), fp.Bitoff, bits-fp.Bits)
	case n.Expr.Declarator != nil:
		g.w(" *(")
		g.lvalue(n.Expr)
		g.w(") = %s(", g.typ(n.Expr.Operand.Type))
		g.convert(n.Expr, op.Type)
	default:
		g.w("{ p := ")
		g.lvalue(n.Expr)
		g.w("; *p = %s(%s(*p)", g.typ(n.Expr.Operand.Type), g.typ(op.Type))
	}
	switch n.Token.Rune {
	case cc.ANDASSIGN:
		g.w("&")
	case cc.ADDASSIGN:
		g.w("+")
	case cc.SUBASSIGN:
		g.w("-")
	case cc.MULASSIGN:
		g.w("*")
	case cc.DIVASSIGN:
		g.w("/")
	case cc.ORASSIGN:
		g.w("|")
	case cc.RSHASSIGN:
		g.w(">>")
		op.Type = cc.UInt
	case cc.XORASSIGN:
		g.w("^")
	case cc.MODASSIGN:
		g.w("%%")
	case cc.LSHASSIGN:
		g.w("<<")
		op.Type = cc.UInt
	default:
		todo("", g.position(n), cc.TokSrc(n.Token))
	}
	if n.Expr.Operand.Bits() != 0 {
		g.w("(")
	}
	g.convert(n.Expr2, op.Type)
	switch {
	case lhs.Bits() != 0:
		g.w("))<<%d&%#x) }", lhs.FieldProperties.Bitoff, mask)
	case n.Expr.Declarator != nil:
		g.w(")")
	default:
		g.w(")}")
	}
}

func (g *gen) assignmentValue(n *cc.Expr) {
	switch op := n.Expr.Operand; {
	case op.Bits() != 0:
		fp := op.FieldProperties
		g.w("%s(&", g.registerHelper("set%db", "=", g.typ(op.Type), g.typ(fp.PackedType), g.model.Sizeof(op.Type)*8, fp.Bits, fp.Bitoff))
		g.value(n.Expr, true)
		g.w(", ")
		g.convert(n.Expr2, n.Operand.Type)
		g.w(")")
	default:
		g.w("%s(", g.registerHelper("set%d", "", g.typ(op.Type)))
		g.lvalue(n.Expr)
		g.w(", ")
		g.convert(n.Expr2, n.Operand.Type)
		g.w(")")
	}
}

func (g *ngen) assignmentValue(n *cc.Expr) {
	switch op := n.Expr.Operand; {
	case op.Bits() != 0:
		fp := op.FieldProperties
		g.w("%s(&", g.registerHelper("setb%d", g.typ(fp.PackedType), g.typ(op.Type), g.typ(n.Expr2.Operand.Type), fp.Bitoff, fp.Bits, g.model.Sizeof(op.Type)*8))
		g.value(n.Expr, true)
		g.w(", ")
		g.value(n.Expr2, false)
		g.w(")")
	default:
		g.w("%s(", g.registerHelper("set%d", "", g.typ(op.Type)))
		g.lvalue(n.Expr)
		g.w(", ")
		g.convert(n.Expr2, n.Operand.Type)
		g.w(")")
	}
}

func (g *gen) binop(n *cc.Expr) {
	l, r := n.Expr.Operand.Type, n.Expr2.Operand.Type
	if l.IsArithmeticType() && r.IsArithmeticType() {
		op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
		l, r = op.Type, op.Type
	}
	switch {
	case
		l.Kind() == cc.Ptr && n.Operand.Type.IsArithmeticType(),
		n.Operand.Type.Kind() == cc.Ptr && l.IsArithmeticType():

		g.convert(n.Expr, n.Operand.Type)
	default:
		g.convert(n.Expr, l)
	}
	g.w(" %s ", cc.TokSrc(n.Token))
	switch {
	case
		r.Kind() == cc.Ptr && n.Operand.Type.IsArithmeticType(),
		n.Operand.Type.Kind() == cc.Ptr && r.IsArithmeticType():

		g.convert(n.Expr2, n.Operand.Type)
	default:
		g.convert(n.Expr2, r)
	}
}

func (g *ngen) binop(n *cc.Expr) {
	l, r := n.Expr.Operand.Type, n.Expr2.Operand.Type
	if l.IsArithmeticType() && r.IsArithmeticType() {
		op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
		l, r = op.Type, op.Type
	}
	switch {
	case
		l.Kind() == cc.Ptr && n.Operand.Type.IsArithmeticType(),
		n.Operand.Type.Kind() == cc.Ptr && l.IsArithmeticType():

		g.convert(n.Expr, n.Operand.Type)
	default:
		g.convert(n.Expr, l)
	}
	g.w(" %s ", cc.TokSrc(n.Token))
	switch {
	case
		r.Kind() == cc.Ptr && n.Operand.Type.IsArithmeticType(),
		n.Operand.Type.Kind() == cc.Ptr && r.IsArithmeticType():

		g.convert(n.Expr2, n.Operand.Type)
	default:
		g.convert(n.Expr2, r)
	}
}

func (g *gen) relop(n *cc.Expr) {
	g.needBool2int++
	g.w(" bool2int(")
	l, r := n.Expr.Operand.Type, n.Expr2.Operand.Type
	if l.IsArithmeticType() && r.IsArithmeticType() {
		op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
		l, r = op.Type, op.Type
	}
	switch {
	case l.Kind() == cc.Ptr || r.Kind() == cc.Ptr:
		g.value(n.Expr, false)
		g.w(" %s ", cc.TokSrc(n.Token))
		g.value(n.Expr2, false)
		g.w(")")
	default:
		g.convert(n.Expr, l)
		g.w(" %s ", cc.TokSrc(n.Token))
		g.convert(n.Expr2, r)
		g.w(")")
	}
}

func (g *ngen) relop(n *cc.Expr) {
	g.w(" bool2int(")
	l, r := n.Expr.Operand.Type, n.Expr2.Operand.Type
	if l.IsArithmeticType() && r.IsArithmeticType() {
		op, _ := cc.UsualArithmeticConversions(g.model, n.Expr.Operand, n.Expr2.Operand)
		l, r = op.Type, op.Type
	}
	switch {
	case l.Kind() == cc.Ptr || r.Kind() == cc.Ptr:
		g.value(n.Expr, false)
		g.w(" %s ", cc.TokSrc(n.Token))
		g.value(n.Expr2, false)
		g.w(")")
	default:
		g.convert(n.Expr, l)
		g.w(" %s ", cc.TokSrc(n.Token))
		g.convert(n.Expr2, r)
		g.w(")")
	}
}

func (g *gen) convert(n *cc.Expr, t cc.Type) {
	if n.Case == cc.ExprPExprList {
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.convert(l[0], t)
		default:
			g.w("func() %v {", g.typ(t))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], t)
			g.w("}()")
		}
		return
	}

	if t.Kind() == cc.Function {
		ft := cc.UnderlyingType(t)
		switch n.Case {
		case cc.ExprIdent: // IDENTIFIER
			d := g.normalizeDeclarator(n.Declarator)
			g.enqueue(d)
			dt := cc.UnderlyingType(d.Type)
			if dt.Equal(ft) {
				g.w("%s", g.mangleDeclarator(d))
				return
			}

			if cc.UnderlyingType(n.Operand.Type).Equal(&cc.PointerType{Item: ft}) {
				switch {
				case d.Type.Kind() == cc.Ptr:
					if g.escaped(d) {
						g.w("%s(*(*uintptr)(unsafe.Pointer(%s)))", g.registerHelper("fn%d", g.typ(ft)), g.mangleDeclarator(n.Declarator))
						break
					}

					g.w("%s(%s)", g.registerHelper("fn%d", g.typ(ft)), g.mangleDeclarator(n.Declarator))
				default:
					g.w("%s", g.mangleDeclarator(n.Declarator))
				}
				return
			}

			todo("", g.position0(n))
		case cc.ExprCast: // '(' TypeName ')' Expr
			if d := g.normalizeDeclarator(n.Expr.Declarator); d != nil {
				g.enqueue(d)
				if d.Type.Equal(t) {
					g.w("%s", g.mangleDeclarator(d))
					return
				}

				g.w("%s(%s(%s))", g.registerHelper("fn%d", g.typ(t)), g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
				return
			}

			g.w("%s(", g.registerHelper("fn%d", g.typ(ft)))
			g.value(n, false)
			g.w(")")
		default:
			g.w("%s(", g.registerHelper("fn%d", g.typ(t)))
			g.value(n, false)
			g.w(")")
		}
		return
	}

	if isVaList(n.Operand.Type) && !isVaList(t) {
		g.w("%sVA%s(", crt, g.typ(cc.UnderlyingType(t)))
		g.value(n, false)
		g.w(")")
		return
	}

	if t.Kind() == cc.Ptr {
		switch {
		case n.Operand.Value != nil && isVaList(t):
			g.w("%s", ap)
		case n.Operand.Type.Kind() == cc.Ptr:
			g.value(n, false)
		case isVaList(t):
			switch x := n.Operand.Value.(type) {
			case *ir.Int64Value:
				if x.Value == 1 {
					g.w("%s", ap)
					return
				}
			default:
				todo("%v, %T, %v %v -> %v", g.position0(n), x, n.Case, n.Operand, t)
			}
			todo("", g.position0(n))
		case n.Operand.Type.IsIntegerType():
			if n.Operand.Value != nil && g.voidCanIgnore(n) {
				t0 := n.Operand.Type
				n.Operand.Type = t
				g.constant(n)
				n.Operand.Type = t0
				return
			}

			g.w(" uintptr(")
			g.value(n, false)
			g.w(")")
		default:
			todo("%v: %v -> %v, %T, %v", g.position0(n), n.Operand, t, t, cc.UnderlyingType(t))
		}
		return
	}

	if n.Operand.Type.Equal(t) {
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			g.w(" %s(", g.typ(t))
			g.constant(n)
			g.w(")")
		default:
			g.value(n, false)
		}
		return
	}

	if cc.UnderlyingType(t).IsArithmeticType() {
		if n.Operand.Value == nil && t.IsIntegerType() {
			switch n.Operand.Type.Kind() {
			case cc.Float:
				switch {
				case t.IsUnsigned():
					switch g.model.Sizeof(t) {
					case 8:
						g.w("%s(", g.registerHelper("float2int%d", g.typ(t), math.Nextafter32(math.MaxUint64, 0)))
						g.value(n, false)
						g.w(")")
						return
					}
				}
			}
		}

		g.w(" %s(", g.typ(t))

		defer g.w(")")

		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			if n.Operand.Type.Kind() == cc.Double && t.IsIntegerType() {
				v := cc.ConvertFloat64(n.Operand.Value.(*ir.Float64Value).Value, t, g.model)
				switch {
				case t.IsUnsigned():
					g.w("%v", uint64(v))
				default:
					g.w("%v", v)
				}
				return
			}

			t0 := n.Operand.Type
			n.Operand.Type = t
			g.constant(n)
			n.Operand.Type = t0
		default:
			g.value(n, false)
		}
		return
	}

	todo("%v: %v -> %v, %T, %v", g.position0(n), n.Operand, t, t, cc.UnderlyingType(t))
}

func (g *ngen) convert(n *cc.Expr, t cc.Type) {
	if g.escaped(n.Declarator) {
		g.convertEscaped(n, t)
		return
	}

	if n.Case == cc.ExprPExprList {
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.convert(l[0], t)
		default:
			g.w("func() %v {", g.typ(t))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], t)
			g.w("}()")
		}
		return
	}

	if t.Kind() == cc.Function {
		ft := cc.UnderlyingType(t)
		switch n.Case {
		case cc.ExprIdent: // IDENTIFIER
			d := n.Declarator
			g.enqueue(d)
			dt := cc.UnderlyingType(d.Type)
			if dt.Equal(ft) {
				g.w("%s", g.mangleDeclarator(d))
				return
			}

			if cc.UnderlyingType(n.Operand.Type).Equal(&cc.PointerType{Item: ft}) {
				switch {
				case d.Type.Kind() == cc.Ptr:
					if g.escaped(d) {
						g.w("%s(*(*uintptr)(unsafe.Pointer(%s)))", g.registerHelper("fn%d", g.typ(ft)), g.mangleDeclarator(n.Declarator))
						break
					}

					g.w("%s(%s)", g.registerHelper("fn%d", g.typ(ft)), g.mangleDeclarator(n.Declarator))
				default:
					g.w("%s", g.mangleDeclarator(n.Declarator))
				}
				return
			}

			todo("", g.position(n))
		case cc.ExprCast: // '(' TypeName ')' Expr
			if d := n.Expr.Declarator; d != nil {
				g.enqueue(d)
				if d.Type.Equal(t) {
					g.w("%s", g.mangleDeclarator(d))
					return
				}

				g.w("%s(%s(%s))", g.registerHelper("fn%d", g.typ(t)), g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(d))
				return
			}

			g.w("%s(", g.registerHelper("fn%d", g.typ(ft)))
			g.value(n, false)
			g.w(")")
		default:
			g.w("%s(", g.registerHelper("fn%d", g.typ(t)))
			g.value(n, false)
			g.w(")")
		}
		return
	}

	//TODO- if isVaList(n.Operand.Type) && !isVaList(t) {
	//TODO- 	g.w("%sVA%s(", g.crtPrefix, g.typ(cc.UnderlyingType(t)))
	//TODO- 	g.value(n, false)
	//TODO- 	g.w(")")
	//TODO- 	return
	//TODO- }

	if t.Kind() == cc.Ptr {
		switch {
		//TODO- case n.Operand.Value != nil && isVaList(t):
		//TODO- 	g.w("%s", ap)
		case n.Operand.Type.Kind() == cc.Ptr:
			g.value(n, false)
		//TODO- case isVaList(t):
		//TODO- 	switch x := n.Operand.Value.(type) {
		//TODO- 	case *ir.Int64Value:
		//TODO- 		if x.Value == 1 {
		//TODO- 			g.w("%s", ap)
		//TODO- 			return
		//TODO- 		}
		//TODO- 	default:
		//TODO- 		todo("%v, %T, %v %v -> %v", g.position(n), x, n.Case, n.Operand, t)
		//TODO- 	}
		//TODO- 	todo("", g.position(n))
		case n.Operand.Type.IsIntegerType():
			if n.Operand.Value != nil && g.voidCanIgnore(n) {
				t0 := n.Operand.Type
				n.Operand.Type = t
				g.constant(n)
				n.Operand.Type = t0
				return
			}

			g.w(" uintptr(")
			g.value(n, false)
			g.w(")")
		default:
			todo("%v: %v -> %v, %T, %v", g.position(n), n.Operand, t, t, cc.UnderlyingType(t))
		}
		return
	}

	ut := cc.UnderlyingType(t)
	if n.Operand.Type.Equal(t) {
		switch {
		case n.Operand.Value != nil && g.voidCanIgnore(n):
			g.w(" %s(", g.typ(t))
			g.constant(n)
			g.w(")")
			return
		case !ut.IsArithmeticType():
			g.value(n, false)
			return
		}
	}

	if ut.IsArithmeticType() {
		g.convert2ArithmeticType(n, t)
		return
	}

	todo("%v: %v -> %v, %T, %v", g.position(n), n.Operand, t, t, cc.UnderlyingType(t))
}

func (g *ngen) convert2ArithmeticType(n *cc.Expr, t cc.Type) {
	if n.Operand.Value == nil && t.IsIntegerType() {
		switch n.Operand.Type.Kind() {
		case cc.Float:
			switch {
			case t.IsUnsigned():
				switch g.model.Sizeof(t) {
				case 8:
					g.w("%s(", g.registerHelper("float2int%d", g.typ(t), math.Nextafter32(math.MaxUint64, 0)))
					g.value(n, false)
					g.w(")")
					return
				}
			}
		}
	}

	more := ""
	switch un, ut := cc.UnderlyingType(n.Operand.Type), cc.UnderlyingType(t); {
	case un.Kind() == cc.Float:
		switch ut.Kind() {
		case cc.FloatComplex:

			g.w(" complex(")
			more = ", 0"
		default:
			g.w(" %s(", g.typ(t))
		}
	case un.Kind() == cc.Double, un.Kind() == cc.LongDouble:
		switch ut.Kind() {
		case cc.DoubleComplex, cc.LongDoubleComplex:
			g.w(" complex(")
			more = ", 0"
		default:
			g.w(" %s(", g.typ(t))
		}
	case un.Kind() == cc.FloatComplex:
		switch ut.Kind() {
		case cc.Float:
			g.w(" real(")
		default:
			g.w(" %s(", g.typ(t))
		}
	case un.Kind() == cc.DoubleComplex, un.Kind() == cc.LongDoubleComplex:
		switch ut.Kind() {
		case cc.Double, cc.LongDouble:
			g.w(" real(")
		default:
			g.w(" %s(", g.typ(t))
		}
	case un.IsIntegerType():
		switch ut.Kind() {
		case cc.FloatComplex:
			g.w(" complex(float32(")
			more = "), 0"
		case cc.DoubleComplex, cc.LongDoubleComplex:
			g.w(" complex(float64(")
			more = "), 0"
		default:
			g.w(" %s(", g.typ(t))
		}
	default:
		g.w(" %s(", g.typ(t))
	}

	defer g.w("%s)", more)

	switch {
	case n.Operand.Value != nil && g.voidCanIgnore(n):
		if n.Operand.Type.Kind() == cc.Double && t.IsIntegerType() {
			v := cc.ConvertFloat64(n.Operand.Value.(*ir.Float64Value).Value, t, g.model)
			switch {
			case t.IsUnsigned():
				g.w("%v", uint64(v))
			default:
				g.w("%v", v)
			}
			return
		}

		t0 := n.Operand.Type
		n.Operand.Type = t
		g.constant(n)
		n.Operand.Type = t0
	default:
		g.value(n, false)
	}
}

func (g *ngen) convertEscaped(n *cc.Expr, t cc.Type) {
	d := n.Declarator
	g.enqueue(d)
	switch n.Case {
	case cc.ExprIdent: // IDENTIFIER
		switch x := underlyingType(d.Type, false).(type) {
		case *cc.ArrayType:
			if t.Kind() == cc.Ptr {
				g.w("%s ", g.mangleDeclarator(d))
				return
			}

			if t.IsIntegerType() {
				g.w("%s(%s) ", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.FunctionType: // d is a function declarator.
			if d.Type.Equal(t) {
				g.w("%s ", g.mangleDeclarator(d))
				return
			}

			if t.Kind() == cc.Ptr {
				g.w("%s(%s)", g.registerHelper("fp%d", g.typ(d.Type)), g.mangleDeclarator(n.Declarator))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.PointerType:
			if x.Item.Kind() == cc.Function && x.Item.Equal(t) {
				g.w("%s(*(*uintptr)(unsafe.Pointer(%s)))", g.registerHelper("fn%d", g.typ(t)), g.mangleDeclarator(n.Declarator))
				return
			}

			if t.Kind() == cc.Ptr {
				g.w(" *(*uintptr)(unsafe.Pointer(%s))", g.mangleDeclarator(d))
				return
			}

			if t.IsIntegerType() {
				g.w(" %s(*(*uintptr)(unsafe.Pointer(%s)))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.StructType:
			if d.Type.Equal(t) {
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.TaggedStructType:
			if d.Type.Equal(t) {
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.TaggedUnionType:
			if d.Type.Equal(t) {
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case cc.TypeKind:
			if d.Type.Equal(t) {
				g.value(n, false)
				return
			}

			if t.IsArithmeticType() {
				g.convert2ArithmeticType(n, t)
				return
			}

			if t.Kind() == cc.Ptr {
				g.w(" uintptr(*(*%s)(unsafe.Pointer(%s)))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		case *cc.UnionType:
			if d.Type.Equal(t) {
				g.w(" *(*%s)(unsafe.Pointer(%s))", g.typ(t), g.mangleDeclarator(d))
				return
			}

			todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		default:
			todo("%v: %T, %v, op %v, d %v, t %v, %q %v:", g.position(n), x, n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
		}
	case cc.ExprPExprList:
		switch l := g.pexprList(n.ExprList); {
		case len(l) == 1:
			g.convert(l[0], t)
		default:
			g.w("func() %v {", g.typ(t))
			for _, v := range l[:len(l)-1] {
				g.void(v)
				g.w(";")
			}
			g.w("return ")
			g.convert(l[len(l)-1], t)
			g.w("}()")
		}
	default:
		todo("%v: %v, op %v, d %v, t %v, %q %v:", g.position(n), n.Case, n.Operand.Type, d.Type, t, dict.S(d.Name()), g.position(d))
	}

}

func (g *gen) int64ToUintptr(n int64) uint64 {
	switch g.model[cc.Ptr].Size {
	case 4:
		return uint64(uint32(n))
	case 8:
		return uint64(n)
	}
	panic("unreachable")
}

func (g *ngen) int64ToUintptr(n int64) uint64 {
	switch g.model[cc.Ptr].Size {
	case 4:
		return uint64(uint32(n))
	case 8:
		return uint64(n)
	}
	panic("unreachable")
}

func (g *gen) convertInt64(n int64, t cc.Type) string {
	v := cc.ConvertInt64(n, t, g.model)
	switch {
	case t.IsUnsigned():
		return fmt.Sprint(uint64(v))
	default:
		return fmt.Sprint(v)
	}
}

func (g *ngen) convertInt64(n int64, t cc.Type) string {
	v := cc.ConvertInt64(n, t, g.model)
	switch {
	case t.IsUnsigned():
		return fmt.Sprint(uint64(v))
	default:
		return fmt.Sprint(v)
	}
}
