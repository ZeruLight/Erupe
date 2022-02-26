// Copyright 2016 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc"

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"modernc.org/mathutil"
	"modernc.org/xc"
)

type (
	// StringLitID is the type of an Expression.Value representing the numeric
	// ID of a string literal.
	StringLitID int

	// LongStringLitID is the type of an Expression.Value representing the
	// numeric ID of a long string literal.
	LongStringLitID int

	// StringLitID is the type of an Expression.Value representing the numeric
	// ID of a label name used in &&label.
	ComputedGotoID int
)

var (
	maxConvF32I32 = math.Nextafter32(math.MaxInt32, 0) // https://github.com/golang/go/issues/19405
	maxConvF32U32 = math.Nextafter32(math.MaxUint32, 0)
)

// ModelItem is a single item of a model.
//
// Note about StructAlign: To provide GCC ABI compatibility set, for example,
// Align of Double to 8 and StructAlign of Double to 4.
type ModelItem struct {
	Size        int         // Size of the entity in bytes.
	Align       int         // Alignment of the entity when it's not a struct field.
	StructAlign int         // Alignment of the entity when it's a struct field.
	More        interface{} // Optional user data.
}

// Model describes size and align requirements of predeclared types.
type Model struct {
	Items map[Kind]ModelItem

	BoolType              Type
	CharType              Type
	DoubleComplexType     Type
	DoubleType            Type
	FloatComplexType      Type
	FloatType             Type
	IntType               Type
	LongDoubleComplexType Type
	LongDoubleType        Type
	LongLongType          Type
	LongType              Type
	ShortType             Type
	UCharType             Type
	UIntType              Type
	ULongLongType         Type
	ULongType             Type
	UShortType            Type
	UintPtrType           Type
	VoidType              Type
	longStrType           Type
	ptrDiffType           Type
	sizeType              Type
	strType               Type

	initialized bool
	tweaks      *tweaks
	intConvRank [kindMax]int
	Signed      [kindMax]bool // Signed[Kind] reports whether Kind is a signed integer type.
	promoteTo   [kindMax]Kind
}

func (m *Model) initialize(lx *lexer) {
	m.BoolType = m.makeType(lx, 0, tsBool)
	m.CharType = m.makeType(lx, 0, tsChar)
	m.DoubleComplexType = m.makeType(lx, 0, tsComplex, tsDouble)
	m.DoubleType = m.makeType(lx, 0, tsDouble)
	m.FloatComplexType = m.makeType(lx, 0, tsComplex, tsFloat)
	m.FloatType = m.makeType(lx, 0, tsFloat)
	m.IntType = m.makeType(lx, 0, tsInt)
	m.LongDoubleComplexType = m.makeType(lx, 0, tsComplex, tsDouble, tsLong)
	m.LongDoubleType = m.makeType(lx, 0, tsLong, tsDouble)
	m.LongLongType = m.makeType(lx, 0, tsLong, tsLong)
	m.LongType = m.makeType(lx, 0, tsLong)
	m.ShortType = m.makeType(lx, 0, tsShort)
	m.UCharType = m.makeType(lx, 0, tsUnsigned, tsChar)
	m.UIntType = m.makeType(lx, 0, tsUnsigned)
	m.ULongLongType = m.makeType(lx, 0, tsUnsigned, tsLong, tsLong)
	m.ULongType = m.makeType(lx, 0, tsUnsigned, tsLong)
	m.UShortType = m.makeType(lx, 0, tsUnsigned, tsShort)
	m.UintPtrType = m.makeType(lx, 0, tsUintptr) // Pseudo type.
	m.VoidType = m.makeType(lx, 0, tsVoid)
	m.strType = m.makeType(lx, 0, tsChar).Pointer()

	// [0], 6.3.1.1.
	m.intConvRank = [kindMax]int{
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
		UintPtr:   7,
	}
	m.Signed = [kindMax]bool{
		Char:     true,
		SChar:    true,
		Short:    true,
		Int:      true,
		Long:     true,
		LongLong: true,
	}
	m.promoteTo = [kindMax]Kind{}
	for i := range m.promoteTo {
		m.promoteTo[i] = Kind(i)
	}
	switch {
	case m.tweaks.enableWideEnumValues:
		m.intConvRank[Enum] = m.intConvRank[LongLong]
	default:
		m.intConvRank[Enum] = m.intConvRank[Int]
	}
	for k := Kind(0); k < kindMax; k++ {
		r := m.intConvRank[k]
		if r == 0 || r > m.intConvRank[Int] {
			continue
		}

		// k is an integer type whose conversion rank is less than or
		// equal to the rank of int and unsigned int.
		switch {
		case m.Items[k].Size < m.Items[Int].Size || m.Signed[k]:
			// If an int can represent all values of the original
			// type, the value is converted to an int;
			m.promoteTo[k] = Int
		default:
			// otherwise, it is converted to an unsigned int.
			m.promoteTo[k] = UInt
		}
	}

	m.initialized = true
}

