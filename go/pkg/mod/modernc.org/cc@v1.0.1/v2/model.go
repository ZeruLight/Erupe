// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

import (
	"fmt"
	"runtime"

	"modernc.org/ir"
	"modernc.org/mathutil"
)

// Model describes properties of scalar Types.
type Model map[TypeKind]ModelItem

// ModelItem describers properties of a particular Type.
type ModelItem struct {
	Size        int
	Align       int
	StructAlign int
}

// NewModel returns the model appropriate for the current OS and architecture
// or according to the environment variables GOOS and GOARCH, if set.
func NewModel() (m Model, err error) {
	switch arch := env("GOARCH", runtime.GOARCH); arch {
	case "arm":
		return Model{
			Bool:      {1, 1, 1},
			Char:      {1, 1, 1},
			Int:       {4, 4, 4},
			Long:      {4, 4, 4},
			LongLong:  {8, 4, 4},
			SChar:     {1, 1, 1},
			Short:     {2, 2, 2},
			UChar:     {1, 1, 1},
			UInt:      {4, 4, 4},
			ULong:     {4, 4, 4},
			ULongLong: {8, 4, 4},
			UShort:    {2, 2, 2},

			Float:      {4, 4, 4},
			Double:     {8, 4, 4},
			LongDouble: {8, 4, 4},

			FloatImaginary:      {4, 4, 4},
			DoubleImaginary:     {8, 4, 4},
			LongDoubleImaginary: {8, 4, 4},

			FloatComplex:      {8, 4, 4},
			DoubleComplex:     {16, 4, 4},
			LongDoubleComplex: {16, 4, 4},

			Void: {1, 1, 1},
			Ptr:  {4, 4, 4},
		}, nil
	case "386":
		return Model{
			Bool:      {1, 1, 1},
			Char:      {1, 1, 1},
			Int:       {4, 4, 4},
			Long:      {4, 4, 4},
			LongLong:  {8, 8, 4},
			SChar:     {1, 1, 1},
			Short:     {2, 2, 2},
			UChar:     {1, 1, 1},
			UInt:      {4, 4, 4},
			ULong:     {4, 4, 4},
			ULongLong: {8, 8, 4},
			UShort:    {2, 2, 2},

			Float:      {4, 4, 4},
			Double:     {8, 8, 4},
			LongDouble: {8, 8, 4},

			FloatImaginary:      {4, 4, 4},
			DoubleImaginary:     {8, 8, 4},
			LongDoubleImaginary: {8, 8, 4},

			FloatComplex:      {8, 8, 4},
			DoubleComplex:     {16, 8, 4},
			LongDoubleComplex: {16, 8, 4},

			Void: {1, 1, 1},
			Ptr:  {4, 4, 4},
		}, nil
	case "amd64":
		var longLength = 8
		if env("GOOS", runtime.GOOS) == "windows" {
			longLength = 4
		}

		model := Model{
			Bool:      {1, 1, 1},
			Char:      {1, 1, 1},
			Int:       {4, 4, 4},
			Long:      {longLength, longLength, longLength},
			LongLong:  {8, 8, 8},
			SChar:     {1, 1, 1},
			Short:     {2, 2, 2},
			UChar:     {1, 1, 1},
			UInt:      {4, 4, 4},
			ULong:     {longLength, longLength, longLength},
			ULongLong: {8, 8, 8},
			UShort:    {2, 2, 2},

			Float:      {4, 4, 4},
			Double:     {8, 8, 8},
			LongDouble: {8, 8, 8},

			FloatImaginary:      {4, 4, 4},
			DoubleImaginary:     {8, 8, 8},
			LongDoubleImaginary: {8, 8, 8},

			FloatComplex:      {8, 8, 4},
			DoubleComplex:     {16, 8, 4},
			LongDoubleComplex: {16, 8, 4},

			Void: {1, 1, 1},
			Ptr:  {8, 8, 8},
		}

		return model, nil
	default:
		return nil, fmt.Errorf("unknown/unsupported architecture %s", arch)
	}
}

// Equal returns whether m equals n.
func (m Model) Equal(n Model) bool {
	if len(m) != len(n) {
		return false
	}

	for k, v := range m {
		if v != n[k] {
			return false
		}
	}
	return true
}

// Sizeof returns the size in bytes of a variable of type t.
func (m Model) Sizeof(t Type) int64 {
	switch x := UnderlyingType(t).(type) {
	case *ArrayType:
		if x.Size.Value != nil { // T[42]
			return m.Sizeof(x.Item) * x.Size.Value.(*ir.Int64Value).Value
		}

		if x.Length != nil {
			panic(fmt.Errorf("Sizeof(%v): variable length array", t))
		}

		panic(fmt.Errorf("Sizeof(%v): incomplete array", t))
	case *EnumType:
		return m.Sizeof(x.Enums[0].Operand.Type)
	case *NamedType:
		return m.Sizeof(x.Type)
	case
		*FunctionType,
		*PointerType:

		return int64(m[Ptr].Size)
	case *StructType:
		layout := m.Layout(x)
		if len(layout) == 0 {
			return 0
		}

		lf := layout[len(layout)-1]
		return lf.Offset + lf.Size + int64(lf.Padding)
	case *TaggedStructType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}

		return m.Sizeof(u)
	case TypeKind:
		return int64(m[x].Size)
	case *UnionType:
		var sz int64
		for _, v := range m.Layout(x) {
			if v.Size > sz {
				sz = v.Size
			}
		}
		return roundup(sz, int64(m.Alignof(x)))
	case nil:
		panic("internal error")
	default:
		panic(x)
	}
}

