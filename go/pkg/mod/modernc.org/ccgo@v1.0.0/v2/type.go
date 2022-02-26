// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v2"

import (
	"bytes"
	"fmt"

	"modernc.org/cc/v2"
	"modernc.org/ir"
)

type tCacheKey struct {
	cc.Type
	bool
}

func isVaList(t cc.Type) bool {
	x, ok := t.(*cc.NamedType)
	return ok && (x.Name == idVaList || x.Name == idBuiltinVaList)
}

func (g *gen) typ(t cc.Type) string { return g.ptyp(t, true, 0) }

func (g *ngen) typ(t cc.Type) string { return g.ptyp(t, true, 0) }

func (g *gen) ptyp(t cc.Type, ptr2uintptr bool, lvl int) (r string) {
	k := tCacheKey{t, ptr2uintptr}
	if s, ok := g.tCache[k]; ok {
		return s
	}

	defer func() { g.tCache[k] = r }()

	if ptr2uintptr {
		if t.Kind() == cc.Ptr && !isVaList(t) {
			if _, ok := t.(*cc.NamedType); !ok {
				g.enqueue(t)
				return "uintptr"
			}
		}

		if x, ok := t.(*cc.ArrayType); ok && x.Size.Value == nil {
			return "uintptr"
		}
	}

	switch x := t.(type) {
	case *cc.ArrayType:
		if x.Size.Value == nil {
			return fmt.Sprintf("*%s", g.ptyp(x.Item, ptr2uintptr, lvl))
		}

		return fmt.Sprintf("[%d]%s", x.Size.Value.(*ir.Int64Value).Value, g.ptyp(x.Item, ptr2uintptr, lvl))
	case *cc.FunctionType:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "func(%sTLS", crt)
		switch {
		case len(x.Params) == 1 && x.Params[0].Kind() == cc.Void:
			// nop
		default:
			for _, v := range x.Params {
				switch underlyingType(v, true).(type) {
				case *cc.ArrayType:
					fmt.Fprintf(&buf, ", uintptr")
				default:
					fmt.Fprintf(&buf, ", %s", g.typ(v))
				}
			}
		}
		if x.Variadic {
			fmt.Fprintf(&buf, ", ...interface{}")
		}
		buf.WriteString(")")
		if x.Result != nil && x.Result.Kind() != cc.Void {
			buf.WriteString(" " + g.typ(x.Result))
		}
		return buf.String()
	case *cc.NamedType:
		if isVaList(x) {
			if ptr2uintptr {
				return "*[]interface{}"
			}

			return fmt.Sprintf("%s", dict.S(x.Name))
		}

		g.enqueue(t)
		t := x.Type
		for {
			if x, ok := t.(*cc.NamedType); ok {
				t = x.Type
				continue
			}

			break
		}
		return g.ptyp(t, ptr2uintptr, lvl)
	case *cc.PointerType:
		if x.Item.Kind() == cc.Void {
			return "uintptr"
		}

		switch {
		case x.Kind() == cc.Function:
			todo("")
		default:
			return fmt.Sprintf("*%s", g.ptyp(x.Item, ptr2uintptr, lvl+1))
		}
	case *cc.StructType:
		var buf bytes.Buffer
		buf.WriteString("struct{")
		layout := g.model.Layout(x)
		for i, v := range x.Fields {
			if v.Bits < 0 {
				continue
			}

			if v.Bits != 0 {
				if layout[i].Bitoff == 0 {
					fmt.Fprintf(&buf, "F%d %s;", layout[i].Offset, g.typ(layout[i].PackedType))
					if lvl == 0 {
						fmt.Fprintf(&buf, "\n")
					}
				}
				continue
			}

			switch {
			case v.Name == 0:
				fmt.Fprintf(&buf, "_ ")
			default:
				fmt.Fprintf(&buf, "%s ", mangleIdent(v.Name, true))
			}
			fmt.Fprintf(&buf, "%s;", g.ptyp(v.Type, ptr2uintptr, lvl+1))
			if lvl == 0 && ptr2uintptr && v.Type.Kind() == cc.Ptr {
				fmt.Fprintf(&buf, "// %s\n", g.ptyp(v.Type, false, lvl+1))
			}
		}
		buf.WriteByte('}')
		return buf.String()
	case *cc.EnumType:
		if x.Tag == 0 {
			return g.typ(x.Enums[0].Operand.Type)
		}

		g.enqueue(x)
		return fmt.Sprintf("E%s", dict.S(x.Tag))
	case *cc.TaggedEnumType:
		g.enqueue(x)
		return fmt.Sprintf("E%s", dict.S(x.Tag))
	case *cc.TaggedStructType:
		g.enqueue(x)
		return fmt.Sprintf("S%s", dict.S(x.Tag))
	case *cc.TaggedUnionType:
		g.enqueue(x)
		return fmt.Sprintf("U%s", dict.S(x.Tag))
	case cc.TypeKind:
		switch x {
		case
			cc.Char,
			cc.Int,
			cc.Long,
			cc.LongLong,
			cc.SChar,
			cc.Short:

			return fmt.Sprintf("int%d", g.model[x].Size*8)
		case
			cc.UChar,
			cc.UShort,
			cc.UInt,
			cc.ULong,
			cc.ULongLong:

			return fmt.Sprintf("uint%d", g.model[x].Size*8)
		case cc.Float:
			return fmt.Sprintf("float32")
		case
			cc.Double,
			cc.LongDouble:

			return fmt.Sprintf("float64")
		default:
			todo("", x)
		}
	case *cc.UnionType:
		al := int64(g.model.Alignof(x))
		sz := g.model.Sizeof(x)
		switch {
		case al == sz:
			return fmt.Sprintf("struct{X int%d}", 8*sz)
		default:
			return fmt.Sprintf("struct{X int%d; _ [%d]byte}", 8*al, sz-al) //TODO use precomputed padding from model layout?
		}
	default:
		todo("%v %T %v\n%v", t, x, ptr2uintptr, pretty(x))
	}
	panic("unreachable")
}