func (m *Model) typ(k Kind) Type {
	switch k {
	case Undefined:
		return undefined
	case Bool:
		return m.BoolType
	case Char:
		return m.CharType
	case Double:
		return m.DoubleType
	case Float:
		return m.FloatType
	case Int:
		return m.IntType
	case LongDouble:
		return m.LongDoubleType
	case LongLong:
		return m.LongLongType
	case Long:
		return m.LongType
	case Short:
		return m.ShortType
	case UChar:
		return m.UCharType
	case UInt:
		return m.UIntType
	case ULongLong:
		return m.ULongLongType
	case ULong:
		return m.ULongType
	case UShort:
		return m.UShortType
	case UintPtr:
		return m.UintPtrType
	case FloatComplex:
		return m.FloatComplexType
	case DoubleComplex:
		return m.DoubleComplexType
	case LongDoubleComplex:
		return m.LongDoubleComplexType
	case Enum:
		switch {
		case m.tweaks.enableWideEnumValues:
			return m.LongLongType
		default:
			return m.IntType
		}
	default:
		panic(k)
	}
}

func (m *Model) enumValueToInt(v interface{}) (interface{}, bool) {
	intSize := m.Items[Int].Size
	if m.tweaks.enableWideEnumValues {
		intSize = m.Items[LongLong].Size
	}
	switch x := v.(type) {
	case byte, int8, int16, uint16, int32:
		return m.MustConvert(x, m.IntType), true
	case uint32:
		switch intSize {
		case 4:
			return m.MustConvert(x, m.IntType), x <= math.MaxUint32
		case 8:
			return m.MustConvert(x, m.IntType), true
		default:
			panic(intSize)
		}
	case int64:
		switch intSize {
		case 4:
			return m.MustConvert(x, m.IntType), x <= math.MaxUint32
		case 8:
			return m.MustConvert(x, m.IntType), true
		default:
			panic(intSize)
		}
	case uint64:
		switch intSize {
		case 4:
			return m.MustConvert(x, m.IntType), x <= math.MaxUint32
		case 8:
			return m.MustConvert(x, m.IntType), x <= math.MaxUint64
		default:
			panic(intSize)
		}
	default:
		panic(fmt.Errorf("%T", x))
	}
}

