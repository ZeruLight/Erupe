// Copyright 2017 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v2"

import (
	"unsafe"

	"modernc.org/cc/v2"
	"modernc.org/ir"
	"modernc.org/xc"
)

func (g *gen) isZeroInitializer(n *cc.Initializer) bool {
	if n == nil {
		return true
	}

	if n.Case == cc.InitializerExpr { // Expr
		return n.Expr.IsZero()
	}

	// '{' InitializerList CommaOpt '}'
	for l := n.InitializerList; l != nil; l = l.InitializerList {
		if !g.isZeroInitializer(l.Initializer) {
			return false
		}
	}
	return true
}

func (g *ngen) isZeroInitializer(n *cc.Initializer) bool {
	if n == nil {
		return true
	}

	if n.Case == cc.InitializerExpr { // Expr
		return n.Expr.IsZero() && g.voidCanIgnore(n.Expr)
	}

	// '{' InitializerList CommaOpt '}'
	for l := n.InitializerList; l != nil; l = l.InitializerList {
		if !g.isZeroInitializer(l.Initializer) {
			return false
		}
	}
	return true
}

func (g *gen) isConstInitializer(t cc.Type, n *cc.Initializer) bool {
	switch n.Case {
	case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if !g.isConstInitializer(x.Item, l.Initializer) {
					return false
				}
			}
			return true
		case *cc.StructType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position0(n))
					}

					fld = l[0]
				}

				if !g.isConstInitializer(layout[fld].Type, l.Initializer) {
					return false
				}

				fld++
			}
			return true
		case *cc.UnionType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position0(n))
					}

					fld = l[0]
				}

				if !g.isConstInitializer(layout[fld].Type, l.Initializer) {
					return false
				}

				fld++
			}
			return true
		default:
			todo("%v: %T %v", g.position0(n), x, t)
		}
	case cc.InitializerExpr: // Expr
		op := n.Expr.Operand
		if op.Value == nil || !g.voidCanIgnore(n.Expr) {
			return false
		}

		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			switch y := op.Value.(type) {
			case *ir.StringValue:
				if x.Size.Value != nil {
					switch x.Item.Kind() {
					case cc.Char, cc.SChar, cc.UChar:
						return true
					default:
						return false
					}
				}

				return false
			default:
				todo("%v: %T %v %v", g.position0(n), y, t, op)
			}
		case *cc.EnumType:
			return true
		case *cc.PointerType:
			_, ok := op.Value.(*ir.Int64Value)
			return ok
		case cc.TypeKind:
			if x.IsArithmeticType() {
				return true
			}
		default:
			todo("%v: %T %v %v", g.position0(n), x, t, op)
		}
	default:
		todo("%v: %v", g.position0(n), n.Case)
	}
	panic("unreachable")
}

func (g *ngen) isConstInitializer(t cc.Type, n *cc.Initializer) bool {
	switch n.Case {
	case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if !g.isConstInitializer(x.Item, l.Initializer) {
					return false
				}
			}
			return true
		case *cc.StructType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}

				if !g.isConstInitializer(layout[fld].Type, l.Initializer) {
					return false
				}

				fld++
			}
			return true
		case *cc.UnionType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}

				if !g.isConstInitializer(layout[fld].Type, l.Initializer) {
					return false
				}

				fld++
			}
			return true
		default:
			todo("%v: %T %v", g.position(n), x, t)
		}
	case cc.InitializerExpr: // Expr
		op := n.Expr.Operand
		if op.Value == nil || !g.voidCanIgnore(n.Expr) {
			return false
		}

		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			switch y := op.Value.(type) {
			case *ir.StringValue:
				if x.Size.Value != nil {
					switch x.Item.Kind() {
					case cc.Char, cc.SChar, cc.UChar:
						return true
					default:
						return false
					}
				}

				return false
			default:
				todo("%v: %T %v %v", g.position(n), y, t, op)
			}
		case *cc.EnumType:
			return true
		case *cc.PointerType:
			_, ok := op.Value.(*ir.Int64Value)
			return ok
		case cc.TypeKind:
			if x.IsArithmeticType() {
				return true
			}
		case *cc.StructType:
			switch x := op.Value.(type) {
			case *ir.Int64Value:
				return true
			default:
				todo("%v: %T %v %v", g.position(n), x, t, op)
			}
		case *cc.UnionType:
			switch x := op.Value.(type) {
			case *ir.Int64Value:
				return true
			default:
				todo("%v: %T %v %v", g.position(n), x, t, op)
			}
		default:
			todo("%v: %T %v %v", g.position(n), x, t, op)
		}
	default:
		todo("%v: %v", g.position(n), n.Case)
	}
	panic("unreachable")
}