// FieldProperties describe a struct/union field.
type FieldProperties struct {
	Bitoff     int // Zero based bit number of a bitfield
	Bits       int // Width of a bit field or zero otherwise.
	Declarator *Declarator
	Offset     int64 // Byte offset relative to start of the struct/union.
	PackedType Type  // Bits != 0: Storage type holding the bit field.
	Padding    int   // Adjustment to enforce proper alignment.
	Size       int64 // Field size for copying.
	Type       Type

	Anonymous       bool
	IsFlexibleArray bool
}

// Mask returns the bit mask of bit field described by f.
func (f *FieldProperties) Mask() uint64 {
	if f.Bits == 0 {
		return 1<<64 - 1
	}

	return (1<<uint(f.Bits) - 1) << uint(f.Bitoff)
}

// Layout computes the memory layout of t.
func (m Model) Layout(t Type) (r []FieldProperties) {
	//TODO memoize
	switch x := UnderlyingType(t).(type) {
	case *StructType:
		if len(x.Fields) == 0 {
			return nil
		}

		if x.layout != nil {
			return x.layout
		}

		defer func() { x.layout = r }()

		r = make([]FieldProperties, len(x.Fields))
		var off int64
		bitoff := 0
		for i, v := range x.Fields {
			switch {
			case v.Bits != 0:
				switch {
				case bitoff == 0 && v.Bits > 0:
					r[i] = FieldProperties{Offset: off, Bitoff: bitoff, Bits: v.Bits, Declarator: v.Declarator, Type: v.Type}
					bitoff = v.Bits
				default:
					if v.Bits < 0 {
						if n := m.packBits(bitoff, i-1, off, r); bitoff != 0 {
							off = n
						}
						r[i] = FieldProperties{Offset: off, Bits: -1, Declarator: v.Declarator, Type: v.Type}
						bitoff = 0
						break
					}

					n := bitoff + v.Bits
					if n > 32 {
						off = m.packBits(bitoff, i-1, off, r)
						r[i] = FieldProperties{Offset: off, Bits: v.Bits, Declarator: v.Declarator, Type: v.Type}
						bitoff = v.Bits
						break
					}

					r[i] = FieldProperties{Offset: off, Bitoff: bitoff, Bits: v.Bits, Declarator: v.Declarator, Type: v.Type}
					bitoff = n
				}
			default:
				if bitoff != 0 {
					off = m.packBits(bitoff, i-1, off, r)
					bitoff = 0
				}
				var sz int64
				if !v.IsFlexibleArray {
					sz = m.Sizeof(v.Type)
				}
				a := m.StructAlignof(v.Type)
				z := off
				if a != 0 {
					off = roundup(off, int64(a))
				}
				if off != z {
					r[i-1].Padding = int(off - z)
				}
				r[i] = FieldProperties{Offset: off, Size: sz, Declarator: v.Declarator, Type: v.Type, Anonymous: v.Anonymous, IsFlexibleArray: v.IsFlexibleArray}
				off += sz
			}
		}
		i := len(r) - 1
		if bitoff != 0 {
			off = m.packBits(bitoff, i, off, r)
		}
		for i, v := range r {
			if v.Bits > 0 {
				x.Fields[i].PackedType = v.PackedType
			}
		}
		align := 0
		for i, v := range x.Fields {
			if r[i].Bits < 0 {
				continue
			}

			t := v.Type
			if v.PackedType != nil {
				t = v.PackedType
			}
			align = mathutil.Max(align, m.StructAlignof(t))
		}
		z := off
		off = roundup(off, int64(align))
		if off != z {
			r[len(r)-1].Padding = int(off - z)
		}
		return r
	case *UnionType:
		if len(x.Fields) == 0 {
			return nil
		}

		if x.layout != nil {
			return x.layout
		}

		defer func() { x.layout = r }()

		r = make([]FieldProperties, len(x.Fields))
		for i, v := range x.Fields {
			switch {
			case v.Bits < 0:
				m.packBits(v.Bits, i, 0, r)
			case v.Bits > 0:
				r[i] = FieldProperties{Bits: v.Bits, Declarator: v.Declarator, Type: v.Type}
				m.packBits(v.Bits, i, 0, r)
				x.Fields[i].PackedType = r[i].PackedType
			default:
				sz := m.Sizeof(v.Type)
				r[i] = FieldProperties{Size: sz, Bits: v.Bits, Declarator: v.Declarator, Type: v.Type}
			}
		}
		for i, v := range r {
			if v.Bits != 0 {
				x.Fields[i].PackedType = v.PackedType
			}
		}
		return r
	case nil:
		panic("internal error")
	default:
		panic(x)
	}
}