// sanityCheck reports model errors, if any.
func (m *Model) sanityCheck() error {
	if len(m.Items) == 0 {
		return fmt.Errorf("model has no items")
	}

	tab := map[Kind]struct {
		minSize, maxSize   int
		minAlign, maxAlign int
	}{
		Ptr:               {4, 8, 4, 8},
		UintPtr:           {4, 8, 4, 8},
		Void:              {0, 0, 1, 1},
		Char:              {1, 1, 1, 1},
		SChar:             {1, 1, 1, 1},
		UChar:             {1, 1, 1, 1},
		Short:             {2, 2, 2, 2},
		UShort:            {2, 2, 2, 2},
		Int:               {4, 4, 4, 4},
		UInt:              {4, 4, 4, 4},
		Long:              {4, 8, 4, 8},
		ULong:             {4, 8, 4, 8},
		LongLong:          {8, 8, 8, 8},
		ULongLong:         {8, 8, 8, 8},
		Float:             {4, 4, 4, 4},
		Double:            {8, 8, 8, 8},
		LongDouble:        {8, 16, 8, 16},
		Bool:              {1, 1, 1, 1},
		FloatComplex:      {8, 8, 8, 8},
		DoubleComplex:     {16, 16, 8, 16},
		LongDoubleComplex: {16, 32, 8, 16},
	}
	a := []int{}
	required := map[Kind]bool{}
	seen := map[Kind]bool{}
	for k := range tab {
		required[k] = true
		a = append(a, int(k))
	}
	sort.Ints(a)
	for k, v := range m.Items {
		if seen[k] {
			return fmt.Errorf("model has duplicate item: %s", k)
		}

		seen[k] = true
		if !required[k] {
			return fmt.Errorf("model has invalid type: %s: %#v", k, v)
		}

		for typ, t := range tab {
			if typ == k {
				if v.Size < t.minSize {
					return fmt.Errorf("size %d too small: %s", v.Size, k)
				}

				if v.Size > t.maxSize {
					return fmt.Errorf("size %d too big: %s", v.Size, k)
				}

				if v.Size != 0 && mathutil.PopCount(v.Size) != 1 {
					return fmt.Errorf("size %d is not a power of two: %s", v.Size, k)
				}

				if v.Align < t.minAlign {
					return fmt.Errorf("align %d too small: %s", v.Align, k)
				}

				if v.Align > t.maxAlign {
					return fmt.Errorf("align %d too big: %s", v.Align, k)
				}

				if v.Align < v.Size && v.Align < t.minAlign {
					return fmt.Errorf("align is smaller than size: %s", k)
				}

				if v.StructAlign < 1 {
					return fmt.Errorf("struct align %d too small: %s", v.StructAlign, k)
				}

				if v.StructAlign > t.maxAlign {
					return fmt.Errorf("struct align %d too big: %s", v.Align, k)
				}

				if mathutil.PopCount(v.Align) != 1 {
					return fmt.Errorf("align %d is not a power of two: %s", v.Align, k)
				}

				break
			}
		}
	}
	w := m.Items[Ptr].Size
	if m.Items[Short].Size < w &&
		m.Items[Int].Size < w &&
		m.Items[Long].Size < w &&
		m.Items[LongLong].Size < w {
		return fmt.Errorf("model has no integer type suitable for pointer difference and sizeof")
	}

	for _, typ := range a {
		if !seen[Kind(typ)] {
			return fmt.Errorf("model has no item for type %s", Kind(typ))
		}
	}

	if g, e := w, m.Items[UintPtr].Size; g != e {
		return fmt.Errorf("model uintptr has different sizes than ptr has")
	}
	return nil
}

