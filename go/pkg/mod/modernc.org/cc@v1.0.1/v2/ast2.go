// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"fmt"
	"go/token"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"modernc.org/ir"
	"modernc.org/mathutil"
)

// Node represents an AST node.
type Node interface {
	Pos() token.Pos
}

// Pos implements Node.
func (n *TranslationUnit) Pos() token.Pos { return token.Pos(0) }

func (d *Declarator) IsVolatile() bool { return d != nil && d.DeclarationSpecifier.IsVolatile() }

// DeclarationSpecifier describes declaration specifiers.
type DeclarationSpecifier struct {
	Parent                 *DeclarationSpecifier
	AlignmentSpecifiers    []*AlignmentSpecifier
	FunctionSpecifiers     []*FunctionSpecifier
	StorageClassSpecifiers []*StorageClassSpecifier
	TypeQualifiers         []*TypeQualifier
	TypeSpecifiers         []*TypeSpecifier
	typeSpecifiers         []TypeSpecifierCase
}

func (d *DeclarationSpecifier) typ(ctx *context) Type {
	if d.typeSpecifiers == nil {
		d.typeSpecifiers = make([]TypeSpecifierCase, len(d.TypeSpecifiers))
		for i, v := range d.TypeSpecifiers {
			d.typeSpecifiers[i] = v.Case
		}
		sort.Slice(d.typeSpecifiers, func(i, j int) bool { return d.typeSpecifiers[i] < d.typeSpecifiers[j] })
	}

	// [0]6.7.2-2
	if len(d.typeSpecifiers) == 1 {
		switch d.typeSpecifiers[0] {
		case -1:
			return d.TypeSpecifiers[0].typ
		case TypeSpecifierBool:
			return Bool
		case TypeSpecifierChar:
			return Char
		case TypeSpecifierDouble:
			return Double
		case TypeSpecifierFloat:
			return Float
		case TypeSpecifierInt:
			return Int
		case TypeSpecifierLong:
			return Long
		case TypeSpecifierName:
			ts := d.TypeSpecifiers[0]
			r := &NamedType{Name: ts.Token.Val}
			switch x := ts.scope.LookupIdent(ts.Token.Val).(type) {
			case *Declarator:
				if !x.DeclarationSpecifier.IsTypedef() {
					//dbg("", ctx.position(x))
					panic("internal error 1")
				}

				r.Type = x.Type
				t := r.Type
				for {
					switch x := t.(type) {
					case
						*ArrayType,
						*EnumType,
						*FunctionType,
						*StructType,
						*UnionType:

						return r
					case *NamedType:
						t = x.Type
					case *PointerType:
						t = x.Item
					case *TaggedStructType:
						x.getType()
						return r
					case *TaggedUnionType:
						x.getType()
						return r
					case *TaggedEnumType:
						x.getType()
						return r
					case TypeKind:
						switch x {
						case
							Char,
							Double,
							Float,
							Int,
							Long,
							LongDouble,
							LongLong,
							SChar,
							Short,
							UChar,
							UInt,
							ULong,
							ULongLong,
							UShort,
							Void:

							return r
						default:
							panic(x)
						}
					default:
						panic(fmt.Errorf("%T %v", x, x))
					}
				}
			default:
				panic(fmt.Errorf("%T", x))
			}
		case TypeSpecifierShort:
			return Short
		case TypeSpecifierStruct:
			t := d.TypeSpecifiers[0].StructOrUnionSpecifier.typ
			switch x := t.(type) {
			case *TaggedStructType:
				x.getType()
			case *TaggedUnionType:
				x.getType()
			}
			return t
		case TypeSpecifierSigned:
			return Int
		case TypeSpecifierUnsigned:
			return UInt
		case TypeSpecifierVoid:
			return Void
		case TypeSpecifierEnum:
			t := d.TypeSpecifiers[0].EnumSpecifier.typ
			if x, ok := t.(*TaggedEnumType); ok {
				x.getType()
			}
			return t
		case
			TypeSpecifierTypeof,
			TypeSpecifierTypeofExpr:

			return d.TypeSpecifiers[0].typ
		default:
			panic(d.typeSpecifiers)
		}
	}

	switch {
	case d.is(TypeSpecifierChar, TypeSpecifierSigned):
		return SChar
	case d.is(TypeSpecifierChar, TypeSpecifierUnsigned):
		return UChar
	case d.is(TypeSpecifierComplex, TypeSpecifierDouble):
		return DoubleComplex
	case d.is(TypeSpecifierComplex, TypeSpecifierDouble, TypeSpecifierLong):
		return LongDoubleComplex
	case d.is(TypeSpecifierComplex, TypeSpecifierFloat):
		return FloatComplex
	case d.is(TypeSpecifierDouble, TypeSpecifierLong):
		return LongDouble
	case d.is(TypeSpecifierImaginary, TypeSpecifierDouble):
		return DoubleImaginary
	case d.is(TypeSpecifierImaginary, TypeSpecifierFloat):
		return FloatImaginary
	case d.is(TypeSpecifierImaginary, TypeSpecifierDouble, TypeSpecifierLong):
		return LongDoubleImaginary
	case d.is(TypeSpecifierInt, TypeSpecifierLong):
		return Long
	case d.is(TypeSpecifierInt, TypeSpecifierLong, TypeSpecifierLong):
		return LongLong
	case d.is(TypeSpecifierInt, TypeSpecifierLong, TypeSpecifierLong, TypeSpecifierUnsigned):
		return ULongLong
	case d.is(TypeSpecifierInt, TypeSpecifierLong, TypeSpecifierSigned):
		return Long
	case d.is(TypeSpecifierInt, TypeSpecifierLong, TypeSpecifierLong, TypeSpecifierSigned):
		return LongLong
	case d.is(TypeSpecifierInt, TypeSpecifierLong, TypeSpecifierUnsigned):
		return ULong
	case d.is(TypeSpecifierInt, TypeSpecifierShort):
		return Short
	case d.is(TypeSpecifierInt, TypeSpecifierShort, TypeSpecifierSigned):
		return Short
	case d.is(TypeSpecifierInt, TypeSpecifierShort, TypeSpecifierUnsigned):
		return UShort
	case d.is(TypeSpecifierInt, TypeSpecifierSigned):
		return Int
	case d.is(TypeSpecifierInt, TypeSpecifierUnsigned):
		return UInt
	case d.is(TypeSpecifierLong, TypeSpecifierLong):
		return LongLong
	case d.is(TypeSpecifierLong, TypeSpecifierSigned):
		return Long
	case d.is(TypeSpecifierLong, TypeSpecifierUnsigned):
		return ULong
	case d.is(TypeSpecifierLong, TypeSpecifierLong, TypeSpecifierSigned):
		return LongLong
	case d.is(TypeSpecifierLong, TypeSpecifierLong, TypeSpecifierUnsigned):
		return ULongLong
	case d.is(TypeSpecifierShort, TypeSpecifierUnsigned):
		return UShort
	case d.is(TypeSpecifierShort, TypeSpecifierSigned):
		return Short
	default:
		panic(d.typeSpecifiers)
	}
}

func (d *DeclarationSpecifier) is(a ...TypeSpecifierCase) bool {
	if d == nil {
		return false
	}

	if len(d.typeSpecifiers) != len(a) {
		return false
	}

	for i, v := range a {
		if v != d.typeSpecifiers[i] {
			return false
		}
	}
	return true
}

// IsTypedef reports whether the storage class specifier "typedef" is present.
func (d *DeclarationSpecifier) IsTypedef() bool {
	if d == nil {
		return false
	}

	for _, v := range d.StorageClassSpecifiers {
		if v.Case == StorageClassSpecifierTypedef {
			return true
		}
	}
	return false
}

// IsVolatile reports whether the type qualifier specifier "volatile" is present.
func (d *DeclarationSpecifier) IsVolatile() bool {
	if d == nil {
		return false
	}

	for _, v := range d.TypeQualifiers {
		if v.Case == TypeQualifierVolatile {
			return true
		}
	}
	return d.Parent.IsVolatile()
}

// IsStatic reports whether the storage class specifier "static" is present.
func (d *DeclarationSpecifier) IsStatic() bool {
	if d == nil {
		return false
	}

	for _, v := range d.StorageClassSpecifiers {
		if v.Case == StorageClassSpecifierStatic {
			return true
		}
	}
	return false
}

// IsExtern reports whether the storage class specifier "extern" is present.
func (d *DeclarationSpecifier) IsExtern() bool {
	if d == nil {
		return false
	}

	for _, v := range d.StorageClassSpecifiers {
		if v.Case == StorageClassSpecifierExtern {
			return true
		}
	}
	return false
}

func (n *ConstExpr) eval(ctx *context) Operand {
	if n.Operand.Type == nil {
		var dummy bool
		n.Operand = n.Expr.eval(ctx, true, nil, nil, nil, &dummy)
		if n.Operand.Value == nil { // not a constant expression
			panic(fmt.Errorf("TODO247 %v", ctx.position(n)))
		}
	}
	return n.Operand
}