func (m *Model) packBits(bitoff, i int, off int64, r []FieldProperties) int64 {
	var t Type
	switch {
	case bitoff <= 8:
		t = UChar
	case bitoff <= 16:
		t = UShort
	case bitoff <= 32:
		t = UInt
	case bitoff <= 64:
		t = ULongLong
	default:
		panic("internal error")
	}
	sz := m.Sizeof(t)
	a := m.StructAlignof(t)
	z := off
	if a != 0 {
		off = roundup(off, int64(a))
	}
	var first int
	for first = i; first >= 0 && r[first].Bits > 0 && r[first].PackedType == nil; first-- {
	}
	first++
	if off != z {
		r[first-1].Padding = int(off - z)
	}
	for j := first; j <= i; j++ {
		r[j].Offset = off
		r[j].Size = sz
		r[j].PackedType = t
	}
	return off + sz
}

// Alignof computes the memory alignment requirements of t. One is returned
// for a struct/union type with no fields.
func (m Model) Alignof(t Type) int {
	switch x := t.(type) {
	case *ArrayType:
		return m.Alignof(x.Item)
	case *EnumType:
		return m.Alignof(x.Enums[0].Operand.Type)
	case *NamedType:
		return m.Alignof(x.Type)
	case *PointerType:
		return m[Ptr].Align
	case *StructType:
		m.Layout(x)
		r := 1
		for _, v := range x.Fields {
			t := v.Type
			if v.Bits < 0 {
				continue
			}

			if v.Bits > 0 {
				t = v.PackedType
			}
			if a := m.StructAlignof(t); a > r {
				r = a
			}
		}
		return r
	case *TaggedEnumType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.Alignof(u)
	case *TaggedStructType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.Alignof(u)
	case *TaggedUnionType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.Alignof(u)
	case TypeKind:
		return m[x].Align
	case *UnionType:
		m.Layout(x)
		r := 1
		for _, v := range x.Fields {
			t := v.Type
			if v.Bits < 0 {
				continue
			}

			if v.Bits > 0 {
				t = v.PackedType
			}
			if a := m.StructAlignof(t); a > r {
				r = a
			}
		}
		return r
	case nil:
		panic("internal error")
	default:
		panic(x)
	}
}

// StructAlignof computes the memory alignment requirements of t when its
// instance is a struct field. One is returned for a struct/union type with no
// fields.
func (m Model) StructAlignof(t Type) int {
	switch x := t.(type) {
	case *ArrayType:
		return m.StructAlignof(x.Item)
	case *EnumType:
		return m.StructAlignof(x.Enums[0].Operand.Type)
	case *NamedType:
		return m.StructAlignof(x.Type)
	case *PointerType:
		return m[Ptr].StructAlign
	case *StructType:
		m.Layout(x)
		r := 1
		for _, v := range x.Fields {
			t := v.Type
			if v.Bits < 0 {
				continue
			}

			if v.Bits > 0 {
				t = v.PackedType
			}
			if a := m.StructAlignof(t); a > r {
				r = a
			}
		}
		return r
	case *TaggedEnumType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.StructAlignof(u)
	case *TaggedStructType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.StructAlignof(u)
	case *TaggedUnionType:
		u := x.getType()
		if u == x {
			panic("TODO")
		}
		return m.StructAlignof(u)
	case TypeKind:
		return m[x].StructAlign
	case *UnionType:
		m.Layout(x)
		r := 1
		for _, v := range x.Fields {
			t := v.Type
			if v.Bits < 0 {
				continue
			}

			if v.Bits > 0 {
				t = v.PackedType
			}
			if a := m.StructAlignof(t); a > r {
				r = a
			}
		}
		return r
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func roundup(n, to int64) int64 {
	if r := n % to; r != 0 {
		return n + to - r
	}

	return n
}

func (m Model) defaultArgumentPromotion(op Operand) (r Operand) {
	u := op.Type
	for {
		switch x := u.(type) {
		case *EnumType:
			u = x.Enums[0].Operand.Type
		case *NamedType:
			u = x.Type
		case
			*PointerType,
			*StructType,
			*TaggedStructType,
			*TaggedUnionType,
			*UnionType:

			op.Type = x
			return op
		case *TaggedEnumType:
			u = x.getType()
		case TypeKind:
			op.Type = x
			switch x {
			case Float:
				return op.ConvertTo(m, Double)
			case
				Double,
				LongDouble:

				return op
			case
				Char,
				Int,
				Long,
				LongLong,
				SChar,
				Short,
				UChar,
				UInt,
				ULong,
				ULongLong,
				UShort:

				return op.integerPromotion(m)
			default:
				panic(x)
			}
		default:
			panic(x)
		}
	}
}
