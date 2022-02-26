// Copyright 2016 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc"

import (
	"bytes"
	"fmt"
	"go/token"
	"path/filepath"
	"strings"

	"modernc.org/golex/lex"
	"modernc.org/mathutil"
	"modernc.org/strutil"
	"modernc.org/xc"
)

var (
	_ Specifier = (*DeclarationSpecifiers)(nil)
	_ Specifier = (*SpecifierQualifierList)(nil)
	_ Specifier = (*spec)(nil)

	_ Type = (*ctype)(nil)
)

var (
	noTypedefNameAfter = map[rune]bool{
		'*':         true,
		'.':         true,
		ARROW:       true,
		BOOL:        true,
		CHAR:        true,
		COMPLEX:     true,
		DOUBLE:      true,
		ENUM:        true,
		FLOAT:       true,
		GOTO:        true,
		INT:         true,
		LONG:        true,
		SHORT:       true,
		SIGNED:      true,
		STRUCT:      true,
		TYPEDEFNAME: true,
		UNION:       true,
		UNSIGNED:    true,
		VOID:        true,
	}

	undefined        = &ctype{}
	debugTypeStrings bool
)

// EnumConstant represents the name/value pair defined by an Enumerator.
type EnumConstant struct {
	DefTok xc.Token    // Enumeration constant name definition token.
	Value  interface{} // Value represented by name. Type of Value is C int.
	Tokens []xc.Token  // The tokens the constant expression consists of.
}

// Specifier describes a combination of {Function,StorageClass,Type}Specifiers
// and TypeQualifiers.
type Specifier interface {
	IsAuto() bool                       // StorageClassSpecifier "auto" present.
	IsConst() bool                      // TypeQualifier "const" present.
	IsExtern() bool                     // StorageClassSpecifier "extern" present.
	IsInline() bool                     // FunctionSpecifier "inline" present.
	IsRegister() bool                   // StorageClassSpecifier "register" present.
	IsRestrict() bool                   // TypeQualifier "restrict" present.
	IsStatic() bool                     // StorageClassSpecifier "static" present.
	IsTypedef() bool                    // StorageClassSpecifier "typedef" present.
	IsVolatile() bool                   // TypeQualifier "volatile" present.
	TypedefName() int                   // TypedefName returns the typedef name ID used, if any, zero otherwise.
	attrs() int                         // Encoded attributes.
	firstTypeSpecifier() *TypeSpecifier //
	kind() Kind                         //
	member(int) (*Member, error)        //
	str() string                        //
	typeSpecifiers() int                // Encoded TypeSpecifier combination.
}

// Type decribes properties of a C type.
type Type interface {
	// AlignOf returns the alignment in bytes of a value of this type when
	// allocated in memory not as a struct field. Incomplete struct types
	// have no alignment and the value returned will be < 0.
	AlignOf() int

	// Bits returns the bit width of the type's value. For non integral
	// types the returned value will < 0.
	Bits() int

	// SetBits returns a type instance with the value Bits() will return
	// equal to n. SetBits panics for n < 0.
	SetBits(n int) Type

	// CanAssignTo returns whether this type can be assigned to dst.
	CanAssignTo(dst Type) bool

	// Declarator returns the full Declarator which defined an entity of
	// this type. The returned declarator is possibly artificial.
	Declarator() *Declarator

	// RawDeclarator returns the typedef declarator associated with a type
	// if this type is a typedef name. Otherwise the normal declarator is
	// returned.
	RawDeclarator() *Declarator

	// Element returns the type this Ptr type points to or the element type
	// of an Array type.
	Element() Type

	// Elements returns the number of elements an Array type has. The
	// returned value is < 0 if this type is not an Array or if the array
	// is not of a constant size.
	Elements() int

	// EnumeratorList returns the enumeration constants defined by an Enum
	// type, if any.
	EnumeratorList() []EnumConstant

	// Kind returns one of Ptr, Void, Int, ...
	Kind() Kind

	// Member returns the type of a member of this Struct or Union type,
	// having numeric name identifier nm.
	Member(nm int) (*Member, error)

	// Members returns the members of a Struct or Union type in declaration
	// order. Returned members are valid iff non nil.
	//
	// Note: Non nil members of length 0 means the struct/union has no
	// members or the type is incomplete, which is indicated by the
	// isIncomplete return value.
	//
	// Note 2: C99 standard does not allow empty structs/unions, but GCC
	// supports that as an extension.
	Members() (members []Member, isIncomplete bool)

	// Parameters returns the parameters of a Function type in declaration
	// order. Result is valid iff params is not nil.
	//
	// Note: len(params) == 0 is fine and just means the function has no
	// parameters.
	Parameters() (params []Parameter, isVariadic bool)

	// Pointer returns a type that points to this type.
	Pointer() Type

	// Result returns the result type of a Function type.
	Result() Type

	// Sizeof returns the number of bytes needed to store a value of this
	// type. Incomplete struct types have no size and the value returned
	// will be < 0.
	SizeOf() int

	// Specifier returns the Specifier of this type.
	Specifier() Specifier

	// String returns a C-like type specifier of this type.
	String() string

	// StructAlignOf returns the alignment in bytes of a value of this type
	// when allocated in memory as a struct field. Incomplete struct types
	// have no alignment and the value returned will be < 0.
	StructAlignOf() int

	// Tag returns the ID of a tag of a Struct, Union or Enum type, if any.
	// Otherwise the returned value is zero.
	Tag() int

	sizeOf(*lexer) int
	structAlignOf(*lexer) int
}

// Member describes a member of a struct or union.
//
// BitFieldGroup represents the ordinal number of the packed bit fields:
//
//	struct foo {
//		int i;
//		int j:1;	// BitFieldGroup: 0
//		int k:2;	// BitFieldGroup: 0
//		double l;
//		int m:1;	// BitFieldGroup: 1
//		int n:2;	// BitFieldGroup: 1
//	}
type Member struct {
	BitFieldType  Type
	BitFieldGroup int         // Ordinal number of the packed bits field.
	BitOffsetOf   int         // Bit field starting bit.
	Bits          int         // Size in bits for bit fields, 0 otherwise.
	Declarator    *Declarator // Possibly nil for bit fields.
	Name          int
	OffsetOf      int
	Padding       int // Number of unused bytes added to the end of the field to force proper alignment requirements.
	Type          Type
}

// Parameter describes a function argument.
type Parameter struct {
	Declarator *Declarator
	Name       int
	Type       Type
}

// PrettyString pretty prints things produced by this package.
func PrettyString(v interface{}) string {
	return strutil.PrettyString(v, "", "", printHooks)
}

func position(pos token.Pos) token.Position { return fset.Position(pos) }