func (g *ngen) ptyp(t cc.Type, ptr2uintptr bool, lvl int) (r string) {
	k := tCacheKey{t, ptr2uintptr}
	if s, ok := g.tCache[k]; ok {
		return s
	}

	defer func() { g.tCache[k] = r }()

	if ptr2uintptr {
		if t.Kind() == cc.Ptr {
			return "uintptr"
		}

		if x, ok := t.(*cc.ArrayType); ok && x.Size.Value == nil {
			return "uintptr"
		}
	}

	switch x := t.(type) {
	case *cc.ArrayType:
		if x.Size.Value == nil {
			return fmt.Sprintf("*%s", g.ptyp(x.Item, ptr2uintptr, lvl))
		}

		return fmt.Sprintf("[%d]%s", x.Size.Value.(*ir.Int64Value).Value, g.ptyp(x.Item, ptr2uintptr, lvl))
	case *cc.FunctionType:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "func(%sTLS", g.crtPrefix)
		switch {
		case len(x.Params) == 1 && x.Params[0].Kind() == cc.Void:
			// nop
		default:
			for _, v := range x.Params {
				switch underlyingType(v, true).(type) {
				case *cc.ArrayType:
					fmt.Fprintf(&buf, ", uintptr")
				default:
					fmt.Fprintf(&buf, ", %s", g.typ(v))
				}
			}
		}
		if x.Variadic {
			fmt.Fprintf(&buf, ", ...interface{}")
		}
		buf.WriteString(")")
		if x.Result != nil && x.Result.Kind() != cc.Void {
			buf.WriteString(" " + g.typ(x.Result))
		}
		return buf.String()
	case *cc.NamedType:
		g.enqueue(x)
		t := x.Type
		for {
			if x, ok := t.(*cc.NamedType); ok {
				g.enqueue(x)
				t = x.Type
				continue
			}

			break
		}
		return g.ptyp(t, ptr2uintptr, lvl)
	case *cc.PointerType:
		if x.Item.Kind() == cc.Void {
			return "uintptr"
		}

		switch {
		case x.Kind() == cc.Function:
			todo("")
		default:
			return fmt.Sprintf("*%s", g.ptyp(x.Item, ptr2uintptr, lvl+1))
		}
	case *cc.StructType:
		var buf bytes.Buffer
		buf.WriteString("struct{")
		layout := g.model.Layout(x)
		for i, v := range x.Fields {
			if v.Bits < 0 {
				continue
			}

			if v.Bits != 0 {
				if layout[i].Bitoff == 0 {
					fmt.Fprintf(&buf, "F%d %s;", layout[i].Offset, g.typ(layout[i].PackedType))
					if lvl == 0 {
						fmt.Fprintf(&buf, "\n")
					}
				}
				continue
			}

			switch {
			case v.Name == 0:
				fmt.Fprintf(&buf, "_ ")
			default:
				fmt.Fprintf(&buf, "F%s ", dict.S(v.Name))
			}
			fmt.Fprintf(&buf, "%s;", g.ptyp(v.Type, ptr2uintptr, lvl+1))
			if lvl == 0 && ptr2uintptr && v.Type.Kind() == cc.Ptr {
				fmt.Fprintf(&buf, "// %s\n", g.typeComment(v.Type))
			}
		}
		buf.WriteByte('}')
		return buf.String()
	case *cc.EnumType:
		if x.Tag == 0 {
			return g.typ(x.Enums[0].Operand.Type)
		}

		g.enqueue(x)
		return fmt.Sprintf("E%s", dict.S(x.Tag))
	case *cc.TaggedEnumType:
		g.enqueue(x)
		return fmt.Sprintf("E%s", dict.S(x.Tag))
	case *cc.TaggedStructType:
		g.enqueue(x)
		return fmt.Sprintf("S%s", dict.S(x.Tag))
	case *cc.TaggedUnionType:
		g.enqueue(x)
		return fmt.Sprintf("U%s", dict.S(x.Tag))
	case cc.TypeKind:
		switch x {
		case
			cc.Char,
			cc.Int,
			cc.Long,
			cc.LongLong,
			cc.SChar,
			cc.Short:

			return fmt.Sprintf("int%d", g.model[x].Size*8)
		case
			cc.Bool,
			cc.UChar,
			cc.UShort,
			cc.UInt,
			cc.ULong,
			cc.ULongLong:

			return fmt.Sprintf("uint%d", g.model[x].Size*8)
		case cc.Float:
			return fmt.Sprintf("float32")
		case
			cc.Double,
			cc.LongDouble:

			return fmt.Sprintf("float64")
		case
			cc.DoubleComplex,
			cc.LongDoubleComplex:

			return fmt.Sprintf("complex128")
		case cc.FloatComplex:
			return fmt.Sprintf("complex64")
		default:
			todo("", x)
		}
	case *cc.UnionType:
		var buf bytes.Buffer
		buf.WriteString("struct{")
		layout := g.model.Layout(x)
		for i, v := range x.Fields {
			if v.Bits < 0 {
				continue
			}

			if v.Bits != 0 {
				if layout[i].Bitoff == 0 {
					fmt.Fprintf(&buf, "F%d [0]%s;", i, g.typ(layout[i].PackedType))
					if lvl == 0 {
						fmt.Fprintf(&buf, "\n")
					}
				}
				continue
			}

			switch {
			case v.Name == 0:
				fmt.Fprintf(&buf, "_ ")
			default:
				fmt.Fprintf(&buf, "F%s ", dict.S(v.Name))
			}
			fmt.Fprintf(&buf, "[0]%s;", g.ptyp(v.Type, ptr2uintptr, lvl+1))
			if lvl == 0 && ptr2uintptr && v.Type.Kind() == cc.Ptr {
				fmt.Fprintf(&buf, "// %s\n", g.typeComment(v.Type))
			}
		}
		al := int64(g.model.Alignof(x))
		sz := g.model.Sizeof(x)
		switch {
		case al == sz:
			fmt.Fprintf(&buf, "F int%d}", 8*sz)
		default:
			fmt.Fprintf(&buf, "F int%d; _ [%d]byte}", 8*al, sz-al) //TODO use precomputed padding from model layout?
		}
		return buf.String()
	default:
		todo("%v %T %v\n%v", t, x, ptr2uintptr, pretty(x))
	}
	panic("unreachable")
}