func (g *gen) allocBSS(t cc.Type) int64 {
	g.bss = roundup(g.bss, int64(g.model.Alignof(t)))
	r := g.bss
	g.bss += g.model.Sizeof(t)
	return r
}

func (g *gen) allocDS(t cc.Type, n *cc.Initializer) int64 {
	up := roundup(int64(len(g.ds)), int64(g.model.Alignof(t)))
	if n := up - int64(len(g.ds)); n != 0 {
		g.ds = append(g.ds, make([]byte, n)...)
	}
	r := len(g.ds)
	b := make([]byte, g.model.Sizeof(t))
	if !g.isConstInitializer(t, n) {
		todo("%v: %v", g.position0(n), t)
	}
	g.renderInitializer(b, t, n)
	g.ds = append(g.ds, b...)
	return int64(r)
}

func (g *ngen) allocDS(t cc.Type, n *cc.Initializer) []byte {
	b := make([]byte, g.model.Sizeof(t))
	if !g.isConstInitializer(t, n) {
		todo("%v: %v", g.position(n), t)
	}
	g.renderInitializer(b, t, n)
	return b
}

func (g *gen) initializer(d *cc.Declarator) { // non TLD
	n := d.Initializer
	if n.Case == cc.InitializerExpr { // Expr
		switch {
		case g.escaped(d):
			g.w("\n*(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
		default:
			g.w("\n%s", g.mangleDeclarator(d))
		}
		g.w(" = ")
		g.literal(d.Type, n)
		return
	}

	if g.isConstInitializer(d.Type, n) {
		b := make([]byte, g.model.Sizeof(d.Type))
		g.renderInitializer(b, d.Type, n)
		switch {
		case g.escaped(d):
			g.w("\n%sCopy(%s, ts+%d, %d)", crt, g.mangleDeclarator(d), g.allocString(dict.ID(b)), len(b))
		default:
			g.w("\n%s = *(*%s)(unsafe.Pointer(ts+%d))", g.mangleDeclarator(d), g.typ(d.Type), g.allocString(dict.ID(b)))
		}
		return
	}

	switch {
	case g.initializerHasBitFields(d.Type, d.Initializer):
		switch n.Case {
		case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
			switch x := underlyingType(d.Type, true).(type) {
			case *cc.StructType:
				layout := g.model.Layout(x)
				var fld int64
				fields := x.Fields
				for l := n.InitializerList; l != nil; l = l.InitializerList {
					if fld < int64(len(layout)) {
						for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
							fld++
						}
					}
					if d := l.Designation; d != nil {
						l := d.List
						if len(l) != 1 {
							todo("", g.position0(n))
						}

						fld = l[0]
					}

					switch n := l.Initializer; n.Case {
					case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
						todo("", g.position0(n))
					case cc.InitializerExpr: // Expr
						fp := x.Field(fields[fld].Name) //TODO mixed index fieds vs layout
						e := &cc.Expr{
							Case: cc.ExprAssign,
							Expr: &cc.Expr{
								Case: cc.ExprSelect,
								Expr: &cc.Expr{
									Case:       cc.ExprIdent,
									Declarator: d,
									Scope:      d.Scope,
									Token:      xc.Token{Val: d.Name()},
								},
								Operand: cc.Operand{Type: fp.Type, FieldProperties: fp},
								Token2:  xc.Token{Val: fields[fld].Name}, //TODO mixed index fieds vs layout
							},
							Expr2:   n.Expr,
							Operand: cc.Operand{Type: fp.Declarator.Type},
						}
						g.w("\n")
						g.void(e)
					}

					fld++
				}
			default:
				todo("%v: %T", g.position0(n), x)
			}
		case cc.InitializerExpr: // Expr
			todo("", g.position0(n))
		}
	default:
		switch {
		case g.escaped(d):
			g.w("\n*(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
		default:
			g.w("\n%s", g.mangleDeclarator(d))
		}
		g.w(" = ")
		g.literal(d.Type, n)
	}
}

func (g *ngen) initializer(d *cc.Declarator) {
	n := d.Initializer
	if n.Case == cc.InitializerExpr { // Expr
		switch {
		case g.escaped(d):
			g.w("\n*(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
		default:
			g.w("\n%s", g.mangleDeclarator(d))
		}
		g.w(" = ")
		g.literal(d.Type, n)
		return
	}

	if g.isConstInitializer(d.Type, n) {
		b := make([]byte, g.model.Sizeof(d.Type))
		if !g.isZeroInitializer(n) {
			g.renderInitializer(b, d.Type, n)
		}
		switch {
		case g.escaped(d):
			g.w("\n%sCopy(%s, %q, %d)", g.crtPrefix, g.mangleDeclarator(d), b, len(b))
		default:
			g.w("\n%s = *(*%s)(unsafe.Pointer(%q))", g.mangleDeclarator(d), g.typ(d.Type), b)
		}
		return
	}

	switch {
	case g.initializerHasBitFields(d.Type, d.Initializer):
		switch n.Case {
		case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
			switch x := underlyingType(d.Type, true).(type) {
			case *cc.StructType:
				layout := g.model.Layout(x)
				var fld int64
				fields := x.Fields
				for l := n.InitializerList; l != nil; l = l.InitializerList {
					if fld < int64(len(layout)) {
						for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
							fld++
						}
					}
					if d := l.Designation; d != nil {
						l := d.List
						if len(l) != 1 {
							todo("", g.position(n))
						}

						fld = l[0]
					}

					switch n := l.Initializer; n.Case {
					case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
						todo("", g.position(n))
					case cc.InitializerExpr: // Expr
						fp := x.Field(fields[fld].Name) //TODO mixed index fieds vs layout
						e := &cc.Expr{
							Case: cc.ExprAssign,
							Expr: &cc.Expr{
								Case: cc.ExprSelect,
								Expr: &cc.Expr{
									Case:       cc.ExprIdent,
									Declarator: d,
									Scope:      d.Scope,
									Token:      xc.Token{Val: d.Name()},
								},
								Operand: cc.Operand{Type: fp.Type, FieldProperties: fp},
								Token2:  xc.Token{Val: fields[fld].Name}, //TODO mixed index fieds vs layout
							},
							Expr2:   n.Expr,
							Operand: cc.Operand{Type: fp.Declarator.Type},
						}
						g.w("\n")
						g.void(e)
					}

					fld++
				}
			default:
				todo("%v: %T", g.position(n), x)
			}
		case cc.InitializerExpr: // Expr
			todo("", g.position(n))
		}
	default:
		switch {
		case g.escaped(d):
			g.w("\n*(*%s)(unsafe.Pointer(%s))", g.typ(d.Type), g.mangleDeclarator(d))
		default:
			g.w("\n%s", g.mangleDeclarator(d))
		}
		g.w(" = ")
		g.literal(d.Type, n)
	}
}

func (g *gen) initializerHasBitFields(t cc.Type, n *cc.Initializer) bool {
	switch n.Case {
	case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			index := 0
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if l.Designation != nil {
					todo("", g.position0(n))
				}
				if g.initializerHasBitFields(x.Item, l.Initializer) {
					return true
				}

				index++
			}
			return false
		case *cc.StructType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position0(n))
					}

					fld = l[0]
				}

				if layout[fld].Bits > 0 {
					return true
				}

				if g.initializerHasBitFields(layout[fld].Type, l.Initializer) {
					return true
				}

				fld++
			}
			return false
		default:
			todo("%v: %T", g.position0(n), x)
		}
	case cc.InitializerExpr: // Expr
		switch x := underlyingType(t, true).(type) {
		case
			*cc.EnumType,
			*cc.PointerType,
			*cc.StructType:

			return false
		case cc.TypeKind:
			if x.IsScalarType() {
				return false
			}

			todo("%v: %v", g.position0(n), x)
		default:
			todo("%v: %T", g.position0(n), x)
		}
	}
	panic("unreachable")
}