// MustConvert returns v converted to the type of typ. If the conversion is
// impossible, the method panics.
//
// Conversion an integer type to any pointer type yields an uintptr.
func (m *Model) MustConvert(v interface{}, typ Type) interface{} {
	if typ.Kind() == Enum {
		typ = m.IntType
	}
	mi, ok := m.Items[typ.Kind()]
	if !ok && typ.Kind() != Function {
		panic(fmt.Errorf("internal error: no model item for %s, %s", typ, typ.Kind()))
	}

	w := mi.Size
	switch typ.Kind() {
	case Short:
		switch x := v.(type) {
		case int32:
			switch w {
			case 2:
				return int16(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 2:
				return int16(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case UShort:
		switch x := v.(type) {
		case uint16:
			switch w {
			case 2:
				return x
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 2:
				return uint16(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 2:
				return uint16(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Int:
		switch x := v.(type) {
		case int8:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case byte:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case int16:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case uint16:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 4:
				return x
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		case float32:
			switch w {
			case 4:
				switch {
				case x > maxConvF32I32:
					return int32(math.MaxInt32)
				default:
					return int32(x)
				}
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 4:
				return int32(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case UInt:
		switch x := v.(type) {
		case uint8:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case int16:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case uint16:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return x
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case uintptr:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		case float32:
			switch w {
			case 4:
				switch {
				case x > maxConvF32U32:
					return uint32(math.MaxUint32)
				default:
					return uint32(x)
				}
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 4:
				return uint32(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Long:
		switch x := v.(type) {
		case int16:
			switch w {
			case 4:
				return int32(x)
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 4:
				return x
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return int32(x)
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 4:
				return int32(x)
			case 8:
				return x
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 4:
				return int32(x)
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case uintptr:
			switch w {
			case 4:
				return int32(x)
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case LongLong:
		switch x := v.(type) {
		case int32:
			switch w {
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 8:
				return x
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 8:
				return int64(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case ULong:
		switch x := v.(type) {
		case uint8:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case int:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return x
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return x
			default:
				panic(w)
			}
		case uintptr:
			switch w {
			case 4:
				return uint32(x)
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case ULongLong:
		switch x := v.(type) {
		case int32:
			switch w {
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 8:
				return x
			default:
				panic(w)
			}
		case uintptr:
			switch w {
			case 8:
				return uint64(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Float:
		switch x := v.(type) {
		case int32:
			switch w {
			case 4:
				return float32(x)
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return float32(x)
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 4:
				return float32(x)
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 4:
				return float32(x)
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case float32:
			switch w {
			case 4:
				return x
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 4:
				return float32(x)
			case 8:
				return x
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Double:
		switch x := v.(type) {
		case int32:
			switch w {
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case float32:
			switch w {
			case 8:
				return float64(x)
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 8:
				return x
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Ptr, Function:
		switch x := v.(type) {
		case int32:
			return uintptr(x)
		case uint32:
			return uintptr(x)
		case int64:
			return uintptr(x)
		case uint64:
			return uintptr(x)
		case uintptr:
			return x
		case StringLitID:
			return nil
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Void:
		return nil
	case Char, SChar:
		switch x := v.(type) {
		case int32:
			switch w {
			case 1:
				return int8(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 1:
				return int8(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case UChar:
		switch x := v.(type) {
		case uint8:
			switch w {
			case 1:
				return x
			default:
				panic(w)
			}
		case int32:
			switch w {
			case 1:
				return byte(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 1:
				return byte(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case UintPtr:
		switch x := v.(type) {
		case int32:
			switch w {
			case 4, 8:
				return uintptr(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 4:
				return uintptr(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 8:
				return uintptr(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case LongDouble:
		switch x := v.(type) {
		case int32:
			switch w {
			case 8, 16:
				return float64(x)
			default:
				panic(w)
			}
		case uint32:
			switch w {
			case 8, 16:
				return float64(x)
			default:
				panic(w)
			}
		case int64:
			switch w {
			case 8, 16:
				return float64(x)
			default:
				panic(w)
			}
		case uint64:
			switch w {
			case 8, 16:
				return float64(x)
			default:
				panic(w)
			}
		case float32:
			switch w {
			case 8, 16:
				return float64(x)
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 8, 16:
				return x
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case Bool:
		switch x := v.(type) {
		case int32:
			if x != 0 {
				return int32(1)
			}

			return int32(0)
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case FloatComplex:
		switch x := v.(type) {
		case float32:
			switch w {
			case 8:
				return complex(x, 0)
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 8:
				return complex(float32(x), 0)
			default:
				panic(w)
			}
		case complex64:
			switch w {
			case 8:
				return x
			default:
				panic(w)
			}
		case complex128:
			switch w {
			case 8:
				return complex64(x)
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case DoubleComplex:
		switch x := v.(type) {
		case int32:
			switch w {
			case 16:
				return complex(float64(x), 0)
			default:
				panic(w)
			}
		case float64:
			switch w {
			case 16:
				return complex(x, 0)
			default:
				panic(w)
			}
		case complex128:
			switch w {
			case 16:
				return x
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	case LongDoubleComplex:
		switch x := v.(type) {
		case float64:
			switch w {
			case 16:
				return complex(x, 0)
			default:
				panic(w)
			}
		case complex128:
			switch w {
			case 16:
				return x
			default:
				panic(w)
			}
		default:
			panic(fmt.Errorf("internal error %T", x))
		}
	default:
		panic(fmt.Errorf("internal error %s, %s", typ, typ.Kind()))
	}
}

func (m *Model) value2(v interface{}, typ Type) (interface{}, Type) {
	return m.MustConvert(v, typ), typ
}

func (m *Model) charConst(lx *lexer, t xc.Token) (interface{}, Type) {
	switch t.Rune {
	case CHARCONST:
		s := string(t.S())
		typ := m.IntType
		var r rune
		s = s[1 : len(s)-1] // Remove outer 's.
		if len(s) == 1 {
			return rune(s[0]), m.IntType
		}

		runes := []rune(s)
		switch runes[0] {
		case '\\':
			r, _ = decodeEscapeSequence(runes)
			if r < 0 {
				r = -r
			}
		default:
			r = runes[0]
		}
		return r, typ
	case LONGCHARCONST:
		s := t.S()
		typ := m.LongType
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
			lx.report.Err(t.Pos(), "invalid character literal %s", t.S())
			return 0, typ
		}

		return runes[0], typ
	default:
		panic("internal error")
	}
}

func (m *Model) getSizeType(lx *lexer) Type {
	if t := m.sizeType; t != nil {
		return t
	}

	b := lx.scope.Lookup(NSIdentifiers, xc.Dict.SID("size_t"))
	if b.Node == nil {
		w := m.Items[Ptr].Size
		switch {
		case m.Items[Short].Size >= w:
			return m.ShortType
		case m.Items[Int].Size >= w:
			return m.IntType
		case m.Items[Long].Size >= w:
			return m.LongType
		default:
			return m.LongLongType
		}
	}

	d := b.Node.(*DirectDeclarator)
	if !d.TopDeclarator().RawSpecifier().IsTypedef() {
		lx.report.Err(d.Pos(), "size_t is not a typedef name")
		m.sizeType = undefined
		return undefined
	}

	m.sizeType = b.Node.(*DirectDeclarator).top().declarator.Type
	return m.sizeType
}

func (m *Model) getPtrDiffType(lx *lexer) Type {
	if t := m.ptrDiffType; t != nil {
		return t
	}

	b := lx.scope.Lookup(NSIdentifiers, xc.Dict.SID("ptrdiff_t"))
	if b.Node == nil {
		w := m.Items[Ptr].Size
		switch {
		case m.Items[Short].Size >= w:
			return m.ShortType
		case m.Items[Int].Size >= w:
			return m.IntType
		case m.Items[Long].Size >= w:
			return m.LongType
		default:
			return m.LongLongType
		}
	}

	d := b.Node.(*DirectDeclarator)
	if !d.TopDeclarator().RawSpecifier().IsTypedef() {
		lx.report.Err(d.Pos(), "ptrdiff_t is not a typedef name")
		m.ptrDiffType = undefined
		return undefined
	}

	m.ptrDiffType = b.Node.(*DirectDeclarator).top().declarator.Type
	return m.ptrDiffType
}

func (m *Model) getLongStrType(lx *lexer, tok xc.Token) Type {
	if t := m.longStrType; t != nil {
		return t
	}

	b := lx.scope.Lookup(NSIdentifiers, xc.Dict.SID("wchar_t"))
	if b.Node == nil {
		m.longStrType = m.IntType.Pointer()
		return m.longStrType
	}

	d := b.Node.(*DirectDeclarator)
	if !d.TopDeclarator().RawSpecifier().IsTypedef() {
		lx.report.Err(d.Pos(), "wchar_t is not a typedef name")
		m.longStrType = undefined
		return m.longStrType
	}

	m.longStrType = b.Node.(*DirectDeclarator).top().declarator.Type.Pointer()
	return m.longStrType
}

func (m *Model) strConst(lx *lexer, t xc.Token) (interface{}, Type) {
	s := t.S()
	typ := m.strType
	var buf bytes.Buffer
	switch t.Rune {
	case LONGSTRINGLITERAL:
		typ = m.getLongStrType(lx, t)
		s = s[1:] // Remove leading 'L'.
		fallthrough
	case STRINGLITERAL:
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
	default:
		panic("internal error")
	}
	s = buf.Bytes()
	switch t.Rune {
	case LONGSTRINGLITERAL:
		return LongStringLitID(xc.Dict.ID(s)), typ
	case STRINGLITERAL:
		return StringLitID(xc.Dict.ID(s)), typ
	default:
		panic("internal error")
	}
}

func (m *Model) floatConst(lx *lexer, t xc.Token) (interface{}, Type) {
	const (
		f = 1 << iota
		j
		l
	)
	k := 0
	s := t.S()
	i := len(s) - 1
more:
	switch c := s[i]; c {
	case 'i', 'j':
		k |= j
		i--
		goto more
	case 'l', 'L':
		k |= l
		i--
		goto more
	case 'f', 'F':
		k |= f
		i--
		goto more
	}
	if k&j != 0 && !lx.tweaks.enableImaginarySuffix {
		lx.report.Err(t.Pos(), "imaginary suffixes not enabled")
		k &^= j
	}
	ss := string(s[:i+1])
	var v float64
	var err error
	switch {
	case strings.Contains(ss, "p"):
		var bf *big.Float
		bf, _, err = big.ParseFloat(ss, 0, 53, big.ToNearestEven)
		switch {
		case err != nil:
			lx.report.Err(t.Pos(), "invalid floating point constant %s", ss)
			v = 0
		default:
			v, _ = bf.Float64()
		}
	default:
		v, err = strconv.ParseFloat(ss, 64)
	}
	if err != nil {
		lx.report.Err(t.Pos(), "invalid floating point constant %s", ss)
		v = 0
	}
	switch k {
	case 0:
		return m.value2(v, m.DoubleType)
	case l:
		return m.value2(v, m.LongDoubleType)
	case j:
		return m.value2(complex(0, v), m.DoubleComplexType)
	case j | l:
		return m.value2(complex(0, v), m.LongDoubleComplexType)
	case f:
		return m.value2(v, m.FloatType)
	case f | j:
		return m.value2(complex(0, v), m.FloatComplexType)
	default:
		lx.report.Err(t.Pos(), "invalid literal %s", t.S())
		return 0.0, m.DoubleType
	}
}

func (m *Model) intConst(lx *lexer, t xc.Token) (interface{}, Type) {
	const (
		l = 1 << iota
		ll
		u
	)
	k := 0
	s := t.S()
	i := len(s) - 1
more:
	switch c := s[i]; c {
	case 'u', 'U':
		k |= u
		i--
		goto more
	case 'l', 'L':
		if i > 0 && (s[i-1] == 'l' || s[i-1] == 'L') {
			k |= ll
			i -= 2
			goto more
		}

		k |= l
		i--
		goto more
	}
	n, err := strconv.ParseUint(string(s[:i+1]), 0, 64)
	if err != nil {
		lx.report.Err(t.Pos(), "invalid integer constant: %s", s)
	}

	switch k {
	case 0:
		switch b := mathutil.BitLenUint64(n); {
		case b < 32:
			return m.value2(n, m.IntType)
		case b < 33:
			return m.value2(n, m.UIntType)
		case b < 64:
			if m.Items[Long].Size == 8 {
				return m.value2(n, m.LongType)
			}

			return m.value2(n, m.LongLongType)
		default:
			if m.Items[ULong].Size == 8 {
				return m.value2(n, m.ULongType)
			}

			return m.value2(n, m.ULongLongType)
		}
	case l:
		return m.value2(n, m.LongType)
	case ll:
		return m.value2(n, m.LongLongType)
	case u:
		return m.value2(n, m.UIntType)
	case u | l:
		return m.value2(n, m.ULongType)
	case u | ll:
		return m.value2(n, m.ULongLongType)
	default:
		panic("internal error")
	}
}

func (m *Model) cBool(v bool) interface{} {
	if v {
		return m.MustConvert(int32(1), m.IntType)

	}
	return m.MustConvert(int32(0), m.IntType)
}

func (m *Model) binOp(lx *lexer, a, b operand) (interface{}, interface{}, Type) {
	av, at := a.eval(lx)
	bv, bt := b.eval(lx)
	t := at
	if IsArithmeticType(at) && IsArithmeticType(bt) {
		t = m.BinOpType(at, bt)
	}
	if av == nil || bv == nil || t.Kind() == Undefined {
		return nil, nil, t
	}

	return m.MustConvert(av, t), m.MustConvert(bv, t), t
}

// BinOpType returns the evaluation type of a binop b, ie. the type operands
// are converted to before performing the operation. Operands must be
// arithmetic types.
//
// See [0], 6.3.1.8 - Usual arithmetic conversions.
func (m *Model) BinOpType(a, b Type) Type {
	ak := a.Kind()
	bk := b.Kind()

	if ak == LongDoubleComplex || bk == LongDoubleComplex {
		return m.LongDoubleComplexType
	}

	if ak == DoubleComplex || bk == DoubleComplex {
		return m.DoubleComplexType
	}

	if ak == FloatComplex || bk == FloatComplex {
		return m.FloatComplexType
	}

	// First, if the corresponding real type of either operand is long
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is long double.
	if ak == LongDouble || bk == LongDouble {
		return m.LongDoubleType
	}

	// Otherwise, if the corresponding real type of either operand is
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is double.
	if ak == Double || bk == Double {
		return m.DoubleType
	}

	// Otherwise, if the corresponding real type of either operand is float, the other
	// operand is converted, without change of type domain, to a type whose
	// corresponding real type is float.
	if ak == Float || bk == Float {
		return m.FloatType
	}

	// Otherwise, the integer promotions are performed on both operands.
	ak = m.promoteTo[ak]
	bk = m.promoteTo[bk]

	// Then the following rules are applied to the promoted operands:
	ar := m.intConvRank[ak]
	br := m.intConvRank[bk]

	// If both operands have the same type, then no further conversion is
	// needed.
	if ak == bk {
		return m.typ(ak)
	}

	// Otherwise, if both operands have signed integer types or both have
	// unsigned integer types, the operand with the type of lesser integer
	// conversion rank is converted to the type of the operand with greater
	// rank.
	if m.Signed[ak] == m.Signed[bk] {
		switch {
		case ar < br:
			return m.typ(bk)
		default:
			return m.typ(ak)
		}
	}

	// Make a the unsigned type and b the signed type.
	if m.Signed[ak] {
		a, b = b, a
		ak, bk = bk, ak
		ar, br = br, ar
	}

	// Otherwise, if the operand that has unsigned integer type has rank
	// greater or equal to the rank of the type of the other operand, then
	// the operand with signed integer type is converted to the type of the
	// operand with unsigned integer type.
	if ar >= br {
		return m.typ(ak)
	}

	// Otherwise, if the type of the operand with signed integer type can
	// represent all of the values of the type of the operand with unsigned
	// integer type, then the operand with unsigned integer type is
	// converted to the type of the operand with signed integer type.
	as := m.Items[ak].Size
	bs := m.Items[bk].Size
	if bs > as {
		return m.typ(bk)
	}

	// Otherwise, both operands are converted to the unsigned integer type
	// corresponding to the type of the operand with signed integer type.
	return m.typ(unsigned(bk))
}

func (m *Model) promote(t Type) Type {
	if !IsIntType(t) {
		return t
	}

	return m.BinOpType(t, t)
}

func (m *Model) makeType(lx *lexer, attr int, ts ...int) Type {
	d := m.makeDeclarator(attr, ts...)
	return d.setFull(lx)
}

func (m *Model) makeDeclarator(attr int, ts ...int) *Declarator {
	s := &spec{attr, tsEncode(ts...)}
	d := &Declarator{specifier: s}
	dd := &DirectDeclarator{declarator: d, specifier: s}
	d.DirectDeclarator = dd
	return d
}

func (m *Model) checkArithmeticType(lx *lexer, a ...operand) (r bool) {
	r = true
	for _, v := range a {
		_, t := v.eval(lx)
		if !IsArithmeticType(t) {
			lx.report.Err(v.Pos(), "not an arithmetic type (have '%s')", t)
			r = false
		}
	}
	return r
}

func (m *Model) checkIntegerOrBoolType(lx *lexer, a ...operand) (r bool) {
	r = true
	for _, v := range a {
		_, t := v.eval(lx)
		if !IsIntType(t) && !(t.Kind() == Bool) {
			lx.report.Err(v.Pos(), "not an integer or bool type (have '%s')", t)
			r = false
		}
	}
	return r
}