func (n *Expr) eval(ctx *context, arr2ptr bool, fn *Declarator, seq *int, sc []int, hasLabels *bool) Operand {
	if n.Operand.Type != nil {
		return n.Operand
	}

	defer func() {
		if n.Operand.Type != nil && n.Operand.Type.IsArithmeticType() {
			n.Operand = n.Operand.normalize(ctx.model)
		}
		if n.UseGotos {
			*hasLabels = true
		}
	}()

outer:
	switch n.Case {
	case ExprPreInc: // "++" Expr
		// [0]6.5.3.1
		//
		// The operand of the prefix increment or decrement operator
		// shall have qualified or unqualified real or pointer type and
		// shall be a modifiable lvalue.
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !n.Operand.isScalarType() {
			panic(ctx.position(n))
		}
	case ExprPreDec: // "--" Expr
		// [0]6.5.3.1
		//
		// The operand of the prefix increment or decrement operator
		// shall have qualified or unqualified real or pointer type and
		// shall be a modifiable lvalue.
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !n.Operand.isScalarType() {
			panic(ctx.position(n))
		}
	case ExprAlignofType: // "__alignof__" '(' TypeName ')'
		t := n.TypeName.check(ctx)
		n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: int64(ctx.model.Alignof(t))}}
	case ExprAlignofExpr: // "__alignof__" Expr
		op := n.Expr.eval(ctx, false, fn, seq, sc, &n.UseGotos)
		n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: int64(ctx.model.Alignof(underlyingType(op.Type, false)))}}
	case ExprSizeofType: // "sizeof" '(' TypeName ')'
		t := n.TypeName.check(ctx)
		n.Operand = ctx.sizeof(t)
	case ExprSizeofExpr: // "sizeof" Expr
		// [0]6.5.3.4
		switch t := n.Expr.eval(ctx, false, fn, seq, sc, &n.UseGotos).Type.(type) { // [0]6.3.2.1-3
		case *ArrayType:
			if t.Size.Type == nil {
				panic(fmt.Errorf("%v", ctx.position(n)))
			}

			if t.IsVLA() {
				switch d := n.Expr.Declarator; {
				case d != nil && d.IsFunctionParameter:
					panic(fmt.Errorf("%v", ctx.position(n)))
				default:
					n.Operand = Operand{Type: Int}
					break outer
				}
			}

			switch d := n.Expr.Declarator; {
			case d != nil && d.IsFunctionParameter:
				n.Operand = ctx.sizeof(Ptr)
			default:
				n.Operand = t.Size.mul(ctx, ctx.sizeof(t.Item))
			}
		case
			*PointerType,
			*StructType,
			*TaggedStructType,
			*TaggedUnionType,
			*UnionType:

			n.Operand = ctx.sizeof(t)
		case *NamedType:
			n.Operand = ctx.sizeof(t.Type)
		case TypeKind:
			switch t {
			case
				Char,
				Double,
				DoubleComplex,
				DoubleImaginary,
				Float,
				FloatComplex,
				FloatImaginary,
				Int,
				Long,
				LongDouble,
				LongDoubleComplex,
				LongDoubleImaginary,
				LongLong,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort:

				n.Operand = ctx.sizeof(t)
			default:
				panic(t)
			}
		default:
			//dbg("", ctx.position(n))
			panic(t)
		}
		if n.Operand.Value == nil {
			if x, ok := n.Expr.Operand.Type.(*ArrayType); ok {
				panic(fmt.Errorf("%v: %p %s", ctx.position(n), x, PrettyString(x)))
			}
			panic(fmt.Errorf("%v", ctx.position(n)))
		}
	case ExprNot: // '!' Expr
		n.Operand = Operand{Type: Int}
		a := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if a.IsZero() { //TODO n.Expr.IsZero everywhere
			n.Operand.Value = &ir.Int64Value{Value: 1}
			break
		}

		if a.IsNonZero() { //TODO n.Expr.IsNonZero everywhere
			n.Operand.Value = &ir.Int64Value{Value: 0}
		}
	case ExprAddrof: // '&' Expr
		// [0]6.5.3.2
		op := n.Expr.eval(ctx, false, fn, seq, sc, &n.UseGotos) // [0]6.3.2.1-3
		n.Operand = Operand{Type: &PointerType{op.Type, n.Pos()}}
		if d := n.Expr.Declarator; d != nil && n.Expr.Case != ExprPSelect {
			d.AddressTaken = true
		}
		n.Operand.Value = op.Value
	case ExprPExprList: // '(' ExprList ')'
		n.Operand = n.ExprList.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		n.Declarator = n.ExprList.declarator(ctx)
	case ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}'
		t := n.TypeName.check(ctx)
		n.Operand = Operand{Type: t}
		if n.InitializerList == nil {
			n.InitializerList = &InitializerList{}
		}
		n.InitializerList.check(ctx, t, fn)
		if fn == nil {
			break
		}

		nmTok := n.Token
		nmTok.Char.Rune = IDENTIFIER
		fn.unnamed++
		nmTok.Val = dict.SID(fmt.Sprintf("unnamed%d", fn.unnamed))
		d := &Declarator{
			DirectDeclarator: &DirectDeclarator{
				Case:  DirectDeclaratorIdent,
				Token: nmTok,
			},
			Initializer: &Initializer{
				Case:            InitializerCompLit,
				InitializerList: n.InitializerList,
			},
			Referenced: 1,
			Scope:      n.Scope,
			Type:       t,
		}
		n.Declarator = d
		fn.vars = append(fn.vars, d)
	case ExprCast: // '(' TypeName ')' Expr
		// [0]6.5.4
		t := n.TypeName.check(ctx)
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if t == Void {
			n.Operand = Operand{Type: Void}
			break
		}
	more:
		switch x := t.(type) {
		case *EnumType:
			n.Operand = op.ConvertTo(ctx.model, t)
		case *PointerType:
			//dbg("", ctx.position(n), t, op)
			n.Operand = op.ConvertTo(ctx.model, t)
		case *NamedType:
			t = x.Type
			goto more
		case *TaggedEnumType:
			n.Operand = op.ConvertTo(ctx.model, t)
		case *TaggedStructType:
			n.Operand.Type = t
		case *TaggedUnionType:
			n.Operand = op.ConvertTo(ctx.model, t)
		case TypeKind:
			switch x {
			case
				Char,
				Double,
				DoubleComplex,
				DoubleImaginary,
				Float,
				FloatComplex,
				FloatImaginary,
				Int,
				Long,
				LongDouble,
				LongDoubleComplex,
				LongDoubleImaginary,
				LongLong,
				SChar,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort:

				n.Operand = op.ConvertTo(ctx.model, t)
			default:
				panic(x)
			}
		default:
			panic(fmt.Errorf("%v: %T", ctx.position(n), x))
		}
		if n.Expr.Operand.Value != nil {
			op := n.Expr.Operand.ConvertTo(ctx.model, t)
			n.Operand.Value = op.Value
		}
	case ExprDeref: // '*' Expr
		// [0]6.5.3
		if !arr2ptr && n.Expr.Case == ExprCall { // int *f(); int *p = &*f();
			arr2ptr = true
		}
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		for t, done := op.Type, false; !done; {
			switch x := t.(type) {
			case *ArrayType:
				n.Operand = Operand{Type: x.Item}
				done = true
			case *NamedType:
				t = x.Type
			case *PointerType:
				n.Operand = Operand{Type: x.Item}
				done = true
			default:
				panic(fmt.Errorf("%v: ExprDeref %v, arr2ptr %v", ctx.position(n), op, arr2ptr))
			}
		}
	case ExprUnaryPlus: // '+' Expr
		// [0]6.5.3.3
		// The operand of the unary + or - operator shall have
		// arithmetic type; of the ~ operator, integer type; of the !
		// operator, scalar type.
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !op.isArithmeticType() {
			panic(ctx.position(n))
		}
		n.Operand = op.integerPromotion(ctx.model)
	case ExprUnaryMinus: // '-' Expr
		// [0]6.5.3.3
		// The operand of the unary + or - operator shall have
		// arithmetic type; of the ~ operator, integer type; of the !
		// operator, scalar type.
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !op.isArithmeticType() {
			panic(ctx.position(n))
		}
		n.Operand = op.unaryMinus(ctx)
	case ExprCpl: // '~' Expr
		// [0]6.5.3.3
		// The operand of the unary + or - operator shall have
		// arithmetic type; of the ~ operator, integer type; of the !
		// operator, scalar type.
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).integerPromotion(ctx.model)
		if !op.Type.IsIntegerType() {
			panic(ctx.position(n))
		}
		n.Operand = op.cpl(ctx)
	case ExprChar: // CHARCONST
		n.Operand = ctx.charConst(n.Token)
	case ExprNe: // Expr "!=" Expr
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if n.Expr.Equals(n.Expr2) {
			n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: 0}}
			break
		}

		// [0]6.5.9
		switch {
		// One of the following shall hold:
		case
			// both operands have arithmetic type
			lhs.isArithmeticType() && rhs.isArithmeticType():

			n.Operand = lhs.ne(ctx, rhs)
			// fmt.Printf("TODO497 %v\n", ctx.position(n))
			// n.Operand = n.Operand.normalize(ctx.model)
			// n.DumpOperands("· ") //TODO-
		case
			// one operand is a pointer and the other is a null
			// pointer constant.
			lhs.isPointerType() && rhs.isNullPtrConst():

			n.Operand = Operand{Type: Int}
			switch {
			case n.Expr.IsNonZero():
				n.Operand.Value = &ir.Int64Value{Value: 1}
			case n.Expr.IsZero():
				n.Operand.Value = &ir.Int64Value{Value: 0}
			}
		case
			// one operand is a pointer and the other is a null
			// pointer constant.
			lhs.isNullPtrConst() && rhs.isPointerType():

			n.Operand = Operand{Type: Int}
			switch {
			case n.Expr2.IsNonZero():
				n.Operand.Value = &ir.Int64Value{Value: 1}
			case n.Expr2.IsZero():
				n.Operand.Value = &ir.Int64Value{Value: 0}
			}
		case
			// both operands are pointers to qualified or unqualified versions of compatible types
			lhs.isPointerType() && rhs.isPointerType() && (lhs.Type.IsCompatible(rhs.Type) || ctx.tweaks.EnablePointerCompatibility):

			n.Operand = Operand{Type: Int}
			if n.Expr.Case == ExprAddrof && n.Expr.Expr.Case == ExprIdent &&
				n.Expr2.Case == ExprAddrof && n.Expr2.Expr.Case == ExprIdent {
				var val int64
				if n.Expr.Expr.Token.Val != n.Expr2.Expr.Token.Val {
					val = 1
				}
				n.Operand.Value = &ir.Int64Value{Value: val}
			}
		default:
			panic(fmt.Errorf("%v: %v %v", ctx.position(n), lhs, rhs))
		}
	case ExprModAssign: // Expr "%=" Expr
		// [0]6.5.16.2
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).mod(ctx, n, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprLAnd: // Expr "&&" Expr
		n.Operand = Operand{Type: Int}
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if n.Expr.IsZero() {
			n.Operand.Value = &ir.Int64Value{Value: 0}
			break
		}

		n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if n.Expr2.IsZero() {
			n.Operand.Value = &ir.Int64Value{Value: 0}
			break
		}

		if n.Expr.IsNonZero() && n.Expr2.IsNonZero() {
			n.Operand.Value = &ir.Int64Value{Value: 1}
			break
		}
	case ExprAndAssign: // Expr "&=" Expr
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).and(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprMulAssign: // Expr "*=" Expr
		// [0]6.5.16.2
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).mul(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprPostInc: // Expr "++"
		// [0]6.5.2.4
		//
		// The operand of the postfix increment or decrement operator
		// shall have qualified or unqualified real or pointer type and
		// shall be a modifiable lvalue.
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !n.Operand.isScalarType() {
			panic(ctx.position(n))
		}
	case ExprAddAssign: // Expr "+=" Expr
		// [0]6.5.16.2
		//
		// 1. For the operators += and -= only, either the left operand
		// shall be a pointer to an object type and the right shall
		// have integer type, or the left operand shall have qualified
		// or unqualified arithmetic type and the right shall have
		// arithmetic type.
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		switch {
		case
			lhs.isPointerType() && rhs.isIntegerType(),
			lhs.isArithmeticType() && rhs.isArithmeticType():

			// ok
		default:
			panic(ctx.position(n))
		}
		n.Operand = lhs
	case ExprPostDec: // Expr "--"
		// [0]6.5.2.4
		//
		// The operand of the postfix increment or decrement operator
		// shall have qualified or unqualified real or pointer type and
		// shall be a modifiable lvalue.
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if !n.Operand.isScalarType() {
			panic(ctx.position(n))
		}
	case ExprSubAssign: // Expr "-=" Expr
		// [0]6.5.16.2
		//
		// 1. For the operators += and -= only, either the left operand
		// shall be a pointer to an object type and the right shall
		// have integer type, or the left operand shall have qualified
		// or unqualified arithmetic type and the right shall have
		// arithmetic type.
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		switch {
		case
			lhs.isPointerType() && rhs.isIntegerType(),
			lhs.isArithmeticType() && rhs.isArithmeticType():

			// ok
		default:
			panic(ctx.position(n))
		}
		n.Operand = lhs
	case ExprPSelect: // Expr "->" IDENTIFIER
		n.Expr.AssignedTo = n.AssignedTo
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if d := n.Expr.Declarator; d != nil && n.AssignedTo {
			d.AssignedTo++
		}
		t := op.Type
		for done := false; !done; {
			switch x := t.(type) {
			case *ArrayType:
				t = x.Item
				done = true
			case *NamedType:
				t = x.Type
			case *PointerType:
				t = x.Item
				done = true
			default:
				panic(fmt.Errorf("%v: %T", ctx.position(n), x))
			}
		}
	out:
		for {
			switch x := t.(type) {
			case *NamedType:
				t = x.Type
			case *StructType:
				ctx.model.Layout(x)
				fp := x.Field(n.Token2.Val)
				if fp == nil {
					panic(ctx.position(n))
				}

				n.Operand = Operand{Type: fp.Declarator.Type}
				n.Operand.FieldProperties = fp
				if op.Value == Null {
					n.Operand.Value = &ir.Int64Value{Value: fp.Offset}
				}
				break out
			case *TaggedStructType:
				t = x.getType()
				if t == x {
					panic(fmt.Errorf("%v: %q", ctx.position(n), dict.S(x.Tag)))
				}
			case *TaggedUnionType:
				t = x.getType()
				if t == x {
					panic(fmt.Errorf("%v: %q", ctx.position(n), dict.S(x.Tag)))
				}
			case *UnionType:
				ctx.model.Layout(x)
				fp := x.Field(n.Token2.Val)
				if fp == nil {
					panic(ctx.position(n))
				}

				n.Operand = Operand{Type: fp.Declarator.Type}
				n.Operand.FieldProperties = fp
				if op.Value == Null {
					n.Operand.Value = &ir.Int64Value{}
				}
				break out
			default:
				panic(x)
			}
		}
	case ExprSelect: // Expr '.' IDENTIFIER
		n.Expr.AssignedTo = n.AssignedTo
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if d := n.Expr.Declarator; d != nil && n.AssignedTo {
			d.AssignedTo++
		}
		t := op.Type
	out3:
		for {
			switch x := t.(type) {
			case *NamedType:
				t = x.Type
			case *StructType:
				ctx.model.Layout(x)
				fp := x.Field(n.Token2.Val)
				if fp == nil {
					panic(ctx.position(n))
				}

				n.Operand = Operand{Type: fp.Declarator.Type}
				n.Operand.FieldProperties = fp
				break out3
			case *TaggedStructType:
				y := x.getType()
				if x == y {
					panic("TODO")
				}

				t = y
			case *TaggedUnionType:
				y := x.getType()
				if x == y {
					panic(fmt.Errorf("%v:", ctx.position(n)))
				}

				t = y
			case *UnionType:
				ctx.model.Layout(x)
				fp := x.Field(n.Token2.Val)
				if fp == nil {
					panic(ctx.position(n))
				}

				n.Operand = Operand{Type: fp.Declarator.Type}
				n.Operand.FieldProperties = fp
				break out3
			default:
				//dbg("%v: %T", ctx.position(n), x)
				panic(x)
			}
		}
		if d := n.Expr.Declarator; d != nil {
			d.Referenced++
			d.AddressTaken = true
		}
	case ExprDivAssign: // Expr "/=" Expr
		// [0]6.5.16.2
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).div(ctx, n, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprLsh: // Expr "<<" Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).lsh(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
	case ExprLshAssign: // Expr "<<=" Expr
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).lsh(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprLe: // Expr "<=" Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).le(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.Expr.Equals(n.Expr2) {
			n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: 1}}
		}
	case ExprEq: // Expr "==" Expr
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		switch x := underlyingType(lhs.Type, true).(type) {
		case *EnumType:
			n.Expr2.enum = x
		}
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		n.Operand.Type = Int
		if n.Expr.Equals(n.Expr2) {
			n.Operand.Value = &ir.Int64Value{Value: 1}
			break
		}

		// [0]6.5.9
		switch {
		// One of the following shall hold:
		case
			// both operands have arithmetic type
			lhs.isArithmeticType() && rhs.isArithmeticType():

			n.Operand = lhs.eq(ctx, rhs)
		case
			// one operand is a pointer and the other is a null
			// pointer constant.
			lhs.isPointerType() && rhs.isNullPtrConst():

			switch {
			case n.Expr.IsNonZero():
				n.Operand.Value = &ir.Int64Value{Value: 0}
			case n.Expr.IsZero():
				n.Operand.Value = &ir.Int64Value{Value: 1}
			}
		case
			// one operand is a pointer and the other is a null
			// pointer constant.
			lhs.isNullPtrConst() && rhs.isPointerType():

			switch {
			case n.Expr2.IsNonZero():
				n.Operand.Value = &ir.Int64Value{Value: 0}
			case n.Expr2.IsZero():
				n.Operand.Value = &ir.Int64Value{Value: 1}
			}
		case
			// both operands are pointers to qualified or unqualified versions of compatible types
			lhs.isPointerType() && rhs.isPointerType() && (lhs.Type.IsCompatible(rhs.Type) || ctx.tweaks.EnablePointerCompatibility):

			if n.Expr.Case == ExprAddrof && n.Expr.Expr.Case == ExprIdent &&
				n.Expr2.Case == ExprAddrof && n.Expr2.Expr.Case == ExprIdent {
				var val int64
				if n.Expr.Expr.Token.Val == n.Expr2.Expr.Token.Val {
					val = 1
				}
				n.Operand.Value = &ir.Int64Value{Value: val}
			}
		default:
			panic(fmt.Errorf("%v: %v %v", ctx.position(n), lhs, rhs))
		}
	case ExprGe: // Expr ">=" Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).ge(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.Expr.Equals(n.Expr2) {
			n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: 1}}
		}
	case ExprRsh: // Expr ">>" Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).rsh(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
	case ExprRshAssign: // Expr ">>=" Expr
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).rsh(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprXorAssign: // Expr "^=" Expr
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).xor(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprOrAssign: // Expr "|=" Expr
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).or(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		n.Operand = n.Expr.Operand
	case ExprLOr: // Expr "||" Expr
		n.Operand = Operand{Type: Int}
		n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if n.Expr.IsNonZero() {
			n.Operand.Value = &ir.Int64Value{Value: 1}
			break
		}

		n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if n.Expr2.IsNonZero() {
			n.Operand.Value = &ir.Int64Value{Value: 1}
			break
		}

		if n.Expr.IsZero() && n.Expr2.IsZero() {
			n.Operand.Value = &ir.Int64Value{Value: 0}
		}
	case ExprMod: // Expr '%' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).mod(ctx, n, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)) // [0]6.5.5
	case ExprAnd: // Expr '&' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).and(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.isSideEffectsFree() {
			// x & (x | y) == x
			x, y := n.Expr, n.Expr2
			if y.Case == ExprIdent {
				x, y = y, x
			}
			if x.Case == ExprIdent && y.Case == ExprPExprList && y.ExprList.ExprList == nil {
				if z := y.ExprList.Expr; z.Case == ExprOr {
					if z.Expr.Case == ExprIdent && z.Expr.Token.Val == x.Token.Val ||
						z.Expr2.Case == ExprIdent && z.Expr2.Token.Val == x.Token.Val {
						*n = *x
					}
				}

			}
		}
	case ExprCall: // Expr '(' ArgumentExprListOpt ')'
		if n.Expr.Case == ExprIdent {
			switch n.Expr.Token.Val {
			case idBuiltinTypesCompatible:
				// using #define __builtin_types_compatible_p(type1, type2) __builtin_types_compatible__((type1){}, (type2){})
				o := n.ArgumentExprListOpt
				if o == nil {
					panic("missing arguments of __builtin_types_compatible_p")
				}

				args := o.ArgumentExprList
				arg1 := args.Expr
				if arg1.Case != ExprCompLit { // '(' TypeName ')' '{' InitializerList CommaOpt '}'
					panic("invalid argument of __builtin_types_compatible__p")
				}

				args = args.ArgumentExprList
				if args == nil {
					panic("missing argument of __builtin_types_compatible_p")
				}

				arg2 := args.Expr
				if arg2.Case != ExprCompLit { // '(' TypeName ')' '{' InitializerList CommaOpt '}'
					panic("invalid argument of __builtin_types_compatible__")
				}

				if args.ArgumentExprList != nil {
					panic("too many arguments of __builtin_types_compatible_p")
				}

				t := arg1.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).Type
				u := arg2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).Type
				var v int64
				if t.IsCompatible(u) {
					v = 1
				}
				n.Operand.Type = Int
				n.Operand.Value = &ir.Int64Value{Value: v}
				break outer
			case idBuiltinClasifyType:
				o := n.ArgumentExprListOpt
				if o == nil {
					panic("missing argument of __builtin_classify_type")
				}

				args := o.ArgumentExprList
				if args.ArgumentExprList != nil {
					panic("too many arguments of __builtin_classify_type")
				}

				op := args.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
				v := noTypeClass
				if x, ok := classifyType[op.Type.Kind()]; ok {
					v = x
				}
				n.Case = ExprInt
				n.Operand.Type = Int
				n.Operand.Value = &ir.Int64Value{Value: int64(v)}
				break outer
			}
		}

		// [0]6.5.2.2
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		args := n.ArgumentExprListOpt.eval(ctx, fn, seq, sc, &n.UseGotos)
		ops := make([]Operand, len(args))
		n.CallArgs = ops
		t := checkFn(ctx, op.Type)
		if t == nil {
			if !ctx.tweaks.EnableImplicitDeclarations {
				panic(ctx.position(n))
			}

			if n.Expr.Case == ExprIdent && n.Expr.Token.Val == idAsm {
				panic(ctx.position(n).String())
			}

			n.Operand = Operand{Type: Int}
			break
		}

		if _, ok := t.Result.(*ArrayType); ok {
			panic(ctx.position)
		}

		n.Operand = Operand{Type: t.Result}

		// 2. If the expression that denotes the called function has a
		// type that includes a prototype, the number of arguments
		// shall agree with the number of parameters. Each argument
		// shall have a type such that its value may be assigned to an
		// object with the unqualified version of the type of its
		// corresponding parameter.
		params := t.Params
		var voidParams bool
		if voidParams = len(params) == 1 && params[0].Kind() == Void; voidParams {
			params = nil
		}
		switch {
		case voidParams && len(args) != 0:
			panic(fmt.Errorf("%v: %v args %v params %v variadic %v voidParams %v", ctx.position(n), n.Case, len(args), len(params), t.Variadic, voidParams))
		case len(args) < len(params):
			panic(fmt.Errorf("%v: %v args %v params %v variadic %v voidParams %v", ctx.position(n), n.Case, len(args), len(params), t.Variadic, voidParams))
		case len(args) == len(params):
			for i, rhs := range args {
				switch {
				case rhs.Type.Kind() == Ptr && t.Params[i].IsIntegerType():
					ops[i] = rhs.ConvertTo(ctx.model, t.Params[i])
				default:
					ops[i] = AdjustedParameterType(t.Params[i]).assign(ctx, n, rhs)
				}
			}

		// len(args) > len(params)
		case t.Variadic:
			for i, rhs := range args {
				switch {
				case i >= len(t.Params):
					ops[i] = ctx.model.defaultArgumentPromotion(rhs)
				case rhs.Type.Kind() == Ptr && t.Params[i].IsIntegerType():
					ops[i] = rhs.ConvertTo(ctx.model, t.Params[i])
				default:
					ops[i] = AdjustedParameterType(t.Params[i]).assign(ctx, n, rhs)
				}
			}
		case len(params) == 0:
			for i, rhs := range args {
				ops[i] = ctx.model.defaultArgumentPromotion(rhs)
			}
		default:
			panic(fmt.Errorf("%v: %v args %v params %v variadic %v voidParams %v", ctx.position(n), n.Case, len(args), len(params), t.Variadic, voidParams))
		}

		for i, v := range ops {
			if v.Value != nil {
				ops[i] = v.ConvertTo(ctx.model, ops[i].Type)
			}
		}
	case ExprMul: // Expr '*' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).mul(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
	case ExprAdd: // Expr '+' Expr
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		// [0]6.5.6
		//
		// For addition, either both operands shall have arithmetic
		// type, or one operand shall be a pointer to an object type
		// and the other shall have integer type. (Incrementing is
		// equivalent to adding 1.)
		switch {
		case lhs.isArithmeticType() && rhs.isArithmeticType():
			n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).add(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		case lhs.isPointerType() && rhs.isIntegerType():
			n.Operand = lhs
			n.Operand.Value = nil
		case lhs.isIntegerType() && rhs.isPointerType():
			n.Operand = rhs
			n.Operand.Value = nil
		default:
			panic(ctx.position(n))
		}
	case ExprSub: // Expr '-' Expr
		// [0]6.5.6
		lhs := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		switch {
		// 3. For subtraction, one of the following shall hold:
		case
			// both operands have arithmetic type;
			lhs.isArithmeticType() && rhs.isArithmeticType():

			n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).sub(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
			if n.Operand.Type.IsIntegerType() && n.Expr.Equals(n.Expr2) {
				n.Operand.Value = &ir.Int64Value{Value: 0}
			}
		case
			// both operands are pointers to qualified or
			// unqualified versions of compatible object types;
			lhs.isPointerType() && rhs.isPointerType() && (lhs.Type.IsCompatible(rhs.Type) || ctx.tweaks.EnablePointerCompatibility):

			n.Operand = Operand{Type: ctx.ptrDiff()}
			if rhs.IsZero() {
				n.Operand.Value = lhs.Value
			}
		case
			// the left operand is a pointer to an object type and
			// the right operand has integer type.
			lhs.isPointerType() && rhs.isIntegerType():

			n.Operand = lhs
		default:
			//dbg("", lhs, rhs)
			panic(ctx.position(n))
		}
	case ExprDiv: // Expr '/' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).div(ctx, n, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)) // [0]6.5.5
	case ExprLt: // Expr '<' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).lt(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.Expr.Equals(n.Expr2) {
			n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: 0}}
		}
	case ExprAssign: // Expr '=' Expr
		n.Expr.AssignedTo = true
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		if d := n.Expr.Declarator; d != nil {
			d.AssignedTo++
			if n.Expr.Case == ExprIdent {
				n.Operand.Type = d.Type
			}
		}
		switch x := underlyingType(n.Operand.Type, true).(type) {
		case *EnumType:
			n.Expr2.enum = x
		}
		rhs := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		n.Operand.Type.assign(ctx, n, rhs)
	case ExprGt: // Expr '>' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).gt(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.Expr.Equals(n.Expr2) {
			n.Operand = Operand{Type: Int, Value: &ir.Int64Value{Value: 0}}
		}
	case ExprCond: // Expr '?' ExprList ':' Expr
		// [0]6.5.15
		cond := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		// 2. The first operand shall have scalar type.
		if !cond.isScalarType() {
			panic(ctx.position(n))
		}

		switch {
		case cond.IsNonZero():
			n.Operand = n.ExprList.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
			break outer
		case cond.IsZero():
			n.Operand = n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
			break outer
		}

		a := n.ExprList.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		b := n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		switch {
		// 3. One of the following shall hold for the second and third
		// operands:
		case
			// both operands have arithmetic type;
			a.isArithmeticType() && b.isArithmeticType():

			// 5. If both the second and third operands have
			// arithmetic type, the result type that would be
			// determined by the usual arithmetic conversions, were
			// they applied to those two operands, is the type of
			// the result.
			x, _ := UsualArithmeticConversions(ctx.model, a, b)
			n.Operand = Operand{Type: x.Type}
		case
			// both operands have the same structure or union type
			a.Type.Kind() == Struct && b.Type.Kind() == Struct && a.Type.Equal(b.Type),
			a.Type.Kind() == Union && b.Type.Kind() == Union && a.Type.Equal(b.Type):

			n.Operand = Operand{Type: a.Type}
		case
			// both operands are pointers to qualified or
			// unqualified versions of compatible types;
			a.isPointerType() && b.isPointerType() && (a.Type.IsCompatible(b.Type) || ctx.tweaks.EnablePointerCompatibility):

			n.Operand = Operand{Type: a.Type}
		case
			// one operand is a pointer and the other is a null pointer constant
			a.isNullPtrConst() && b.isPointerType():

			n.Operand = Operand{Type: b.Type}
			if cond.IsNonZero() {
				n.Operand.Value = Null
			}
		case
			// one operand is a pointer and the other is a null pointer constant
			a.isPointerType() && b.isNullPtrConst():

			n.Operand = Operand{Type: a.Type}
			if cond.IsZero() {
				n.Operand.Value = Null
			}
		case a.Type.Kind() == Void && b.Type.Kind() == Void: // assert
			n.Operand = Operand{Type: Void}
		default:
			panic(fmt.Errorf("%v: %v, %v", ctx.position(n), a, b))
		}
	case ExprIndex: // Expr '[' ExprList ']'
		// [0]6.5.2.1
		op := n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos)
		index := n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		t := op.Type
	more3:
		switch x := t.(type) {
		case *ArrayType:
			if arr2ptr {
				panic("internal error")
			}
			n.Operand = Operand{Type: x.Item}
		case *PointerType:
			n.Operand = Operand{Type: x.Item}
		case *NamedType:
			t = x.Type
			goto more3
		case TypeKind:
			if !x.IsIntegerType() {
				panic(fmt.Errorf("%v: %T %v", ctx.position(n), x, x))
			}

			r := UnderlyingType(index.Type)
			switch {
			case r.IsPointerType():
				n.Operand = Operand{Type: r.(*PointerType).Item}
			default:
				panic(fmt.Errorf("%v: %v[%v]", ctx.position(n), op.Type, index))
			}
		default:
			panic(fmt.Errorf("%v: %T %v", ctx.position(n), x, x))
		}
		if !index.isIntegerType() {
			l := UnderlyingType(op.Type)
			switch {
			case l.IsIntegerType() && index.isPointerType():
				n.Operand = Operand{Type: UnderlyingType(index.Type).(*PointerType).Item}
			default:
				panic(fmt.Errorf("%v: %v[%v]", ctx.position(n), op.Type, index))
			}
		}
		if d := n.Expr.Declarator; d != nil {
			d.Referenced++
		}
	case ExprXor: // Expr '^' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).xor(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
		if n.Expr.Equals(n.Expr2) {
			n.Operand.Value = &ir.Int64Value{Value: 0}
		}
	case ExprOr: // Expr '|' Expr
		n.Operand = n.Expr.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos).or(ctx, n.Expr2.eval(ctx, arr2ptr, fn, seq, sc, &n.UseGotos))
	case ExprFloat: // FLOATCONST
		n.floatConst(ctx)
	case ExprIdent: // IDENTIFIER
		builtin := false
		// [0]6.5.1
		nm := n.Token.Val
		var nm2 int
		if ctx.tweaks.EnableImplicitBuiltins {
			nm2 = dict.SID("__builtin_" + string(dict.S(nm)))
			if n.Scope.LookupIdent(nm2) != nil {
				builtin = true
			}
		}
		if n.Scope.LookupIdent(nm) == nil {
			if n.enum != nil {
				if c := n.enum.find(nm); c != nil {
					n.Operand = c.Operand
					return n.Operand
				}
			}

			switch {
			case builtin:
				nm = nm2
			default:
				if !ctx.tweaks.EnableImplicitDeclarations {
					panic(fmt.Errorf("%v: undefined %q", ctx.position(n), dict.S(nm)))
				}
			}
		}
		switch x := n.Scope.LookupIdent(nm).(type) {
		case *Declarator:
			x.IsBuiltin = builtin
			n.Declarator = x
			if arr2ptr {
				x.Referenced++
			}
			t := x.Type
			t0 := t
		more2:
			switch y := t.(type) {
			case *ArrayType:
				if !x.IsFunctionParameter && y.Size.Type == nil {
					switch {
					case x.IsTLD() && x.DeclarationSpecifier.IsExtern():
						// ok
					default:
						panic(fmt.Errorf("%v: %s has incomplete type: %v", ctx.position(x), dict.S(x.Name()), t0))
					}
				}
				n.Operand = Operand{Type: t0}
			case
				*EnumType,
				*PointerType,
				*StructType,
				*TaggedEnumType,
				*TaggedStructType,
				*TaggedUnionType,
				*UnionType:

				n.Operand = Operand{Type: t0}
			case *FunctionType:
				n.Operand = Operand{Type: t0}
				if nm == idBuiltinAlloca {
					fn.Alloca = true
				}
			case *NamedType:
				t = y.Type
				goto more2
			case TypeKind:
				switch y {
				case
					Bool,
					Char,
					Double,
					DoubleComplex,
					DoubleImaginary,
					Float,
					FloatComplex,
					FloatImaginary,
					Int,
					Long,
					LongDouble,
					LongDoubleComplex,
					LongDoubleImaginary,
					LongLong,
					SChar,
					Short,
					UChar,
					UInt,
					ULong,
					ULongLong,
					UShort:

					n.Operand = Operand{Type: t0}
				default:
					//dbg("", ctx.position(n))
					panic(y)
				}
			default:
				//dbg("", ctx.position(n))
				panic(y)
			}
		case *EnumerationConstant:
			n.Operand = x.Operand
		case nil:
			if ctx.tweaks.EnableImplicitDeclarations {
				return Operand{}
			}

			panic(fmt.Errorf("%v: undefined: %q", ctx.position(n), dict.S(nm)))
		default:

			//dbg("%s", dict.S(nm))
			panic(fmt.Errorf("%v: %T", ctx.position(n), x))
		}
	case ExprInt: // INTCONST
		s0 := string(dict.S(n.Token.Val))
		s := s0
		if strings.Contains(s, "p") {
			n.Case = ExprFloat
			n.floatConst(ctx)
			break
		}

	loop:
		for i := len(s) - 1; i > 0; i-- {
			switch s0[i] {
			case 'l', 'L', 'u', 'U':
				s = s[:i]
			default:
				break loop
			}
		}
		decadic := s == "0" || !strings.HasPrefix(s, "0")
		v, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			if ctx.tweaks.EnableBinaryLiterals && (strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B")) {
				decadic = false
				v, err = strconv.ParseUint(s[2:], 2, 64)
			}
			if err != nil {
				panic(fmt.Errorf("%v: %v %v", ctx.position(n), n.Case, err))
			}
		}

		// [0]6.4.4.1
		switch suff := strings.ToUpper(s0[len(s):]); {
		case suff == "" && decadic:
			n.Operand = newIntConst(ctx, n, v, Int, Long, LongLong)
		case suff == "" && !decadic:
			n.Operand = newIntConst(ctx, n, v, Int, UInt, Long, ULong, LongLong, ULongLong)
		case suff == "U":
			n.Operand = newIntConst(ctx, n, v, UInt, ULong, ULongLong)
		case suff == "L" && decadic:
			n.Operand = newIntConst(ctx, n, v, Long, LongLong)
		case suff == "L" && !decadic:
			n.Operand = newIntConst(ctx, n, v, Long, ULong, LongLong, ULongLong)
		case suff == "UL", suff == "LU":
			n.Operand = newIntConst(ctx, n, v, ULong, ULongLong)
		case suff == "LL" && decadic:
			n.Operand = newIntConst(ctx, n, v, LongLong)
		case suff == "LL" && !decadic:
			n.Operand = newIntConst(ctx, n, v, LongLong, ULongLong)
		case suff == "ULL", suff == "LLU":
			n.Operand = newIntConst(ctx, n, v, ULongLong)
		default:
			panic(fmt.Errorf("%v: TODO %q %q decadic: %v\n%s", ctx.position(n), s, suff, decadic, PrettyString(n)))
		}
	case ExprLChar: // LONGCHARCONST
		n.Operand = ctx.charConst(n.Token)
	case ExprLString: // LONGSTRINGLITERAL
		n.Operand = ctx.strConst(n.Token)
	case ExprString: // STRINGLITERAL
		n.Operand = ctx.strConst(n.Token)
	case ExprStatement: // '(' CompoundStmt ')'
		n.Operand = n.CompoundStmt.check(ctx, fn, seq, sc, nil, false, &n.UseGotos)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}

	if !arr2ptr {
		return n.Operand
	}

	// [0]6.3.2.1
	//
	// 3. Except when it is the operand of the sizeof operator or
	// the unary & operator, or is a string literal used to
	// initialize an array, an expression that has type ‘‘array of
	// type’’ is converted to an expression with type ‘‘pointer to
	// type’’ that points to the initial element of the array
	// object and is not an lvalue. If the array object has
	// register storage class, the behavior is undefined.
	t := n.Operand.Type
	for {
		switch x := t.(type) {
		case *ArrayType:
			n.Operand.Type = &PointerType{x.Item, n.Pos()}
			return n.Operand
		case *FunctionType:
			n.Operand.Type = &PointerType{x, n.Pos()}
			return n.Operand
		case *NamedType:
			t = x.Type
		case
			*EnumType,
			*PointerType,
			*StructType,
			*TaggedEnumType,
			*TaggedStructType,
			*TaggedUnionType,
			*UnionType:

			return n.Operand
		case TypeKind:
			switch x {
			case
				Bool,
				Char,
				Double,
				DoubleComplex,
				DoubleImaginary,
				Float,
				FloatComplex,
				FloatImaginary,
				Int,
				Long,
				LongDouble,
				LongDoubleComplex,
				LongDoubleImaginary,
				LongLong,
				SChar,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort,
				Void:

				return n.Operand
			default:
				//dbg("", ctx.position(n))
				panic(x)
			}
		case nil:
			if ctx.tweaks.EnableImplicitDeclarations {
				return n.Operand
			}

			panic(fmt.Errorf("%v: %T", ctx.position(n), x))
		default:
			//dbg("", ctx.position(n))
			panic(fmt.Errorf("%T", x))
		}
	}
}

