// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"fmt"
	"math"
	"math/bits"

	"modernc.org/ir"
)

var (
	// [0]6.3.1.1-1
	//
	// Every integer type has an integer conversion rank defined as
	// follows:
	intConvRank = [maxTypeKind]int{
		Bool:      1,
		Char:      2,
		SChar:     2,
		UChar:     2,
		Short:     3,
		UShort:    3,
		Int:       4,
		UInt:      4,
		Long:      5,
		ULong:     5,
		LongLong:  6,
		ULongLong: 6,
	}

	isSigned = [maxTypeKind]bool{
		Bool:     true,
		Char:     true,
		SChar:    true,
		Short:    true,
		Int:      true,
		Long:     true,
		LongLong: true,
	}

	isArithmeticType = [maxTypeKind]bool{
		Bool:      true,
		Char:      true,
		Enum:      true,
		Int:       true,
		Long:      true,
		LongLong:  true,
		SChar:     true,
		Short:     true,
		UChar:     true,
		UInt:      true,
		ULong:     true,
		ULongLong: true,
		UShort:    true,

		Float:      true,
		Double:     true,
		LongDouble: true,

		FloatImaginary:      true,
		DoubleImaginary:     true,
		LongDoubleImaginary: true,

		FloatComplex:      true,
		DoubleComplex:     true,
		LongDoubleComplex: true,
	}
)

// Address represents the address of a variable.
type Address struct { //TODO-
	Declarator *Declarator
	Offset     uintptr
}

func (a *Address) String() string {
	return fmt.Sprintf("(%s+%d, %s)", dict.S(a.Declarator.Name()), a.Offset, a.Declarator.Linkage)
}