// Binding records the declaration Node of a declared name.
//
// In the NSIdentifiers namespace the dynamic type of Node for declared names
// is always *DirectDeclarator.  The *Declarator associated with the direct
// declarator is available via (*DirectDeclarator).TopDeclarator().
//
//	int* p;
//
// In the NSTags namespace the dynamic type of Node is xc.Token when a tag is
// declared:
//
//	struct foo;
//	enum bar;
//
// When a tag is defined, the dynamic type of Node is *EnumSpecifier or
// *StructOrUnionSpecifier:
//
//	struct foo { int i; };
//	enum bar { a = 1 };
//
type Binding struct {
	Node Node
	enum bool
}

// Bindings record names declared in a scope.
type Bindings struct {
	Identifiers map[int]Binding // NSIdentifiers name space bindings.
	Tags        map[int]Binding // NSTags name space bindings.
	kind        Scope           // ScopeFile, ...
	Parent      *Bindings       // Parent scope or nil for ScopeFile.

	// Scoped helpers.

	mergeScope *Bindings // Fn params.
	specifier  Specifier // To store in full declarators.

	// Struct/union field handling.
	bitFieldGroup        int         // Group ordinal number.
	bitFieldTypes        []Type      //
	bitOffset            int         //
	isUnion              bool        //
	maxAlign             int         //
	maxSize              int         //
	offset               int         //
	prevStructDeclarator *Declarator //
}

func newBindings(parent *Bindings, kind Scope) *Bindings {
	return &Bindings{
		kind:   kind,
		Parent: parent,
	}
}

// Scope retuns the kind of b.
func (b *Bindings) Scope() Scope { return b.kind }

func (b *Bindings) merge(c *Bindings) {
	if b.kind != ScopeBlock || len(b.Identifiers) != 0 || c.kind != ScopeParams {
		panic("internal error")
	}

	b.boot(NSIdentifiers)
	for k, v := range c.Identifiers {
		b.Identifiers[k] = v
	}
}

func (b *Bindings) boot(ns Namespace) map[int]Binding {
	var m *map[int]Binding
	switch ns {
	case NSIdentifiers:
		m = &b.Identifiers
	case NSTags:
		m = &b.Tags
	default:
		panic(fmt.Errorf("internal error %v", ns))
	}

	mp := *m
	if mp == nil {
		mp = make(map[int]Binding)
		*m = mp
	}
	return mp
}

func (b *Bindings) root() *Bindings {
	for b.Parent != nil {
		b = b.Parent
	}
	return b
}

// Lookup returns the Binding of id in ns or any of its parents. If id is
// undeclared, the returned Binding has its Node field set to nil.
func (b *Bindings) Lookup(ns Namespace, id int) Binding {
	r, _ := b.Lookup2(ns, id)
	return r
}

// Lookup2 is like Lookup but addionally it returns also the scope in which id
// was found.
func (b *Bindings) Lookup2(ns Namespace, id int) (Binding, *Bindings) {
	if ns == NSTags {
		b = b.root()
	}
	for b != nil {
		m := b.boot(ns)
		if x, ok := m[id]; ok {
			return x, b
		}

		b = b.Parent
	}

	return Binding{}, nil
}

func (b *Bindings) declareIdentifier(tok xc.Token, d *DirectDeclarator, report *xc.Report) {
	m := b.boot(NSIdentifiers)
	var p *Binding
	if ex, ok := m[tok.Val]; ok {
		p = &ex
	}

	d.prev = p
	m[tok.Val] = Binding{d, false}
}

func (b *Bindings) declareEnumTag(tok xc.Token, report *xc.Report) {
	b = b.root()
	m := b.boot(NSTags)
	if ex, ok := m[tok.Val]; ok {
		if !ex.enum {
			report.ErrTok(tok, "struct tag redeclared as enum tag, previous declaration/definition: %s", position(ex.Node.Pos()))
		}
		return
	}

	m[tok.Val] = Binding{tok, true}
}

func (b *Bindings) defineEnumTag(tok xc.Token, n Node, report *xc.Report) {
	b = b.root()
	m := b.boot(NSTags)
	if ex, ok := m[tok.Val]; ok {
		if !ex.enum {
			report.ErrTok(tok, "struct tag redefined as enum tag, previous declaration/definition: %s", position(ex.Node.Pos()))
			return
		}

		if _, ok := ex.Node.(xc.Token); !ok {
			report.ErrTok(tok, "enum tag redefined, previous definition: %s", position(ex.Node.Pos()))
			return
		}
	}

	m[tok.Val] = Binding{n, true}
}

func (b *Bindings) defineEnumConst(lx *lexer, tok xc.Token, v interface{}) *Declarator {
	b = b.root()
	d := lx.model.makeDeclarator(0, tsInt)
	dd := d.DirectDeclarator
	dd.Token = tok
	dd.EnumVal = v
	d.setFull(lx)
	b.declareIdentifier(tok, dd, lx.report)
	switch x := v.(type) {
	case int16:
		lx.iota = int64(x) + 1
	case int32:
		lx.iota = int64(x) + 1
	case int64:
		lx.iota = x + 1
	default:
		panic(fmt.Errorf("%T", x))
	}
	return d
}

func (b *Bindings) declareStructTag(tok xc.Token, report *xc.Report) {
	b = b.root()
	m := b.boot(NSTags)
	if ex, ok := m[tok.Val]; ok {
		if ex.enum {
			report.ErrTok(tok, "enum tag redeclared as struct tag, previous declaration/definition: %s", position(ex.Node.Pos()))
		}
		return
	}

	m[tok.Val] = Binding{tok, false}
}

func (b *Bindings) defineStructTag(tok xc.Token, n Node, report *xc.Report) {
	b = b.root()
	m := b.boot(NSTags)
	if ex, ok := m[tok.Val]; ok {
		if ex.enum {
			report.ErrTok(tok, "enum tag redefined as struct tag, previous declaration/definition: %s", position(ex.Node.Pos()))
			return
		}

		if _, ok := ex.Node.(xc.Token); !ok {
			if !n.(*StructOrUnionSpecifier).isCompatible(ex.Node.(*StructOrUnionSpecifier)) {
				report.ErrTok(tok, "incompatible struct tag redefinition, previous definition at %s", position(ex.Node.Pos()))
			}
			return
		}
	}

	m[tok.Val] = Binding{n, false}
}

func (b *Bindings) isTypedefName(id int) bool {
	x := b.Lookup(NSIdentifiers, id)
	if dd, ok := x.Node.(*DirectDeclarator); ok {
		return dd.specifier.IsTypedef()
	}

	return false
}

func (b *Bindings) lexerHack(tok, prev xc.Token) xc.Token { // https://en.wikipedia.org/wiki/The_lexer_hack
	if noTypedefNameAfter[prev.Rune] {
		return tok
	}

	if tok.Rune == IDENTIFIER && b.isTypedefName(tok.Val) {
		tok.Char = lex.NewChar(tok.Pos(), TYPEDEFNAME)
	}
	return tok
}