func (g *ngen) initializerHasBitFields(t cc.Type, n *cc.Initializer) bool {
	switch n.Case {
	case cc.InitializerCompLit: // '{' InitializerList CommaOpt '}'
		switch x := underlyingType(t, true).(type) {
		case *cc.ArrayType:
			index := 0
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if l.Designation != nil {
					todo("", g.position(n))
				}
				if g.initializerHasBitFields(x.Item, l.Initializer) {
					return true
				}

				index++
			}
			return false
		case *cc.StructType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}

				if layout[fld].Bits > 0 {
					return true
				}

				if g.initializerHasBitFields(layout[fld].Type, l.Initializer) {
					return true
				}

				fld++
			}
			return false
		case *cc.UnionType:
			layout := g.model.Layout(x)
			var fld int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}

				if layout[fld].Bits > 0 {
					return true
				}

				if g.initializerHasBitFields(layout[fld].Type, l.Initializer) {
					return true
				}

				fld++
			}
			return false
		default:
			todo("%v: %T", g.position(n), x)
		}
	case cc.InitializerExpr: // Expr
		switch x := underlyingType(t, true).(type) {
		case
			*cc.EnumType,
			*cc.PointerType,
			*cc.StructType:

			return false
		case cc.TypeKind:
			if x.IsScalarType() {
				return false
			}

			todo("%v: %v", g.position(n), x)
		default:
			todo("%v: %T", g.position(n), x)
		}
	}
	panic("unreachable")
}