// UsualArithmeticConversions performs transformations of operands of a binary
// operation. The function panics if either of the operands is not an
// artithmetic type.
//
// [0]6.3.1.8
//
// Many operators that expect operands of arithmetic type cause conversions and
// yield result types in a similar way. The purpose is to determine a common
// real type for the operands and result. For the specified operands, each
// operand is converted, without change of type domain, to a type whose
// corresponding real type is the common real type. Unless explicitly stated
// otherwise, the common real type is also the corresponding real type of the
// result, whose type domain is the type domain of the operands if they are the
// same, and complex otherwise. This pattern is called the usual arithmetic
// conversions:
func UsualArithmeticConversions(m Model, a, b Operand) (Operand, Operand) {
	if !a.isArithmeticType() || !b.isArithmeticType() {
		panic(fmt.Sprint(a, b))
	}

	a = a.normalize(m)
	b = b.normalize(m)
	// First, if the corresponding real type of either operand is long
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is long double.
	if a.Type.Kind() == LongDoubleComplex || b.Type.Kind() == LongDoubleComplex {
		return a.ConvertTo(m, LongDoubleComplex), b.ConvertTo(m, LongDoubleComplex)
	}

	if a.Type.Kind() == LongDouble || b.Type.Kind() == LongDouble {
		return a.ConvertTo(m, LongDouble), b.ConvertTo(m, LongDouble)
	}

	// Otherwise, if the corresponding real type of either operand is
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is double.
	if a.Type.Kind() == DoubleComplex || b.Type.Kind() == DoubleComplex {
		return a.ConvertTo(m, DoubleComplex), b.ConvertTo(m, DoubleComplex)
	}

	if a.Type.Kind() == Double || b.Type.Kind() == Double {
		return a.ConvertTo(m, Double), b.ConvertTo(m, Double)
	}

	// Otherwise, if the corresponding real type of either operand is
	// float, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is float.)
	if a.Type.Kind() == FloatComplex || b.Type.Kind() == FloatComplex {
		return a.ConvertTo(m, FloatComplex), b.ConvertTo(m, FloatComplex)
	}

	if a.Type.Kind() == Float || b.Type.Kind() == Float {
		return a.ConvertTo(m, Float), b.ConvertTo(m, Float)
	}

	// Otherwise, the integer promotions are performed on both operands.
	// Then the following rules are applied to the promoted operands:
	if !a.isIntegerType() || !b.isIntegerType() {
		//dbg("", a)
		//dbg("", b)
		panic("TODO")
	}

	a = a.integerPromotion(m)
	b = b.integerPromotion(m)

	// If both operands have the same type, then no further conversion is
	// needed.
	if a.Type.Equal(b.Type) {
		return a, b
	}

	// Otherwise, if both operands have signed integer types or both have
	// unsigned integer types, the operand with the type of lesser integer
	// conversion rank is converted to the type of the operand with greater
	// rank.
	if a.isSigned() == b.isSigned() {
		t := a.Type
		if intConvRank[b.Type.Kind()] > intConvRank[a.Type.Kind()] {
			t = b.Type
		}
		return a.ConvertTo(m, t), b.ConvertTo(m, t)
	}

	// Otherwise, if the operand that has unsigned integer type has rank
	// greater or equal to the rank of the type of the other operand, then
	// the operand with signed integer type is converted to the type of the
	// operand with unsigned integer type.
	switch {
	case a.isSigned(): // b is unsigned
		if intConvRank[b.Type.Kind()] >= intConvRank[a.Type.Kind()] {
			return a.ConvertTo(m, b.Type), b
		}
	case b.isSigned(): // a is unsigned
		if intConvRank[a.Type.Kind()] >= intConvRank[b.Type.Kind()] {
			return a, b.ConvertTo(m, a.Type)
		}
	default:
		panic(fmt.Errorf("TODO %v %v", a, b))
	}

	var signed Type
	// Otherwise, if the type of the operand with signed integer type can
	// represent all of the values of the type of the operand with unsigned
	// integer type, then the operand with unsigned integer type is
	// converted to the type of the operand with signed integer type.
	switch {
	case a.isSigned(): // b is unsigned
		signed = a.Type
		if m.Sizeof(a.Type) > m.Sizeof(b.Type) {
			return a, b.ConvertTo(m, a.Type)
		}
	case b.isSigned(): // a is unsigned
		signed = b.Type
		if m.Sizeof(b.Type) > m.Sizeof(a.Type) {
			return a.ConvertTo(m, b.Type), b
		}
	default:
		panic(fmt.Errorf("TODO %v %v", a, b))
	}

	// Otherwise, both operands are converted to the unsigned integer type
	// corresponding to the type of the operand with signed integer type.
	switch signed.Kind() {
	case Int:
		if a.IsEnumConst || b.IsEnumConst {
			return a, b
		}

		return a.ConvertTo(m, UInt), b.ConvertTo(m, UInt)
	case Long:
		return a.ConvertTo(m, ULong), b.ConvertTo(m, ULong)
	case LongLong:
		return a.ConvertTo(m, ULongLong), b.ConvertTo(m, ULongLong)
	default:
		panic(signed)
	}
}

// Operand represents the type and optionally the value of an expression.
type Operand struct {
	Type Type
	ir.Value
	FieldProperties *FieldProperties

	IsEnumConst bool // Blocks int -> unsigned int promotions. See [0]6.4.4.3/2
}

// Bits return the width of a bit field operand or zero othewise
func (o *Operand) Bits() int {
	if fp := o.FieldProperties; fp != nil {
		return fp.Bits
	}

	return 0
}

func newIntConst(ctx *context, n Node, v uint64, t ...TypeKind) (r Operand) {
	b := bits.Len64(v)
	for _, t := range t {
		sign := 1
		if t.IsUnsigned() {
			sign = 0
		}
		if ctx.model[t].Size*8 >= b+sign {
			return Operand{Type: t, Value: &ir.Int64Value{Value: int64(v)}}.normalize(ctx.model)
		}
	}

	last := t[len(t)-1]
	if ctx.model[last].Size*8 == b {
		return Operand{Type: last, Value: &ir.Int64Value{Value: int64(v)}}.normalize(ctx.model)
	}

	ctx.err(n, "invalid integer constant")
	return Operand{Type: Int}.normalize(ctx.model)
}

func (o Operand) String() string {
	return fmt.Sprintf("(type %v, value %v, fieldProps %+v)", o.Type, o.Value, o.FieldProperties)
}