func errPos(a ...token.Pos) token.Pos {
	for _, v := range a {
		if v.IsValid() {
			return v
		}
	}

	return token.Pos(0)
}

func isZero(v interface{}) bool { return !isNonZero(v) }

func isNonZero(v interface{}) bool {
	switch x := v.(type) {
	case int32:
		return x != 0
	case int:
		return x != 0
	case uint32:
		return x != 0
	case int64:
		return x != 0
	case uint64:
		return x != 0
	case float32:
		return x != 0
	case float64:
		return x != 0
	case StringLitID, LongStringLitID:
		return true
	default:
		panic(fmt.Errorf("internal error: %T", x))
	}
}

func fromSlashes(a []string) []string {
	for i, v := range a {
		a[i] = filepath.FromSlash(v)
	}
	return a
}

type ctype struct {
	bits            int
	dds             []*DirectDeclarator // Expanded.
	dds0            []*DirectDeclarator // Unexpanded, only for typedefs
	model           *Model
	resultAttr      int
	resultSpecifier Specifier
	resultStars     int
	stars           int
}

func (n *ctype) SetBits(b int) Type {
	if b < 0 {
		panic("internal error")
	}

	if b == n.bits {
		return n
	}

	o := *n
	o.bits = b
	return &o
}

func (n *ctype) Bits() int {
	if n.bits > 0 {
		return n.bits
	}

	if !IsIntType(n) {
		return -1
	}

	n.bits = n.model.Items[n.Kind()].Size * 8
	return n.bits
}

func (n *ctype) arrayDecay() *ctype {
	return n.setElements(-1)
}

func (n *ctype) setElements(elems int) *ctype {
	m := *n
	m.dds = append([]*DirectDeclarator(nil), n.dds...)
	for i, dd := range m.dds {
		switch dd.Case {
		case 0: // IDENTIFIER
			// nop
		case 2: // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			dd := dd.clone()
			dd.elements = elems
			m.dds[i] = dd
			return &m
		default:
			//dbg("", position(dd.Pos()), n.str(), elems)
			panic(dd.Case)
		}
	}
	return n
}

func (n *ctype) eq(m *ctype) (r bool) {
	const ignore = saInline | saTypedef | saExtern | saStatic | saAuto | saRegister | saConst | saRestrict | saVolatile | saNoreturn

	if n == m {
		return true
	}

	if len(n.dds) != len(m.dds) || n.resultAttr&^ignore != m.resultAttr&^ignore ||
		n.resultStars != m.resultStars || n.stars != m.stars {
		return false
	}

	for i, n := range n.dds {
		if !n.isCompatible(m.dds[i]) {
			return false
		}
	}

	return n.resultSpecifier.str() == m.resultSpecifier.str()
}

func (n *ctype) isCompatible(m *ctype) (r bool) {
	const ignore = saInline | saTypedef | saExtern | saStatic | saAuto | saRegister | saConst | saRestrict | saVolatile | saNoreturn

	if n == m {
		return true
	}

	if n.Kind() == Array {
		n = n.arrayDecay()
	}

	if m.Kind() == Array {
		m = m.arrayDecay()
	}

	if len(n.dds) != len(m.dds) || n.resultAttr&^ignore != m.resultAttr&^ignore ||
		n.resultStars != m.resultStars || n.stars != m.stars {
		return false
	}

	if n.Kind() == Function && m.Kind() == Function {
		a, va := n.Parameters()
		b, vb := m.Parameters()
		return isCompatibleParameters(a, b, va, vb)
	}

	for i, n := range n.dds {
		if !n.isCompatible(m.dds[i]) {
			return false
		}
	}

	ns := n.resultSpecifier
	ms := m.resultSpecifier
	if ns == ms {
		return true
	}

	if n.Kind() != m.Kind() {
		return false
	}

	switch ns.kind() {
	case Array:
		panic("internal error")
	case Struct, Union:
		return n.structOrUnionSpecifier().isCompatible(m.structOrUnionSpecifier())
	case Enum:
		/*TODO
		6.2.7 Compatible type and composite type

		1 Two types have compatible type if their types are the same.
		Additional rules for determining whether two types are
		compatible are described in 6.7.2 for type specifiers, in 6.7.3
		for type qualifiers, and in 6.7.5 for declarators.46) Moreover,
		two structure, union, or enumerated types declared in separate
		translation units are compatible if their tags and members
		satisfy the following requirements: If one is declared with a
		tag, the other shall be declared with the same tag. If both are
		complete types, then the following additional requirements
		apply: there shall be a one-to-one correspondence between their
		members such that each pair of corresponding members are
		declared with compatible types, and such that if one member of
		a corresponding pair is declared with a name, the other member
		is declared with the same name. For two structures,
		corresponding members shall be declared in the same order. For
		two structures or unions, corresponding bit-fields shall have
		the same widths. For two enumerations, corresponding members
		shall have the same values.

		*/
		return ms.kind() == Enum
	case TypedefName:
		panic("internal error")
	default:
		return true
	}
}

func (n *ctype) index(d int) int { return len(n.dds) - 1 + d }

func (n *ctype) top(d int) *DirectDeclarator {
	return n.dds[n.index(d)]
}

// AlignOf implements Type.
func (n *ctype) AlignOf() int {
	if n == undefined {
		return 1
	}

	if n.Kind() == Array {
		return n.Element().AlignOf()
	}

	switch k := n.Kind(); k {
	case
		Void,
		Ptr,
		Char,
		SChar,
		UChar,
		Short,
		UShort,
		Int,
		UInt,
		Long,
		ULong,
		LongLong,
		ULongLong,
		Float,
		Double,
		LongDouble,
		Bool,
		FloatComplex,
		DoubleComplex,
		LongDoubleComplex:
		return n.model.Items[k].Align
	case Enum:
		return n.model.Items[Int].Align
	case Struct, Union:
		switch sus := n.structOrUnionSpecifier(); sus.Case {
		case 1: // StructOrUnion IDENTIFIER
			return -1 // Incomplete type
		case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
			return sus.alignOf
		default:
			panic(sus.Case)
		}
	default:
		panic(k.String())
	}
}

func (n *ctype) unionCanAssignTo(dst Type) bool {
	m, isIncomplete := n.Members()
	if isIncomplete {
		return false
	}

	for _, v := range m {
		if v.Type.CanAssignTo(dst) {
			return true
		}
	}

	return false
}