func (n *Expr) floatConst(ctx *context) {
	s0 := string(dict.S(n.Token.Val))
	s := s0
loop2:
	for i := len(s) - 1; i > 0; i-- {
		switch s0[i] {
		case 'l', 'L', 'f', 'F':
			s = s[:i]
		default:
			break loop2
		}
	}

	var v float64
	var err error
	switch {
	case strings.Contains(s, "p"):
		var bf *big.Float
		bf, _, err = big.ParseFloat(s, 0, 53, big.ToNearestEven)
		if err == nil {
			v, _ = bf.Float64()
		}
	default:
		v, err = strconv.ParseFloat(s, 64)
	}
	if err != nil {
		switch {
		case !strings.HasPrefix(s, "-") && strings.Contains(err.Error(), "value out of range"):
			// linux_386/usr/include/math.h
			//
			// 	/* Value returned on overflow.  With IEEE 754 floating point, this is
			// 	   +Infinity, otherwise the largest representable positive value.  */
			// 	#if __GNUC_PREREQ (3, 3)
			// 	# define HUGE_VAL (__builtin_huge_val ())
			// 	#else
			// 	/* This may provoke compiler warnings, and may not be rounded to
			// 	   +Infinity in all IEEE 754 rounding modes, but is the best that can
			// 	   be done in ISO C while remaining a constant expression.  10,000 is
			// 	   greater than the maximum (decimal) exponent for all supported
			// 	   floating-point formats and widths.  */
			// 	# define HUGE_VAL 1e10000
			// 	#endif
			v = math.Inf(1)
		default:
			panic(fmt.Errorf("%v: %v", ctx.position(n), err))
		}
	}

	// [0]6.4.4.2
	switch suff := strings.ToUpper(s0[len(s):]); suff {
	case "", "l", "L":
		n.Operand = Operand{Type: Double, Value: &ir.Float64Value{Value: v}}
	case "f", "F":
		n.Operand = Operand{Type: Float, Value: &ir.Float32Value{Value: float32(v)}}
	default:
		panic(fmt.Errorf("%v: TODO %q %q %v", ctx.position(n), s, suff, v))
	}
}

func checkFn(ctx *context, t Type) *FunctionType {
	// 1. The expression that denotes the called function 80) shall
	// have type pointer to function returning void or returning an
	// object type other than an array type.
	for {
		switch x := t.(type) {
		case *FunctionType:
			return x
		case *NamedType:
			t = x.Type
		case *PointerType:
			switch x := x.Item.(type) {
			case *FunctionType:
				return x
			case *NamedType:
				t = x.Type
			default:
				panic(fmt.Errorf("%T", x))
			}
		case nil:
			return nil
		default:
			panic(fmt.Errorf("%T", x))
		}
	}
}