func (g *gen) literal(t cc.Type, n *cc.Initializer) {
	switch x := cc.UnderlyingType(t).(type) {
	case *cc.ArrayType:
		if n.Expr != nil {
			switch x.Item.Kind() {
			case
				cc.Char,
				cc.UChar:

				g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
				switch n.Expr.Case {
				case cc.ExprString:
					s := dict.S(int(n.Expr.Operand.Value.(*ir.StringValue).StringID))
					switch {
					case x.Size.Value == nil:
						g.w("ts+%d", g.allocString(dict.ID(s)))
					default:
						b := make([]byte, x.Size.Value.(*ir.Int64Value).Value)
						copy(b, s)
						if len(b) != 0 && b[len(b)-1] == 0 {
							b = b[:len(b)-1]
						}
						g.w("ts+%d", g.allocString(dict.ID(b)))
					}
				default:
					todo("", g.position0(n), n.Expr.Case)
				}
				g.w("))")
			default:
				todo("", g.position0(n), x.Item.Kind())
			}
			return
		}

		g.w("%s{", g.typ(t))
		g.initializerListNL(n.InitializerList)
		if !g.isZeroInitializer(n) {
			index := 0
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if l.Designation != nil {
					todo("", g.position0(n))
				}
				if !g.isZeroInitializer(l.Initializer) {
					g.w("%d: ", index)
					g.literal(x.Item, l.Initializer)
					g.w(", ")
					g.initializerListNL(n.InitializerList)
				}
				index++
			}
		}
		g.w("}")
	case *cc.PointerType:
		if n.Expr.IsZero() || n.Expr.Operand.Value == cc.Null {
			g.w("0")
			return
		}

		g.value(n.Expr, false)
	case *cc.StructType:
		if n.Expr != nil {
			g.value(n.Expr, false)
			return
		}

		g.w("%s{", g.typ(t))
		g.initializerListNL(n.InitializerList)
		if !g.isZeroInitializer(n) {
			layout := g.model.Layout(t)
			var fld int64
			fields := x.Fields
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position0(n))
					}

					fld = l[0]
				}
				switch {
				case layout[fld].Bits > 0:
					todo("bit field %v", g.position0(n))
				}
				if !g.isZeroInitializer(l.Initializer) {
					d := fields[fld] //TODO mixed index fieds vs layout
					g.w("%s: ", mangleIdent(d.Name, true))
					g.literal(d.Type, l.Initializer)
					g.w(", ")
					g.initializerListNL(n.InitializerList)
				}
				fld++
			}
		}
		g.w("}")
	case *cc.EnumType:
		switch n.Case {
		case cc.InitializerExpr:
			g.value(n.Expr, false)
		default:
			todo("", g.position0(n), n.Case)
		}
	case cc.TypeKind:
		if x.IsArithmeticType() {
			g.convert(n.Expr, t)
			return
		}
		todo("", g.position0(n), x)
	case *cc.UnionType:
		// *(*struct{ X int32 })(unsafe.Pointer(&struct{int32}{int32(1)})),
		if g.isZeroInitializer(n) {
			g.w("%s{}", g.typ(t))
			return
		}

		if n.Expr != nil {
			todo("", g.position0(n), x)
			return
		}

		g.w("*(*%s)(unsafe.Pointer(&struct{", g.typ(t))
		if !g.isZeroInitializer(n) {
			layout := g.model.Layout(t)
			var fld int64
			fields := x.Fields
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position0(n))
					}

					fld = l[0]
				}
				switch {
				case layout[fld].Bits > 0:
					todo("bit field %v", g.position0(n))
				}
				if fld != 0 {
					todo("", g.position0(n))
				}

				d := fields[fld] //TODO mixed index fieds vs layout
				switch pad := g.model.Sizeof(t) - g.model.Sizeof(d.Type); {
				case pad == 0:
					g.w("%s}{", g.typ(d.Type))
				default:
					g.w("f %s; _[%d]byte}{f: ", g.typ(d.Type), pad)
				}
				g.literal(d.Type, l.Initializer)
				fld++
			}
		}
		g.w("}))")
	default:
		todo("%v: %T", g.position0(n), x)
	}
}