// CanAssignTo implements Type.
func (n *ctype) CanAssignTo(dst Type) bool {
	if n == undefined || dst.Kind() == Undefined {
		return false
	}

	if n.Kind() == Bool && IsIntType(dst) {
		return true
	}

	if dst.Kind() == Bool && IsIntType(n) {
		return true
	}

	if n.Kind() == Union && n.unionCanAssignTo(dst) {
		return true
	}

	if dst.Kind() == Union && dst.(*ctype).unionCanAssignTo(n) {
		return true
	}

	if n.Kind() == Function {
		n = n.Pointer().(*ctype)
	}

	if dst.Kind() == Function {
		dst = dst.Pointer().(*ctype)
	}

	if n.Kind() == Array && dst.Kind() == Ptr {
		n = n.arrayDecay()
	}

	if dst.Kind() == Array && n.Kind() == Ptr {
		dst = dst.(*ctype).arrayDecay()
	}

	if IsArithmeticType(n) && IsArithmeticType(dst) {
		return true
	}

	if IsIntType(n) && dst.Kind() == Enum {
		return true
	}

	if n.Kind() == Enum && IsIntType(dst) {
		return true
	}

	if n.Kind() == Ptr && dst.Kind() == Ptr && dst.Element().Kind() == Void {
		return true
	}

	if n.Kind() == Ptr && n.Element().Kind() == Void && dst.Kind() == Ptr {
		return true
	}

	if n.isCompatible(dst.(*ctype)) {
		return true
	}

	if n.Kind() == Ptr && dst.Kind() == Ptr {
		t := Type(n)
		u := dst
		for t.Kind() == Ptr && u.Kind() == Ptr {
			t = t.Element()
			u = u.Element()
		}
		if t.Kind() == Array && unsigned(t.Element().Kind()) == unsigned(u.Kind()) {
			return true
		}

		if t.Kind() == Ptr || u.Kind() == Ptr {
			return false
		}

		if IsIntType(t) && IsIntType(u) && unsigned(t.Kind()) == unsigned(u.Kind()) {
			return true
		}

		if t.Kind() == Function && u.Kind() == Function {
			a, _ := t.Parameters()
			b, _ := u.Parameters()
			if (len(a) == 0) != (len(b) == 0) {
				a := t.Result()
				b := u.Result()
				return a.Kind() == Void && b.Kind() == Void || t.Result().CanAssignTo(u.Result())
			}
		}

		return t.(*ctype).isCompatible(u.(*ctype))
	}

	if n.Kind() == Function && dst.Kind() == Ptr && dst.Element().Kind() == Function {
		return n.isCompatible(dst.Element().(*ctype))
	}

	if dst.Kind() == Ptr {
		if IsIntType(n) {
			return true
		}
	}

	return false
}

// RawDeclarator implements Type.
func (n *ctype) RawDeclarator() *Declarator {
	if len(n.dds0) == 0 {
		return n.dds[0].TopDeclarator()
	}

	return n.dds0[0].TopDeclarator()
}

// Declarator implements Type.
func (n *ctype) Declarator() *Declarator {
	if len(n.dds) == 0 {
		panic("internal error")
	}

	return n.dds[0].TopDeclarator()
}

// Element implements Type.
func (n *ctype) Element() Type {
	if n == undefined {
		return n
	}

	if n.Kind() != Ptr && n.Kind() != Array {
		return undefined
	}

	if len(n.dds) == 1 {
		m := *n
		m.stars--
		return &m
	}

	switch dd := n.dds[1]; dd.Case {
	case 1: // '(' Declarator ')'
		if n.stars == 1 {
			m := *n
			m.dds = append([]*DirectDeclarator{n.dds[0]}, n.dds[2:]...)
			m.dds0 = n.dds0
			switch len(m.dds0) {
			case 0:
				// nop
			case 1:
				nm := m.Declarator().RawSpecifier().TypedefName()
				typedef := m.Declarator().DirectDeclarator.idScope.Lookup(NSIdentifiers, nm)
				if typedef.Node == nil {
					break // undefined
				}

				m.dds0 = typedef.Node.(*DirectDeclarator).TopDeclarator().Type.(*ctype).dds0
				if len(m.dds0) < 3 {
					break
				}

				fallthrough
			default:
				m.dds0 = append([]*DirectDeclarator{m.dds0[0]}, m.dds0[2:]...)
			}
			m.stars--
			return &m
		}

		m := *n
		m.stars--
		return &m
	case 2: // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
		m := *n
		m.dds = append([]*DirectDeclarator{n.dds[0]}, n.dds[2:]...)
		switch {
		case len(m.dds) == 1:
			m.stars += m.resultStars
			m.resultStars = 0
		default:
			if dd := m.dds[1]; dd.Case == 1 { // '(' Declarator ')'
				m.stars = dd.Declarator.stars()
				if dd.Declarator.stars() == 0 {
					m.dds = append([]*DirectDeclarator{m.dds[0]}, m.dds[2:]...)
				}
			}
		}
		return &m
	default:
		//dbg("", position(n.dds[0].Pos()), n, n.Kind())
		//dbg("", n.str())
		panic(dd.Case)
	}
}

// Kind implements Type.
func (n *ctype) Kind() Kind {
	if n == undefined {
		return Undefined
	}

	if n.stars > 0 {
		return Ptr
	}

	if len(n.dds) == 1 {
		return n.resultSpecifier.kind()
	}

	i := 1
	for {
		switch dd := n.dds[i]; dd.Case {
		//TODO case 1: // '(' Declarator ')'
		case 2: // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			if dd.elements < 0 {
				return Ptr
			}

			return Array
		case
			6, // DirectDeclarator '(' ParameterTypeList ')'
			7: // DirectDeclarator '(' IdentifierListOpt ')'
			return Function
		default:
			//dbg("", position(n.Declarator().Pos()))
			//dbg("", n)
			//dbg("", n.str())
			panic(dd.Case)
		}
	}
}

// Member implements Type.
func (n *ctype) Member(nm int) (*Member, error) {
	if n == undefined {
		return nil, fmt.Errorf("not a struct/union (have '%s')", n)
	}

	if n.Kind() == Array {
		panic("TODO")
	}

	if k := n.Kind(); k != Struct && k != Union {
		return nil, fmt.Errorf("request for member %s in something not a structure or union (have '%s')", xc.Dict.S(nm), n)
	}

	a, _ := n.Members()
	for i := range a {
		if a[i].Name == nm {
			return &a[i], nil
		}
	}

	return nil, fmt.Errorf("%s has no member named %s", Type(n), xc.Dict.S(nm))
}

// Returns nil if type kind != Enum
func (n *ctype) enumSpecifier() *EnumSpecifier {
	return n.resultSpecifier.firstTypeSpecifier().EnumSpecifier
}

func (n *ctype) structOrUnionSpecifier() *StructOrUnionSpecifier {
	if k := n.Kind(); k != Struct && k != Union {
		return nil
	}

	ts := n.resultSpecifier.firstTypeSpecifier()
	if ts.Case != 11 { // StructOrUnionSpecifier
		panic("internal error")
	}

	switch sus := ts.StructOrUnionSpecifier; sus.Case {
	case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
		return sus
	case 1: // StructOrUnion IDENTIFIER
		b := sus.scope.Lookup(NSTags, sus.Token.Val)
		switch x := b.Node.(type) {
		case nil:
			return sus
		case *StructOrUnionSpecifier:
			return x
		case xc.Token:
			return sus
		default:
			panic("internal error")
		}
	case 2: // StructOrUnion IdentifierOpt '{' '}'                        // Case 2
		return sus
	default:
		panic(sus.Case)
	}
}