func (n *ArgumentExprListOpt) eval(ctx *context, fn *Declarator, seq *int, sc []int, hasLabels *bool) []Operand {
	if n == nil {
		return nil
	}

	return n.ArgumentExprList.eval(ctx, fn, seq, sc, hasLabels)
}

func (n *ArgumentExprList) eval(ctx *context, fn *Declarator, seq *int, sc []int, hasLabels *bool) (r []Operand) {
	for ; n != nil; n = n.ArgumentExprList {
		r = append(r, n.Expr.eval(ctx, true, fn, seq, sc, hasLabels))
	}
	return r
}

func (n *TypeName) check(ctx *context) Type {
	// SpecifierQualifierList AbstractDeclaratorOpt
	ds := &DeclarationSpecifier{}
	n.SpecifierQualifierList.check(ctx, ds)
	if n.AbstractDeclaratorOpt == nil {
		n.Type = ds.typ(ctx)
		return n.Type
	}

	n.AbstractDeclaratorOpt.check(ctx, ds, ds.typ(ctx))
	n.Type = n.AbstractDeclaratorOpt.AbstractDeclarator.Type
	return n.Type
}

func (n *ExprListOpt) eval(ctx *context, arr2ptr bool, fn *Declarator, seq *int, sc []int, hasLabels *bool) Operand {
	if n == nil {
		return Operand{}
	}

	return n.ExprList.eval(ctx, arr2ptr, fn, seq, sc, hasLabels)
}

func (n *ExprList) eval(ctx *context, arr2ptr bool, fn *Declarator, seq *int, sc []int, hasLabels *bool) Operand {
	if n.Operand.Type == nil {
		for l := n; l != nil; l = l.ExprList {
			n.Operand = l.Expr.eval(ctx, arr2ptr, fn, seq, sc, hasLabels)
		}
	}
	return n.Operand
}

func (n *ExprList) declarator(ctx *context) (r *Declarator) {
	for l := n; ; l = l.ExprList {
		if l.ExprList == nil {
			return l.Expr.Declarator
		}
	}
}

func (n *ExprOpt) eval(ctx *context, arr2ptr bool, fn *Declarator, seq *int, sc []int, hasLabels *bool) Operand {
	if n == nil {
		return Operand{}
	}

	return n.Expr.eval(ctx, arr2ptr, fn, seq, sc, hasLabels)
}

// Name returns the ID of the declared name.
func (n *Declarator) Name() int { return n.DirectDeclarator.nm() }

func (n *DirectDeclarator) nm() int {
	switch n.Case {
	case
		DirectDeclaratorArray,
		DirectDeclaratorArraySize,
		DirectDeclaratorArrayVar,
		DirectDeclaratorParameters:

		return n.DirectDeclarator.nm()
	case DirectDeclaratorIdent:
		return n.Token.Val
	case DirectDeclaratorParen:
		return n.Declarator.Name()
	default:
		panic(fmt.Errorf("TODO %v", n.Case))
	}
}

func (n *ExternalDeclarationList) check(ctx *context) (err error) {
	for ; n != nil; n = n.ExternalDeclarationList {
		n.ExternalDeclaration.check(ctx)
	}
	return nil
}

func (n *ExternalDeclaration) check(ctx *context) {
	switch n.Case {
	case ExternalDeclarationDecl: // Declaration
		n.Declaration.check(ctx, nil, nil, nil, nil, false)
	case ExternalDeclarationFunc: // FunctionDefinition
		n.FunctionDefinition.check(ctx)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

// LocalVariables returns all local variables declared in a function in the
// order of appearance. The result does not include function parameters.
func (n *FunctionDefinition) LocalVariables() []*Declarator { return n.Declarator.vars }

func (n *FunctionDefinition) check(ctx *context) {
	ds := &DeclarationSpecifier{}
	switch n.Case {
	case FunctionDefinitionSpec: // DeclarationSpecifiers Declarator DeclarationListOpt FunctionBody
		n.DeclarationSpecifiers.check(ctx, ds)
		if len(ds.TypeSpecifiers) == 0 { // [0]6.7.2-2
			ds.typeSpecifiers = []TypeSpecifierCase{TypeSpecifierInt}
		}
	case FunctionDefinitionInt: // Declarator DeclarationListOpt FunctionBody
		ds.typeSpecifiers = []TypeSpecifierCase{TypeSpecifierInt}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	switch o := n.DeclarationListOpt; {
	case o != nil:
		sc := n.Declarator.fpScope(ctx)
		params := n.DeclarationListOpt.check(ctx, sc)
		n.Declarator.check0(ctx, ds, ds.typ(ctx), false, nil, nil, params)
		if n.Declarator.Type.Kind() != Function {
			panic("TODO")
		}

		d := n.Declarator
		n.FunctionBody.check(ctx, n.Declarator)
		for _, v := range d.ParameterNames() {
			p, _ := sc.Idents[v].(*Declarator)
			p.IsFunctionParameter = true
			d.Parameters = append(d.Parameters, p)
		}
	default:
		n.Declarator.check(ctx, ds, ds.typ(ctx), false, nil, nil)
		if n.Declarator.Type.Kind() != Function {
			panic("TODO")
		}

		d := n.Declarator
		sc := d.fpScope(ctx)
		n.FunctionBody.check(ctx, n.Declarator)
		for _, v := range d.ParameterNames() {
			p, _ := sc.Idents[v].(*Declarator)
			d.Parameters = append(d.Parameters, p)
		}
	}
}

func (n *DeclarationListOpt) check(ctx *context, scope *Scope) (r []*Declarator) {
	if n == nil {
		return
	}

	for l := n.DeclarationList; l != nil; l = l.DeclarationList {
		r = append(r, l.Declaration.check(ctx, nil, nil, nil, scope, true)...)
	}
	return r
}

func (n *FunctionBody) check(ctx *context, fn *Declarator) {
	// CompoundStmt *CompoundStmt
	seq := -1
	var dummy bool
	n.CompoundStmt.check(ctx, fn, &seq, nil, nil, false, &dummy)
}

func (n *CompoundStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) Operand {
	// '{' BlockItemListOpt '}'
	*seq++
	sc = append(sc, *seq)
	if *seq == 0 { // Pull function parameters into the outermost block scope.
		for _, v := range fn.fpScope(ctx).Idents {
			d := v.(*Declarator)
			nm := d.Name()
			if ex := n.scope.Idents[nm]; ex != nil {
				panic("TODO") // redeclared
			}

			n.scope.insertDeclarator(ctx, d)
		}
	}
	r := n.BlockItemListOpt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	if n.UseGotos {
		*hasLabels = true
	}
	return r
}

func (n *Declarator) fpScope(ctx *context) *Scope { return n.DirectDeclarator.fpScope(ctx) }

func (n *DirectDeclarator) fpScope(ctx *context) *Scope {
	switch n.Case {
	//TODO case DirectDeclaratorParen: // '(' Declarator ')'
	case DirectDeclaratorParameters:
		switch p := n.Parameters; p.Case {
		case ParametersIdentList: // IdentifierListOpt
			switch n.DirectDeclarator.Case {
			case DirectDeclaratorIdent:
				return n.paramScope
			default:
				panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.DirectDeclarator.Case))
			}
		case ParametersParamTypes: // ParameterTypeList
			switch n.DirectDeclarator.Case {
			case DirectDeclaratorParen:
				if n.DirectDeclarator.Declarator.DirectDeclarator.Case == DirectDeclaratorIdent {
					return n.paramScope
				}

				return n.DirectDeclarator.Declarator.DirectDeclarator.fpScope(ctx)
			case DirectDeclaratorIdent:
				return n.paramScope
			default:
				panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.DirectDeclarator.Case))
			}
		default:
			panic(fmt.Errorf("%v: TODO %v", ctx.position(p), p.Case))
		}
	//TODO case DirectDeclaratorArraySize: // DirectDeclarator '[' "static" TypeQualifierListOpt Expr ']'
	//TODO case DirectDeclaratorArraySize2: // DirectDeclarator '[' TypeQualifierList "static" Expr ']'
	//TODO case DirectDeclaratorArrayVar: // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
	//TODO case DirectDeclaratorArray: // DirectDeclarator '[' TypeQualifierListOpt ExprOpt ']'
	case DirectDeclaratorIdent: // IDENTIFIER
		return nil
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

// ParameterNames returns a list of IDs of names of parameters of n. The
// function panics if n is not function type.
func (n *Declarator) ParameterNames() []int { return n.DirectDeclarator.parameterNames() }

func (n *DirectDeclarator) parameterNames() (r []int) {
	switch n.Case {
	//TODO case DirectDeclaratorParen: // '(' Declarator ')'
	case DirectDeclaratorParameters:
		switch p := n.Parameters; p.Case {
		case ParametersIdentList: // IdentifierListOpt
			switch n.DirectDeclarator.Case {
			case DirectDeclaratorIdent:
				return p.IdentifierListOpt.check()
			default:
				panic(n.DirectDeclarator.Case)
			}
		case ParametersParamTypes: // ParameterTypeList
			switch n.DirectDeclarator.Case {
			case DirectDeclaratorIdent:
				for l := p.ParameterTypeList.ParameterList; l != nil; l = l.ParameterList {
					switch n := l.ParameterDeclaration; n.Case {
					case ParameterDeclarationAbstract: // DeclarationSpecifiers AbstractDeclaratorOpt
						r = append(r, 0)
					case ParameterDeclarationDeclarator: // DeclarationSpecifiers Declarator
						r = append(r, n.Declarator.Name())
					default:
						panic(n.Case)
					}
				}
				return r
			case DirectDeclaratorParen:
				if n.DirectDeclarator.Declarator.DirectDeclarator.Case == DirectDeclaratorIdent {
					for l := p.ParameterTypeList.ParameterList; l != nil; l = l.ParameterList {
						switch n := l.ParameterDeclaration; n.Case {
						case ParameterDeclarationAbstract: // DeclarationSpecifiers AbstractDeclaratorOpt
							r = append(r, 0)
						case ParameterDeclarationDeclarator: // DeclarationSpecifiers Declarator
							r = append(r, n.Declarator.Name())
						default:
							panic(n.Case)
						}
					}
					return r
				}

				return n.DirectDeclarator.Declarator.DirectDeclarator.parameterNames()
			default:
				panic(n.DirectDeclarator.Case)
			}
		default:
			panic(fmt.Errorf("TODO %v", p.Case))
		}
	//TODO case DirectDeclaratorArraySize: // DirectDeclarator '[' "static" TypeQualifierListOpt Expr ']'
	//TODO case DirectDeclaratorArraySize2: // DirectDeclarator '[' TypeQualifierList "static" Expr ']'
	//TODO case DirectDeclaratorArrayVar: // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
	//TODO case DirectDeclaratorArray: // DirectDeclarator '[' TypeQualifierListOpt ExprOpt ']'
	//TODO case DirectDeclaratorIdent: // IDENTIFIER
	default:
		panic(n.Case)
	}
}

func (n *BlockItemListOpt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) Operand {
	if n == nil {
		return Operand{Type: Void}
	}

	return n.BlockItemList.check(ctx, fn, seq, sc, inSwitch, inLoop, hasLabels)
}

func (n *BlockItemList) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) (r Operand) {
	for ; n != nil; n = n.BlockItemList {
		op := n.BlockItem.check(ctx, fn, seq, sc, inSwitch, inLoop, hasLabels)
		if n.BlockItemList == nil {
			r = op
		}
	}
	return r
}

func (n *BlockItem) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) (r Operand) {
	r = Operand{Type: Void}
	switch n.Case {
	case BlockItemDecl: // Declaration
		n.Declaration.check(ctx, seq, sc, fn, nil, false)
	case BlockItemStmt: // Stmt
		r = n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, hasLabels)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	return r
}

func (n *Stmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) (r Operand) {
	r = Operand{Type: Void}
	switch n.Case {
	case StmtBlock: // CompoundStmt
		n.CompoundStmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case StmtExpr: // ExprStmt
		r = n.ExprStmt.check(ctx, fn, seq, sc, &n.UseGotos)
	case StmtIter: // IterationStmt
		n.IterationStmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case StmtJump: // JumpStmt
		n.JumpStmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case StmtLabeled: // LabeledStmt
		n.LabeledStmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case StmtSelect: // SelectionStmt
		n.SelectionStmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	if n.UseGotos {
		*hasLabels = true
	}
	return r
}

func (n *LabeledStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) {
	//[0]6.8.1
	switch n.Case {
	case LabeledStmtSwitchCase: // "case" ConstExpr ':' Stmt
		op := n.ConstExpr.eval(ctx).ConvertTo(ctx.model, inSwitch.SwitchOp.Type)
		n.ConstExpr.Operand = op
		if op.Value == nil {
			panic("TODO")
		}
		if inSwitch == nil {
			panic("TODO")
		}
		inSwitch.Cases = append(inSwitch.Cases, n)
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case LabeledStmtDefault: // "default" ':' Stmt
		if inSwitch == nil {
			panic("TODO")
		}
		inSwitch.Cases = append(inSwitch.Cases, n)
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case LabeledStmtLabel: // IDENTIFIER ':' Stmt
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case LabeledStmtLabel2: // TYPEDEF_NAME ':' Stmt
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	n.UseGotos = true
	*hasLabels = true
}

func (n *SelectionStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) {
	switch n.Case {
	case SelectionStmtIfElse: // "if" '(' ExprList ')' Stmt "else" Stmt
		if !n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos).isScalarType() {
			panic("TODO")
		}
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
		n.Stmt2.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case SelectionStmtIf: // "if" '(' ExprList ')' Stmt
		if !n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos).isScalarType() {
			panic("TODO")
		}
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, inLoop, &n.UseGotos)
	case SelectionStmtSwitch: // "switch" '(' ExprList ')' Stmt
		// [0]6.8.4.2
		if !n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos).isIntegerType() {
			panic("TODO")
		}
		n.SwitchOp = n.ExprList.Operand.integerPromotion(ctx.model)
		n.Stmt.check(ctx, fn, seq, sc, n, inLoop, &n.UseGotos)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	if n.UseGotos {
		*hasLabels = true
	}
}

func (n *IterationStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) {
	switch n.Case {
	case IterationStmtDo: // "do" Stmt "while" '(' ExprList ')' ';'
		if !n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos).isScalarType() {
			panic(ctx.position)
		}
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, true, &n.UseGotos)
	case IterationStmtForDecl: // "for" '(' Declaration ExprListOpt ';' ExprListOpt ')' Stmt
		n.Declaration.check(ctx, seq, sc, fn, n.Declaration.Scope, false)
		n.ExprListOpt.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		if e := n.ExprListOpt.eval(ctx, true, fn, seq, sc, &n.UseGotos); e.Type != nil && !e.isScalarType() {
			panic(ctx.position(n))
		}
		n.ExprListOpt2.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, true, &n.UseGotos)
	case IterationStmtFor: // "for" '(' ExprListOpt ';' ExprListOpt ';' ExprListOpt ')' Stmt
		// [0]6.8.5.3
		n.ExprListOpt.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		if e := n.ExprListOpt2.eval(ctx, true, fn, seq, sc, &n.UseGotos); e.Type != nil && !e.isScalarType() {
			panic(ctx.position(n))
		}
		n.ExprListOpt3.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, true, &n.UseGotos)
	case IterationStmtWhile: // "while" '(' ExprList ')' Stmt
		if e := n.ExprList.eval(ctx, true, fn, seq, sc, &n.UseGotos); e.Type != nil && !e.isScalarType() {
			panic(ctx.position(n))
		}
		n.Stmt.check(ctx, fn, seq, sc, inSwitch, true, &n.UseGotos)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	if n.UseGotos {
		*hasLabels = true
	}
}