func prefer(d *cc.Declarator) bool {
	if d.DeclarationSpecifier.IsExtern() {
		return false
	}

	if d.Initializer != nil || d.FunctionDefinition != nil {
		return true
	}

	t := d.Type
	for {
		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			return x.Size.Type != nil
		case *cc.FunctionType:
			return false
		case
			*cc.EnumType,
			*cc.StructType:

			return true
		case *cc.PointerType:
			t = x.Item
		case *cc.TaggedStructType:
			return x.Type != nil
		case cc.TypeKind:
			if x.IsScalarType() || x == cc.Void {
				return true
			}

			todo("", x)
		default:
			todo("", x)
		}
	}
}

func underlyingType(t cc.Type, enums bool) cc.Type {
	for {
		switch x := t.(type) {
		case
			*cc.ArrayType,
			*cc.FunctionType,
			*cc.PointerType,
			*cc.StructType,
			*cc.UnionType:

			return x
		case *cc.EnumType:
			if enums {
				return x
			}

			return x.Enums[0].Operand.Type
		case *cc.NamedType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *cc.TaggedEnumType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *cc.TaggedStructType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case *cc.TaggedUnionType:
			if x.Type == nil {
				return x
			}

			t = x.Type
		case cc.TypeKind:
			switch x {
			case
				cc.Char,
				cc.Double,
				cc.DoubleComplex,
				cc.Float,
				cc.FloatComplex,
				cc.Int,
				cc.Long,
				cc.LongDouble,
				cc.LongDoubleComplex,
				cc.LongLong,
				cc.SChar,
				cc.Short,
				cc.UChar,
				cc.UInt,
				cc.ULong,
				cc.ULongLong,
				cc.UShort,
				cc.Void:

				return x
			default:
				todo("%v", x)
			}
		default:
			todo("%T", x)
		}
	}
}