func (n *ctype) members(p *[]Member, l *StructDeclarationList) {
	r := *p
	defer func() { *p = r }()

	for ; l != nil; l = l.StructDeclarationList {
		switch sdn := l.StructDeclaration; sdn.Case {
		case 0: // SpecifierQualifierList StructDeclaratorList ';'
			for l := sdn.StructDeclaratorList; l != nil; l = l.StructDeclaratorList {
				var d *Declarator
				var bits int
				switch sd := l.StructDeclarator; sd.Case {
				case 0: // Declarator
					d = sd.Declarator
				case 1: // DeclaratorOpt ':' ConstantExpression
					if o := sd.DeclaratorOpt; o != nil {
						d = o.Declarator
					}
					switch x := sd.ConstantExpression.Value.(type) {
					case int32:
						bits = int(x)
					case int64:
						if x <= int64(n.model.Items[Int].Size*8) {
							bits = int(x)
							break
						}

						panic("internal error")
					case uint64:
						if x <= uint64(n.model.Items[Int].Size*8) {
							bits = int(x)
							break
						}

						panic("internal error")
					default:
						panic("internal error")
					}
				default:
					panic(sd.Case)
				}
				var id, off, pad, bitoff, group int
				t := n.model.IntType
				var bt Type
				if d != nil {
					id, _ = d.Identifier()
					t = d.Type
					off = d.offsetOf
					pad = d.padding
					bitoff = d.bitOffset
					bt = d.bitFieldType
					group = d.bitFieldGroup
				}
				r = append(r, Member{
					BitFieldGroup: group,
					BitFieldType:  bt,
					BitOffsetOf:   bitoff,
					Bits:          bits,
					Declarator:    d,
					Name:          id,
					OffsetOf:      off,
					Padding:       pad,
					Type:          t,
				})
			}
		case 1: // SpecifierQualifierList ';'                       // Case 1
			d := sdn.SpecifierQualifierList.TypeSpecifier.StructOrUnionSpecifier.declarator
			t := d.Type
			r = append(r, Member{
				Declarator: d,
				OffsetOf:   d.offsetOf,
				Padding:    d.padding,
				Type:       t,
			})
		case 2: // StaticAssertDeclaration                          // Case 2
			//nop
		default:
			panic("internal error")
		}
	}
}

// Members implements Type.
func (n *ctype) Members() (r []Member, isIncomplete bool) {
	if k := n.Kind(); k != Struct && k != Union {
		return nil, false
	}

	switch sus := n.structOrUnionSpecifier(); sus.Case {
	case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
		n.members(&r, sus.StructDeclarationList)
		return r, false
	case 1: // StructOrUnion IDENTIFIER
		return []Member{}, true
	case 2: // StructOrUnion IdentifierOpt '{' '}'                        // Case 2
		return []Member{}, false
	default:
		panic(sus.Case)
	}
}

// Parameters implements Type.
func (n *ctype) Parameters() ([]Parameter, bool) {
	if n == undefined || n.Kind() != Function {
		return nil, false
	}

	switch dd := n.dds[1]; dd.Case {
	case 6: // DirectDeclarator '(' ParameterTypeList ')'
		l := dd.ParameterTypeList
		return l.params, l.Case == 1 // ParameterList ',' "..."
	case 7: // DirectDeclarator '(' IdentifierListOpt ')'
		o := dd.IdentifierListOpt
		if o == nil {
			return make([]Parameter, 0), false
		}

		return o.params, false
	default:
		//dbg("", dd.Case)
		panic("internal error")
	}
}

// Pointer implements Type.
func (n *ctype) Pointer() Type {
	if n == undefined {
		return n
	}

	if len(n.dds) == 1 {
		m := *n
		m.stars++
		return &m
	}

	switch dd := n.dds[1]; dd.Case {
	case
		2, // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'        // Case 2
		3, // DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'  // Case 3
		4, // DirectDeclarator '[' TypeQualifierList "static" Expression ']'     // Case 4
		5, // DirectDeclarator '[' TypeQualifierListOpt '*' ']'                  // Case 5
		6, // DirectDeclarator '(' ParameterTypeList ')'
		7: // DirectDeclarator '(' IdentifierListOpt ')'
		dd := &DirectDeclarator{
			Case: 1, // '(' Declarator ')'
			Declarator: &Declarator{
				DirectDeclarator: &DirectDeclarator{},
				PointerOpt: &PointerOpt{
					Pointer: &Pointer{},
				},
			},
		}
		m := *n
		m.dds = append(append([]*DirectDeclarator{n.dds[0]}, dd), n.dds[1:]...)
		m.stars++
		return &m
	default:
		m := *n
		m.stars++
		return &m
	}
}

// Result implements Type.
func (n *ctype) Result() Type {
	if n == undefined {
		return n
	}

	if n.Kind() != Function {
		//dbg("", n, n.Kind())
		//dbg("", n.str())
		panic("TODO")
	}

	i := 1
	for {
		switch dd := n.dds[i]; dd.Case {
		case
			6, // DirectDeclarator '(' ParameterTypeList ')'
			7: // DirectDeclarator '(' IdentifierListOpt ')'
			if i == len(n.dds)-1 { // Outermost function.
				if i == 1 {
					m := *n
					m.dds = m.dds[:1:1]
					m.stars += m.resultStars
					m.resultStars = 0
					return &m
				}

				//dbg("", n)
				//dbg("", n.str())
				panic("TODO")
			}

			m := *n
			m.dds = append([]*DirectDeclarator{n.dds[0]}, n.dds[i+1:]...)
			if dd := m.dds[1]; dd.Case == 1 { // '(' Declarator ')'
				m.stars = dd.Declarator.stars()
			}
			return &m
		default:
			//dbg("", position(n.dds[0].Pos()), n)
			//dbg("", n.str())
			panic(dd.Case)
		}

	}
}

// Elements implements Type.
func (n *ctype) Elements() int {
	done := false
loop:
	for _, dd := range n.dds {
	more:
		switch dd.Case {
		case 0: // IDENTIFIER
		case 1: // '(' Declarator ')'
			dd = dd.Declarator.DirectDeclarator
			done = true
			goto more
		case
			2, // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			3, // DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'
			4, // DirectDeclarator '[' TypeQualifierList "static" Expression ']'
			5: // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
			return dd.elements
		case 6: // DirectDeclarator '(' ParameterTypeList ')'                         // Case 6
			break loop
		default:
			//dbg("", position(n.dds[0].Pos()), n)
			//dbg("", n.str())
			panic(dd.Case)
		}
		if done {
			break
		}
	}
	return -1
}