func (g *ngen) literal(t cc.Type, n *cc.Initializer) {
	switch x := cc.UnderlyingType(t).(type) {
	case *cc.ArrayType:
		if n.Expr != nil {
			it := x.Item
			switch y := it.(type) {
			case cc.TypeKind:
				switch y {
				case
					cc.Char,
					cc.UChar:

					g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
					switch n.Expr.Case {
					case cc.ExprString:
						s := dict.S(int(n.Expr.Operand.Value.(*ir.StringValue).StringID))
						switch {
						case x.Size.Value == nil:
							g.w("%q", s)
						default:
							b := make([]byte, x.Size.Value.(*ir.Int64Value).Value)
							copy(b, s)
							if len(b) != 0 && b[len(b)-1] == 0 {
								b = b[:len(b)-1]
							}
							g.w("%q", string(b))
						}
					default:
						todo("", g.position(n), n.Expr.Case)
					}
					g.w("))")
				}
			case *cc.NamedType:
				switch {
				case y.Name == idWcharT:
					switch it := underlyingType(y.Type, false); it.Kind() {
					case cc.Int:
						sz := g.model.Sizeof(it)
						g.w("*(*%s)(unsafe.Pointer(", g.typ(t))
						switch n.Expr.Case {
						case cc.ExprLString:
							s := n.Expr.Operand.Value.(*ir.WideStringValue).Value
							switch {
							case x.Size.Value == nil:
								todo("", g.position(n))
							default:
								b := make([]byte, len(s)*int(sz)+3)
								for i, v := range s {
									switch sz {
									case 4:
										*(*uint32)(unsafe.Pointer(&b[4*i])) = uint32(v)
									default:
										todo("", g.position(n), sz)
									}
								}
								g.w("%q", string(b))
							}
						default:
							todo("", g.position(n), n.Expr.Case)
						}
						g.w("))")
					default:
						todo("%v: %v", g.position(n), it)
					}
				default:
					todo("%v:", g.position(n))
				}
			default:
				todo("%v: %T", g.position(n), y)
			}
			return
		}

		g.w("%s{", g.typ(t))
		g.initializerListNL(n.InitializerList)
		if !g.isZeroInitializer(n) {
			var index int64
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					index = l[0]
				}
				if !g.isZeroInitializer(l.Initializer) {
					g.w("%d: ", index)
					g.literal(x.Item, l.Initializer)
					g.w(", ")
					g.initializerListNL(n.InitializerList)
				}
				index++
			}
		}
		g.w("}")
	case *cc.PointerType:
		if n.Expr.IsZero() && g.voidCanIgnore(n.Expr) || n.Expr.Operand.Value == cc.Null {
			g.w("0")
			return
		}

		g.value(n.Expr, false)
	case *cc.StructType:
		if n.Expr != nil {
			g.value(n.Expr, false)
			return
		}

		g.w("%s{", g.typ(t))
		g.initializerListNL(n.InitializerList)
		if !g.isZeroInitializer(n) {
			layout := g.model.Layout(t)
			var fld int64
			fields := x.Fields
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}
				switch {
				case layout[fld].Bits > 0:
					todo("bit field %v", g.position(n))
				}
				if !g.isZeroInitializer(l.Initializer) {
					d := fields[fld] //TODO mixed index fieds vs layout
					g.w("F%s: ", dict.S(d.Name))
					g.literal(d.Type, l.Initializer)
					g.w(", ")
					g.initializerListNL(n.InitializerList)
				}
				fld++
			}
		}
		g.w("}")
	case *cc.EnumType:
		switch n.Case {
		case cc.InitializerExpr:
			todo("", g.position(n))
			//TODO g.value(n.Expr, false)
		default:
			todo("", g.position(n), n.Case)
		}
	case cc.TypeKind:
		if x.IsArithmeticType() {
			g.convert(n.Expr, t)
			return
		}

		todo("", g.position(n), x)
	case *cc.UnionType:
		// *(*struct{ X int32 })(unsafe.Pointer(&struct{int32}{int32(1)})),
		if g.isZeroInitializer(n) {
			g.w("%s{}", g.typ(t))
			return
		}

		if n.Expr != nil {
			g.value(n.Expr, false)
			return
		}

		g.w("*(*%s)(unsafe.Pointer(&struct{", g.typ(t))
		if !g.isZeroInitializer(n) {
			layout := g.model.Layout(t)
			var fld int64
			fields := x.Fields
			for l := n.InitializerList; l != nil; l = l.InitializerList {
				if fld < int64(len(layout)) {
					for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
						fld++
					}
				}
				if d := l.Designation; d != nil {
					l := d.List
					if len(l) != 1 {
						todo("", g.position(n))
					}

					fld = l[0]
				}
				switch {
				case layout[fld].Bits > 0:
					todo("bit field %v", g.position(n))
				}

				d := fields[fld] //TODO mixed index fieds vs layout
				switch pad := g.model.Sizeof(t) - g.model.Sizeof(d.Type); {
				case pad == 0:
					g.w("f %s}{", g.typ(d.Type))
				default:
					g.w("f %s; _[%d]byte}{f: ", g.typ(d.Type), pad)
				}
				g.literal(d.Type, l.Initializer)
				fld++
			}
		}
		g.w("}))")
	default:
		todo("%v: %T", g.position(n), x)
	}
}