func (o Operand) isArithmeticType() bool { return o.Type.IsArithmeticType() }
func (o Operand) isIntegerType() bool    { return o.Type.IsIntegerType() }
func (o Operand) isPointerType() bool    { return o.Type.IsPointerType() }
func (o Operand) isScalarType() bool     { return o.Type.IsScalarType() } // [0]6.2.5-21
func (o Operand) isSigned() bool         { return isSigned[o.Type.Kind()] }

func (o Operand) add(ctx *context, p Operand) (r Operand) {
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if p.IsZero() {
		return o.normalize(ctx.model)
	}

	if o.Value == nil || p.Value == nil {
		o.Value = nil
		return o.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value + p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	case *ir.Float64Value:
		return Operand{Type: o.Type, Value: &ir.Float64Value{Value: x.Value + p.Value.(*ir.Float64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T %v %v", x, o, p))
	}
}

func (o Operand) and(ctx *context, p Operand) (r Operand) {
	if !o.isIntegerType() || !p.isIntegerType() {
		panic(fmt.Errorf("TODO %v & %v", o, p))
	}

	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if o.IsZero() || p.IsZero() {
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: 0}}.normalize(ctx.model)
	}

	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value & p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

// ConvertTo converts o to type t.
func (o Operand) ConvertTo(m Model, t Type) (r Operand) {
	if o.Type.Equal(t) {
		return o.normalize(m)
	}

	switch x := t.(type) {
	case
		*EnumType,
		*PointerType,
		*TaggedEnumType,
		*TaggedUnionType:

		// ok
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
			UShort:

			// ok
		default:
			panic(x)
		}
	case *NamedType:
		return o.ConvertTo(m, x.Type)
	default:
		panic(fmt.Errorf("%T", x))
	}

	if o.Value == nil {
		o.Type = t
		return o.normalize(m)
	}

	if o.isIntegerType() {
		v := *o.Value.(*ir.Int64Value)
		if t.IsIntegerType() {
			return Operand{Type: t, Value: &v}.normalize(m)
		}

		if t.IsPointerType() {
			// [0]6.3.2.3
			if o.IsZero() {
				// 3. An integer constant expression with the
				// value 0, or such an expression cast to type
				// void *, is called a null pointer constant.
				// If a null pointer constant is converted to a
				// pointer type, the resulting pointer, called
				// a null pointer, is guaranteed to compare
				// unequal to a pointer to any object or
				// function.
				return Operand{Type: t, Value: Null}
			}

			return Operand{Type: t, Value: &v}.normalize(m)
		}

		if t.Kind() == Union {
			if v.Value != 0 {
				panic("TODO")
			}

			return Operand{Type: t}
		}

		switch {
		case o.Type.IsUnsigned():
			val := uint64(v.Value)
			switch t.Kind() {
			case Double, LongDouble:
				return Operand{Type: t, Value: &ir.Float64Value{Value: float64(val)}}.normalize(m)
			case DoubleComplex:
				return Operand{Type: t, Value: &ir.Complex128Value{Value: complex(float64(val), 0)}}.normalize(m)
			case Float:
				return Operand{Type: t, Value: &ir.Float32Value{Value: float32(val)}}.normalize(m)
			case FloatComplex:
				return Operand{Type: t, Value: &ir.Complex64Value{Value: complex(float32(val), 0)}}.normalize(m)
			default:
				panic(t)
			}
		default:
			val := v.Value
			switch t.Kind() {
			case Double, LongDouble:
				return Operand{Type: t, Value: &ir.Float64Value{Value: float64(val)}}.normalize(m)
			case DoubleComplex:
				return Operand{Type: t, Value: &ir.Complex128Value{Value: complex(float64(val), 0)}}.normalize(m)
			case Float:
				return Operand{Type: t, Value: &ir.Float32Value{Value: float32(val)}}.normalize(m)
			case FloatComplex:
				return Operand{Type: t, Value: &ir.Complex64Value{Value: complex(float32(val), 0)}}.normalize(m)
			default:
				panic(t)
			}
		}
	}

	if o.Type.Kind() == Double {
		v := o.Value.(*ir.Float64Value).Value
		if t.IsIntegerType() {
			return Operand{Type: t, Value: &ir.Int64Value{Value: ConvertFloat64(v, t, m)}}.normalize(m)
		}

		switch x := t.(type) {
		case TypeKind:
			switch x {
			case Float:
				return Operand{Type: t, Value: &ir.Float32Value{Value: float32(o.Value.(*ir.Float64Value).Value)}}.normalize(m)
			case LongDouble:
				v := *o.Value.(*ir.Float64Value)
				return Operand{Type: t, Value: &v}.normalize(m)
			default:
				panic(x)
			}
		default:
			panic(x)
		}
	}

	if o.Type.Kind() == Float {
		v := o.Value.(*ir.Float32Value).Value
		if t.IsIntegerType() {
			return Operand{Type: t, Value: &ir.Int64Value{Value: ConvertFloat64(float64(v), t, m)}}.normalize(m)
		}

		switch x := t.(type) {
		case TypeKind:
			switch x {
			case
				Double,
				LongDouble:

				return Operand{Type: t, Value: &ir.Float64Value{Value: float64(v)}}.normalize(m)
			default:
				panic(x)
			}
		default:
			panic(x)
		}
	}

	if o.isPointerType() && t.IsPointerType() {
		o.Type = t
		return o.normalize(m)
	}

	if o.isPointerType() && t.IsIntegerType() {
		o.Type = t
		switch x := o.Value.(type) {
		case *ir.AddressValue:
			if x.NameID != 0 {
				o.Value = nil
				break
			}

			o.Value = &ir.Int64Value{Value: int64(x.Offset)}
		case
			*ir.Int64Value,
			*ir.StringValue:

			// nop
		default:
			//fmt.Printf("TODO405 %T %v -> %v\n", x, o, t) //TODO-
			panic(fmt.Errorf("%T %v -> %v", x, o, t))
		}
		return o.normalize(m)
	}

	panic(fmt.Errorf("%T(%v) -> %T(%v)", o.Type, o, t, t))
}

func (o Operand) cpl(ctx *context) Operand {
	if o.isIntegerType() {
		o = o.integerPromotion(ctx.model)
	}

	switch x := o.Value.(type) {
	case nil:
		return o
	case *ir.Int64Value:
		o.Value = &ir.Int64Value{Value: ^o.Value.(*ir.Int64Value).Value}
		return o.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) div(ctx *context, n Node, p Operand) (r Operand) {
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if o.Value == nil || p.Value == nil {
		o.Value = nil
		return o.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		if p.IsZero() {
			ctx.err(n, "division by zero")
			return Operand{Type: o.Type}.normalize(ctx.model)
		}

		switch {
		case o.Type.IsUnsigned():
			return Operand{Type: o.Type, Value: &ir.Int64Value{Value: int64(uint64(x.Value) / uint64(p.Value.(*ir.Int64Value).Value))}}.normalize(ctx.model)
		default:
			return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value / p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
		}
	case *ir.Float32Value:
		return Operand{Type: o.Type, Value: &ir.Float32Value{Value: x.Value / p.Value.(*ir.Float32Value).Value}}.normalize(ctx.model)
	case *ir.Float64Value:
		return Operand{Type: o.Type, Value: &ir.Float64Value{Value: x.Value / p.Value.(*ir.Float64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) eq(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		if x.Value == p.Value.(*ir.Int64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value == p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

func (o Operand) ge(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case o.isSigned():
			if x.Value >= p.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			if uint64(x.Value) >= uint64(p.Value.(*ir.Int64Value).Value) {
				val = 1
			}
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value >= p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

func (o Operand) gt(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case o.isSigned():
			if x.Value > p.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			if uint64(x.Value) > uint64(p.Value.(*ir.Int64Value).Value) {
				val = 1
			}
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value > p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

// integerPromotion computes the integer promotion of o.
//
// [0]6.3.1.1-2
//
// If an int can represent all values of the original type, the value is
// converted to an int; otherwise, it is converted to an unsigned int. These
// are called the integer promotions. All other types are unchanged by the
// integer promotions.
func (o Operand) integerPromotion(m Model) Operand {
	t := o.Type
	for {
		switch x := t.(type) {
		case *EnumType:
			t = x.Enums[0].Operand.Type
		case *NamedType:
			t = x.Type
		case *TaggedEnumType:
			t = x.getType().(*EnumType).Enums[0].Operand.Type
		case TypeKind:
			// github.com/gcc-mirror/gcc/gcc/testsuite/gcc.c-torture/execute/bf-sign-2.c
			//
			// This test checks promotion of bitfields.  Bitfields
			// should be promoted very much like chars and shorts:
			//
			// Bitfields (signed or unsigned) should be promoted to
			// signed int if their value will fit in a signed int,
			// otherwise to an unsigned int if their value will fit
			// in an unsigned int, otherwise we don't promote them
			// (ANSI/ISO does not specify the behavior of bitfields
			// larger than an unsigned int).
			if x.IsIntegerType() && o.Bits() != 0 {
				bits := m[Int].Size * 8
				switch {
				case x.IsUnsigned():
					if o.Bits() < bits {
						return o.ConvertTo(m, Int)
					}
				default:
					if o.Bits() < bits-1 {
						return o.ConvertTo(m, Int)
					}
				}
			}

			switch x {
			case
				Double,
				Float,
				Int,
				Long,
				LongDouble,
				LongLong,
				UInt,
				ULong,
				ULongLong:

				return o
			case
				Char,
				SChar,
				Short,
				UChar,
				UShort:

				return o.ConvertTo(m, Int)
			default:
				panic(x)
			}
		default:
			panic(x)
		}
	}
}

// IsNonZero returns true when the value of o is known to be non-zero.
func (o Operand) IsNonZero() bool {
	switch x := o.Value.(type) {
	case nil:
		return false
	case *ir.Float32Value:
		return x.Value != 0
	case *ir.Float64Value:
		return x.Value != 0
	case *ir.Int64Value:
		return x.Value != 0
	case *ir.StringValue:
		return true
	case *ir.AddressValue:
		return x != Null
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

// IsZero returns true when the value of o is known to be zero.
func (o Operand) IsZero() bool {
	switch x := o.Value.(type) {
	case nil:
		return false
	case *ir.Complex128Value:
		return x.Value == 0
	case *ir.Float32Value:
		return x.Value == 0
	case *ir.Float64Value:
		return x.Value == 0
	case *ir.Int64Value:
		return x.Value == 0
	case
		*ir.StringValue,
		*ir.WideStringValue:

		return false
	case *ir.AddressValue:
		return x == Null
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) isNullPtrConst() bool {
	return o.isIntegerType() && o.IsZero() || o.Value == Null
}

func (o Operand) le(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case o.isSigned():
			if x.Value <= p.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			if uint64(x.Value) <= uint64(p.Value.(*ir.Int64Value).Value) {
				val = 1
			}
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value <= p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

func (o Operand) lsh(ctx *context, p Operand) (r Operand) { // [0]6.5.7
	// 2. Each of the operands shall have integer type.
	if !o.isIntegerType() || !p.isIntegerType() {
		panic("TODO")
	}

	// 3. The integer promotions are performed on each of the operands. The
	// type of the result is that of the promoted left operand. If the
	// value of the right operand is negative or is greater than or equal
	// to the width of the promoted left operand, the behavior is
	// undefined.
	o = o.integerPromotion(ctx.model)
	p = p.integerPromotion(ctx.model)
	if o.IsZero() {
		return o.normalize(ctx.model)
	}

	m := uint64(32)
	if ctx.model.Sizeof(o.Type) > 4 {
		m = 64
	}
	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value << (uint64(p.Value.(*ir.Int64Value).Value) % m)}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) lt(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case o.isSigned():
			if x.Value < p.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			if uint64(x.Value) < uint64(p.Value.(*ir.Int64Value).Value) {
				val = 1
			}
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value < p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

func (o Operand) mod(ctx *context, n Node, p Operand) (r Operand) {
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if p.IsZero() {
		ctx.err(n, "division by zero")
		return p.normalize(ctx.model)
	}

	if o.IsZero() { // 0 % x == 0
		return o.normalize(ctx.model)
	}

	if y, ok := p.Value.(*ir.Int64Value); ok && (y.Value == 1 || y.Value == -1) {
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: 0}}.normalize(ctx.model) //  y % {1,-1} == 0
	}

	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value % p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) mul(ctx *context, p Operand) (r Operand) {
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if o.IsZero() || p.IsZero() {
		switch x := UnderlyingType(o.Type).(type) {
		case TypeKind:
			if x.IsIntegerType() {
				return Operand{Type: o.Type, Value: &ir.Int64Value{Value: 0}}.normalize(ctx.model)
			}
		default:
			panic(fmt.Errorf("TODO %T", x))
		}
	}

	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value * p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	case *ir.Float32Value:
		return Operand{Type: o.Type, Value: &ir.Float32Value{Value: x.Value * p.Value.(*ir.Float32Value).Value}}
	case *ir.Float64Value:
		return Operand{Type: o.Type, Value: &ir.Float64Value{Value: x.Value * p.Value.(*ir.Float64Value).Value}}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) ne(ctx *context, p Operand) (r Operand) {
	r = Operand{Type: Int}
	if o.isArithmeticType() && p.isArithmeticType() {
		o, p = UsualArithmeticConversions(ctx.model, o, p)
	}
	if o.Value == nil || p.Value == nil {
		return r.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		var val int64
		if x.Value != p.Value.(*ir.Int64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float32Value:
		var val int64
		if x.Value != p.Value.(*ir.Float32Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	case *ir.Float64Value:
		var val int64
		if x.Value != p.Value.(*ir.Float64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r.normalize(ctx.model)
}

// ConvertFloat64 converts v to t, which must be an integer type.
func ConvertFloat64(v float64, t Type, m Model) int64 {
	if !t.IsIntegerType() {
		panic(fmt.Errorf("ConvertFloat64: %T", t))
	}

	switch sz := m.Sizeof(t); {
	case t.IsUnsigned():
		switch sz {
		case 1:
			if v > math.Nextafter(math.MaxUint8, 0) {
				return math.MaxUint8
			}

			if v <= 0 {
				return 0
			}
		case 2:
			if v > math.Nextafter(math.MaxUint16, 0) {
				return math.MaxUint16
			}

			if v <= 0 {
				return 0
			}
		case 4:
			if v > math.Nextafter(math.MaxUint32, 0) {
				return math.MaxUint32
			}

			if v <= 0 {
				return 0
			}
		case 8:
			if v > math.Nextafter(math.MaxUint64, 0) {
				return -1 // int64(math,MaxUint64)
			}

			if v <= 0 {
				return 0
			}
		default:
			panic(sz)
		}
	default:
		switch sz {
		case 1:
			if v > math.Nextafter(math.MaxInt8, 0) {
				return math.MaxInt8
			}

			if v < math.Nextafter(math.MinInt8, 0) {
				return math.MinInt8
			}
		case 2:
			if v > math.Nextafter(math.MaxInt16, 0) {
				return math.MaxInt16
			}

			if v < math.Nextafter(math.MinInt16, 0) {
				return math.MinInt16
			}
		case 4:
			if v > math.Nextafter(math.MaxInt32, 0) {
				return math.MaxInt32
			}

			if v < math.Nextafter(math.MinInt32, 0) {
				return math.MinInt32
			}
		case 8:
			if v > math.Nextafter(math.MaxInt64, 0) {
				return math.MaxInt64
			}

			if v < math.Nextafter(math.MinInt64, 0) {
				return math.MinInt64
			}
		default:
			panic(sz)
		}
	}
	return int64(v)
}

// ConvertInt64 converts n to t, which must be an integer or enum type, doing
// masking and/or sign extending as appropriate.
func ConvertInt64(n int64, t Type, m Model) int64 {
	switch x := UnderlyingType(t).(type) {
	case *EnumType:
		t = x.Enums[0].Operand.Type
	}
	signed := !t.IsUnsigned()
	switch sz := m[UnderlyingType(t).Kind()].Size; sz {
	case 1:
		switch {
		case signed:
			switch {
			case int8(n) < 0:
				return n | ^math.MaxUint8
			default:
				return n & math.MaxUint8
			}
		default:
			return n & math.MaxUint8
		}
	case 2:
		switch {
		case signed:
			switch {
			case int16(n) < 0:
				return n | ^math.MaxUint16
			default:
				return n & math.MaxUint16
			}
		default:
			return n & math.MaxUint16
		}
	case 4:
		switch {
		case signed:
			switch {
			case int32(n) < 0:
				return n | ^math.MaxUint32
			default:
				return n & math.MaxUint32
			}
		default:
			return n & math.MaxUint32
		}
	case 8:
		return n
	default:
		panic(fmt.Errorf("TODO %v %T %v", sz, t, t))
	}
}

func (o Operand) normalize(m Model) (r Operand) {
	switch x := o.Value.(type) {
	case *ir.Int64Value:
		if v := ConvertInt64(x.Value, o.Type, m); v != x.Value {
			n := *x
			n.Value = v
			x = &n
			o.Value = x
		}
	case nil:
		// nop
	case
		*ir.AddressValue,
		*ir.Complex128Value,
		*ir.Complex64Value,
		*ir.Float32Value,
		*ir.Float64Value,
		*ir.StringValue,
		*ir.WideStringValue:

		// nop
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return o
}

func (o Operand) or(ctx *context, p Operand) (r Operand) {
	if !o.isIntegerType() || !p.isIntegerType() {
		panic("TODO")
	}
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	r.Type = o.Type
	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value | p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) rsh(ctx *context, p Operand) (r Operand) { // [0]6.5.7
	// 2. Each of the operands shall have integer type.
	if !o.isIntegerType() || !p.isIntegerType() {
		panic("TODO")
	}

	// 3. The integer promotions are performed on each of the operands. The
	// type of the result is that of the promoted left operand. If the
	// value of the right operand is negative or is greater than or equal
	// to the width of the promoted left operand, the behavior is
	// undefined.
	o = o.integerPromotion(ctx.model)
	p = p.integerPromotion(ctx.model)
	r.Type = o.Type
	m := uint64(32)
	if ctx.model.Sizeof(o.Type) > 4 {
		m = 64
	}
	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		switch {
		case o.isSigned():
			return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value >> (uint64(p.Value.(*ir.Int64Value).Value) % m)}}.normalize(ctx.model)
		default:
			return Operand{Type: o.Type, Value: &ir.Int64Value{Value: int64(uint64(x.Value) >> (uint64(p.Value.(*ir.Int64Value).Value) % m))}}.normalize(ctx.model)
		}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) sub(ctx *context, p Operand) (r Operand) {
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if p.IsZero() {
		return o.normalize(ctx.model)
	}

	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}.normalize(ctx.model)
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value - p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	case *ir.Float32Value:
		return Operand{Type: o.Type, Value: &ir.Float32Value{Value: x.Value - p.Value.(*ir.Float32Value).Value}}.normalize(ctx.model)
	case *ir.Float64Value:
		return Operand{Type: o.Type, Value: &ir.Float64Value{Value: x.Value - p.Value.(*ir.Float64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) unaryMinus(ctx *context) Operand {
	if o.isIntegerType() {
		o = o.integerPromotion(ctx.model)
	}

	switch x := o.Value.(type) {
	case nil:
		return o
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: -x.Value}}.normalize(ctx.model)
	case *ir.Float32Value:
		return Operand{Type: o.Type, Value: &ir.Float32Value{Value: -x.Value}}
	case *ir.Float64Value:
		return Operand{Type: o.Type, Value: &ir.Float64Value{Value: -x.Value}}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (o Operand) xor(ctx *context, p Operand) (r Operand) {
	if !o.isIntegerType() || !p.isIntegerType() {
		panic("TODO")
	}
	o, p = UsualArithmeticConversions(ctx.model, o, p)
	if o.Value == nil || p.Value == nil {
		return Operand{Type: o.Type}
	}

	switch x := o.Value.(type) {
	case *ir.Int64Value:
		return Operand{Type: o.Type, Value: &ir.Int64Value{Value: x.Value ^ p.Value.(*ir.Int64Value).Value}}.normalize(ctx.model)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}