// EnumeratorList implements Type
func (n *ctype) EnumeratorList() (r []EnumConstant) {
	if n.Kind() != Enum {
		return nil
	}

	switch es := n.enumSpecifier(); es.Case {
	case 0: // "enum" IdentifierOpt '{' EnumeratorList CommaOpt '}'
		for l := es.EnumeratorList; l != nil; l = l.EnumeratorList {
			e := l.Enumerator
			if e.ConstantExpression != nil {
				r = append(r, EnumConstant{
					DefTok: e.EnumerationConstant.Token,
					Value:  e.Value,
					Tokens: e.ConstantExpression.toks})
				continue
			}
			r = append(r, EnumConstant{
				DefTok: e.EnumerationConstant.Token,
				Value:  e.Value,
			})
		}
		return r
	case 1: // "enum" IDENTIFIER
		return nil
	default:
		panic(es.Case)
	}
}

// SizeOf implements Type.
func (n *ctype) SizeOf() int {
	if n == undefined {
		return 1
	}

	if n.Kind() == Array {
		switch nelem := n.Elements(); {
		case nelem < 0:
			return n.model.Items[Ptr].Size
		default:
			return nelem * n.Element().SizeOf()
		}
	}

	switch k := n.Kind(); k {
	case
		Void,
		Ptr,
		Char,
		SChar,
		UChar,
		Short,
		UShort,
		Int,
		UInt,
		Long,
		ULong,
		LongLong,
		ULongLong,
		Float,
		Double,
		LongDouble,
		Bool,
		FloatComplex,
		DoubleComplex,
		LongDoubleComplex:
		return n.model.Items[k].Size
	case Enum:
		return n.model.Items[Int].Size
	case Struct, Union:
		switch sus := n.structOrUnionSpecifier(); sus.Case {
		case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
			return sus.sizeOf
		case 1: // StructOrUnion IDENTIFIER
			return -1 // Incomplete type
		case 2: // StructOrUnion IdentifierOpt '{' '}'                        // Case 2
			return 0
		default:
			panic(sus.Case)
		}
	case Function:
		return n.model.Items[Ptr].Size
	default:
		return -1
	}
}

func (n *ctype) sizeOf(lx *lexer) int {
	r := n.SizeOf()
	if r < 0 {
		lx.report.Err(n.Declarator().Pos(), "cannot determine size of %v", n)
		r = 1
	}
	return r
}

// Specifier implements Type.
func (n *ctype) Specifier() Specifier { return &spec{n.resultAttr, n.resultSpecifier.typeSpecifiers()} }

// String implements Type.
func (n *ctype) String() string {
	if n == undefined {
		return "<undefined>"
	}

	var buf bytes.Buffer
	s := attrString(n.resultAttr)
	buf.WriteString(s)
	if s != "" {
		buf.WriteString(" ")
	}
	s = specifierString(n.resultSpecifier)
	buf.WriteString(s)
	buf.WriteString(strings.Repeat("*", n.resultStars))

	params := func(p []Parameter) {
		for i, v := range p {
			fmt.Fprintf(&buf, "%s", v.Type)
			if i != len(p)-1 {
				buf.WriteByte(',')
			}
		}
	}

	var f func(int)
	starsWritten := false
	f = func(x int) {
		switch dd := n.top(x); dd.Case {
		case 0: // IDENTIFIER
			if debugTypeStrings {
				id := dd.Token.Val
				if id == 0 {
					id = idID
				}
				fmt.Fprintf(&buf, "<%s>", xc.Dict.S(id))
			}
			if !starsWritten {
				buf.WriteString(strings.Repeat("*", n.stars))
			}
		case 1: // '(' Declarator ')'
			buf.WriteString("(")
			s := 0
			switch dd2 := n.top(x - 1); dd2.Case {
			case 0: // IDENTIFIER
				s = n.stars
				starsWritten = true
			default:
				s = dd.Declarator.stars()
			}
			buf.WriteString(strings.Repeat("*", s))
			f(x - 1)
			buf.WriteString(")")
		case 2: // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			f(x - 1)
			buf.WriteString("[")
			sep := ""
			if o := dd.TypeQualifierListOpt; o != nil {
				buf.WriteString(attrString(o.TypeQualifierList.attr))
				sep = " "
			}
			if e := dd.elements; e > 0 {
				buf.WriteString(sep)
				fmt.Fprint(&buf, e)
			}
			buf.WriteString("]")
		case 6: // DirectDeclarator '(' ParameterTypeList ')'
			f(x - 1)
			buf.WriteString("(")
			params(dd.ParameterTypeList.params)
			buf.WriteString(")")
		case 7: // DirectDeclarator '(' IdentifierListOpt ')'
			f(x - 1)
			buf.WriteString("(")
			if o := dd.IdentifierListOpt; o != nil {
				params(o.params)
			}
			buf.WriteString(")")
		default:
			panic(dd.Case)
		}
	}
	f(0)
	return buf.String()
}

// StructAlignOf implements Type.
func (n *ctype) StructAlignOf() int {
	if n == undefined {
		return 1
	}

	if n.Kind() == Array {
		return n.Element().StructAlignOf()
	}

	switch k := n.Kind(); k {
	case
		Void,
		Ptr,
		Char,
		SChar,
		UChar,
		Short,
		UShort,
		Int,
		UInt,
		Long,
		ULong,
		LongLong,
		ULongLong,
		Float,
		Double,
		LongDouble,
		Bool,
		FloatComplex,
		DoubleComplex,
		LongDoubleComplex:
		return n.model.Items[k].StructAlign
	case Enum:
		return n.model.Items[Int].StructAlign
	case Struct, Union:
		switch sus := n.structOrUnionSpecifier(); sus.Case {
		case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
			return sus.alignOf
		case 1: // StructOrUnion IDENTIFIER
			return -1 // Incomplete type
		case 2: // StructOrUnion IdentifierOpt '{' '}'                        // Case 2
			return 1
		default:
			panic(sus.Case)
		}
	default:
		return -1
	}
}

func (n *ctype) structAlignOf(lx *lexer) int {
	r := n.StructAlignOf()
	if r < 0 {
		lx.report.Err(n.Declarator().Pos(), "cannot determine struct align of %v", n)
		r = 1
	}
	return r
}

// Tag implements Type.
func (n *ctype) Tag() int {
	switch k := n.Kind(); k {
	case Struct, Union:
		switch sus := n.structOrUnionSpecifier(); sus.Case {
		case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
			if o := sus.IdentifierOpt; o != nil {
				return o.Token.Val
			}

			return 0
		case 1: // StructOrUnion IDENTIFIER
			return sus.Token.Val
		default:
			panic(sus.Case)
		}
	case Enum:
		es := n.enumSpecifier()
		if es == nil {
			return 0
		}

		switch es.Case {
		case 0: // "enum" IdentifierOpt '{' EnumeratorList CommaOpt '}'
			if o := es.IdentifierOpt; o != nil {
				return o.Token.Val
			}

			return 0
		case 1: // "enum" IDENTIFIER
			return es.Token2.Val
		default:
			panic(es.Case)
		}
	default:
		return 0
	}
}

