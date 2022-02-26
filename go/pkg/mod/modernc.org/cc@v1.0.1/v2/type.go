// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"bytes"
	"fmt"
	"go/token"

	"modernc.org/ir"
)

var (
	_ Type = (*ArrayType)(nil)
	_ Type = (*EnumType)(nil)
	_ Type = (*FunctionType)(nil)
	_ Type = (*NamedType)(nil)
	_ Type = (*PointerType)(nil)
	_ Type = (*StructType)(nil)
	_ Type = (*TaggedStructType)(nil)
	_ Type = (*TaggedUnionType)(nil)
	_ Type = (*TaggedEnumType)(nil)
	_ Type = (*UnionType)(nil)
	_ Type = TypeKind(0)

	_ fieldFinder = (*StructType)(nil)
	_ fieldFinder = (*UnionType)(nil)
)

type fieldFinder interface {
	findField(int) *FieldProperties
}

// Type represents a C type.
type Type interface {
	Node
	Equal(Type) bool
	IsArithmeticType() bool
	IsCompatible(Type) bool // [0]6.2.7
	IsIntegerType() bool
	IsPointerType() bool
	IsScalarType() bool
	IsUnsigned() bool
	IsVoidPointerType() bool
	Kind() TypeKind
	String() string
	assign(ctx *context, n Node, op Operand) Operand
}

// TypeKind represents a particular type kind.
type TypeKind int

func (t TypeKind) Pos() token.Pos { return token.Pos(0) }

// TypeKind values.
const (
	_ TypeKind = iota

	Bool
	Char
	Int
	Long
	LongLong
	SChar
	Short
	UChar
	UInt
	ULong
	ULongLong
	UShort

	Float
	Double
	LongDouble

	FloatComplex
	DoubleComplex
	LongDoubleComplex

	FloatImaginary
	DoubleImaginary
	LongDoubleImaginary

	Array
	Enum
	EnumTag
	Function
	Ptr
	Struct
	StructTag
	TypedefName
	Union
	UnionTag
	Void

	maxTypeKind
)

// IsUnsigned implements Type.
func (t TypeKind) IsUnsigned() bool { return t.IsIntegerType() && !isSigned[t] }

// Kind implements Type.
func (t TypeKind) Kind() TypeKind { return t }

// assign implements Type.
func (t TypeKind) assign(ctx *context, n Node, op Operand) Operand {
	// [0]6.5.16.1
	switch {
	// One of the following shall hold:
	case
		// the left operand has qualified or unqualified arithmetic
		// type and the right has arithmetic type;
		t.IsArithmeticType() && op.Type.IsArithmeticType():
		return op.ConvertTo(ctx.model, t)
	default:
		panic(fmt.Sprintf("%v: %v <- %v", ctx.position(n), t, op))
	}
}

// IsPointerType implements Type.
func (t TypeKind) IsPointerType() bool {
	if t.IsArithmeticType() {
		return false
	}

	switch t {
	case Void:
		return false
	default:
		panic(t)
	}
}

// IsIntegerType implements Type.
func (t TypeKind) IsIntegerType() bool {
	switch t {
	case
		Bool,
		Char,
		DoubleImaginary,
		FloatImaginary,
		Int,
		Long,
		LongDoubleImaginary,
		LongLong,
		SChar,
		Short,
		UChar,
		UInt,
		ULong,
		ULongLong,
		UShort:

		return true
	case
		Double,
		DoubleComplex,
		Float,
		FloatComplex,
		LongDouble,
		LongDoubleComplex,
		Void:

		return false
	default:
		panic(t)
	}
}

// IsScalarType implements Type.
func (t TypeKind) IsScalarType() bool {
	switch t {
	case
		Char,
		Double,
		DoubleComplex,
		Float,
		FloatComplex,
		Int,
		Long,
		LongDouble,
		LongDoubleComplex,
		LongLong,
		SChar,
		Short,
		UChar,
		UInt,
		ULong,
		ULongLong,
		UShort:

		return true
	case
		Bool,
		Void:

		return false
	default:
		panic(t)
	}
}

// IsVoidPointerType implements Type.
func (t TypeKind) IsVoidPointerType() bool {
	if t != Ptr {
		return false
	}

	panic(t)
}

// IsArithmeticType implements Type.
func (t TypeKind) IsArithmeticType() bool { return isArithmeticType[t] }

// IsCompatible implements Type.
func (t TypeKind) IsCompatible(u Type) bool {
	for {
		switch x := u.(type) {
		case *PointerType:
			switch t {
			case
				Char,
				Int:

				return false
			default:
				panic(fmt.Errorf("%v %v", t, x))
			}
		case *NamedType:
			u = x.Type
		case *EnumType:
			u = x.Enums[0].Operand.Type
		case TypeKind:
			if t.IsIntegerType() && u.IsIntegerType() {
				return intConvRank[t] == intConvRank[x] && isSigned[t] == isSigned[x]
			}

			switch x {
			case
				Double,
				DoubleComplex,
				Float,
				FloatComplex,
				LongDouble,
				LongDoubleComplex:

				return t == x
			case Void:
				return t == x
			default:
				panic(fmt.Errorf("%v", x))
			}
		default:
			panic(fmt.Errorf("%v %T", t, x))
		}
	}
}