func (n *JumpStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, inSwitch *SelectionStmt, inLoop bool, hasLabels *bool) {
	switch n.Case {
	case JumpStmtBreak: // "break" ';'
		// [0]6.8.6.3
		//
		// 1. A break statement shall appear only in or as a switch
		// body or loop body.
		if inSwitch == nil && !inLoop {
			panic(ctx.position)
		}
	case JumpStmtContinue: // "continue" ';'
		// [0]6.8.6.2
		//
		// 1. A continue statement shall appear only in or as a loop
		// body.
		if !inLoop {
			panic(ctx.position(n))
		}
	case JumpStmtGoto: // "goto" IDENTIFIER ';'
		n.UseGotos = true
		if nm := n.Token2.Val; n.scope.LookupLabel(nm) == nil {
			panic(ctx.position(n))
		}
	case JumpStmtReturn: // "return" ExprListOpt ';'
		// [0]6.8.6.4
		op := n.ExprListOpt.eval(ctx, true, fn, seq, sc, &n.UseGotos)
		switch t := fn.Type.(*FunctionType).Result; t.Kind() {
		case Void:
			if op.Type != nil {
				if ctx.tweaks.EnableReturnExprInVoidFunc {
					break
				}

				panic(ctx.position(n))
			}
		default:
			if op.Type == nil {
				panic(ctx.position(n))
			}
			n.ReturnOperand = op.ConvertTo(ctx.model, t)
		}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	if n.UseGotos {
		*hasLabels = true
	}
}

func (n *ExprStmt) check(ctx *context, fn *Declarator, seq *int, sc []int, hasLabels *bool) Operand {
	// ExprListOpt ';'
	r := n.ExprListOpt.eval(ctx, true, fn, seq, sc, &n.UseGotos)
	if n.UseGotos {
		*hasLabels = true
	}
	return r
}

func (n *Declaration) check(ctx *context, seq *int, sc []int, fn *Declarator, scope *Scope, fnParam bool) []*Declarator {
	// DeclarationSpecifiers InitDeclaratorListOpt ';'
	ds := &DeclarationSpecifier{}
	n.DeclarationSpecifiers.check(ctx, ds)
	if len(ds.TypeSpecifiers) == 0 { // [0]6.7.2-2
		panic("TODO")
	}
	r := n.InitDeclaratorListOpt.check(ctx, ds, seq, sc, fn, scope)
	for _, v := range r {
		if fnParam {
			v.IsFunctionParameter = true
			continue
		}

		if v.Linkage == LinkageExternal && v.DeclarationSpecifier.IsExtern() {
			continue
		}

		switch x := UnderlyingType(v.Type).(type) {
		case *ArrayType:
			if x.Size.Type == nil {
				panic(fmt.Errorf("%v: %s has incomplete type: %v", ctx.position(v), dict.S(v.Name()), v.Type))
			}
		}
	}
	return r
}

func (n *InitDeclaratorListOpt) check(ctx *context, ds *DeclarationSpecifier, seq *int, sc []int, fn *Declarator, scope *Scope) []*Declarator {
	if n == nil {
		return nil
	}

	return n.InitDeclaratorList.check(ctx, ds, seq, sc, fn, scope)
}

func (n *InitDeclaratorList) check(ctx *context, ds *DeclarationSpecifier, seq *int, sc []int, fn *Declarator, scope *Scope) (r []*Declarator) {
	for ; n != nil; n = n.InitDeclaratorList {
		r = append(r, n.InitDeclarator.check(ctx, ds, seq, sc, fn, scope))
	}
	return r
}

func (n *InitDeclarator) check(ctx *context, ds *DeclarationSpecifier, seq *int, sc []int, fn *Declarator, scope *Scope) *Declarator {
	if scope != nil {
		n.Declarator.Scope = scope
	}
	switch n.Case {
	case InitDeclaratorBase: // Declarator
		n.Declarator.check(ctx, ds, ds.typ(ctx), !ds.IsTypedef(), sc, fn)
	case InitDeclaratorInit: // Declarator '=' Initializer
		if ds.IsTypedef() || ds.IsExtern() {
			panic(ctx.position(n)) // error
		}
		n.Declarator.Initializer = n.Initializer
		n.Declarator.check(ctx, ds, ds.typ(ctx), true, sc, fn)
		n.Initializer.check(ctx, n.Declarator.Type, fn, false, nil, seq, sc)
		if ex := n.Declarator.Scope.Idents[n.Declarator.Name()].(*Declarator); ex != nil && ex != n.Declarator {
			switch {
			case ex.Initializer == nil:
				ex.Initializer = n.Initializer
			default:
				if n.Initializer != nil {
					panic(fmt.Errorf("%v: existing initializer at %v", ctx.position(n), ctx.position(ex))) // More than one initializer
				}
			}
		}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	return n.Declarator
}

func (n *Initializer) check(ctx *context, t Type, fn *Declarator, field bool, arr *ArrayType, seq *int, sc []int) (r Operand) {
	var dummy bool
	// [0]6.7.8
	switch n.Case {
	case InitializerCompLit: // '{' InitializerList CommaOpt '}'
		return n.InitializerList.check(ctx, t, fn)
	case InitializerExpr: // Expr
		op := n.Expr.eval(ctx, true, fn, seq, sc, &dummy)
		switch {
		case t.Kind() == Function && op.Type.IsPointerType() && UnderlyingType(op.Type).(*PointerType).Item.Kind() == Function:
			t.assign(ctx, n, op)
			return n.Expr.Operand
		case op.isPointerType() && !t.IsPointerType() && arr != nil:
			t = arr
		case t.IsScalarType():
			// 11. The initializer for a scalar shall be a single
			// expression, optionally enclosed in braces. The
			// initial value of the object is that of the
			// expression (after conversion); the same type
			// constraints and conversions as for simple assignment
			// apply, taking the type of the scalar to be the
			// unqualified version of its declared type.
			t.assign(ctx, n, op)
			return n.Expr.Operand
		case op.IsZero():
			switch k := UnderlyingType(t).Kind(); k {
			case
				Array,
				Struct,
				Union:

				return op
			default:
				panic(fmt.Errorf("%v: %v", ctx.position(n.Expr), k))
			}
		}

		if t.Kind() == Struct || t.Kind() == Union {
			// 13. The initializer for a structure or union object
			// that has automatic storage duration shall be either
			// an initializer list as described below, or a single
			// expression that has compatible structure or union
			// type. In the latter case, the initial value of the
			// object, including unnamed members, is that of the
			// expression.
			if t.IsCompatible(op.Type) {
				return Operand{}
			}

			switch t.Kind() {
			case Union:
				x := UnderlyingType(t).(*UnionType)
				if x.Fields[0].Type.Kind() == Array && n.Expr.IsZero() {
					return op
				}

				if x.Fields[0].Type.IsCompatible(op.Type) {
					return op
				}
			}

			panic(fmt.Errorf("%v: typ %v, op type %v", ctx.position(n), t, op.Type))
		}

		if t.Kind() == Array {
			at := UnderlyingType(t).(*ArrayType)
			it := at.Item
			switch x := it.(type) {
			case TypeKind:
				switch x {
				case Char, UChar:
					if op.isPointerType() && op.Type.(*PointerType).Item == Char {
						// 14. An array of character type may be initialized by
						// a character string literal, optionally enclosed in
						// braces. Successive characters of the character
						// string literal (including the terminating null
						// character if there is room or if the array is of
						// unknown size) initialize the elements of the array.
						x := t.(*ArrayType)
						if !field && x.Size.Type == nil {
							switch y := op.Value.(type) {
							case *ir.StringValue:
								x.Size = newIntConst(ctx, n, uint64(len(dict.S(int(y.StringID)))+1), UInt, ULong, ULongLong)
							default:
								panic(fmt.Errorf("%v: TODO", ctx.position(n)))
							}
						}
						return op
					}
					panic(fmt.Errorf("%v:", ctx.position(n)))
				default:
					panic(fmt.Errorf("%v:", ctx.position(n)))
				}
			case *NamedType:
				switch {
				case x.Name == idWcharT:
					if op.isPointerType() && op.Type.(*PointerType).Item == UnderlyingType(x) {
						// 14. An array of character type may be initialized by
						// a character string literal, optionally enclosed in
						// braces. Successive characters of the character
						// string literal (including the terminating null
						// character if there is room or if the array is of
						// unknown size) initialize the elements of the array.
						if !field && at.Size.Type == nil {
							switch y := op.Value.(type) {
							case *ir.WideStringValue:
								at.Size = newIntConst(ctx, n, uint64(len(y.Value)+1), UInt, ULong, ULongLong)
							default:
								panic(fmt.Errorf("%v: TODO", ctx.position(n)))
							}
						}
						return op
					}
					panic(fmt.Errorf("%v:", ctx.position(n)))
				default:
					panic(fmt.Errorf("%v:", ctx.position(n)))
				}
			case *ArrayType:
				at2 := UnderlyingType(it).(*ArrayType)
				switch UnderlyingType(at2.Item).Kind() {
				case Char, UChar:
					if op.isPointerType() && op.Type.(*PointerType).Item == Char {
						// 14. An array of character type may be initialized by
						// a character string literal, optionally enclosed in
						// braces. Successive characters of the character
						// string literal (including the terminating null
						// character if there is room or if the array is of
						// unknown size) initialize the elements of the array.
						if !field && at2.Size.Type == nil {
							switch y := op.Value.(type) {
							case *ir.StringValue:
								at2.Size = newIntConst(ctx, n, uint64(len(dict.S(int(y.StringID)))+1), UInt, ULong, ULongLong)
							default:
								panic(fmt.Errorf("%v: TODO", ctx.position(n)))
							}
						}
						return op
					}
				case Array:
					panic(fmt.Errorf("%v: TODO Initializer t %v, op %v", ctx.position(n), at2, op))
				default:
					panic(fmt.Errorf("%v:", ctx.position(n)))
				}
			default:
				panic(fmt.Errorf("%v: %T", ctx.position(n), x))
			}
		}

		panic(fmt.Errorf("%v: TODO Initializer t %v, op %v", ctx.position(n), t, op))
	default:
		panic(fmt.Errorf("%v: TODO Initializer %v", ctx.position(n), n.Case))
	}
}

func (n *InitializerList) check(ctx *context, t Type, fn *Declarator) Operand {
	// InitializerList:
	//         /* empty */                                  // Case 0
	// |       Initializer                                  // Case 1
	// |       Designation Initializer                      // Case 2
	// |       InitializerList ',' Initializer              // Case 3
	// |       InitializerList ',' Designation Initializer  // Case 4
	r := &ir.CompositeValue{}
	n.Operand = Operand{Type: t, Value: r}
	n0 := n
	for {
		switch x := t.(type) {
		case *ArrayType:
			var index, maxIndex int64 = 0, -1
			for ; n != nil; n = n.InitializerList {
				n0.Len++
				if n.Designation != nil {
					dst, nt := n.Designation.check(ctx, x)
					_ = nt //TODO
					if len(dst) != 1 {
						panic(ctx.position(n))
					}

					index = dst[0]
					switch nv := int64(len(r.Values)); {
					case nv < index:
						r.Values = append(r.Values, make([]ir.Value, index-nv)...)
					case nv > index:
						r.Values[index] = n.Initializer.check(ctx, x.Item, fn, false, x, nil, nil)
						index++
						continue
					}
				}

				r.Values = append(r.Values, n.Initializer.check(ctx, x.Item, fn, false, x, nil, nil))
				if index > maxIndex {
					maxIndex = index
				}
				index++
			}
			if x.Size.Type == nil {
				x.Size = newIntConst(ctx, n0, uint64(maxIndex+1), UInt, ULong, ULongLong)
			}
			return Operand{Type: t, Value: r}
		case *NamedType:
			t = x.Type
		case *StructType:
			field := 0
			for ; n != nil; n = n.InitializerList {
				if n.Designation != nil {
					dst, nt := n.Designation.check(ctx, x)
					field = int(dst[0])
					switch nv := len(r.Values); {

					case nv < field:
						r.Values = append(r.Values, make([]ir.Value, field-nv)...)
					case nv > field:
						r.Values[field] = n.Initializer.check(ctx, nt, fn, true, nil, nil, nil)
						field++
						continue
					}
				}

				for x.Fields[field].Bits < 0 {
					field++
				}
				switch {
				case field < len(x.Fields):
					r.Values = append(r.Values, n.Initializer.check(ctx, x.Fields[field].Type, fn, true, nil, nil, nil))
					field++
				default:
					panic(fmt.Errorf("%v: TODO", ctx.position(n.Initializer)))
				}
				n0.Len++
			}
			return Operand{Type: t, Value: r}
		case *TaggedStructType:
			t = x.getType()
			if t == x {
				panic("TODO")
			}
		case *TaggedUnionType:
			t = x.getType()
			if t == x {
				panic("TODO")
			}
		case TypeKind:
			switch {
			case x.IsArithmeticType():
				switch n.Case {
				case 0: //  /* empty */
					if x.IsIntegerType() {
						return Operand{Type: t, Value: &ir.Int64Value{}}
					}

					switch t {
					case Double, LongDouble:
						return Operand{Type: t, Value: &ir.Float64Value{}}
					case Float:
						return Operand{Type: t, Value: &ir.Float64Value{}}
					default:
						panic(fmt.Errorf("%v: %v", ctx.position(n), t))
					}
				case 1: // Initializer
					return n.Initializer.check(ctx, t, fn, false, nil, nil, nil)
				default:
					panic(fmt.Errorf("%v: %v", ctx.position(n), n.Case))
				}
			default:
				panic("TODO")
			}
		case *UnionType:
			field := 0
			for ; n != nil; n = n.InitializerList {
				if n.Designation != nil {
					dst, nt := n.Designation.check(ctx, x)
					_ = nt //TODO
					if len(dst) != 1 {
						panic(ctx.position(n))
					}

					field = int(dst[0])
					switch nv := len(r.Values); {
					case nv < field:
						r.Values = append(r.Values, make([]ir.Value, field-nv)...)
					case nv > field:
						r.Values[field] = n.Initializer.check(ctx, x.Fields[field].Type, fn, true, nil, nil, nil)
						field++
						continue
					}
				}

				for x.Fields[field].Bits < 0 {
					field++
				}
				switch {
				case field < len(x.Fields):
					r.Values = append(r.Values, n.Initializer.check(ctx, x.Fields[field].Type, fn, true, nil, nil, nil))
					field++
				default:
					panic(fmt.Errorf("%v: TODO", ctx.position(n.Initializer)))
				}
				n0.Len++
			}
			return Operand{Type: t, Value: r}
		default:
			panic(fmt.Errorf("%v: TODO %T", ctx.position(n), x))
		}
	}
}

func (n *Designation) check(ctx *context, t Type) (r []int64, _ Type) {
	switch x := underlyingType(t, true).(type) {
	case *ArrayType:
		n.Type = x.Item
		for l := n.DesignatorList; l != nil; l = l.DesignatorList {
			switch d := l.Designator; d.Case {
			case DesignatorField: // '.' IDENTIFIER
				panic("TODO")
			case DesignatorIndex: // '[' ConstExpr ']'
				op := d.ConstExpr.eval(ctx)
				switch x := op.Value.(type) {
				case *ir.Int64Value:
					if v := x.Value; v < mathutil.MinInt || v > mathutil.MaxInt {
						panic("TODO")
					}

					r = append(r, x.Value)
				default:
					panic(fmt.Errorf("%v: %T", ctx.position(d), x))
				}
			default:
				panic("TODO")
			}
		}
	case *StructType:
		t := Type(x)
		for l := n.DesignatorList; l != nil; l = l.DesignatorList {
			switch d := l.Designator; d.Case {
			case DesignatorField: // '.' IDENTIFIER
				nm := d.Token2.Val
				var f *FieldProperties
				switch x := t.(type) {
				case *StructType:
					ctx.model.Layout(t)
					f = x.Field(nm)
					if f == nil || f.Type == nil {
						panic(fmt.Errorf("(C) %v: TODO %q", ctx.position(d.Token2), dict.S(nm)))
					}

					if fi := f.Declarator.Field; fi > len(x.Fields) || x.Fields[fi].Name != nm {
						panic("TODO")
					}
				case *UnionType:
					ctx.model.Layout(t)
					f = x.Field(nm)
					if f == nil || f.Type == nil {
						panic(fmt.Errorf("(C) %v: TODO %q", ctx.position(d.Token2), dict.S(nm)))
					}

					if fi := f.Declarator.Field; fi > len(x.Fields) || x.Fields[fi].Name != nm {
						panic("TODO")
					}
				default:
					panic(fmt.Errorf("%v", ctx.position(d)))
				}

				r = append(r, int64(f.Declarator.Field))
				t = f.Type
				n.Type = t
			case DesignatorIndex: // '[' ConstExpr ']'
				op := d.ConstExpr.eval(ctx)
				switch x := UnderlyingType(t).(type) {
				case *ArrayType:
					n.Type = x.Item
					switch y := op.Value.(type) {
					case *ir.Int64Value:
						if x.Size.Value == nil {
							panic("TODO")
						}

						if y.Value >= x.Size.Value.(*ir.Int64Value).Value {
							panic("TODO")
						}
						r = append(r, y.Value)
					default:
						panic(fmt.Errorf("%v: %v %T %v", ctx.position(d), t, y, op))
					}
				default:
					panic(fmt.Errorf("%v: %v %T %v", ctx.position(d), t, x, op))
				}
			default:
				panic("TODO")
			}
		}
	case *UnionType:
		t := Type(x)
		for l := n.DesignatorList; l != nil; l = l.DesignatorList {
			switch d := l.Designator; d.Case {
			case DesignatorField: // '.' IDENTIFIER
				nm := d.Token2.Val
				var f *FieldProperties
				switch x := t.(type) {
				//TODO case *StructType:
				//TODO 	f = x.Field(nm)
				//TODO 	if f == nil || f.Type == nil {
				//TODO 		panic(fmt.Errorf("(C) %v: TODO %q", ctx.position(n.Token2), dict.S(nm)))
				//TODO 	}

				//TODO 	if fi := f.Declarator.Field; fi > len(x.Fields) || x.Fields[fi].Name != nm {
				//TODO 		panic("TODO")
				//TODO 	}
				case *UnionType:
					ctx.model.Layout(t)
					f = x.Field(nm)
					if f == nil || f.Type == nil {
						panic(fmt.Errorf("(C) %v: TODO %q", ctx.position(d.Token2), dict.S(nm)))
					}

					if fi := f.Declarator.Field; fi > len(x.Fields) || x.Fields[fi].Name != nm {
						panic("TODO")
					}
				default:
					panic(fmt.Errorf("%v: %T", ctx.position(d), x))
				}

				r = append(r, int64(f.Declarator.Field))
				t = f.Type
				n.Type = t
			//TODO case DesignatorIndex: // '[' ConstExpr ']'
			//TODO 	panic("TODO")
			default:
				panic(d.Case)
			}
		}
		if len(r) > 1 {
			panic("TODO")
		}
	default:
		panic(fmt.Errorf("%v: %T", ctx.position(n), x))
	}
	n.List = r
	return r, n.Type
}

func (n *Declarator) check(ctx *context, ds *DeclarationSpecifier, t Type, isObject bool, sc []int, fn *Declarator) (r Type) {
	return n.check0(ctx, ds, t, isObject, sc, fn, nil)
}

func (n *Declarator) check0(ctx *context, ds *DeclarationSpecifier, t Type, isObject bool, sc []int, fn *Declarator, fnParams []*Declarator) (r Type) {
	// PointerOpt DirectDeclarator
	if l := len(sc); l != 0 {
		n.ScopeNum = sc[l-1]
	}
	n.DeclarationSpecifier = ds
	t = n.PointerOpt.check(ctx, t, &n.TypeQualifiers)
	n.Type = n.DirectDeclarator.check(ctx, t, sc, fn)
	isFunction := n.Type.Kind() == Function && !ds.IsTypedef()
	if fn != nil && !n.Embedded && !isFunction {
		fn.vars = append(fn.vars, n)
	}
	if isFunction && fnParams != nil {
		t := underlyingType(n.Type, true).(*FunctionType)
		if n := len(t.Params); n != 0 && n != len(fnParams) {
			panic("TODO")
		}

		n.Parameters = nil
		t.Params = make([]Type, len(fnParams))
	outer:
		for _, v := range fnParams {
			nm := v.Name()
			for i, w := range t.params {
				if nm == w {
					v.IsFunctionParameter = true
					t.Params[i] = v.Type
					continue outer
				}
			}
			panic("TODO")
		}
	}
	if n.Embedded {
		return n.Type
	}

	// [0]6.2.2
	switch {
	case
		// 3. If the declaration of a file scope identifier for an
		// object or a function contains the storage-class specifier
		// static, the identifier has internal linkage.
		(isObject || isFunction) && n.Scope == ctx.scope && ds.IsStatic():

		n.Linkage = LinkageInternal
	case
		// 4. For an identifier declared with the storage-class
		// specifier extern in a scope in which a prior declaration of
		// that identifier is visible, if the prior declaration
		// specifies internal or external linkage, the linkage of the
		// identifier at the later declaration is the same as the
		// linkage specified at the prior declaration. If no prior
		// declaration is visible, or if the prior declaration
		// specifies no linkage, then the identifier has external
		// linkage.
		(isObject || isFunction) && ds.IsExtern():

		n.Linkage = LinkageExternal
	case
		// 5. If the declaration of an identifier for a function has no
		// storage-class specifier, its linkage is determined exactly
		// as if it were declared with the storage-class specifier
		// extern. If the declaration of an identifier for an object
		// has file scope and no storage-class specifier, its linkage
		// is external.
		(isObject || isFunction) && (n.Scope == ctx.scope || isFunction) && len(ds.StorageClassSpecifiers) == 0:

		n.Linkage = LinkageExternal
	case
		// 6. The following identifiers have no linkage: an identifier
		// declared to be anything other than an object or a function;
		// an identifier declared to be a function parameter; a block
		// scope identifier for an object declared without the
		// storage-class specifier extern.
		!(isObject || isFunction),
		n.IsFunctionParameter,
		n.Scope != ctx.scope && !ds.IsExtern():

		n.Linkage = LinkageNone
	default:
		panic(ctx.position(n))
	}

	// [0]6.2.4
	switch {
	case
		// 3. An object whose identifier is declared with external or
		// internal linkage, or with the storage-class specifier static
		// has static storage duration. Its lifetime is the entire
		// execution of the program and its stored value is initialized
		// only once, prior to program startup.
		n.Linkage == LinkageExternal,
		n.Linkage == LinkageInternal,
		ds.IsStatic():

		n.StorageDuration = StorageDurationStatic
	case
		// 4. An object whose identifier is declared with no linkage
		// and without the storage-class specifier static has automatic
		// storage duration.
		n.Linkage == LinkageNone && !ds.IsStatic():

		n.StorageDuration = StorageDurationAutomatic
	default:
		panic(ctx.position(n))
	}

	if (ds.IsExtern() || isFunction) && n.Scope != ctx.scope {
		if ds.IsExtern() {
			n.insert(ctx, isFunction)
		}
		n.Scope = ctx.scope
	}
	n.insert(ctx, isFunction)
	return n.Type
}

func (n *Declarator) insert(ctx *context, isFunction bool) {
	nm := n.Name()
	switch ex := n.Scope.Idents[n.Name()]; ex := ex.(type) {
	case nil:
		n.Scope.insertDeclarator(ctx, n)
	case *Declarator:
		switch ex.Linkage {
		case LinkageNone:
			switch n.Linkage {
			case LinkageNone:
				if ex.DeclarationSpecifier.IsTypedef() && n.DeclarationSpecifier.IsTypedef() && ex.Type.IsCompatible(n.Type) {
					break
				}

				panic(fmt.Errorf("%v: %s %v, %v %v:", ctx.position(ex), dict.S(nm), ex.Type, n.Type, ctx.position(n)))
			default:
				panic(n.Linkage)
			}
		case LinkageExternal:
			switch n.Linkage {
			case LinkageExternal:
				if isFunction && (n.Scope.Parent != nil || n.FunctionDefinition == nil) {
					break
				}

				if !ex.Type.IsCompatible(n.Type) {
					if ex.Type.Kind() == Function && ex.FunctionDefinition == nil && n.FunctionDefinition != nil {
						if len(ex.Type.(*FunctionType).Params) == 0 {
							n.Scope.Idents[nm] = n
							break
						}
					}

					if !(n.Name() == idMain && n.Scope.Parent == nil && n.Type.Kind() == Function) {
						//IncompatibleTypeDiff(ex.Type, n.Type) //TODO-
						panic(fmt.Errorf("%v: %v\n%v: %v", ctx.position(ex), ex.Type, ctx.position(n), n.Type))
					}
					break
				}

				if prefer(n) || !prefer(ex) {
					n.Scope.Idents[nm] = n
				}
			case LinkageInternal:
				n.Scope.Idents[nm] = n // musl-1.1.20 src/time/__tz.c:23:41
			default:
				panic(fmt.Errorf("%v: %v %v %v:", ctx.position(n), n.Linkage, isFunction, ctx.position(ex)))
			}
		case LinkageInternal:
			switch n.Linkage {
			case LinkageInternal:
				if !ex.Type.IsCompatible(n.Type) {
					//IncompatibleTypeDiff(ex.Type, n.Type) //TODO-
					panic(fmt.Errorf("%v: %v\n%v: %v", ctx.position(ex), ex.Type, ctx.position(n), n.Type))
				}
			case LinkageExternal:
				ex.Linkage = n.Linkage
				n.Scope.Idents[nm] = n
			default:
				// Keep it internal.
			}
		default:
			panic(ex.Linkage)
		}

		if isFunction {
			switch {
			case n.FunctionDefinition != nil:
				if ex.FunctionDefinition != nil {
					panic(ctx.position(n))
				}

				ex.Definition = n
				n.Scope.Idents[nm] = n
			case ex.FunctionDefinition != nil:
				n.Definition = ex
			}
			break
		}

		switch {
		case n.Initializer != nil:
			if ex.Initializer != nil {
				panic(ctx.position(n))
			}

			ex.Definition = n
			n.Scope.Idents[nm] = n
		case ex.Initializer != nil:
			n.Definition = ex
		}
	default:
		panic(ctx.position(n))
	}
}

func (n *PointerOpt) check(ctx *context, t Type, tq *[]*TypeQualifier) Type {
	if n == nil {
		return t
	}

	return n.Pointer.check(ctx, t, tq)
}

func (n *Pointer) check(ctx *context, t Type, tq *[]*TypeQualifier) Type {
	n.TypeQualifierListOpt.check(ctx, tq)
	switch n.Case {
	case PointerBase: // '*' TypeQualifierListOpt
		return &PointerType{t, n.Pos()}
	case PointerPtr: // '*' TypeQualifierListOpt Pointer
		return n.Pointer.check(ctx, &PointerType{t, n.Pos()}, tq)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *TypeQualifierListOpt) check(ctx *context, tq *[]*TypeQualifier) {
	if n == nil {
		return
	}

	n.TypeQualifierList.check(ctx, tq)
}

func (n *TypeQualifierList) check(ctx *context, tq *[]*TypeQualifier) {
	s := *tq
	for ; n != nil; n = n.TypeQualifierList {
		s = append(s, n.TypeQualifier)
	}
	*tq = s
}

func (n *DirectDeclarator) check(ctx *context, t Type, sc []int, fn *Declarator) Type {
	var dummy bool
	switch n.Case {
	case DirectDeclaratorParen: // '(' Declarator ')'
		return n.Declarator.check(ctx, nil, t, false, sc, fn)
	case DirectDeclaratorParameters:
		switch p := n.Parameters; p.Case {
		case ParametersIdentList: // IdentifierListOpt
			t := &FunctionType{
				params: p.IdentifierListOpt.check(),
				Result: t,
				pos:    n.Pos(),
			}
			return n.DirectDeclarator.check(ctx, t, sc, fn)
		case ParametersParamTypes: // ParameterTypeList
			fp, variadic := p.ParameterTypeList.check(ctx)
			t := &FunctionType{
				Params:   fp,
				Result:   t,
				Variadic: variadic,
				pos:      n.Pos(),
			}
			return n.DirectDeclarator.check(ctx, t, sc, fn)
		default:
			panic(fmt.Errorf("%v: TODO %v", ctx.position(p), p.Case))
		}
	case DirectDeclaratorArraySize: // DirectDeclarator '[' "static" TypeQualifierListOpt Expr ']'
		var tq []*TypeQualifier
		n.TypeQualifierListOpt.check(ctx, &tq)
		sz := n.Expr.eval(ctx, true, fn, nil, nil, &dummy)
		t := &ArrayType{
			Item:           t,
			Length:         n.Expr,
			Size:           sz,
			TypeQualifiers: tq,
			pos:            n.Pos(),
		}
		return n.DirectDeclarator.check(ctx, t, sc, fn)
	//TODO case DirectDeclaratorArraySize2: // DirectDeclarator '[' TypeQualifierList "static" Expr ']'
	case DirectDeclaratorArrayVar: // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
		var tq []*TypeQualifier
		n.TypeQualifierListOpt.check(ctx, &tq)
		t := &ArrayType{
			Item:           t,
			TypeQualifiers: tq,
			pos:            n.Pos(),
		}
		return n.DirectDeclarator.check(ctx, t, sc, fn)
	case DirectDeclaratorArray: // DirectDeclarator '[' TypeQualifierListOpt ExprOpt ']'
		var tq []*TypeQualifier
		n.TypeQualifierListOpt.check(ctx, &tq)
		n.ExprOpt.eval(ctx, true, fn, nil, nil, &dummy)
		var sz Operand
		var e *Expr
		if o := n.ExprOpt; o != nil {
			e = o.Expr
			sz = e.Operand
		}
		t := &ArrayType{
			Item:           t,
			Length:         e,
			Size:           sz,
			TypeQualifiers: tq,
			pos:            n.Pos(),
		}
		return n.DirectDeclarator.check(ctx, t, sc, fn)
	case DirectDeclaratorIdent: // IDENTIFIER
		return t
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *IdentifierListOpt) check() []int {
	if n == nil {
		return nil
	}

	return n.IdentifierList.check()
}

func (n *IdentifierList) check() (r []int) {
	m := map[int]struct{}{}
	for ; n != nil; n = n.IdentifierList {
		nm := n.Token.Val
		if n.Token.Rune == ',' {
			nm = n.Token2.Val
		}

		if _, ok := m[nm]; ok {
			panic("TODO")
		}

		m[nm] = struct{}{}
		r = append(r, nm)
	}
	return r
}

func (n *ParameterTypeList) check(ctx *context) ([]Type, bool) {
	switch n.Case {
	case ParameterTypeListBase: // ParameterList
		return n.ParameterList.check(ctx), false
	case ParameterTypeListDots: // ParameterList ',' "..."
		return n.ParameterList.check(ctx), true
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *ParameterList) check(ctx *context) (r []Type) {
	for ; n != nil; n = n.ParameterList {
		r = append(r, n.ParameterDeclaration.check(ctx))
	}
	return r
}

func (n *ParameterDeclaration) check(ctx *context) Type {
	switch n.Case {
	case ParameterDeclarationAbstract: // DeclarationSpecifiers AbstractDeclaratorOpt
		ds := &DeclarationSpecifier{}
		n.DeclarationSpecifiers.check(ctx, ds)
		n.AbstractDeclaratorOpt.check(ctx, ds, ds.typ(ctx))
		if n.AbstractDeclaratorOpt == nil {
			return ds.typ(ctx)
		}

		return n.AbstractDeclaratorOpt.AbstractDeclarator.Type
	case ParameterDeclarationDeclarator: // DeclarationSpecifiers Declarator
		ds := &DeclarationSpecifier{}
		n.DeclarationSpecifiers.check(ctx, ds)
		n.Declarator.IsFunctionParameter = true
		return n.Declarator.check(ctx, ds, ds.typ(ctx), true, nil, nil)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *AbstractDeclaratorOpt) check(ctx *context, ds *DeclarationSpecifier, t Type) {
	if n == nil {
		return
	}

	n.AbstractDeclarator.check(ctx, ds, t)
}

func (n *AbstractDeclarator) check(ctx *context, ds *DeclarationSpecifier, t Type) Type {
	n.DeclarationSpecifier = ds
	switch n.Case {
	case AbstractDeclaratorPointer: // Pointer
		n.Type = n.Pointer.check(ctx, t, &n.TypeQualifiers)
	case AbstractDeclaratorAbstract: // PointerOpt DirectAbstractDeclarator
		t = n.PointerOpt.check(ctx, t, &n.TypeQualifiers)
		n.Type = n.DirectAbstractDeclarator.check(ctx, t)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	return n.Type
}

func (n *DirectAbstractDeclarator) check(ctx *context, t Type) Type {
	var dummy bool
	switch n.Case {
	case DirectAbstractDeclaratorAbstract: // '(' AbstractDeclarator ')'
		return n.AbstractDeclarator.check(ctx, nil, t)
	case DirectAbstractDeclaratorParamList: // '(' ParameterTypeListOpt ')'
		fp, variadic := n.ParameterTypeListOpt.check(ctx)
		t := &FunctionType{
			Params:   fp,
			Result:   t,
			Variadic: variadic,
			pos:      n.Pos(),
		}
		return t
	case DirectAbstractDeclaratorDFn: // DirectAbstractDeclarator '(' ParameterTypeListOpt ')'
		fp, variadic := n.ParameterTypeListOpt.check(ctx)
		t := &FunctionType{
			Params:   fp,
			Result:   t,
			Variadic: variadic,
			pos:      n.Pos(),
		}
		return n.DirectAbstractDeclarator.check(ctx, t)
	case DirectAbstractDeclaratorDArrSize: // DirectAbstractDeclaratorOpt '[' "static" TypeQualifierListOpt Expr ']'
		var tq []*TypeQualifier
		n.TypeQualifierListOpt.check(ctx, &tq)
		sz := n.Expr.eval(ctx, true, nil, nil, nil, &dummy)
		t := &ArrayType{
			Item:   t,
			Length: n.Expr,
			Size:   sz,
			pos:    n.Pos(),
		}
		if n.DirectAbstractDeclaratorOpt == nil {
			return t
		}

		return n.DirectAbstractDeclaratorOpt.DirectAbstractDeclarator.check(ctx, t)
	case DirectAbstractDeclaratorDArrVL: // DirectAbstractDeclaratorOpt '[' '*' ']'
		return nil //TODO
	case DirectAbstractDeclaratorDArr: // DirectAbstractDeclaratorOpt '[' ExprOpt ']'
		n.ExprOpt.eval(ctx, true, nil, nil, nil, &dummy)
		var sz Operand
		var e *Expr
		if o := n.ExprOpt; o != nil {
			e = o.Expr
			sz = e.Operand
		}
		t := &ArrayType{
			Item:   t,
			Length: e,
			Size:   sz,
			pos:    n.Pos(),
		}
		if n.DirectAbstractDeclaratorOpt == nil {
			return t
		}

		return n.DirectAbstractDeclaratorOpt.DirectAbstractDeclarator.check(ctx, t)
	//TODO case DirectAbstractDeclaratorDArrSize2: // DirectAbstractDeclaratorOpt '[' TypeQualifierList "static" Expr ']'
	case DirectAbstractDeclaratorDArr2: // DirectAbstractDeclaratorOpt '[' TypeQualifierList ExprOpt ']'
		var tq []*TypeQualifier
		n.TypeQualifierListOpt.check(ctx, &tq)
		n.ExprOpt.eval(ctx, true, nil, nil, nil, &dummy)
		var sz Operand
		var e *Expr
		if o := n.ExprOpt; o != nil {
			e = o.Expr
			sz = e.Operand
		}
		t := &ArrayType{
			Item:   t,
			Length: e,
			Size:   sz,
			pos:    n.Pos(),
		}
		if n.DirectAbstractDeclaratorOpt == nil {
			return t
		}

		return n.DirectAbstractDeclaratorOpt.DirectAbstractDeclarator.check(ctx, t)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *ParameterTypeListOpt) check(ctx *context) ([]Type, bool) {
	if n == nil {
		return nil, false
	}

	return n.ParameterTypeList.check(ctx)
}

func (n *DeclarationSpecifiers) check(ctx *context, ds *DeclarationSpecifier) {
	switch n.Case {
	case DeclarationSpecifiersFunc: // FunctionSpecifier DeclarationSpecifiersOpt
		n.FunctionSpecifier.check(ctx, ds)
	case DeclarationSpecifiersStorage: // StorageClassSpecifier DeclarationSpecifiersOpt
		n.StorageClassSpecifier.check(ctx, ds)
	case DeclarationSpecifiersQualifier: // TypeQualifier DeclarationSpecifiersOpt
		n.TypeQualifier.check(ctx, ds)
	case DeclarationSpecifiersSpecifier: // TypeSpecifier DeclarationSpecifiersOpt
		n.TypeSpecifier.check(ctx, ds)
	case DeclarationSpecifiersAlignment: // AlignmentSpecifier DeclarationSpecifiersOpt
		n.AlignmentSpecifier.check(ctx, ds)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	n.DeclarationSpecifiersOpt.check(ctx, ds)
}

func (n *AlignmentSpecifier) check(ctx *context, ds *DeclarationSpecifier) {
	switch n.Case {
	case AlignmentSpecifierTypeName: // "_Alignas" '(' TypeName ')'
		n.TypeName.check(ctx)
	case AlignmentSpecifierConstExpr: // "_Alignas" '(' ConstExpr ')'
		n.ConstExpr.eval(ctx)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	ds.AlignmentSpecifiers = append(ds.AlignmentSpecifiers, n)
}

func (n *FunctionSpecifier) check(ctx *context, ds *DeclarationSpecifier) {
	ds.FunctionSpecifiers = append(ds.FunctionSpecifiers, n)
}

func (n *TypeQualifier) check(ctx *context, ds *DeclarationSpecifier) {
	switch n.Case {
	case
		TypeQualifierAtomic,   // "_Atomic"
		TypeQualifierConst,    // "const"
		TypeQualifierRestrict, // "restrict"
		TypeQualifierVolatile: // "volatile"

		//nop
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	ds.TypeQualifiers = append(ds.TypeQualifiers, n)
}

func (n *TypeSpecifier) check(ctx *context, ds *DeclarationSpecifier) {
	var dummy bool
	switch n.Case {
	case TypeSpecifierAtomic: // "_Atomic"
		// _Atomic(int) -> _Atomic int
		ds.TypeQualifiers = append(ds.TypeQualifiers, &TypeQualifier{Case: TypeQualifierAtomic, Token: n.Token})
		n.typ = n.TypeName.check(ctx)
		switch x := n.typ.(type) {
		case TypeKind:
			switch x {
			case Int:
				n.Case = TypeSpecifierInt
			default:
				panic(fmt.Errorf("TODO %v", x))
			}
		case *PointerType:
			n.Case = -1
		default:
			panic(fmt.Errorf("TODO %T", x))
		}
	case
		TypeSpecifierBool,      // "_Bool"
		TypeSpecifierChar,      // "char"
		TypeSpecifierComplex,   // "_Complex"
		TypeSpecifierDouble,    // "double"
		TypeSpecifierFloat,     // "float"
		TypeSpecifierImaginary, // "_Imaginary"
		TypeSpecifierInt,       // "int"
		TypeSpecifierLong,      // "long"
		TypeSpecifierName,      // TYPEDEF_NAME
		TypeSpecifierShort,     // "short"
		TypeSpecifierSigned,    // "signed"
		TypeSpecifierUnsigned,  // "unsigned"
		TypeSpecifierVoid:      // "void"

		// nop
	case TypeSpecifierEnum: // EnumSpecifier
		n.EnumSpecifier.check(ctx)
	case TypeSpecifierStruct: // StructOrUnionSpecifier
		n.StructOrUnionSpecifier.check(ctx, ds)
	case TypeSpecifierTypeof: // "typeof" '(' TypeName ')'
		n.typ = n.TypeName.check(ctx)
	case TypeSpecifierTypeofExpr: // "typeof" '(' Expr ')'
		op := n.Expr.eval(ctx, false, nil, nil, nil, &dummy)
		n.typ = op.Type
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
	ds.TypeSpecifiers = append(ds.TypeSpecifiers, n)
}

func (n *EnumSpecifier) check(ctx *context) { // [0]6.7.2.2
	n.Tag = n.Token2.Val
	if n.IdentifierOpt != nil {
		n.Tag = n.IdentifierOpt.Token.Val
	}
	switch n.Case {
	case EnumSpecifierTag: // "enum" IDENTIFIER
		n.typ = &TaggedEnumType{Tag: n.Tag, scope: n.scope, pos: n.Pos()}
	case EnumSpecifierDefine: // "enum" IdentifierOpt '{' EnumeratorList CommaOpt '}'
		t := n.EnumeratorList.check(ctx, n.scope)
		t.Tag = n.Tag
		n.typ = t
		var min int64
		var max uint64
		for _, v := range t.Enums {
			w := v.Operand.Value.(*ir.Int64Value).Value
			u := uint64(w)
			if !v.Operand.Type.IsUnsigned() && w < 0 {
				if w < min {
					min = w
				}
				u = uint64(-w)
			}
			if u > max {
				max = u
			}
		}
		t.Min = min
		t.Max = max
		var x Operand
		//TODO- switch {
		//TODO- case min < 0:
		if u := uint64(-min); u > max {
			max = u
		}
		x = newIntConst(ctx, n, max /*TODO- Char, */, Int, Long, LongLong, ULongLong)
		//TODO- default:
		//TODO- 	x = newIntConst(ctx, n, max /*TODO- UChar, */, UInt, ULong, ULongLong)
		//TODO- }
		for i := range t.Enums {
			t.Enums[i].Operand.Type = x.Type
			t.Enums[i].Operand.IsEnumConst = true
		}
		if n.IdentifierOpt != nil {
			n.scope.insertEnumTag(ctx, n.IdentifierOpt.Token.Val, n)
		}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *EnumSpecifier) isCompatible(m *EnumSpecifier) bool {
	if n.Tag != m.Tag || (n.EnumeratorList != nil) != (m.EnumeratorList != nil) {
		return false
	}

	if n.typ != nil && m.typ != nil {
		return n.typ.IsCompatible(m.typ)
	}

	panic(fmt.Errorf("%v, %v", n, m))
}

func (n *EnumeratorList) check(ctx *context, s *Scope) *EnumType {
	r := &EnumType{pos: n.Pos()}
	iota := int64(-1)
	for ; n != nil; n = n.EnumeratorList {
		r.Enums = append(r.Enums, n.Enumerator.check(ctx, s, &iota))
	}
	return r
}

func (n *Enumerator) check(ctx *context, s *Scope, iota *int64) *EnumerationConstant {
	c := n.EnumerationConstant
	switch n.Case {
	case EnumeratorBase: // EnumerationConstant
		*iota++
		c.Operand = Operand{Value: &ir.Int64Value{Value: *iota}, Type: Int}
		s.insertEnumerationConstant(ctx, c)
		return c
	case EnumeratorInit: // EnumerationConstant '=' ConstExpr
		c.Operand = n.ConstExpr.eval(ctx)
		s.insertEnumerationConstant(ctx, c)
		switch x := c.Operand.Value.(type) {
		case *ir.Int64Value:
			*iota = x.Value
		default:
			panic(ctx.position(n))
		}
		return c
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *StructOrUnionSpecifier) check(ctx *context, parent *DeclarationSpecifier) {
	var tag int
	if n.IdentifierOpt != nil {
		tag = n.IdentifierOpt.Token.Val
	}
	switch n.Case {
	case StructOrUnionSpecifierTag: // StructOrUnion IDENTIFIER
		switch n.StructOrUnion.Case {
		case StructOrUnionStruct:
			n.typ = &TaggedStructType{Tag: n.Token.Val, scope: n.scope, pos: n.Pos()}
		case StructOrUnionUnion:
			n.typ = &TaggedUnionType{Tag: n.Token.Val, scope: n.scope, pos: n.Pos()}
		default:
			panic(ctx.position(n))
		}
	case StructOrUnionSpecifierEmpty: // StructOrUnion IdentifierOpt '{' '}'
		switch n.StructOrUnion.Case {
		case StructOrUnionStruct:
			n.typ = &StructType{structBase{Tag: tag}, n.Pos()}
		default:
			panic(ctx.position(n))
		}
		if tag != 0 {
			n.scope.Parent.insertStructTag(ctx, n)
		}
	case StructOrUnionSpecifierDefine: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
		switch n.StructOrUnion.Case {
		case StructOrUnionStruct:
			t := &StructType{structBase{Tag: tag, Fields: n.StructDeclarationList.check(ctx, parent), scope: n.scope}, n.Pos()}
			for i, v := range t.Fields {
				if x, ok := underlyingType(v.Type, true).(*ArrayType); ok {
					if x.Size.Value == nil && x.Length == nil {
						t.Fields[i].IsFlexibleArray = true
						if t.HasFlexibleArrayMember {
							ctx.err(v.Declarator, "struct has multiple flexible array members")
						}
						t.HasFlexibleArrayMember = true
						if i != len(t.Fields)-1 {
							ctx.err(v.Declarator, "flexible array member must be last")
						}
					}
				}
			}
			n.typ = t
		default:
			n.typ = &UnionType{structBase{Tag: tag, Fields: n.StructDeclarationList.check(ctx, parent), scope: n.scope}, n.Pos()}
		}
		if tag != 0 {
			n.scope.Parent.insertStructTag(ctx, n)
		}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *StructDeclarationList) check(ctx *context, parent *DeclarationSpecifier) (r []Field) {
	field := 0
	for ; n != nil; n = n.StructDeclarationList {
		r = append(r, n.StructDeclaration.check(ctx, parent, &field)...)
	}
	for len(r) > 0 && r[len(r)-1].Bits < 0 {
		r = r[:len(r)-1]
	}
	return r
}

func (n *StructDeclaration) check(ctx *context, parent *DeclarationSpecifier, field *int) []Field {
	switch n.Case {
	case StructDeclarationBase: // SpecifierQualifierList StructDeclaratorList ';'
		ds := &DeclarationSpecifier{Parent: parent}
		n.SpecifierQualifierList.check(ctx, ds)
		return n.StructDeclaratorList.check(ctx, ds, field)
	case StructDeclarationAnon: // SpecifierQualifierList ';'
		ds := &DeclarationSpecifier{Parent: parent}
		n.SpecifierQualifierList.check(ctx, ds)
		*field++
		return []Field{{Type: ds.typ(ctx), Anonymous: true}}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *StructDeclaratorList) check(ctx *context, ds *DeclarationSpecifier, field *int) (r []Field) {
	for ; n != nil; n = n.StructDeclaratorList {
		r = append(r, n.StructDeclarator.check(ctx, ds, *field))
		*field++
	}
	return r
}

func (n *StructDeclarator) check(ctx *context, ds *DeclarationSpecifier, field int) Field {
	switch n.Case {
	case StructDeclaratorBase: // Declarator
		f := Field{Type: n.Declarator.check(ctx, ds, ds.typ(ctx), false, nil, nil), Name: n.Declarator.Name(), Declarator: n.Declarator}
		n.Declarator.IsField = true
		n.Declarator.Field = field
		return f
	case StructDeclaratorBits: // DeclaratorOpt ':' ConstExpr
		var d *Declarator
		var nm int
		t := ds.typ(ctx)
		if n.DeclaratorOpt != nil {
			d = n.DeclaratorOpt.Declarator
			nm = d.Name()
			d.IsField = true
			d.Field = field
			t = d.check(ctx, ds, t, false, nil, nil)
		}
		op := n.ConstExpr.eval(ctx)
		if op.Value == nil {
			panic(ctx.position)
		}
		if !op.isIntegerType() {
			panic(ctx.position)
		}
		bits := op.Value.(*ir.Int64Value).Value
		if bits > 64 {
			panic(ctx.position)
		}

		if bits == 0 {
			bits = -1
		}
		n.Bits = int(bits)
		if d != nil {
			d.Bits = n.Bits
		}
		return Field{Type: t, Name: nm, Bits: n.Bits, Declarator: d}
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *SpecifierQualifierList) check(ctx *context, ds *DeclarationSpecifier) {
	switch n.Case {
	case SpecifierQualifierListQualifier: // TypeQualifier SpecifierQualifierListOpt
		n.TypeQualifier.check(ctx, ds)
		n.SpecifierQualifierListOpt.check(ctx, ds)
	case SpecifierQualifierListSpecifier: // TypeSpecifier SpecifierQualifierListOpt
		n.TypeSpecifier.check(ctx, ds)
		n.SpecifierQualifierListOpt.check(ctx, ds)
	default:
		panic(fmt.Errorf("%v: TODO %v", ctx.position(n), n.Case))
	}
}

func (n *SpecifierQualifierListOpt) check(ctx *context, ds *DeclarationSpecifier) {
	if n == nil {
		return
	}

	n.SpecifierQualifierList.check(ctx, ds)
}

func (n *StorageClassSpecifier) check(ctx *context, ds *DeclarationSpecifier) {
	if len(ds.StorageClassSpecifiers) != 0 {
		panic("TODO") // [0]6.7.1-2
	}
	ds.StorageClassSpecifiers = []*StorageClassSpecifier{n}
}

func (n *DeclarationSpecifiersOpt) check(ctx *context, ds *DeclarationSpecifier) {
	if n == nil {
		return
	}

	n.DeclarationSpecifiers.check(ctx, ds)
}

// IsTLD reports whether n is declared in file scope.
func (n *Declarator) IsTLD() bool { return n.Scope.Parent == nil }

func (n *Expr) isSideEffectsFree() bool {
	return false //TODO-

	switch n.Case {
	case
		ExprChar,       // CHARCONST
		ExprFloat,      // FLOATCONST
		ExprIdent,      // IDENTIFIER
		ExprInt,        // INTCONST
		ExprLChar,      // LONGCHARCONST
		ExprSizeofExpr, // "sizeof" Expr
		ExprSizeofType, // "sizeof" '(' TypeName ')'
		ExprString:     // STRINGLITERAL

		return true
	case ExprPExprList: // '(' ExprList ')'
		return n.ExprList.isSideEffectsFree()
	case
		ExprAddAssign, // Expr "+=" Expr
		ExprAndAssign, // Expr "&=" Expr
		ExprAssign,    // Expr '=' Expr
		ExprCall,      // Expr '(' ArgumentExprListOpt ')'
		ExprCompLit,   // '(' TypeName ')' '{' InitializerList CommaOpt '}' //TODO we can do better
		ExprDivAssign, // Expr "/=" Expr
		ExprLshAssign, // Expr "<<=" Expr
		ExprModAssign, // Expr "%=" Expr
		ExprMulAssign, // Expr "*=" Expr
		ExprOrAssign,  // Expr "|=" Expr
		ExprPostDec,   // Expr "--"
		ExprPostInc,   // Expr "++"
		ExprPreDec,    // "--" Expr
		ExprPreInc,    // "++" Expr
		ExprRshAssign, // Expr ">>=" Expr
		ExprStatement, // '(' CompoundStmt ')' //TODO we can do better
		ExprSubAssign, // Expr "-=" Expr
		ExprXorAssign: // Expr "^=" Expr

		return false
	case ExprCast: // '(' TypeName ')' Expr
		return !isVaList(n.Expr.Operand.Type) && n.Expr.isSideEffectsFree()
	case ExprCond: // Expr '?' ExprList ':' Expr
		if !n.Expr.isSideEffectsFree() {
			return false
		}

		switch {
		case n.Expr.Operand.IsNonZero():
			return n.ExprList.isSideEffectsFree()
		case n.Expr.Operand.IsZero():
			return n.Expr2.isSideEffectsFree()
		}
		return false
	case
		ExprAdd, // Expr '+' Expr
		ExprAnd, // Expr '&' Expr
		ExprDiv, // Expr '/' Expr
		ExprEq,  // Expr "==" Expr
		ExprGe,  // Expr ">=" Expr
		ExprGt,  // Expr ">" Expr
		ExprLe,  // Expr "<=" Expr
		ExprLsh, // Expr "<<" Expr
		ExprLt,  // Expr '<' Expr
		ExprMod, // Expr '%' Expr
		ExprMul, // Expr '*' Expr
		ExprNe,  // Expr "!=" Expr
		ExprOr,  // Expr '|' Expr
		ExprRsh, // Expr ">>" Expr
		ExprSub, // Expr '-' Expr
		ExprXor: // Expr '^' Expr

		return n.Expr.isSideEffectsFree() && n.Expr2.isSideEffectsFree()
	case ExprLAnd: // Expr "&&" Expr
		return n.Expr.isSideEffectsFree() && n.Expr2.isSideEffectsFree()
	case ExprLOr: // Expr "||" Expr
		return n.Expr.isSideEffectsFree() && n.Expr2.isSideEffectsFree()
	case
		ExprAddrof,     // '&' Expr
		ExprCpl,        // '~' Expr
		ExprDeref,      // '*' Expr
		ExprNot,        // '!' Expr
		ExprPSelect,    // Expr "->" IDENTIFIER
		ExprSelect,     // Expr '.' IDENTIFIER
		ExprUnaryMinus, // '-' Expr
		ExprUnaryPlus:  // '+' Expr

		return n.Expr.isSideEffectsFree()
	case ExprIndex: // Expr '[' ExprList ']'
		return n.Expr.isSideEffectsFree() && n.ExprList.isSideEffectsFree()
	default:
		panic(fmt.Sprint(n.Case, n.Operand))
	}
}

func (n *ExprList) isSideEffectsFree() bool {
	for l := n; l != nil; l = l.ExprList {
		if !l.Expr.isSideEffectsFree() {
			return false
		}
	}

	return true
}

// IsNonZero reports whether n is provably non zero.
func (n *ExprList) IsNonZero() bool {
	for n.ExprList != nil {
		n = n.ExprList
	}
	return n.Expr.IsNonZero()
}

// IsNonZero reports whether n is provably non zero.
func (n *Expr) IsNonZero() bool {
	return n.Operand.IsNonZero() //TODO-

	if n.Operand.IsNonZero() {
		return true
	}

	switch n.Case {
	case
		ExprAdd,        // Expr '+' Expr
		ExprAddAssign,  // Expr "+=" Expr
		ExprAnd,        // Expr '&' Expr
		ExprAndAssign,  // Expr "&=" Expr
		ExprCall,       // Expr '(' ArgumentExprListOpt ')'
		ExprChar,       // CHARCONST
		ExprDeref,      // '*' Expr
		ExprDiv,        // Expr '/' Expr
		ExprEq,         // Expr "==" Expr
		ExprFloat,      // FLOATCONST
		ExprGe,         // Expr ">=" Expr
		ExprGt,         // Expr '>' Expr
		ExprIdent,      // IDENTIFIER
		ExprIndex,      // Expr '[' ExprList ']'
		ExprInt,        // INTCONST
		ExprLe,         // Expr "<=" Expr
		ExprLsh,        // Expr "<<" Expr
		ExprLt,         // Expr '<' Expr
		ExprMod,        // Expr '%' Expr
		ExprMul,        // Expr '*' Expr
		ExprPSelect,    // Expr "->" IDENTIFIER
		ExprPostDec,    // Expr "--"
		ExprPostInc,    // Expr "++"
		ExprPreDec,     // "--" Expr
		ExprPreInc,     // Expr "++"
		ExprRsh,        // Expr ">>" Expr
		ExprRshAssign,  // Expr ">>=" Expr
		ExprSelect,     // Expr '.' IDENTIFIER
		ExprSizeofExpr, // "sizeof" Expr
		ExprSizeofType, // "sizeof" '(' TypeName ')'
		ExprString,     // STRINGLITERAL
		ExprSubAssign,  // Expr "-=" Expr
		ExprSub:        // Expr '-' Expr

		return false
	case ExprAddrof: // '&' Expr
		n = n.Expr
		for {
			switch n.Case {
			case ExprIdent: // IDENTIFIER
				return true
			case ExprIndex: // Expr '[' ExprList ']'
				n = n.Expr
			case ExprPSelect: // Expr "->" IDENTIFIER
				return false
			default:
				panic(fmt.Errorf("Expr.IsNonZero %s", n.Case))
			}
		}
	case ExprAssign: // Expr '=' Expr
		return n.Expr2.IsNonZero()
	case ExprCast: // '(' TypeName ')' Expr
		return n.Expr.IsNonZero()
	case ExprCond: // Expr '?' ExprList ':' Expr
		return n.Expr.IsNonZero() && n.ExprList.IsNonZero() || n.Expr.IsZero() && n.Expr2.IsNonZero()
	case ExprCpl: // '~' Expr
		return n.Expr.IsZero()
	case ExprLAnd: // Expr "&&" Expr
		return n.Expr.IsNonZero() && n.Expr2.IsNonZero()
	case ExprLOr: // Expr "||" Expr
		return n.Expr.IsNonZero() || n.Expr2.IsNonZero()
	case ExprNe: // Expr "!=" Expr
		return n.Expr.IsZero() && n.Expr2.IsNonZero() || n.Expr.IsNonZero() && n.Expr2.IsZero()
	case ExprNot: // '!' Expr
		return n.Expr.IsZero()
	case ExprOr: // Expr '|' Expr
		return n.Expr.IsNonZero() || n.Expr2.IsNonZero()
	case ExprOrAssign: // Expr "|=" Expr
		return n.Expr.IsNonZero() || n.Expr2.IsNonZero()
	case ExprPExprList: // '(' ExprList ')'
		return n.ExprList.IsNonZero()
	case ExprUnaryMinus: // '-' Expr
		return n.Expr.IsNonZero()
	case ExprUnaryPlus: // '+' Expr
		return n.Expr.IsNonZero()
	case ExprXor: // Expr '^' Expr
		return n.Expr.IsNonZero() && n.Expr2.IsZero() || n.Expr.IsZero() && n.Expr2.IsNonZero()
	case ExprXorAssign: // Expr "^=" Expr
		return n.Expr.IsNonZero() && n.Expr2.IsZero() || n.Expr.IsZero() && n.Expr2.IsNonZero()
	default:
		panic(fmt.Errorf("%T.IsNonZero %v", n, n.Case))
	}
}

// IsZero reports whether n is provably zero.
func (n *ExprList) IsZero() bool {
	for n.ExprList != nil {
		n = n.ExprList
	}
	return n.Expr.IsZero()
}

// IsZero reports whether n is provably zero.
func (n *Expr) IsZero() bool {
	return n.Operand.IsZero() //TODO-

	if n.Operand.IsZero() {
		return true
	}

	switch n.Case {
	case
		ExprAdd,        // Expr '+' Expr
		ExprAddAssign,  // Expr "+=" Expr
		ExprAddrof,     // '&' Expr
		ExprCall,       // Expr '(' ArgumentExprListOpt ')'
		ExprChar,       // CHARCONST
		ExprCompLit,    // '(' TypeName ')' '{' InitializerList CommaOpt '}' //TODO we can do better
		ExprCpl,        // '~' Expr
		ExprDeref,      // '*' Expr
		ExprDiv,        // Expr '/' Expr
		ExprEq,         // Expr "==" Expr
		ExprFloat,      // FLOATCONST
		ExprGe,         // Expr ">=" Expr
		ExprGt,         // Expr '>' Expr
		ExprIdent,      // IDENTIFIER
		ExprIndex,      // Expr '[' ExprList ']'
		ExprInt,        // INTCONST
		ExprLe,         // Expr "<=" Expr
		ExprLsh,        // Expr "<<" Expr
		ExprLt,         // Expr '<' Expr
		ExprPSelect,    // Expr "->" IDENTIFIER
		ExprPostDec,    // "--" Expr
		ExprPostInc,    // Expr "++"
		ExprPreDec,     // "--" Expr
		ExprPreInc,     // Expr "++"
		ExprRsh,        // Expr ">>" Expr
		ExprRshAssign,  // Expr ">>=" Expr
		ExprSubAssign,  // Expr "-=" Expr
		ExprSelect,     // Expr '.' IDENTIFIER
		ExprSizeofExpr, // "sizeof" Expr
		ExprSizeofType, // "sizeof" '(' TypeName ')'
		ExprString,     // STRINGLITERAL
		ExprSub:        // Expr '-' Expr

		return false
	case ExprAnd: // Expr '&' Expr
		return n.Expr.IsZero() || n.Expr2.IsZero()
	case ExprAndAssign: // Expr "&=" Expr
		return n.Expr.IsZero() || n.Expr2.IsZero()
	case ExprAssign: // Expr '=' Expr
		return n.Expr2.IsZero()
	case ExprCast: // '(' TypeName ')' Expr
		return n.Expr.IsZero()
	case ExprCond: // Expr '?' ExprList ':' Expr
		return n.Expr.IsNonZero() && n.ExprList.IsZero() || n.Expr.IsNonZero() && n.Expr2.IsZero()
	case ExprLAnd: // Expr "&&" Expr
		return n.Expr.IsZero() || n.Expr2.IsZero()
	case ExprLOr: // Expr "||" Expr
		return n.Expr.IsZero() && n.Expr2.IsZero()
	case ExprMod: // Expr '%' Expr
		return n.Expr.IsZero()
	case ExprMul: // Expr '*' Expr
		return n.Expr.IsZero() || n.Expr2.IsZero()
	case ExprNe: // Expr "!=" Expr
		return n.Expr.IsZero() && n.Expr2.IsZero()
	case ExprNot: // '!' Expr
		return n.Expr.IsNonZero()
	case ExprOr: // Expr '|' Expr
		return n.Expr.IsZero() && n.Expr2.IsZero()
	case ExprOrAssign: // Expr "|=" Expr
		return n.Expr.IsZero() && n.Expr2.IsZero()
	case ExprPExprList: // '(' ExprList ')'
		return n.ExprList.IsZero()
	case ExprUnaryMinus: // '-' Expr
		return n.Expr.IsZero()
	case ExprUnaryPlus: // '+' Expr
		return n.Expr.IsZero()
	case ExprXor: // Expr '^' Expr
		return n.Expr.IsZero() && n.Expr2.IsZero() || n.Expr.Equals(n.Expr2)
	case ExprXorAssign: // Expr "^=" Expr
		return n.Expr.IsZero() && n.Expr2.IsZero() || n.Expr.Equals(n.Expr2)
	default:
		panic(fmt.Errorf("%T.IsZero %v", n, n.Case))
	}
}

func (n *ExprList) equals(m *ExprList) bool {
	if (n.ExprList == nil) != (m.ExprList == nil) {
		return false
	}

	for l, k := n, m; l != nil; l, k = l.ExprList, k.ExprList {
		if (l.ExprList == nil) != (k.ExprList == nil) || !l.Expr.Equals(k.Expr) {
			return false
		}
	}
	return true
}

func (n *ExprList) equals2(m *Expr) bool {
	for n.ExprList != nil {
		n = n.ExprList
	}
	return n.Expr.Equals(m)
}

// Equals reports whether s equals m.
func (n *Expr) Equals(m *Expr) bool {
	return false //TODO-

	if n.Operand.Type != m.Operand.Type || n.Operand.Value != m.Operand.Value || !n.isSideEffectsFree() || !m.isSideEffectsFree() {
		return false
	}

	if n.Case != m.Case {
		switch {
		case n.Case == ExprPExprList:
			return n.ExprList.equals2(m)
		case m.Case == ExprPExprList:
			return m.ExprList.equals2(n)
		}
		return false
	}

	switch n.Case {
	case
		ExprPSelect, // Expr "->" IDENTIFIER
		ExprSelect:  // Expr '.' IDENTIFIER

		return n.Expr.Equals(m.Expr) && n.Token.Val == m.Token.Val && n.Token2.Val == m.Token2.Val
	case ExprIdent: // IDENTIFIER
		return n.Token.Val == m.Token.Val
	case ExprPExprList: // '(' ExprList ')'
		return n.ExprList.equals(m.ExprList)
	case ExprInt: // INTCONST
		return true
	case // unary
		ExprCast,      // '(' TypeName ')' Expr
		ExprCpl,       // '~' Expr
		ExprDeref,     // '*' Expr
		ExprNot,       // '!' Expr
		ExprUnaryPlus: // '+' Expr

		return n.Expr.Equals(m.Expr)
	case ExprIndex: // Expr '[' ExprList ']'
		return n.Expr.Equals(m.Expr) && n.ExprList.equals(m.ExprList)
	case // binary commutative
		ExprAdd, // Expr '+' Expr
		ExprAnd, // Expr '&' Expr
		ExprEq,  // Expr "==" Expr
		ExprGe,  // Expr "==" Expr
		ExprLe,  // Expr "==" Expr
		ExprLt,  // Expr "==" Expr
		ExprMul, // Expr '*' Expr
		ExprNe,  // Expr "==" Expr
		ExprOr,  // Expr '|' Expr
		ExprXor: // Expr '^' Expr

		return n.Expr.Equals(m.Expr) && n.Expr2.Equals(m.Expr2) || n.Expr.Equals(m.Expr2) && n.Expr2.Equals(m.Expr)
	case // binary
		ExprDiv,  // Expr '/' Expr
		ExprGt,   // Expr '>' Expr
		ExprLAnd, // Expr "&&" Expr
		ExprLOr,  // Expr "||" Expr
		ExprLsh,  // Expr "<<" Expr
		ExprMod,  // Expr '%' Expr
		ExprRsh,  // Expr ">>" Expr
		ExprSub:  // Expr '-' Expr

		return n.Expr.Equals(m.Expr) && n.Expr2.Equals(m.Expr2)
	case ExprCall: // Expr '(' ArgumentExprListOpt ')'
		if !n.Expr.Equals(m.Expr) {
			return false
		}

		if (n.ArgumentExprListOpt == nil) != (m.ArgumentExprListOpt == nil) {
			return false
		}

		if n.ArgumentExprListOpt == nil {
			return true
		}

		for l, k := n.ArgumentExprListOpt.ArgumentExprList, m.ArgumentExprListOpt.ArgumentExprList; l != nil; l, k = l.ArgumentExprList, k.ArgumentExprList {
			if (l.ArgumentExprList == nil) != (k.ArgumentExprList == nil) || !l.Expr.Equals(k.Expr) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Errorf("%T.Equal %v", n, n.Case))
	}
}

func (n *EnumerationConstant) equal(m *EnumerationConstant) (r bool) {
	return n.Token.Val == m.Token.Val && n.Operand.Value.(*ir.Int64Value).Value == m.Operand.Value.(*ir.Int64Value).Value
}

// DumpOperands dumps the operands of n to stdout.
func (n *ExprList) DumpOperands(s string) {
	for l := n; l != nil; l = l.ExprList {
		l.Expr.DumpOperands(s + "· ")
	}
}

// DumpOperands dumps the operands of n to stdout.
func (n *Expr) DumpOperands(s string) {
	z := ""
	switch {
	case n.IsZero():
		z = ".Z"
	case n.IsNonZero():
		z = ".NZ"
	}
	fmt.Printf("%s%v%v %v\n", s, n.Case, z, n.Operand)
	switch n.Case {
	case
		ExprAddrof,
		ExprCast,
		ExprCpl,
		ExprDeref,
		ExprIndex,
		ExprNot,
		ExprPSelect,
		ExprPostDec,
		ExprPostInc,
		ExprPreDec,
		ExprPreInc,
		ExprSelect,
		ExprSizeofExpr,
		ExprUnaryMinus,
		ExprUnaryPlus:

		n.Expr.DumpOperands(s + "· ")
	case
		ExprAdd,
		ExprAnd,
		ExprAndAssign,
		ExprAssign,
		ExprDiv,
		ExprEq,
		ExprGe,
		ExprGt,
		ExprLAnd,
		ExprLOr,
		ExprLe,
		ExprLsh,
		ExprLt,
		ExprMod,
		ExprMul,
		ExprNe,
		ExprOr,
		ExprOrAssign,
		ExprRsh,
		ExprSub,
		ExprXor,
		ExprXorAssign:

		n.Expr.DumpOperands(s + "· ")
		n.Expr2.DumpOperands(s + "· ")
	case ExprPExprList:
		n.ExprList.DumpOperands(s + "· ")
	case ExprCond:
		n.Expr.DumpOperands(s + "· ")
		n.ExprList.DumpOperands(s + "· ")
		n.Expr2.DumpOperands(s + "· ")
	case
		ExprCall,
		ExprChar,
		ExprFloat,
		ExprIdent,
		ExprInt,
		ExprSizeofType,
		ExprString:

		// nop
	default:
		panic(n.Case.String())
	}
}