type spec struct {
	attr int
	ts   int
}

func (s *spec) IsAuto() bool                       { return s.attr&saAuto != 0 }
func (s *spec) IsConst() bool                      { return s.attr&saConst != 0 }
func (s *spec) IsExtern() bool                     { return s.attr&saExtern != 0 }
func (s *spec) IsInline() bool                     { return s.attr&saInline != 0 }
func (s *spec) IsRegister() bool                   { return s.attr&saRegister != 0 }
func (s *spec) IsRestrict() bool                   { return s.attr&saRestrict != 0 }
func (s *spec) IsStatic() bool                     { return s.attr&saStatic != 0 }
func (s *spec) IsTypedef() bool                    { return s.attr&saTypedef != 0 }
func (s *spec) IsVolatile() bool                   { return s.attr&saVolatile != 0 }
func (s *spec) TypedefName() int                   { return 0 }
func (s *spec) attrs() int                         { return s.attr }
func (s *spec) firstTypeSpecifier() *TypeSpecifier { panic("TODO") }
func (s *spec) kind() Kind                         { return tsValid[s.ts] }
func (s *spec) member(int) (*Member, error)        { panic("TODO") }
func (s *spec) str() string                        { return specifierString(s) }
func (s *spec) typeSpecifiers() int                { return s.ts }

func specifierString(sp Specifier) string {
	if sp == nil {
		return ""
	}

	var buf bytes.Buffer
	switch k := sp.kind(); k {
	case Enum:
		switch ts := sp.firstTypeSpecifier(); ts.Case {
		case 12: // EnumSpecifier
			es := ts.EnumSpecifier
			switch es.Case {
			case 0: // "enum" IdentifierOpt '{' EnumeratorList CommaOpt '}'
				buf.WriteString("enum")
				if o := es.IdentifierOpt; o != nil {
					buf.WriteString(" " + string(xc.Dict.S(o.Token.Val)))
				}
				buf.WriteString(" { ... }")
			case 1: // "enum" IDENTIFIER
				fmt.Fprintf(&buf, "enum %s", xc.Dict.S(es.Token2.Val))
			default:
				panic(es.Case)
			}
		default:
			panic(ts.Case)
		}
	case Function:
		panic("TODO Function")
	case Struct, Union:
		switch ts := sp.firstTypeSpecifier(); ts.Case {
		case 11: // StructOrUnionSpecifier
			sus := ts.StructOrUnionSpecifier
			buf.WriteString(sus.StructOrUnion.str())
			switch sus.Case {
			case 0: // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
				if o := sus.IdentifierOpt; o != nil {
					buf.WriteString(" ")
					buf.Write(o.Token.S())
					break
				}

				buf.WriteString("{")
				outerFirst := true
				for l := sus.StructDeclarationList; l != nil; l = l.StructDeclarationList {
					if !outerFirst {
						buf.WriteString("; ")
					}
					outerFirst = false
					first := true
					for l := l.StructDeclaration.StructDeclaratorList; l != nil; l = l.StructDeclaratorList {
						if !first {
							buf.WriteString(", ")
						}
						first = false
						switch sd := l.StructDeclarator; sd.Case {
						case 0: // Declarator
							buf.WriteString(sd.Declarator.Type.String())
						case 1: // DeclaratorOpt ':' ConstantExpression
							if o := sd.DeclaratorOpt; o != nil {
								buf.WriteString(o.Declarator.Type.String())
							}
							buf.WriteByte(':')
							fmt.Fprintf(&buf, "%v", sd.ConstantExpression.Value)
						default:
							fmt.Fprintf(&buf, "specifierString_TODO%v", sd.Case)
						}
					}
				}
				buf.WriteString(";}")
			case 1: // StructOrUnion IDENTIFIER
				buf.WriteString(" ")
				buf.Write(sus.Token.S())
			case 2: // StructOrUnion IdentifierOpt '{' '}'                        // Case 2
				if o := sus.IdentifierOpt; o != nil {
					buf.WriteString(" ")
					buf.Write(o.Token.S())
				}
				buf.WriteString("{}")
			default:
				panic(sus.Case)
			}
		default:
			panic(ts.Case)
		}
	default:
		buf.WriteString(k.CString())
	}
	return buf.String()
}

func align(off, algn int) int {
	r := off % algn
	if r != 0 {
		off += algn - r
	}
	return off
}

func finishBitField(n Node, lx *lexer) {
	sc := lx.scope
	maxLLBits := lx.model.LongLongType.SizeOf() * 8
	bits := sc.bitOffset
	if bits > maxLLBits || bits == 0 {
		panic(fmt.Errorf("%s: internal error %v", position(n.Pos()), bits)) //TODO split group.
	}

	var bytes, al int
	for _, k := range []Kind{Char, Short, Int, Long, LongLong} {
		bytes = lx.model.Items[k].Size
		al = lx.model.Items[k].StructAlign
		if bytes*8 >= bits {
			var t Type
			switch k {
			case Char:
				t = lx.model.CharType
			case Short:
				t = lx.model.ShortType
			case Int:
				t = lx.model.IntType
			case Long:
				t = lx.model.LongType
			case LongLong:
				t = lx.model.LongLongType
			default:
				panic("internal error")
			}
			sc.bitFieldTypes = append(sc.bitFieldTypes, t)
			break
		}
	}
	switch {
	case sc.isUnion:
		off := 0
		sc.offset = align(sc.offset, al)
		if pd := sc.prevStructDeclarator; pd != nil {
			pd.padding = sc.offset - off
			pd.offsetOf = sc.offset
		}
		sc.bitOffset = 0
		sc.bitFieldGroup++
	default:
		off := sc.offset
		sc.offset = align(sc.offset, al)
		if pd := sc.prevStructDeclarator; pd != nil {
			pd.padding = sc.offset - off
			pd.offsetOf = sc.offset
		}
		sc.offset += bytes
		sc.bitOffset = 0
		sc.bitFieldGroup++
	}
	sc.maxAlign = mathutil.Max(sc.maxAlign, al)
}

// IsArithmeticType reports wheter t.Kind() is one of UintPtr, Char, SChar,
// UChar, Short, UShort, Int, UInt, Long, ULong, LongLong, ULongLong, Float,
// Double, LongDouble, FloatComplex, DoubleComplex, LongDoubleComplex, Bool or
// Enum.
func IsArithmeticType(t Type) bool {
	switch t.Kind() {
	case
		UintPtr,
		Char,
		SChar,
		UChar,
		Short,
		UShort,
		Int,
		UInt,
		Long,
		ULong,
		LongLong,
		ULongLong,
		Float,
		Double,
		LongDouble,
		FloatComplex,
		DoubleComplex,
		LongDoubleComplex,
		Bool,
		Enum:
		return true
	default:
		return false
	}
}