// Equal implements Type.
func (t TypeKind) Equal(u Type) bool {
	switch x := u.(type) {
	case
		*ArrayType,
		*EnumType,
		*FunctionType,
		*PointerType,
		*StructType,
		*TaggedEnumType,
		*TaggedStructType,
		*TaggedUnionType,
		*UnionType:

		if t.IsScalarType() || t == Void {
			return false
		}

		panic(t)
	case *NamedType:
		return t.Equal(x.Type)
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

			return t == x
		default:
			panic(x)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func (t TypeKind) String() string {
	switch t {
	case Bool:
		return "bool"
	case Char:
		return "char"
	case Int:
		return "int"
	case Long:
		return "long"
	case LongLong:
		return "long long"
	case SChar:
		return "signed char"
	case Short:
		return "short"
	case UChar:
		return "unsigned char"
	case UInt:
		return "unsigned"
	case ULong:
		return "unsigned long"
	case ULongLong:
		return "unsigned long long"
	case UShort:
		return "unsigned short"
	case Float:
		return "float"
	case Double:
		return "double"
	case LongDouble:
		return "long double"
	case FloatImaginary:
		return "float imaginary"
	case DoubleImaginary:
		return "double imaginary"
	case LongDoubleImaginary:
		return "long double imaginary"
	case FloatComplex:
		return "float complex"
	case DoubleComplex:
		return "double complex"
	case LongDoubleComplex:
		return "long double complex"
	case Array:
		return "array"
	case Enum:
		return "enum"
	case EnumTag:
		return "enum tag"
	case Function:
		return "function"
	case Ptr:
		return "ptr"
	case Struct:
		return "struct"
	case StructTag:
		return "struct tag"
	case TypedefName:
		return "typedef name"
	case Union:
		return "union"
	case Void:
		return "void"
	default:
		return fmt.Sprintf("TypeKind(%v)", int(t))
	}
}

// ArrayType type represents an array type.
type ArrayType struct {
	Item           Type
	Length         *Expr
	Size           Operand
	TypeQualifiers []*TypeQualifier // Eg. double a[restrict 3][5], see 6.7.5.3-21.
	pos            token.Pos
}

func (t *ArrayType) Pos() token.Pos { return t.pos }

func (t *ArrayType) IsVLA() bool { return t.Length != nil && t.Length.Operand.Value == nil }

// IsUnsigned implements Type.
func (t *ArrayType) IsUnsigned() bool { panic("TODO") }

// IsVoidPointerType implements Type.
func (t *ArrayType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *ArrayType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *ArrayType) IsCompatible(u Type) bool {
	// [0]6.7.5.2
	//
	// 6. For two array types to be compatible, both shall have compatible
	// element types, and if both size specifiers are present, and are
	// integer constant expressions, then both size specifiers shall have
	// the same constant value. If the two array types are used in a
	// context which requires them to be compatible, it is undefined
	// behavior if the two size specifiers evaluate to unequal values.
	switch x := u.(type) {
	case *ArrayType:
		if !t.Item.IsCompatible(x.Item) {
			return false
		}

		if t.Size.Type != nil && x.Size.Type != nil {
			return t.Size.Value.(*ir.Int64Value).Value == x.Size.Value.(*ir.Int64Value).Value
		}

		return true
	case *NamedType:
		return x.IsCompatible(t)
	case *PointerType:
		return t.Size.Type == nil && t.Item.IsCompatible(x.Item)
	case TypeKind:
		if x.IsArithmeticType() {
			return false
		}

		panic(fmt.Errorf("%T\n%v\n%v", x, t, u))
	default:
		panic(fmt.Errorf("%T\n%v\n%v", x, t, u))
	}
}

// Equal implements Type.
func (t *ArrayType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case *ArrayType:
		if !t.Item.Equal(x.Item) {
			return false
		}

		switch {
		case t.Size.Type != nil && x.Size.Type != nil:
			return x.Size.Type != nil && t.Size.Value.(*ir.Int64Value).Value == x.Size.Value.(*ir.Int64Value).Value
		case t.Size.Type == nil && x.Size.Type == nil:
			return true
		default:
			panic(fmt.Errorf("%v, %v", t, u))
		}
	case *NamedType:
		return x.Type.Equal(t)
	case
		*FunctionType,
		*PointerType,
		*StructType,
		*TaggedStructType:

		return false
	case TypeKind:
		if x.IsScalarType() || x == Void {
			return false
		}

		panic(x)
	default:
		panic(fmt.Errorf("%T %v", x, x))
	}
}

// Kind implements Type.
func (t *ArrayType) Kind() TypeKind { return Array }

// assign implements Type.
func (t *ArrayType) assign(ctx *context, n Node, op Operand) Operand {
	switch {
	case t.Size.Value == nil:
		return (&PointerType{Item: t.Item}).assign(ctx, n, op)
	default:
		panic("TODO")
	}
}

// IsPointerType implements Type.
func (t *ArrayType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *ArrayType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *ArrayType) IsScalarType() bool { return false }

func (t *ArrayType) String() string {
	switch {
	case t.Size.Type != nil && t.Size.Value != nil:
		return fmt.Sprintf("array of %v %v", t.Size.Value, t.Item)
	default:
		return fmt.Sprintf("array of %v", t.Item)
	}
}

// EnumType represents an enum type.
type EnumType struct {
	Tag   int
	Enums []*EnumerationConstant
	Min   int64
	Max   uint64
	pos   token.Pos
}

func (t *EnumType) Pos() token.Pos { return t.pos }

func (t *EnumType) find(nm int) *EnumerationConstant {
	for _, v := range t.Enums {
		if v.Token.Val == nm {
			return v
		}
	}
	return nil
}

// IsUnsigned implements Type.
func (t *EnumType) IsUnsigned() bool { return t.Enums[0].Operand.Type.IsUnsigned() }

// IsVoidPointerType implements Type.
func (t *EnumType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *EnumType) IsArithmeticType() bool { return true }

// IsCompatible implements Type.
func (t *EnumType) IsCompatible(u Type) bool {
	switch x := underlyingType(u, true).(type) {
	case *EnumType:
		if len(t.Enums) != len(x.Enums) {
			return false
		}

		for i, v := range t.Enums {
			if !v.equal(x.Enums[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// Equal implements Type.
func (t *EnumType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := UnderlyingType(u).(type) {
	case *EnumType:
		if t.Tag != x.Tag || len(t.Enums) != len(x.Enums) {
			return false
		}

		for i, v := range t.Enums {
			w := x.Enums[i]
			if !v.Operand.Type.Equal(w.Operand.Type) || v.Operand.Value.(*ir.Int64Value).Value != w.Operand.Value.(*ir.Int64Value).Value {
				return false
			}
		}

		return true
	case TypeKind:
		if x.IsScalarType() {
			return false
		}

		panic(fmt.Errorf("%v %T %v", t, x, x))
	default:
		panic(fmt.Errorf("%v %T %v", t, x, x))
	}
}

// Kind implements Type.
func (t *EnumType) Kind() TypeKind { return Enum }

// assign implements Type.
func (t *EnumType) assign(ctx *context, n Node, op Operand) Operand {
	return op.ConvertTo(ctx.model, t)
}

// IsPointerType implements Type.
func (t *EnumType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *EnumType) IsIntegerType() bool { return true }

// IsScalarType implements Type.
func (t *EnumType) IsScalarType() bool { return true }

func (t *EnumType) String() string {
	return fmt.Sprintf("%s enumeration %s", t.Enums[0].Operand.Type.String(), dict.S(t.Tag))
}

// Field represents a struct/union field.
type Field struct {
	Bits       int
	Declarator *Declarator
	Name       int
	PackedType Type // Bits != 0: underlaying struct field type
	Type       Type

	Anonymous       bool
	IsFlexibleArray bool
}

func (f Field) equal(g Field) bool {
	return f.Name == g.Name && f.Type.Equal(g.Type) && f.Bits == g.Bits
}

func (f Field) isCompatiblel(g Field) bool {
	return f.Name == g.Name && f.Type.IsCompatible(g.Type) && f.Bits == g.Bits
}

func (f Field) String() string { return fmt.Sprintf("%s %v", dict.S(f.Name), f.Type) }

// FunctionType represents a function type.
type FunctionType struct {
	Params   []Type
	Result   Type
	Variadic bool
	params   []int
	pos      token.Pos
}

func (t *FunctionType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *FunctionType) IsUnsigned() bool { panic("TODO") }

// IsVoidPointerType implements Type.
func (t *FunctionType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *FunctionType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *FunctionType) IsCompatible(u Type) bool {
	switch x := u.(type) {
	case *FunctionType:
		if t.Variadic != x.Variadic || !t.Result.IsCompatible(x.Result) {
			return false
		}

		if len(t.Params) != len(x.Params) {
			switch {
			case len(t.Params) == 1 && len(x.Params) == 0 && t.Params[0].Kind() == Void:
				return true
			case len(t.Params) == 0 && len(x.Params) == 1 && x.Params[0].Kind() == Void:
				return true
			}
			return false
		}

		for i, t := range t.Params {
			if !t.IsCompatible(x.Params[i]) {
				return false
			}
		}
		return true
	case *NamedType:
		return x.Type.IsCompatible(t)
	case *PointerType:
		return x.Item.IsCompatible(t)
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// Equal implements Type.
func (t *FunctionType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case *FunctionType:
		if len(t.Params) != len(x.Params) || t.Variadic != x.Variadic || !t.Result.Equal(x.Result) {
			return false
		}

		for i, t := range t.Params {
			if !t.Equal(x.Params[i]) {
				return false
			}
		}
		return true
	case
		*NamedType,
		*PointerType,
		*StructType,
		*TaggedStructType:

		return false
	case TypeKind:
		switch x {
		case
			Char,
			Int,
			Void:

			return false
		default:
			panic(x)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// Kind implements Type.
func (t *FunctionType) Kind() TypeKind { return Function }

// assign implements Type.
func (t *FunctionType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := UnderlyingType(op.Type).(type) {
	case *PointerType:
		switch y := UnderlyingType(x.Item).(type) {
		case *FunctionType:
			if !t.Equal(y) {
				panic(fmt.Errorf("%v: %v != %v", ctx.position(n), t, y))
			}

			return op
		default:
			panic(fmt.Errorf("%v: %T", ctx.position(n), y))
		}
	default:
		panic(fmt.Errorf("%v: %T", ctx.position(n), x))
	}
}

// IsPointerType implements Type.
func (t *FunctionType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *FunctionType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *FunctionType) IsScalarType() bool { return false }

func (t *FunctionType) String() string {
	var buf bytes.Buffer
	buf.WriteString("function (")
	switch {
	case len(t.Params) == 1 && t.Params[0].Kind() == Void:
		// nop
	default:
		for i, v := range t.Params {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(v.String())
		}
	}
	fmt.Fprintf(&buf, ") returning %v", t.Result)
	return buf.String()
}

// NamedType represents a type described by a typedef name.
type NamedType struct {
	Name int
	Type Type // The type Name refers to.
	pos  token.Pos
}

func (t *NamedType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *NamedType) IsUnsigned() bool { return t.Type.IsUnsigned() }

// IsVoidPointerType implements Type.
func (t *NamedType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *NamedType) IsArithmeticType() bool { return t.Type.IsArithmeticType() }

// IsCompatible implements Type.
func (t *NamedType) IsCompatible(u Type) (r bool) {
	return UnderlyingType(t.Type).IsCompatible(UnderlyingType(u))
}

// Equal implements Type.
func (t *NamedType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case *ArrayType:
		return x.Equal(t.Type)
	case *EnumType:
		return x.Equal(t.Type)
	case *NamedType:
		return t.Name == x.Name && t.Type.Equal(x.Type)
	case
		*FunctionType,
		*PointerType:

		return x.Equal(t.Type)
	case *StructType:
		return t.Type.Equal(x)
	case *TaggedEnumType:
		v := x.getType()
		return v != x && t.Type.Equal(v)
	case *TaggedStructType:
		v := x.getType()
		return v != x && t.Type.Equal(v)
	case *TaggedUnionType:
		v := x.getType()
		return v != x && t.Type.Equal(v)
	case TypeKind:
		switch x {
		case
			Bool,
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

			return x.Equal(t.Type)
		default:
			panic(x)
		}
	case *UnionType:
		return t.Type.Equal(x)
	default:
		panic(fmt.Errorf("%T: %v, %v", x, t.Type, u))
	}
}

// Kind implements Type.
func (t *NamedType) Kind() TypeKind { return t.Type.Kind() }

// assign implements Type.
func (t *NamedType) assign(ctx *context, n Node, op Operand) Operand { return t.Type.assign(ctx, n, op) }

// IsPointerType implements Type.
func (t *NamedType) IsPointerType() bool { return t.Type.IsPointerType() }

// IsIntegerType implements Type.
func (t *NamedType) IsIntegerType() bool { return t.Type.IsIntegerType() }

// IsScalarType implements Type.
func (t *NamedType) IsScalarType() bool { return t.Type.IsScalarType() }

func (t *NamedType) String() string { return string(dict.S(t.Name)) }

// PointerType represents a pointer type.
type PointerType struct {
	Item Type
	pos  token.Pos
}

func (t *PointerType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *PointerType) IsUnsigned() bool { return true }

// IsVoidPointerType implements Type.
func (t *PointerType) IsVoidPointerType() bool { return UnderlyingType(t.Item) == Void }

// IsArithmeticType implements Type.
func (t *PointerType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *PointerType) IsCompatible(u Type) bool {
	if t.Equal(u) {
		return true
	}

	var ai Type
	switch x := UnderlyingType(t.Item).(type) {
	case *ArrayType:
		ai = x.Item
	}

	switch x := UnderlyingType(u).(type) {
	case *ArrayType:
		return t.Item.IsCompatible(x.Item)
	case *NamedType:
		return t.IsCompatible(x.Type)
	case *FunctionType:
		return x.IsCompatible(t)
	case *PointerType:
		// [0]6.3.2.3
		//
		// 1. A pointer to void may be converted to or from a pointer to any
		// incomplete or object type. A pointer to any incomplete or object
		// type may be converted to a pointer to void and back again; the
		// result shall compare equal to the original pointer.
		return t.Item == Void || x.Item == Void ||
			ai != nil && ai.IsCompatible(x.Item) ||
			underlyingType(t.Item, true).IsCompatible(underlyingType(x.Item, true)) ||
			Unsigned(t.Item) == Unsigned(x.Item)
	case *StructType:
		return false
	case TypeKind:
		if u.IsArithmeticType() {
			return false
		}

		panic(fmt.Errorf("%T\n%v\n%v", x, t, u))
	default:
		panic(fmt.Errorf("%T\n%v\n%v", x, t, u))
	}
}

func Unsigned(t Type) TypeKind {
	switch k := underlyingType(t, false).Kind(); k {
	case Char, SChar:
		return UChar
	case Short:
		return UShort
	case Int:
		return UInt
	case Long:
		return ULong
	case LongLong:
		return ULongLong
	default:
		return k
	}
}

// Equal implements Type.
func (t *PointerType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case
		*ArrayType,
		*FunctionType,
		*StructType,
		*TaggedStructType,
		*UnionType:

		return false
	case *NamedType:
		return t.Equal(x.Type)
	case *PointerType:
		return t.Item.Equal(x.Item)
	case TypeKind:
		switch x {
		case
			Char,
			Double,
			Float,
			Int,
			Long,
			LongLong,
			Short,
			UChar,
			UInt,
			ULong,
			ULongLong,
			UShort,
			Void:

			return false
		default:
			panic(x)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// Kind implements Type.
func (t *PointerType) Kind() TypeKind { return Ptr }

// assign implements Type.
func (t *PointerType) assign(ctx *context, n Node, op Operand) (r Operand) {
	u := UnderlyingType(t.Item)
	var v Type
	if op.Type.IsPointerType() {
		v = UnderlyingType(UnderlyingType(op.Type).(*PointerType).Item)
	}
	// [0]6.5.16.1
	switch {
	// One of the following shall hold:
	case ctx.tweaks.EnablePointerCompatibility && op.Type.IsPointerType():
		return op.ConvertTo(ctx.model, t)
	case
		// both operands are pointers to qualified or unqualified
		// versions of compatible types, and the type pointed to by the
		// left has all the qualifiers of the type pointed to by the
		// right;
		op.Type.IsPointerType() && t.IsCompatible(op.Type),
		u.IsIntegerType() && v != nil && v.IsIntegerType() && ctx.model.Sizeof(u) == ctx.model.Sizeof(v) && u.IsUnsigned() == v.IsUnsigned():

		return op.ConvertTo(ctx.model, t)
	case
		// one operand is a pointer to an object or incomplete type and
		// the other is a pointer to a qualified or unqualified version
		// of void, and the type pointed to by the left has all the
		// qualifiers of the type pointed to by the right;
		t.IsPointerType() && op.Type.IsVoidPointerType():

		panic("TODO")
	case
		// the left operand is a pointer and the right is a null
		// pointer constant;
		op.isNullPtrConst():

		return Operand{Type: t, Value: Null}
	default:
		panic(fmt.Errorf("%v: lhs type %v <- operand %v, op.IsPointerType %v op.IsCompatible %v", ctx.position(n), t, op, op.Type.IsPointerType(), t.IsCompatible(op.Type)))
	}
}

// IsPointerType implements Type.
func (t *PointerType) IsPointerType() bool { return true }

// IsIntegerType implements Type.
func (t *PointerType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *PointerType) IsScalarType() bool { return true }

func (t *PointerType) String() string { return fmt.Sprintf("pointer to %v", t.Item) }

type structBase struct {
	Fields                 []Field
	HasFlexibleArrayMember bool
	Tag                    int
	scope                  *Scope
	layout                 []FieldProperties
}

func (s *structBase) findField(nm int) *FieldProperties {
	fps := s.layout
	if x, ok := s.scope.Idents[nm].(*Declarator); ok {
		r := fps[x.Field]
		return &r
	}

	for _, v := range fps {
		if v.Declarator != nil {
			continue
		}

		x, ok := v.Type.(fieldFinder)
		if !ok {
			continue
		}

		if fp := x.findField(nm); fp != nil {
			r := *fp
			r.Offset += v.Offset
			return &r
		}
	}
	return nil
}

// StructType represents a struct type.
type StructType struct {
	structBase
	pos token.Pos
}

func (t *StructType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *StructType) IsUnsigned() bool { panic("TODO") }

// Field returns the properties of field nm or nil if the field does not exist.
func (t *StructType) Field(nm int) *FieldProperties {
	return t.findField(nm)
}

// IsVoidPointerType implements Type.
func (t *StructType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *StructType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *StructType) IsCompatible(u Type) bool {
	if t.Equal(u) {
		return true
	}

	// [0]6.2.7.1
	//
	// Moreover, two structure, union, or enumerated types declared in
	// separate translation units are compatible if their tags and members
	// satisfy the following requirements: If one is declared with a tag,
	// the other shall be declared with the same tag. If both are complete
	// types, then the following additional requirements apply: there shall
	// be a one-to-one correspondence between their members such that each
	// pair of corresponding members are declared with compatible types,
	// and such that if one member of a corresponding pair is declared with
	// a name, the other member is declared with the same name. For two
	// structures, corresponding members shall be declared in the same
	// order. For two structures or unions, corresponding bit-fields shall
	// have the same widths.
	switch x := UnderlyingType(u).(type) {
	case *StructType:
		//TODO if len(t.Fields) != len(x.Fields) || t.Tag != x.Tag { // Fails in _sqlite/src/test_fs.c
		if len(t.Fields) != len(x.Fields) {
			return false
		}

		for i, v := range t.Fields {
			if !v.isCompatiblel(x.Fields[i]) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// Equal implements Type.
func (t *StructType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case *NamedType:
		return t.Equal(x.Type)
	case
		*ArrayType,
		*FunctionType,
		*PointerType:

		return false
	case *StructType:
		if len(t.Fields) != len(x.Fields) || t.Tag != x.Tag {
			return false
		}

		for i, v := range t.Fields {
			if !v.equal(x.Fields[i]) {
				return false
			}
		}
		return true
	case *TaggedStructType:
		v := x.getType()
		if v == u {
			return false
		}

		return t.Equal(v)
	case TypeKind:
		switch x {
		case
			Char,
			Int,
			Long,
			UChar,
			UInt,
			UShort,
			Void:

			return false
		default:
			panic(x)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// Kind implements Type.
func (t *StructType) Kind() TypeKind { return Struct }

// assign implements Type.
func (t *StructType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := op.Type.(type) {
	case *StructType:
		if !t.IsCompatible(x) {
			panic("TODO")
		}

		return Operand{Type: t}
	case *NamedType:
		if !t.IsCompatible(x.Type) {
			panic("TODO")
		}

		return Operand{Type: t}
	default:
		panic(fmt.Errorf("%v: %T %v", ctx.position(n), x, x))
	}
}

// IsPointerType implements Type.
func (t *StructType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *StructType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *StructType) IsScalarType() bool { return false }

func (t *StructType) String() string {
	var buf bytes.Buffer
	buf.WriteString("struct{")
	for i, v := range t.Fields {
		if i != 0 {
			buf.WriteString("; ")
		}
		fmt.Fprintf(&buf, "%s %s", dict.S(v.Name), v.Type)
		if v.Bits != 0 {
			fmt.Fprintf(&buf, ".%d", v.Bits)
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

// TaggedEnumType represents an enum type described by a tag name.
type TaggedEnumType struct {
	Tag   int
	Type  Type
	scope *Scope
	pos   token.Pos
}

func (t *TaggedEnumType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *TaggedEnumType) IsUnsigned() bool { return t.Type.IsUnsigned() }

// Equal implements Type.
func (t *TaggedEnumType) Equal(u Type) bool {
	switch x := UnderlyingType(u).(type) {
	case *EnumType:
		switch y := t.getType(); {
		case y == t:
			panic("TODO")
		default:
			return y.Equal(u)
		}
	case *TaggedEnumType:
		if t.Tag == x.Tag {
			return true
		}

		panic("TODO")
	case TypeKind:
		if x.IsScalarType() {
			return false
		}

		panic(fmt.Errorf("%v", x))
	case *PointerType:
		return false
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// IsArithmeticType implements Type.
func (t *TaggedEnumType) IsArithmeticType() bool { return true }

// IsCompatible implements Type.
func (t *TaggedEnumType) IsCompatible(u Type) bool {
	switch x := t.getType(); {
	case x == t:
		panic("TODO")
	default:
		return underlyingType(x, true).IsCompatible(underlyingType(u, true))
	}
}

// IsIntegerType implements Type.
func (t *TaggedEnumType) IsIntegerType() bool { return true }

// IsPointerType implements Type.
func (t *TaggedEnumType) IsPointerType() bool { return false }

// IsScalarType implements Type.
func (t *TaggedEnumType) IsScalarType() bool { return true }

// IsVoidPointerType implements Type.
func (t *TaggedEnumType) IsVoidPointerType() bool { panic("TODO") }

// Kind implements Type.
func (t *TaggedEnumType) Kind() TypeKind { return Int }

func (t *TaggedEnumType) String() string { return fmt.Sprintf("enum %s", dict.S(t.Tag)) }

// assign implements Type.
func (t *TaggedEnumType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := UnderlyingType(op.Type).(type) {
	case *EnumType:
		//TODO use IsCompatible
		op.Type = t
		return op
	case TypeKind:
		if x.IsIntegerType() {
			op.Type = t
			return op
		}

		panic(x)
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func (t *TaggedEnumType) getType() Type {
	if t.Type != nil {
		return t.Type
	}

	s := t.scope.lookupEnumTag(t.Tag)
	if s == nil {
		return t
	}

	t.Type = s.typ
	return t.Type
}

// TaggedStructType represents a struct type described by a tag name.
type TaggedStructType struct {
	Tag   int
	Type  Type
	scope *Scope
	pos   token.Pos
}

func (t *TaggedStructType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *TaggedStructType) IsUnsigned() bool { panic("TODO") }

// IsVoidPointerType implements Type.
func (t *TaggedStructType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *TaggedStructType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *TaggedStructType) IsCompatible(u Type) bool {
	if t == u {
		return true
	}

	if x, ok := u.(*TaggedStructType); ok && t.Tag == x.Tag {
		return true
	}

	return false
}

// Equal implements Type.
func (t *TaggedStructType) Equal(u Type) bool {
	if t == u {
		return true
	}

	if x, ok := u.(*TaggedStructType); ok && t.Tag == x.Tag {
		return true
	}

	switch x := t.getType().(type) {
	case *StructType:
		return x.Equal(u)
	case *TaggedStructType:
		if x == t {
			switch y := u.(type) {
			case *NamedType:
				return t.Equal(y.Type)
			case *StructType:
				return false
			case *TaggedStructType:
				return t.Tag == y.Tag
			case TypeKind:
				switch y {
				case Void:
					return false
				default:
					panic(y)
				}
			default:
				panic(fmt.Errorf("%T", y))
			}
		}

		panic("TODO")
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func (t *TaggedStructType) getType() Type {
	if t.Type != nil {
		return t.Type
	}

	s := t.scope.lookupStructTag(t.Tag)
	if s == nil {
		return t
	}

	t.Type = s.typ
	return t.Type
}

// Kind implements Type.
func (t *TaggedStructType) Kind() TypeKind { return Struct }

// assign implements Type.
func (t *TaggedStructType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := op.Type.(type) {
	case *NamedType:
		op.Type = x.Type
		return t.assign(ctx, n, op)
	case *TaggedStructType:
		t2 := t.getType()
		u2 := x.getType()
		if t2 != t && u2 != x {
			// [0]6.5.16.1
			//
			// the left operand has a qualified or unqualified
			// version of a structure or union type compatible with
			// the type of the right;
			if t2.Equal(u2) {
				return op
			}

			panic("TODO")
		}
		panic("TODO")
	case *StructType:
		t2 := t.getType()
		if t2.Equal(x) {
			return op
		}
		panic(fmt.Errorf("%v: %T %v, %T %v", ctx.position(n), t2, t2, x, x))
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// IsPointerType implements Type.
func (t *TaggedStructType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *TaggedStructType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *TaggedStructType) IsScalarType() bool { return false }

func (t *TaggedStructType) String() string { return fmt.Sprintf("struct %s", dict.S(t.Tag)) }

// UnionType represents a union type.
type UnionType struct {
	structBase
	pos token.Pos
}

func (t *UnionType) Pos() token.Pos { return t.pos }

// Field returns the properties of field nm or nil if the field does not exist.
func (t *UnionType) Field(nm int) *FieldProperties {
	return t.findField(nm)
}

// IsUnsigned implements Type.
func (t *UnionType) IsUnsigned() bool { panic("TODO") }

// TaggedUnionType represents a struct type described by a tag name.
type TaggedUnionType struct {
	Tag   int
	Type  Type
	scope *Scope
	pos   token.Pos
}

func (t *TaggedUnionType) Pos() token.Pos { return t.pos }

// IsUnsigned implements Type.
func (t *TaggedUnionType) IsUnsigned() bool { panic("TODO") }

// IsVoidPointerType implements Type.
func (t *TaggedUnionType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *TaggedUnionType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *TaggedUnionType) IsCompatible(u Type) bool { return t.Equal(u) }

// Equal implements Type.
func (t *TaggedUnionType) Equal(u Type) bool {
	if t == u {
		return true
	}

	if x, ok := u.(*TaggedUnionType); ok && t.Tag == x.Tag {
		return true
	}

	switch x := t.getType().(type) {
	case *UnionType:
		return x.Equal(u)
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func (t *TaggedUnionType) getType() Type {
	if t.Type != nil {
		return t.Type
	}

	s := t.scope.lookupStructTag(t.Tag)
	if s == nil {
		return t
	}

	t.Type = s.typ
	return t.Type
}

// Kind implements Type.
func (t *TaggedUnionType) Kind() TypeKind { return Union }

// assign implements Type.
func (t *TaggedUnionType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := op.Type.(type) {
	case *TaggedUnionType:
		t2 := t.getType()
		u2 := x.getType()
		if t2 != t && u2 != x {
			// [0]6.5.16.1
			//
			// the left operand has a qualified or unqualified
			// version of a structure or union type compatible with
			// the type of the right;
			if t2.Equal(u2) {
				return op
			}

			panic("TODO")
		}
		panic("TODO")
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// IsPointerType implements Type.
func (t *TaggedUnionType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *TaggedUnionType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *TaggedUnionType) IsScalarType() bool { return false }

func (t *TaggedUnionType) String() string { return fmt.Sprintf("union %s", dict.S(t.Tag)) }

// IsVoidPointerType implements Type.
func (t *UnionType) IsVoidPointerType() bool { panic("TODO") }

// IsArithmeticType implements Type.
func (t *UnionType) IsArithmeticType() bool { return false }

// IsCompatible implements Type.
func (t *UnionType) IsCompatible(u Type) bool {
	if t.Equal(u) {
		return true
	}

	switch x := UnderlyingType(u).(type) {
	case *PointerType:
		return false
	case TypeKind:
		if x.IsScalarType() {
			return false
		}
		panic(fmt.Errorf("%v %T %v", t, x, x))
	default:
		panic(fmt.Errorf("%v %T %v", t, x, x))
	}
}

// Equal implements Type.
func (t *UnionType) Equal(u Type) bool {
	if t == u {
		return true
	}

	switch x := u.(type) {
	case *NamedType:
		return t.Equal(x.Type)
	case *PointerType:
		return false
	case *TaggedUnionType:
		v := x.Type
		if v == x {
			panic(x)
		}

		return t.Equal(v)
	case TypeKind:
		switch x {
		case
			Char,
			Int,
			UInt,
			ULongLong,
			Void:

			return false
		default:
			panic(x)
		}
	case *UnionType:
		if len(t.Fields) != len(x.Fields) || t.Tag != x.Tag {
			return false
		}

		for i, v := range t.Fields {
			if !v.equal(x.Fields[i]) {
				return false
			}
		}
		return true
	default:
		panic(x)
	}
}

// Kind implements Type.
func (t *UnionType) Kind() TypeKind { return Union }

// assign implements Type.
func (t *UnionType) assign(ctx *context, n Node, op Operand) Operand {
	switch x := op.Type.(type) {
	case *NamedType:
		if !t.IsCompatible(x.Type) {
			panic("TODO")
		}

		return Operand{Type: t}
	case *UnionType:
		if !t.IsCompatible(x) {
			panic("TODO")
		}

		return Operand{Type: t}
	default:
		panic(fmt.Errorf("%v: %T %v", ctx.position(n), x, x))
	}
}

// IsPointerType implements Type.
func (t *UnionType) IsPointerType() bool { return false }

// IsIntegerType implements Type.
func (t *UnionType) IsIntegerType() bool { return false }

// IsScalarType implements Type.
func (t *UnionType) IsScalarType() bool { return false }

func (t *UnionType) String() string {
	var buf bytes.Buffer
	buf.WriteString("union{")
	for i, v := range t.Fields {
		if i != 0 {
			buf.WriteString("; ")
		}
		fmt.Fprintf(&buf, "%s %s", dict.S(v.Name), v.Type)
		if v.Bits != 0 {
			fmt.Fprintf(&buf, ".%d", v.Bits)
		}

	}
	buf.WriteByte('}')
	return buf.String()
}

// AdjustedParameterType returns the type of an expression when used as an
// argument of a function, see [0]6.9.1-10.
func AdjustedParameterType(t Type) Type {
	u := t
	for {
		switch x := u.(type) {
		case *ArrayType:
			return &PointerType{t, t.Pos()}
		case *NamedType:
			if isVaList(x) {
				return x
			}

			u = x.Type
		case
			*EnumType,
			*PointerType,
			*StructType,
			*TaggedEnumType,
			*TaggedStructType,
			*TaggedUnionType,
			*UnionType:

			return t
		case TypeKind:
			switch x {
			case
				Char,
				Double,
				DoubleComplex,
				LongDouble,
				Float,
				FloatComplex,
				Int,
				Long,
				LongDoubleComplex,
				LongLong,
				SChar,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort:

				return t
			default:
				panic(x)
			}
		default:
			panic(fmt.Errorf("%T", x))
		}
	}
}

// UnderlyingType returns the concrete type of t, if posible.
func UnderlyingType(t Type) Type { return underlyingType(t, false) }

func underlyingType(t Type, enums bool) Type {
	for {
		switch x := t.(type) {
		case
			*ArrayType,
			*FunctionType,
			*PointerType,
			*StructType,
			*UnionType:

			return x
		case *EnumType:
			if enums {
				return x
			}

			return x.Enums[0].Operand.Type
		case *NamedType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *TaggedEnumType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *TaggedStructType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *TaggedUnionType:
			if x.Type == nil {
				return x
			}

			t = x.Type
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
				Ptr,
				SChar,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort,
				Void:

				return x
			default:
				panic(fmt.Errorf("%v", x))
			}
		default:
			panic(fmt.Errorf("%T", x))
		}
	}
}