func (g *gen) initializerListNL(n *cc.InitializerList) {
	if n.Len > 1 {
		g.w("\n")
	}
}

func (g *ngen) initializerListNL(n *cc.InitializerList) {
	if n.Len > 1 {
		g.w("\n")
	}
}

func (g *gen) renderInitializer(b []byte, t cc.Type, n *cc.Initializer) {
	switch x := cc.UnderlyingType(t).(type) {
	case *cc.ArrayType:
		if n.Expr != nil {
			switch y := n.Expr.Operand.Value.(type) {
			case *ir.StringValue:
				switch z := x.Item.Kind(); z {
				case
					cc.Char,
					cc.UChar:

					copy(b, dict.S(int(y.StringID)))
				default:
					todo("", g.position0(n), z)
				}
			default:
				todo("%v: %T", g.position0(n), y)
			}
			return
		}

		itemSz := g.model.Sizeof(x.Item)
		var index int64
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if l.Designation != nil {
				todo("", g.position0(n))
			}
			lo := index * itemSz
			hi := lo + itemSz
			g.renderInitializer(b[lo:hi:hi], x.Item, l.Initializer)
			index++
		}
	case *cc.PointerType:
		switch {
		case n.Expr.IsNonZero():
			*(*uintptr)(unsafe.Pointer(&b[0])) = uintptr(n.Expr.Operand.Value.(*ir.Int64Value).Value)
		case n.Expr.IsZero():
			// nop
		default:
			todo("", g.position0(n), n.Expr.Operand)
		}
	case *cc.StructType:
		if n.Expr != nil {
			todo("", g.position0(n))
		}

		layout := g.model.Layout(t)
		var fld int64
		fields := x.Fields
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if fld < int64(len(layout)) {
				for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
					fld++
				}
			}
			if d := l.Designation; d != nil {
				l := d.List
				if len(l) != 1 {
					todo("", g.position0(n))
				}

				fld = l[0]
			}
			fp := layout[fld]
			lo := fp.Offset
			hi := lo + fp.Size
			switch {
			case fp.Bits > 0:
				v := uint64(l.Initializer.Expr.Operand.Value.(*ir.Int64Value).Value)
				switch sz := g.model.Sizeof(fp.PackedType); sz {
				case 1:
					m := fp.Mask()
					x := uint64(b[lo])
					x = x&^m | v<<uint(fp.Bitoff)&m
					b[lo] = byte(x)
				case 2:
					m := fp.Mask()
					x := uint64(*(*uint16)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint16)(unsafe.Pointer(&b[lo])) = uint16(x)
				case 4:
					m := fp.Mask()
					x := uint64(*(*uint32)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint32)(unsafe.Pointer(&b[lo])) = uint32(x)
				case 8:
					m := fp.Mask()
					x := *(*uint64)(unsafe.Pointer(&b[lo]))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint64)(unsafe.Pointer(&b[lo])) = x
				default:
					todo("", g.position0(n), sz, v)
				}
			default:
				g.renderInitializer(b[lo:hi:hi], fields[fld].Type, l.Initializer) //TODO mixed index fieds vs layout
			}
			fld++
		}
	case cc.TypeKind:
		if x.IsIntegerType() {
			var v int64
			switch y := n.Expr.Operand.Value.(type) {
			case *ir.Float64Value:
				v = int64(y.Value)
			case *ir.Int64Value:
				v = y.Value
			default:
				todo("%v: %T", g.position0(n), y)
			}
			switch sz := g.model[x].Size; sz {
			case 1:
				*(*int8)(unsafe.Pointer(&b[0])) = int8(v)
			case 2:
				*(*int16)(unsafe.Pointer(&b[0])) = int16(v)
			case 4:
				*(*int32)(unsafe.Pointer(&b[0])) = int32(v)
			case 8:
				*(*int64)(unsafe.Pointer(&b[0])) = v
			default:
				todo("", g.position0(n), sz)
			}
			return
		}

		switch x {
		case cc.Float:
			switch x := n.Expr.Operand.Value.(type) {
			case *ir.Float32Value:
				*(*float32)(unsafe.Pointer(&b[0])) = x.Value
			case *ir.Float64Value:
				*(*float32)(unsafe.Pointer(&b[0])) = float32(x.Value)
			}
		case
			cc.Double,
			cc.LongDouble:

			switch x := n.Expr.Operand.Value.(type) {
			case *ir.Float64Value:
				*(*float64)(unsafe.Pointer(&b[0])) = x.Value
			case *ir.Int64Value:
				*(*float64)(unsafe.Pointer(&b[0])) = float64(x.Value)
			}
		default:
			todo("", g.position0(n), x)
		}
	case *cc.UnionType:
		if n.Expr != nil {
			todo("", g.position0(n))
		}

		layout := g.model.Layout(t)
		var fld int64
		fields := x.Fields
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if fld < int64(len(layout)) {
				for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
					fld++
				}
			}
			if d := l.Designation; d != nil {
				l := d.List
				if len(l) != 1 {
					todo("", g.position0(n))
				}

				fld = l[0]
			}
			if fld != 0 {
				todo("%v", g.position0(n))
			}
			fp := layout[fld]
			lo := fp.Offset
			hi := lo + fp.Size
			switch {
			case layout[fld].Bits > 0:
				v := uint64(l.Initializer.Expr.Operand.Value.(*ir.Int64Value).Value)
				switch sz := g.model.Sizeof(fp.PackedType); sz {
				case 1:
					m := fp.Mask()
					x := uint64(b[lo])
					x = x&^m | v<<uint(fp.Bitoff)&m
					b[lo] = byte(x)
				case 2:
					m := fp.Mask()
					x := uint64(*(*uint16)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint16)(unsafe.Pointer(&b[lo])) = uint16(x)
				case 4:
					m := fp.Mask()
					x := uint64(*(*uint32)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint32)(unsafe.Pointer(&b[lo])) = uint32(x)
				case 8:
					m := fp.Mask()
					x := *(*uint64)(unsafe.Pointer(&b[lo]))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint64)(unsafe.Pointer(&b[lo])) = x
				default:
					todo("", g.position0(n), sz, v)
				}
			default:
				g.renderInitializer(b[lo:hi:hi], fields[fld].Type, l.Initializer) //TODO mixed index fieds vs layout
			}
			fld++
		}
	default:
		todo("%v: %T", g.position0(n), x)
	}
}