// IsIntType reports t.Kind() is one of Char, SChar, UChar, Short, UShort, Int,
// UInt, Long, ULong, LongLong, ULongLong, Bool or Enum.
func IsIntType(t Type) bool {
	switch t.Kind() {
	case
		Char,
		SChar,
		UChar,
		Short,
		UShort,
		Int,
		UInt,
		Long,
		ULong,
		LongLong,
		ULongLong,
		// [0], 6.2.5/6: The type _Bool and the unsigned integer types
		// that correspond to the standard signed integer types are the
		// standard unsigned integer types.
		Bool,
		Enum:
		return true
	default:
		return false
	}
}

func elements(v interface{}, t Type) (int, error) {
	if !IsIntType(t) {
		return -1, fmt.Errorf("expression shall have integer type")
	}

	if v == nil {
		return -1, nil
	}

	r, err := toInt(v)
	if err != nil {
		return -1, err
	}

	if r < 0 {
		return -1, fmt.Errorf("array size must be positive: %v", v)
	}

	return r, nil
}

func toInt(v interface{}) (int, error) {
	switch x := v.(type) {
	case int8:
		return int(x), nil
	case byte:
		return int(x), nil
	case int16:
		return int(x), nil
	case uint16:
		return int(x), nil
	case int32:
		return int(x), nil
	case uint32:
		return int(x), nil
	case int64:
		if x < mathutil.MinInt || x > mathutil.MaxInt {
			return 0, fmt.Errorf("value out of bounds: %v", x)
		}

		return int(x), nil
	case uint64:
		if x > mathutil.MaxInt {
			return 0, fmt.Errorf("value out of bounds: %v", x)
		}

		return int(x), nil
	case int:
		return x, nil
	default:
		return -1, fmt.Errorf("not a constant integer expression: %v", x)
	}
}

func dedupAbsPaths(a []string) (r []string, _ error) {
	m := map[string]struct{}{}
	for _, v := range a {
		av, err := filepath.Abs(v)
		if err != nil {
			return nil, err
		}

		if _, ok := m[av]; ok {
			continue
		}

		r = append(r, v)
		m[v] = struct{}{}
	}
	return r, nil
}

func isCompatibleParameters(a, b []Parameter, va, vb bool) bool {
	if len(a) != len(b) || va != vb {
		return false
	}

	for i, v := range a {
		if !v.Type.CanAssignTo(b[i].Type) {
			return false
		}
	}

	return true
}

// [0], 6.2.7-3
//
// A composite type can be constructed from two types that are compatible; it
// is a type that is compatible with both of the two types and satisfies the
// following conditions:
//
// — If one type is an array of known constant size, the composite type is an
// array of that size; otherwise, if one type is a variable length array, the
// composite type is that type.
//
// — If only one type is a function type with a parameter type list (a function
// prototype), the composite type is a function prototype with the parameter
// type list.
//
// — If both types are function types with parameter type lists, the type of
// each parameter in the composite parameter type list is the composite type of
// the corresponding parameters.
//
// These rules apply recursively to the types from which the two types are
// derived.
func compositeType(a, b Type) (c Type, isA bool) {
	t, u := a, b
	for t.Kind() == Ptr && u.Kind() == Ptr {
		t = t.Element()
		u = u.Element()
	}

	if t.Kind() == Function && u.Kind() == Function {
		if !t.Result().CanAssignTo(u.Result()) {
			return nil, false
		}

		p, va := t.Parameters()
		q, vb := u.Parameters()
		if va != vb {
			return nil, false
		}

		if len(p) == 0 && len(q) != 0 {
			return b, false
		}

		if len(p) != 0 && len(q) == 0 {
			return a, true
		}

		if len(p) != len(q) {
			return nil, false
		}

		for i, v := range p {
			w := q[i]
			if v.Type != undefined && w.Type == undefined || v.Type.CanAssignTo(w.Type) {
				continue
			}

			return nil, false
		}

		return a, true
	}

	return nil, false
}

func eqTypes(a, b Type) bool { return a.(*ctype).eq(b.(*ctype)) }

func isStrLitID(v interface{}) bool {
	switch v.(type) {
	case StringLitID, LongStringLitID:
		return true
	}

	return false
}

func nElem(t Type) int {
	p := -1
	for {
		n := t.Elements()
		if n < 0 {
			return p
		}

		if p < 0 {
			p = 1
		}
		p *= n
		t = t.Element()
	}
}

func unsigned(k Kind) Kind {
	switch k {
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

func isEnum(tn ...*TypeName) bool {
	for _, tn := range tn {
		t := tn.Type
		if t.Kind() == Enum {
			return true
		}

		ts := tn.SpecifierQualifierList.TypeSpecifier
		if ts == nil {
			continue
		}

		switch ts.Case {
		case 15: // "typeof" '(' TypeName ')'    // Case 15
			nm := ts.TypeName.SpecifierQualifierList.TypedefName()
			if nm == 0 {
				break
			}

			n := ts.TypeName.scope.Lookup(NSIdentifiers, nm)
			switch x := n.Node.(type) {
			case *DirectDeclarator:
				if x.specifier.kind() == Enum {
					return true
				}
			}
		}
	}
	return false
}

func memberOffsetRecursive(t Type, name int) (offset int, ty *Type, err error) {
	members, incomplete := t.Members()
	if incomplete {
		return 0, nil, fmt.Errorf("memberOffsetRecursive: incomplete")
	}
	matches := 0
	for _, member := range members {
		if member.Name == name {
			matches++
			offset = member.OffsetOf
			ty = &member.Type
		}
		if member.Name == 0 {
			moffset, mty, err := memberOffsetRecursive(member.Type, name)
			if err == nil {
				matches++
				offset += member.OffsetOf + moffset
				ty = mty
			}
		}
	}
	if matches > 1 {
		return 0, nil, fmt.Errorf("memberOffsetRecursive: ambigous member %s", string(dict.S(name)))
	}
	if matches == 0 {
		return 0, nil, fmt.Errorf("memberOffsetRecursive: non-existent member %s", string(dict.S(name)))
	}
	return offset, ty, err
}

func comment(tw *tweaks, p ...Node) int {
	for _, v := range p {
		v := v.Pos()
		if n := tw.comments[v]; n != 0 {
			return n
		}

		v -= token.Pos(xc.FileSet.Position(v).Column - 1)
		if n := tw.comments[v]; n != 0 {
			return n
		}
	}
	return 0

}

func fixParams(in []Parameter) {
	for i, v := range in {
		if t := v.Type; t.Kind() == Function {
			in[i].Type = t.Pointer()
		}
	}
}

func clean(paths []string) (r []string) {
	for _, v := range paths {
		a, err := filepath.Abs(v)
		if err != nil {
			a = v
		}
		r = append(r, a)
	}
	return r
}