func (g *ngen) renderInitializer(b []byte, t cc.Type, n *cc.Initializer) {
	switch x := cc.UnderlyingType(t).(type) {
	case *cc.ArrayType:
		if n.Expr != nil {
			switch y := n.Expr.Operand.Value.(type) {
			case *ir.StringValue:
				switch z := x.Item.Kind(); z {
				case
					cc.Char,
					cc.UChar:

					copy(b, dict.S(int(y.StringID)))
				default:
					todo("", g.position(n), z)
				}
			default:
				todo("%v: %T", g.position(n), y)
			}
			return
		}

		itemSz := g.model.Sizeof(x.Item)
		if x.Item.Kind() == cc.Char && n.InitializerList.Len == 1 {
			in := n.InitializerList.Initializer
			if in.Expr != nil {
				switch x := in.Expr.Operand.Value.(type) {
				case *ir.StringValue:
					copy(b, dict.S(int(x.StringID)))
					return
				}
			}
		}

		var index int64
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if d := l.Designation; d != nil {
				l := d.List
				if len(l) != 1 {
					todo("", g.position(n), l)
				}

				index = l[0]
			}
			lo := index * itemSz
			hi := lo + itemSz
			g.renderInitializer(b[lo:hi:hi], x.Item, l.Initializer)
			index++
		}
	case *cc.PointerType:
		switch {
		case n.Expr.IsNonZero() && g.voidCanIgnore(n.Expr):
			*(*uintptr)(unsafe.Pointer(&b[0])) = uintptr(n.Expr.Operand.Value.(*ir.Int64Value).Value)
		case n.Expr.IsZero() && g.voidCanIgnore(n.Expr):
			// nop
		default:
			todo("", g.position(n), n.Expr.Operand)
		}
	case *cc.StructType:
		if n.Expr != nil {
			todo("", g.position(n))
		}

		layout := g.model.Layout(t)
		var fld int64
		fields := x.Fields
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if fld < int64(len(layout)) {
				for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
					fld++
				}
			}
			if d := l.Designation; d != nil {
				l := d.List
				if len(l) != 1 {
					todo("", g.position(n))
				}

				fld = l[0]
			}
			fp := layout[fld]
			lo := fp.Offset
			hi := lo + fp.Size
			switch {
			case fp.Bits > 0:
				v := uint64(l.Initializer.Expr.Operand.Value.(*ir.Int64Value).Value)
				switch sz := g.model.Sizeof(fp.PackedType); sz {
				case 1:
					m := fp.Mask()
					x := uint64(b[lo])
					x = x&^m | v<<uint(fp.Bitoff)&m
					b[lo] = byte(x)
				case 2:
					m := fp.Mask()
					x := uint64(*(*uint16)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint16)(unsafe.Pointer(&b[lo])) = uint16(x)
				case 4:
					m := fp.Mask()
					x := uint64(*(*uint32)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint32)(unsafe.Pointer(&b[lo])) = uint32(x)
				case 8:
					m := fp.Mask()
					x := *(*uint64)(unsafe.Pointer(&b[lo]))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint64)(unsafe.Pointer(&b[lo])) = x
				default:
					todo("", g.position(n), sz, v)
				}
			default:
				g.renderInitializer(b[lo:hi:hi], fields[fld].Type, l.Initializer) //TODO mixed index fieds vs layout
			}
			fld++
		}
	case cc.TypeKind:
		if x.IsIntegerType() {
			var v int64
			switch y := n.Expr.Operand.Value.(type) {
			case *ir.Float64Value:
				v = int64(y.Value)
			case *ir.Int64Value:
				v = y.Value
			default:
				todo("%v: %T", g.position(n), y)
			}
			switch sz := g.model[x].Size; sz {
			case 1:
				*(*int8)(unsafe.Pointer(&b[0])) = int8(v)
			case 2:
				*(*int16)(unsafe.Pointer(&b[0])) = int16(v)
			case 4:
				*(*int32)(unsafe.Pointer(&b[0])) = int32(v)
			case 8:
				*(*int64)(unsafe.Pointer(&b[0])) = v
			default:
				todo("", g.position(n), sz)
			}
			return
		}

		switch x {
		case cc.Float:
			switch x := n.Expr.Operand.Value.(type) {
			case *ir.Float32Value:
				*(*float32)(unsafe.Pointer(&b[0])) = x.Value
			case *ir.Float64Value:
				*(*float32)(unsafe.Pointer(&b[0])) = float32(x.Value)
			}
		case
			cc.Double,
			cc.LongDouble:

			switch x := n.Expr.Operand.Value.(type) {
			case *ir.Float64Value:
				*(*float64)(unsafe.Pointer(&b[0])) = x.Value
			case *ir.Int64Value:
				*(*float64)(unsafe.Pointer(&b[0])) = float64(x.Value)
			}
		default:
			todo("", g.position(n), x)
		}
	case *cc.UnionType:
		layout := g.model.Layout(t)
		if n.Expr != nil {
			fp := layout[0]
			if fp.Bits != 0 {
				todo("", g.position(n))
			}

			lo := fp.Offset
			hi := lo + fp.Size
			g.renderInitializerExpr(b[lo:hi:hi], fp.Type, n.Expr)
			return
		}

		fields := x.Fields
		var fld int64
		for l := n.InitializerList; l != nil; l = l.InitializerList {
			if fld < int64(len(layout)) {
				for !layout[fld].Anonymous && (layout[fld].Bits < 0 || layout[fld].Declarator == nil) {
					fld++
				}
			}
			if d := l.Designation; d != nil {
				l := d.List
				if len(l) != 1 {
					todo("", g.position(n))
				}

				fld = l[0]
			}
			if fld != 0 {
				todo("%v", g.position(n))
			}
			fp := layout[fld]
			lo := fp.Offset
			hi := lo + fp.Size
			switch {
			case layout[fld].Bits > 0:
				v := uint64(l.Initializer.Expr.Operand.Value.(*ir.Int64Value).Value)
				switch sz := g.model.Sizeof(fp.PackedType); sz {
				case 1:
					m := fp.Mask()
					x := uint64(b[lo])
					x = x&^m | v<<uint(fp.Bitoff)&m
					b[lo] = byte(x)
				case 2:
					m := fp.Mask()
					x := uint64(*(*uint16)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint16)(unsafe.Pointer(&b[lo])) = uint16(x)
				case 4:
					m := fp.Mask()
					x := uint64(*(*uint32)(unsafe.Pointer(&b[lo])))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint32)(unsafe.Pointer(&b[lo])) = uint32(x)
				case 8:
					m := fp.Mask()
					x := *(*uint64)(unsafe.Pointer(&b[lo]))
					x = x&^m | v<<uint(fp.Bitoff)&m
					*(*uint64)(unsafe.Pointer(&b[lo])) = x
				default:
					todo("", g.position(n), sz, v)
				}
			default:
				g.renderInitializer(b[lo:hi:hi], fields[fld].Type, l.Initializer) //TODO mixed index fieds vs layout
			}
			fld++
		}
	default:
		todo("%v: %T", g.position(n), x)
	}
}

func (g *ngen) renderInitializerExpr(b []byte, t cc.Type, n *cc.Expr) {
	switch x := cc.UnderlyingType(t).(type) {
	case cc.TypeKind:
		switch x {
		case cc.Int:
			v := uint64(n.Operand.Value.(*ir.Int64Value).Value)
			switch sz := g.model.Sizeof(t); sz {
			case 4:
				*(*uint32)(unsafe.Pointer(&b[0])) = uint32(v)
			default:
				todo("", g.position(n), sz, v)
			}
		default:
			todo("%v: %v", g.position(n), x)
		}
	default:
		todo("%v: %T", g.position(n), x)
	}
}
