%{
// Copyright 2016 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on [0], 6.10. Substantial portions of expression AST size
// optimizations are from [2], license of which follows.

// ----------------------------------------------------------------------------

// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This grammar is derived from the C grammar in the 'ansitize'
// program, which carried this notice:
// 
// Copyright (c) 2006 Russ Cox, 
// 	Massachusetts Institute of Technology
// 
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the
// Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute,
// sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
// 
// The above copyright notice and this permission notice shall
// be included in all copies or substantial portions of the
// Software.
// 
// The software is provided "as is", without warranty of any
// kind, express or implied, including but not limited to the
// warranties of merchantability, fitness for a particular
// purpose and noninfringement.  In no event shall the authors
// or copyright holders be liable for any claim, damages or
// other liability, whether in an action of contract, tort or
// otherwise, arising from, out of or in connection with the
// software or the use or other dealings in the software.

package cc

import (
	"fmt"

	"modernc.org/xc"
	"modernc.org/golex/lex"
)
%}

%union {
	Token			xc.Token
	groupPart		Node
	node			Node
	toks			PPTokenList
}

%token
	/*yy:token "'%c'"            */ CHARCONST		"character constant"
	/*yy:token "1.%d"            */ FLOATCONST		"floating-point constant"
	/*yy:token "%c"              */ IDENTIFIER		"identifier"
	/*yy:token "%c"              */ IDENTIFIER_NONREPL	"non replaceable identifier"
	/*yy:token "%c("             */ IDENTIFIER_LPAREN	"identifier immediatelly followed by '('"
	/*yy:token "%d"              */ INTCONST		"integer constant"
	/*yy:token "L'%c'"           */ LONGCHARCONST		"long character constant"
	/*yy:token "L\"%c\""         */ LONGSTRINGLITERAL	"long string constant"
	/*yy:token "<%c.h>"          */ PPHEADER_NAME		"header name"
	/*yy:token "%d"              */ PPNUMBER		"preprocessing number"
	/*yy:token "\"%c\""          */ STRINGLITERAL		"string literal"

	/*yy:token "\U00100000"      */	PREPROCESSING_FILE	1048576	"preprocessing file prefix"	// 0x100000 = 1048576
	/*yy:token "\U00100001"      */	CONSTANT_EXPRESSION	1048577	"constant expression prefix"
	/*yy:token "\U00100002"      */	TRANSLATION_UNIT	1048578	"translation unit prefix"

	/*yy:token "\n#define"       */	PPDEFINE		"#define"
	/*yy:token "\n#elif"         */	PPELIF			"#elif"
	/*yy:token "\n#else"         */	PPELSE			"#else"
	/*yy:token "\n#endif"        */	PPENDIF			"#endif"
	/*yy:token "\n#error"        */	PPERROR			"#error"
	/*yy:token "\n#"             */	PPHASH_NL		"#"
	/*yy:token "\n#if"           */	PPIF			"#if"
	/*yy:token "\n#ifdef"        */	PPIFDEF			"#ifdef"
	/*yy:token "\n#ifndef"       */	PPIFNDEF		"#ifndef"
	/*yy:token "\n#include"      */	PPINCLUDE		"#include"
	/*yy:token "\n#include_next" */	PPINCLUDE_NEXT		"#include_next"
	/*yy:token "\n#line"         */	PPLINE			"#line"
	/*yy:token "\n#foo"          */	PPNONDIRECTIVE		"#foo"
	/*yy:token "other_%c"        */ PPOTHER			"ppother"
	/*yy:token "\n##"            */	PPPASTE			"##"
	/*yy:token "\n#pragma"       */	PPPRAGMA		"#pragma"
	/*yy:token "\n#undef"        */	PPUNDEF			"#undef"

	ADDASSIGN			"+="
	ALIGNOF				"_Alignof"
	ANDAND				"&&"
	ANDASSIGN			"&="
	ARROW				"->"
	ASM				"asm"
	AUTO				"auto"
	BOOL				"_Bool"
	BREAK				"break"
	CASE				"case"
	CHAR				"char"
	COMPLEX				"_Complex"
	CONST				"const"
	CONTINUE			"continue"
	DDD				"..."
	DEC				"--"
	DEFAULT				"default"
	DIVASSIGN			"/="
	DO				"do"
	DOUBLE				"double"
	ELSE				"else"
	ENUM				"enum"
	EQ				"=="
	EXTERN				"extern"
	FLOAT				"float"
	FOR				"for"
	GEQ				">="
	GOTO				"goto"
	IF				"if"
	INC				"++"
	INLINE				"inline"
	INT				"int"
	LEQ				"<="
	LONG				"long"
	LSH				"<<"
	LSHASSIGN			"<<="
	MODASSIGN			"%="
	MULASSIGN			"*="
	NEQ				"!="
	NORETURN			"_Noreturn"
	ORASSIGN			"|="
	OROR				"||"
	REGISTER			"register"
	RESTRICT			"restrict"
	RETURN				"return"
	RSH				">>"
	RSHASSIGN			">>="
	SHORT				"short"
	SIGNED				"signed"
	SIZEOF				"sizeof"
	STATIC				"static"
	STATIC_ASSERT			"_Static_assert"
	STRUCT				"struct"
	SUBASSIGN			"-="
	SWITCH				"switch"
	TYPEDEF				"typedef"
	TYPEDEFNAME			"typedefname"
	TYPEOF				"typeof"
	UNION				"union"
	UNSIGNED			"unsigned"
	VOID				"void"
	VOLATILE			"volatile"
	WHILE				"while"
	XORASSIGN			"^="

%type	<toks>
	PPTokenList			"token list"
	PPTokenListOpt			"optional token list"
	ReplacementList			"replacement list"
	TextLine			"text line"

%type	<groupPart>
	GroupPart			"group part"

%type	<node>
	AbstractDeclarator		"abstract declarator"
	AbstractDeclaratorOpt		"optional abstract declarator"
	ArgumentExpressionList		"argument expression list"
	ArgumentExpressionListOpt	"optional argument expression list"
	AssemblerInstructions		"assembler instructions"
	AssemblerOperand		"assembler operand"
	AssemblerOperands		"assembler operands"
	AssemblerStatement		"assembler statement"
	AssemblerSymbolicNameOpt	"optional assembler symbolic name"
	BasicAssemblerStatement		"basic assembler statement"
	BlockItem			"block item"
	BlockItemList			"block item list"
	BlockItemListOpt		"optional block item list"
	Clobbers			"clobbers"
	CommaOpt			"optional comma"
	CompoundStatement		"compound statement"
	ConstantExpression		"constant expression"
	ControlLine			"control line"
	Declaration			"declaration"
	DeclarationList			"declaration list"
	DeclarationListOpt		"optional declaration list"
	DeclarationSpecifiers		"declaration specifiers"
	DeclarationSpecifiersOpt	"optional declaration specifiers"
	Declarator			"declarator"
	DeclaratorOpt			"optional declarator"
	Designation			"designation"
	DesignationOpt			"optional designation"
	Designator			"designator"
	DesignatorList			"designator list"
	DirectAbstractDeclarator	"direct abstract declarator"
	DirectAbstractDeclaratorOpt	"optional direct abstract declarator"
	DirectDeclarator		"direct declarator"
	ElifGroup			"elif group"
	ElifGroupList			"elif group list"
	ElifGroupListOpt		"optional elif group list"
	ElseGroup			"else group"
	ElseGroupOpt			"optional else group"
	EndifLine			"endif line"
	EnumSpecifier			"enum specifier"
	EnumerationConstant		"enumearation constant"
	Enumerator			"enumerator"
	EnumeratorList			"enumerator list"
	Expression			"expression"
	ExpressionList			"expression list"
	ExpressionListOpt		"optional expression list"
	ExpressionOpt			"optional expression"
	ExpressionStatement		"expression statement"
	ExternalDeclaration		"external declaration"
	FunctionDefinition		"function definition"
	FunctionBody			"function body"
	FunctionSpecifier		"function specifier"
	GroupList			"group list"
	GroupListOpt			"optional group list"
	IdentifierList			"identifier list"
	IdentifierListOpt		"optional identifier list"
	IdentifierOpt			"optional identifier"
	IfGroup				"if group"
	IfSection			"if section"
	InitDeclarator			"init declarator"
	InitDeclaratorList		"init declarator list"
	InitDeclaratorListOpt		"optional init declarator list"
	Initializer			"initializer"
	InitializerList			"initializer list"
	IterationStatement		"iteration statement"
	JumpStatement			"jump statement"
	LabeledStatement		"labeled statement"
	ParameterDeclaration		"parameter declaration"
	ParameterList			"parameter list"
	ParameterTypeList		"parameter type list"
	ParameterTypeListOpt		"optional parameter type list"
	Pointer				"pointer"
	PointerOpt			"optional pointer"
	PreprocessingFile		"preprocessing file"
	SelectionStatement		"selection statement"
	SpecifierQualifierList		"specifier qualifier list"
	SpecifierQualifierListOpt	"optional specifier qualifier list"
	Statement			"statement"
	StaticAssertDeclaration		"static assert declaration"
	StorageClassSpecifier		"storage class specifier"
	StructDeclaration		"struct declaration"
	StructDeclarationList		"struct declaration list"
	StructDeclarator		"struct declarator"
	StructDeclaratorList		"struct declarator list"
	StructOrUnion			"struct-or-union"
	StructOrUnionSpecifier		"struct-or-union specifier"
	TranslationUnit			"translation unit"
	TypeName			"type name"
	TypeQualifier			"type qualifier"
	TypeQualifierList		"type qualifier list"
	TypeQualifierListOpt		"optional type qualifier list"
	TypeSpecifier			"type specifier"
	VolatileOpt			"optional volatile"


%precedence	NOSEMI
%precedence	';'

%precedence	NOELSE
%precedence	ELSE

%right		'=' ADDASSIGN ANDASSIGN DIVASSIGN LSHASSIGN MODASSIGN MULASSIGN
		ORASSIGN RSHASSIGN SUBASSIGN XORASSIGN
		
%right		':' '?'
%left		OROR
%left		ANDAND
%left		'|'
%left		'^'
%left		'&'
%left		EQ NEQ
%left		'<' '>' GEQ LEQ
%left		LSH RSH
%left		'+' '-' 
%left		'%' '*' '/'
%precedence	CAST
%left		'!' '~' SIZEOF UNARY
%right		'(' '.' '[' ARROW DEC INC

%%

//yy:ignore
Start:
	PREPROCESSING_FILE
	{
		lx.preprocessingFile = nil
	}
	PreprocessingFile
	{
		lx.preprocessingFile = $3.(*PreprocessingFile)
	}
|	CONSTANT_EXPRESSION
	{
		lx.constantExpression = nil
	}
	ConstantExpression
	{
		lx.constantExpression = $3.(*ConstantExpression)
	}
|	TRANSLATION_UNIT
	{
		lx.translationUnit = nil
	}
	TranslationUnit
	{
		if lx.report.Errors(false) == nil && lx.scope.kind != ScopeFile {
			panic("internal error")
		}

		lx.translationUnit = $3.(*TranslationUnit).reverse()
		lx.translationUnit.Declarations = lx.scope
	}

// [0](6.4.4.3)
EnumerationConstant:
	IDENTIFIER

// [0](6.5.2)
ArgumentExpressionList:
	Expression
|	ArgumentExpressionList ',' Expression

ArgumentExpressionListOpt:
	/* empty */ {}
|	ArgumentExpressionList

// [0](6.5.16)
//yy:field	BinOpType	Type		// The type operands of binary expression are coerced into, if different from Type.
//yy:field	Type		Type		// Type of expression.
//yy:field	Value		interface{}	// Non nil for certain constant expressions.
//yy:field	scope		*Bindings	// Case 0: IDENTIFIER resolution scope.
Expression:
	IDENTIFIER %prec NOSEMI
	{
		lhs.scope = lx.scope
	}
|	CHARCONST
|	FLOATCONST
|	INTCONST
|	LONGCHARCONST
|	LONGSTRINGLITERAL
|	STRINGLITERAL
|	'(' ExpressionList ')'
|	Expression '[' ExpressionList ']'
|	Expression '(' ArgumentExpressionListOpt ')'
	{
		o := lhs.ArgumentExpressionListOpt
		if o == nil {
			break
		}

		if lhs.Expression.Case == 0 { // IDENTIFIER
			if lx.tweaks.enableBuiltinConstantP &&lhs.Expression.Token.Val == idBuiltinConstantP {
				break
			}

			b := lhs.Expression.scope.Lookup(NSIdentifiers, lhs.Expression.Token.Val)
			if b.Node == nil && lx.tweaks.enableImplicitFuncDef {
				for l := o.ArgumentExpressionList; l != nil; l = l.ArgumentExpressionList {
					l.Expression.eval(lx)
				}
				break
			}
		}

		lhs.Expression.eval(lx)
		for l := o.ArgumentExpressionList; l != nil; l = l.ArgumentExpressionList {
			l.Expression.eval(lx)
		}
	}
|	Expression '.' IDENTIFIER
|	Expression "->" IDENTIFIER
|	Expression "++"
|	Expression "--"
|	'(' TypeName ')' '{' InitializerList CommaOpt '}'
|	"++" Expression
|	"--" Expression
|	'&' Expression %prec UNARY
|	'*' Expression %prec UNARY
|	'+' Expression %prec UNARY
|	'-' Expression %prec UNARY
|	'~' Expression
|	'!' Expression
|	"sizeof" Expression
|	"sizeof" '(' TypeName ')' %prec SIZEOF
|	'(' TypeName ')' Expression %prec CAST
|	Expression '*' Expression
|	Expression '/' Expression
|	Expression '%' Expression
|	Expression '+' Expression
|	Expression '-' Expression
|	Expression "<<" Expression
|	Expression ">>" Expression
|	Expression '<' Expression
|	Expression '>' Expression
|	Expression "<=" Expression
|	Expression ">=" Expression
|	Expression "==" Expression
|	Expression "!=" Expression
|	Expression '&' Expression
|	Expression '^' Expression
|	Expression '|' Expression
|	Expression "&&" Expression
|	Expression "||" Expression
|	Expression '?' ExpressionList ':' Expression
|	Expression '=' Expression
|	Expression "*=" Expression
|	Expression "/=" Expression
|	Expression "%=" Expression
|	Expression "+=" Expression
|	Expression "-=" Expression
|	Expression "<<=" Expression
|	Expression ">>=" Expression
|	Expression "&=" Expression
|	Expression "^=" Expression
|	Expression "|=" Expression
|	"_Alignof" '(' TypeName ')'
|	'(' CompoundStatement ')'
|	"&&" IDENTIFIER
	{
		if !lx.tweaks.enableComputedGotos {
			lx.report.Err(lhs.Pos(), "computed gotos not enabled")
		}
	}
|	Expression '?' ':' Expression
	{
		if !lx.tweaks.enableOmitConditionalOperand {
			lx.report.Err(lhs.Pos(), "omitting conditional operand not enabled")
		}
	}


ExpressionOpt:
	/* empty */ {}
|	Expression
	{
		lhs.Expression.eval(lx)
	}

// [0](6.5.17)
//yy:field	Type	Type		// Type of expression.
//yy:field	Value	interface{}	// Non nil for certain constant expressions.
//yy:list
ExpressionList:
	Expression
|	ExpressionList ',' Expression

ExpressionListOpt:
	/* empty */ {}
|	ExpressionList
	{
		lhs.ExpressionList.eval(lx)
	}

// [0](6.6)
//yy:field	Type	Type		// Type of expression.
//yy:field	Value	interface{}	// Non nil for certain constant expressions.
//yy:field	toks	[]xc.Token	//
ConstantExpression:
	{
		lx.constExprToks = []xc.Token{lx.last}
	}
	Expression
	{
		lhs.Value, lhs.Type = lhs.Expression.eval(lx)
		if lhs.Value == nil {
			lx.report.Err(lhs.Pos(), "not a constant expression")
		}
		l := lx.constExprToks
		lhs.toks = l[:len(l)-1]
		lx.constExprToks = nil
	}

// [0](6.7)
//yy:field	declarator	*Declarator	// Synthetic declarator when InitDeclaratorListOpt is nil.
Declaration:
	DeclarationSpecifiers InitDeclaratorListOpt ';'
	{
		ts0 := lhs.DeclarationSpecifiers.typeSpecifiers()
		if ts0 == 0 && lx.tweaks.enableImplicitIntType {
			lhs.DeclarationSpecifiers.typeSpecifier = tsEncode(tsInt)
		}
		ts := tsDecode(lhs.DeclarationSpecifiers.typeSpecifiers())
		ok := false
		for _, v := range ts {
			if v == tsStructSpecifier || v == tsUnionSpecifier {
				ok = true
				break
			}
		}
		if ok {
			s := lhs.DeclarationSpecifiers
			d := &Declarator{specifier: s}
			dd := &DirectDeclarator{
				Token: xc.Token{Char: lex.NewChar(lhs.Pos(), 0)},
				declarator: d,
				idScope: lx.scope,
				specifier: s,
			}
			d.DirectDeclarator = dd
			d.setFull(lx)
			for l := lhs.DeclarationSpecifiers; l != nil; {
				ts := l.TypeSpecifier
				if ts != nil && ts.Case == 11 && ts.StructOrUnionSpecifier.Case == 0 { // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
					ts.StructOrUnionSpecifier.declarator = d
					break
				}

				if o := l.DeclarationSpecifiersOpt; o != nil {
					l = o.DeclarationSpecifiers
					continue
				}

				break
			}
		}

		o := lhs.InitDeclaratorListOpt
		if o != nil {
			break
		}

		s := lhs.DeclarationSpecifiers
		d := &Declarator{specifier: s}
		dd := &DirectDeclarator{
			Token: xc.Token{Char: lex.NewChar(lhs.Pos(), 0)},
			declarator: d,
			idScope: lx.scope,
			specifier: s,
		}
		d.DirectDeclarator = dd
		d.setFull(lx)
		lhs.declarator = d
	}
|	StaticAssertDeclaration

// [0](6.7)
//yy:field	attr		int	// tsInline, tsTypedefName, ...
//yy:field	typeSpecifier	int	// Encoded combination of tsVoid, tsInt, ...
DeclarationSpecifiers:
	StorageClassSpecifier DeclarationSpecifiersOpt
	{
		lx.scope.specifier = lhs
		a := lhs.StorageClassSpecifier
		b := lhs.DeclarationSpecifiersOpt
		if b == nil {
			lhs.attr = a.attr
			break
		}

		if a.attr&b.attr != 0 {
			lx.report.Err(a.Pos(), "invalid storage class specifier")
			break
		}

		lhs.attr = a.attr|b.attr
		lhs.typeSpecifier = b.typeSpecifier
		if lhs.StorageClassSpecifier.Case != 0 /* "typedef" */ && lhs.IsTypedef() {
			lx.report.Err(a.Pos(), "invalid storage class specifier")
		}
	}
|	TypeSpecifier DeclarationSpecifiersOpt
	{
		lx.scope.specifier = lhs
		a := lhs.TypeSpecifier
		b := lhs.DeclarationSpecifiersOpt
		if b == nil {
			lhs.typeSpecifier = a.typeSpecifier
			break
		}

		lhs.attr = b.attr
		tsb := tsDecode(b.typeSpecifier)
		if len(tsb) == 1 && tsb[0] == tsTypedefName && lx.tweaks.allowCompatibleTypedefRedefinitions {
			tsb[0] = 0
		}
		ts := tsEncode(append(tsDecode(a.typeSpecifier), tsb...)...)
		if _, ok := tsValid[ts]; !ok {
			ts = tsEncode(tsInt)
			lx.report.Err(a.Pos(), "invalid type specifier")
		}
		lhs.typeSpecifier = ts
	}
|	TypeQualifier DeclarationSpecifiersOpt
	{
		lx.scope.specifier = lhs
		a := lhs.TypeQualifier
		b := lhs.DeclarationSpecifiersOpt
		if b == nil {
			lhs.attr = a.attr
			break
		}
	
		if a.attr&b.attr != 0 {
			lx.report.Err(a.Pos(), "invalid type qualifier")
			break
		}

		lhs.attr = a.attr|b.attr
		lhs.typeSpecifier = b.typeSpecifier
		if lhs.IsTypedef() {
			lx.report.Err(a.Pos(), "invalid type qualifier")
		}
	}
|	FunctionSpecifier DeclarationSpecifiersOpt
	{
		lx.scope.specifier = lhs
		a := lhs.FunctionSpecifier
		b := lhs.DeclarationSpecifiersOpt
		if b == nil {
			lhs.attr = a.attr
			break
		}
	
		if a.attr&b.attr != 0 {
			lx.report.Err(a.Pos(), "invalid function specifier")
			break
		}

		lhs.attr = a.attr|b.attr
		lhs.typeSpecifier = b.typeSpecifier
		if lhs.IsTypedef() {
			lx.report.Err(a.Pos(), "invalid function specifier")
		}
	}

//yy:field	attr		int	// tsInline, tsTypedefName, ...
//yy:field	typeSpecifier	int	// Encoded combination of tsVoid, tsInt, ...
DeclarationSpecifiersOpt:
	/* empty */ {}
|	DeclarationSpecifiers
	{
		lhs.attr = lhs.DeclarationSpecifiers.attr
		lhs.typeSpecifier = lhs.DeclarationSpecifiers.typeSpecifier
	}

// [0](6.7)
InitDeclaratorList:
	InitDeclarator
|	InitDeclaratorList ',' InitDeclarator

InitDeclaratorListOpt:
	/* empty */ {}
|	InitDeclaratorList

// [0](6.7)
InitDeclarator:
	Declarator
	{
		lhs.Declarator.setFull(lx)
	}
|	Declarator
	{
		d := $1.(*Declarator)
		d.setFull(lx)
	}
	'=' Initializer
	{
		d := lhs.Declarator
		lhs.Initializer.typeCheck(&d.Type, d.Type, lhs.Declarator.specifier.IsStatic(), lx)
		if d.Type.Specifier().IsExtern() {
			id, _ := d.Identifier()
			lx.report.Err(d.Pos(), "'%s' initialized and declared 'extern'", dict.S(id))
		}
	}

// [0](6.7.1)
//yy:field	attr	int
StorageClassSpecifier:
	"typedef"
	{
		lhs.attr = saTypedef
	}
|	"extern"
	{
		lhs.attr = saExtern
	}
|	"static"
	{
		lhs.attr = saStatic
	}
|	"auto"
	{
		lhs.attr = saAuto
	}
|	"register"
	{
		lhs.attr = saRegister
	}

// [0](6.7.2)
//yy:field	scope		*Bindings	// If case TYPEDEFNAME.
//yy:field	typeSpecifier	int		// Encoded combination of tsVoid, tsInt, ...
//yy:field	Type		Type		// Type of typeof.
TypeSpecifier:
	"void"
	{
		lhs.typeSpecifier = tsEncode(tsVoid)
	}
|	"char"
	{
		lhs.typeSpecifier = tsEncode(tsChar)
	}
|	"short"
	{
		lhs.typeSpecifier = tsEncode(tsShort)
	}
|	"int"
	{
		lhs.typeSpecifier = tsEncode(tsInt)
	}
|	"long"
	{
		lhs.typeSpecifier = tsEncode(tsLong)
	}
|	"float"
	{
		lhs.typeSpecifier = tsEncode(tsFloat)
	}
|	"double"
	{
		lhs.typeSpecifier = tsEncode(tsDouble)
	}
|	"signed"
	{
		lhs.typeSpecifier = tsEncode(tsSigned)
	}
|	"unsigned"
	{
		lhs.typeSpecifier = tsEncode(tsUnsigned)
	}
|	"_Bool"
	{
		lhs.typeSpecifier = tsEncode(tsBool)
	}
|	"_Complex"
	{
		lhs.typeSpecifier = tsEncode(tsComplex)
	}
|	StructOrUnionSpecifier
	{
		lhs.typeSpecifier = tsEncode(lhs.StructOrUnionSpecifier.typeSpecifiers())
	}
|	EnumSpecifier
	{
		lhs.typeSpecifier = tsEncode(tsEnumSpecifier)
	}
/*yy:example "\U00100002 typedef int i; i j;" */
|	TYPEDEFNAME
	{
		lhs.typeSpecifier = tsEncode(tsTypedefName)
		_, lhs.scope = lx.scope.Lookup2(NSIdentifiers, lhs.Token.Val)
	}
|	"typeof" '(' Expression ')'
	{
		lhs.typeSpecifier = tsEncode(tsTypeof)
		_, lhs.Type = lhs.Expression.eval(lx)
	}
|	"typeof" '(' TypeName ')'
	{
		lhs.typeSpecifier = tsEncode(tsTypeof)
		lhs.Type = undefined
		if t := lhs.TypeName.Type; t != nil {
			lhs.Type = t
		}
	}

// [0](6.7.2.1)
//yy:example	"\U00100002 struct { int i; } ("
//yy:field	alignOf	int
//yy:field	declarator	*Declarator	// Synthetic declarator when tagged struct/union defined inline.
//yy:field	scope		*Bindings
//yy:field	sizeOf		int
StructOrUnionSpecifier:
	StructOrUnion IdentifierOpt
	'{'
	{
		if o := $2.(*IdentifierOpt); o != nil {
			lx.scope.declareStructTag(o.Token, lx.report)
		}
		lx.pushScope(ScopeMembers)
		lx.scope.isUnion = $1.(*StructOrUnion).Case == 1 // "union"
		lx.scope.prevStructDeclarator = nil
	}
	StructDeclarationList '}'
	{
		sc := lx.scope
		lhs.scope = sc
		if sc.bitOffset != 0 {
			finishBitField(lhs, lx)
		}

		i := 0
		var bt Type
		var d *Declarator
		for l := lhs.StructDeclarationList; l != nil; l = l.StructDeclarationList {
			for l := l.StructDeclaration.StructDeclaratorList; l != nil; l = l.StructDeclaratorList {
				switch sd := l.StructDeclarator; sd.Case {
				case 0: // Declarator
					d = sd.Declarator
				case 1: // DeclaratorOpt ':' ConstantExpression
					if o := sd.DeclaratorOpt; o != nil {
						x := o.Declarator
						if x.bitOffset == 0  {
							d = x
							bt = lx.scope.bitFieldTypes[i]
							i++
						}
						x.bitFieldType = bt
					}
				}
			}
		}
		lx.scope.bitFieldTypes = nil

		lhs.alignOf = sc.maxAlign
		switch {
		case sc.isUnion:
			lhs.sizeOf = align(sc.maxSize, sc.maxAlign)
		default:
			off := sc.offset
			lhs.sizeOf = align(sc.offset, sc.maxAlign)
			if d != nil {
				d.padding = lhs.sizeOf-off
			}
		}

		lx.popScope(lhs.Token2)
		if o := lhs.IdentifierOpt; o != nil {
			lx.scope.defineStructTag(o.Token, lhs, lx.report)
		}
	}
|	StructOrUnion IDENTIFIER
	{
		lx.scope.declareStructTag(lhs.Token, lx.report)
		lhs.scope = lx.scope
	}
|	StructOrUnion IdentifierOpt '{' '}'
	{
		if !lx.tweaks.enableEmptyStructs {
			lx.report.Err(lhs.Token.Pos(), "empty structs/unions not allowed")
		}
		if o := $2.(*IdentifierOpt); o != nil {
			lx.scope.declareStructTag(o.Token, lx.report)
		}
		lx.scope.isUnion = $1.(*StructOrUnion).Case == 1 // "union"
		lx.scope.prevStructDeclarator = nil
		lhs.alignOf = 1
		lhs.sizeOf = 0
		if o := lhs.IdentifierOpt; o != nil {
			lx.scope.defineStructTag(o.Token, lhs, lx.report)
		}
	}

// [0](6.7.2.1)
StructOrUnion:
	"struct"
|	"union"

// [0](6.7.2.1)
StructDeclarationList:
	StructDeclaration
|	StructDeclarationList StructDeclaration

// [0](6.7.2.1)
StructDeclaration:
	SpecifierQualifierList StructDeclaratorList ';'
	{
		s := lhs.SpecifierQualifierList
		if k := s.kind(); k != Struct && k != Union {
			break
		}

		d := &Declarator{specifier: s}
		dd := &DirectDeclarator{
			Token: xc.Token{Char: lex.NewChar(lhs.Pos(), 0)},
			declarator: d,
			idScope: lx.scope,
			specifier: s,
		}
		d.DirectDeclarator = dd
		d.setFull(lx)
		for l := lhs.SpecifierQualifierList; l != nil; {
			ts := l.TypeSpecifier
			if ts != nil && ts.Case == 11 && ts.StructOrUnionSpecifier.Case == 0 { // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
				ts.StructOrUnionSpecifier.declarator = d
				break
			}

			if o := l.SpecifierQualifierListOpt; o != nil {
				l = o.SpecifierQualifierList
				continue
			}

			break
		}
	}
|	SpecifierQualifierList ';'
	{
		s := lhs.SpecifierQualifierList
		if !lx.tweaks.enableAnonymousStructFields {
			lx.report.Err(lhs.Token.Pos(), "unnamed fields not allowed")
		} else if k := s.kind(); k != Struct && k != Union {
			lx.report.Err(lhs.Token.Pos(), "only unnamed structs and unions are allowed")
			break
		}

		d := &Declarator{specifier: s}
		dd := &DirectDeclarator{
			Token: xc.Token{Char: lex.NewChar(lhs.Pos(), 0)},
			declarator: d,
			idScope: lx.scope,
			specifier: s,
		}
		d.DirectDeclarator = dd
		d.setFull(lx)

		// we have no struct declarators to parse, so we have to create the case of one implicit declarator
		// because else the size of anonymous members is not included in the struct size!
		dummy := &StructDeclarator{Declarator: d}
		dummy.post(lx)

		for l := lhs.SpecifierQualifierList; l != nil; {
			ts := l.TypeSpecifier
			if ts != nil && ts.Case == 11 && ts.StructOrUnionSpecifier.Case == 0 { // StructOrUnion IdentifierOpt '{' StructDeclarationList '}'
				ts.StructOrUnionSpecifier.declarator = d
				break
			}

			if o := l.SpecifierQualifierListOpt; o != nil {
				l = o.SpecifierQualifierList
				continue
			}

			break
		}
	}
|	StaticAssertDeclaration

// [0](6.7.2.1)
//yy:field	attr		int	// tsInline, tsTypedefName, ...
//yy:field	typeSpecifier	int	// Encoded combination of tsVoid, tsInt, ...
SpecifierQualifierList:
	TypeSpecifier SpecifierQualifierListOpt
	{
		lx.scope.specifier = lhs
		a := lhs.TypeSpecifier
		b := lhs.SpecifierQualifierListOpt
		if b == nil {
			lhs.typeSpecifier = a.typeSpecifier
			break
		}

		lhs.attr = b.attr
		ts := tsEncode(append(tsDecode(a.typeSpecifier), tsDecode(b.typeSpecifier)...)...)
		if _, ok := tsValid[ts]; !ok {
			lx.report.Err(a.Pos(), "invalid type specifier")
			break
		}

		lhs.typeSpecifier = ts
	}
|	TypeQualifier SpecifierQualifierListOpt
	{
		lx.scope.specifier = lhs
		a := lhs.TypeQualifier
		b := lhs.SpecifierQualifierListOpt
		if b == nil {
			lhs.attr = a.attr
			break
		}
	
		if a.attr&b.attr != 0 {
			lx.report.Err(a.Pos(), "invalid type qualifier")
			break
		}

		lhs.attr = a.attr|b.attr
		lhs.typeSpecifier = b.typeSpecifier
	}

//yy:field	attr		int	// tsInline, tsTypedefName, ...
//yy:field	typeSpecifier	int	// Encoded combination of tsVoid, tsInt, ...
SpecifierQualifierListOpt:
	/* empty */ {}
|	SpecifierQualifierList
	{
		lhs.attr = lhs.SpecifierQualifierList.attr
		lhs.typeSpecifier = lhs.SpecifierQualifierList.typeSpecifier
	}

// [0](6.7.2.1)
StructDeclaratorList:
	StructDeclarator
|	StructDeclaratorList ',' StructDeclarator

// [0](6.7.2.1)
StructDeclarator:
	Declarator
	{
		lhs.Declarator.setFull(lx)
		lhs.post(lx)
	}
|	DeclaratorOpt ':' ConstantExpression
	{
		m := lx.model
		e := lhs.ConstantExpression
		if e.Value == nil {
			e.Value, e.Type = m.value2(1, m.IntType)
		}
		if !IsIntType(e.Type) {
			lx.report.Err(e.Pos(), "bit field width not an integer (have '%s')", e.Type)
			e.Value, e.Type = m.value2(1, m.IntType)
		}
		if o := lhs.DeclaratorOpt; o != nil {
			o.Declarator.setFull(lx)
		}
		lhs.post(lx)
	}

CommaOpt:
	/* empty */ {}
|	','

// [0](6.7.2.2)
//yy:field	unsigned	bool
EnumSpecifier:
	"enum" IdentifierOpt
	{
		if o := $2.(*IdentifierOpt); o != nil {
			lx.scope.declareEnumTag(o.Token, lx.report)
		}
		lx.iota = 0
	}
	'{' EnumeratorList  CommaOpt '}'
	{
		if o := lhs.IdentifierOpt; o != nil {
			lx.scope.defineEnumTag(o.Token, lhs, lx.report)
		}
		if !lx.tweaks.enableUnsignedEnums {
			break
		}

		lhs.unsigned = true
	loop:
		for l := lhs.EnumeratorList; l != nil; l = l.EnumeratorList {
			switch e := l.Enumerator; x := e.Value.(type) {
			case int32:
				if x < 0 {
					lhs.unsigned = false
					break loop
				}
			case int64:
				if x < 0 {
					lhs.unsigned = false
					break loop
				}
			default:
				panic(fmt.Errorf("%s: TODO Enumerator.Value type %T", position(e.Pos()), x))
			}
		}
	}
|	"enum" IDENTIFIER
	{
		lx.scope.declareEnumTag(lhs.Token2, lx.report)
	}

// [0](6.7.2.2)
EnumeratorList:
	Enumerator
|	EnumeratorList ',' Enumerator

// [0](6.7.2.2)
//yy:field	Value		interface{}	// Enumerator's value.
Enumerator:
	EnumerationConstant
	{
		m := lx.model
		v := m.MustConvert(lx.iota, m.IntType)
		lhs.Value = v
		lx.scope.defineEnumConst(lx, lhs.EnumerationConstant.Token, v)
	}
|	EnumerationConstant '=' ConstantExpression
	{
		m := lx.model
		e := lhs.ConstantExpression
		var v interface{}
		// [0], 6.7.2.2
		// The expression that defines the value of an enumeration
		// constant shall be an integer constant expression that has a
		// value representable as an int.
		switch {
		case !IsIntType(e.Type):
			lx.report.Err(e.Pos(), "not an integer constant expression (have '%s')", e.Type)
			v = m.MustConvert(int32(0), m.IntType)
		default:
			var ok bool
			if v, ok = m.enumValueToInt(e.Value); !ok {
				lx.report.Err(e.Pos(), "overflow in enumeration value: %v", e.Value)
			}
		}

		lhs.Value = v
		lx.scope.defineEnumConst(lx, lhs.EnumerationConstant.Token, v)
	}

// [0](6.7.3)
//yy:field	attr		int	// tsInline, tsTypedefName, ...
TypeQualifier:
	"const"
	{
		lhs.attr = saConst
	}
|	"restrict"
	{
		lhs.attr = saRestrict
	}
|	"volatile"
	{
		lhs.attr = saVolatile
	}

// [0](6.7.4)
//yy:field	attr		int	// tsInline, tsTypedefName, ...
FunctionSpecifier:
	"inline"
	{
		lhs.attr = saInline
	}
|	"_Noreturn"
	{
		lhs.attr = saNoreturn
	}

// [0](6.7.5)
//yy:field	Linkage		Linkage
//yy:field	Type		Type
//yy:field	bitFieldType	Type
//yy:field	bitFieldGroup	int
//yy:field	bitOffset	int
//yy:field	bits		int
//yy:field	offsetOf	int
//yy:field	padding		int
//yy:field	specifier	Specifier
Declarator:
	PointerOpt DirectDeclarator
	{
		lhs.specifier = lx.scope.specifier
		lhs.DirectDeclarator.declarator = lhs
	}

DeclaratorOpt:
	/* empty */ {}
|	Declarator

// [0](6.7.5)
//yy:field	EnumVal		interface{}	// Non nil if DD declares an enumeration constant.
//yy:field	declarator	*Declarator
//yy:field	elements	int
//yy:field	idScope		*Bindings	// Of case 0: IDENTIFIER.
//yy:field	paramsScope	*Bindings
//yy:field	parent		*DirectDeclarator
//yy:field	prev		*Binding	// Existing declaration in same scope, if any.
//yy:field	specifier	Specifier
//yy:field	visible		*Binding	// Existing declaration of same ident visible in same scope, if any and this DD has storage class extrn.
DirectDeclarator:
	IDENTIFIER
	{
		lhs.specifier = lx.scope.specifier
		lx.scope.declareIdentifier(lhs.Token, lhs, lx.report)
		lhs.idScope = lx.scope
	}
|	'(' Declarator ')'
	{
		lhs.Declarator.specifier = nil
		lhs.Declarator.DirectDeclarator.parent = lhs
	}
|	DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
	{
		lhs.elements = -1
		if o := lhs.ExpressionOpt; o != nil {
			var err error
			if lhs.elements, err = elements(o.Expression.eval(lx)); err != nil {
				lx.report.Err(o.Expression.Pos(), "%s", err)
			}
			
		}
		lhs.DirectDeclarator.parent = lhs
	}
|	DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'
	{
		var err error
		if lhs.elements, err = elements(lhs.Expression.eval(lx)); err != nil {
			lx.report.Err(lhs.Expression.Pos(), "%s", err)
		}
		lhs.DirectDeclarator.parent = lhs
	}
|	DirectDeclarator '[' TypeQualifierList "static" Expression ']'
	{
		var err error
		if lhs.elements, err = elements(lhs.Expression.eval(lx)); err != nil {
			lx.report.Err(lhs.Expression.Pos(), "%s", err)
		}
		lhs.DirectDeclarator.parent = lhs
	}
|	DirectDeclarator '[' TypeQualifierListOpt '*' ']'
	{
		lhs.DirectDeclarator.parent = lhs
		lhs.elements = -1
	}
|	DirectDeclarator '('
	{
		lx.pushScope(ScopeParams)
	}
	ParameterTypeList ')'
	{
		lhs.paramsScope, _ = lx.popScope(lhs.Token2)
		lhs.DirectDeclarator.parent = lhs
	}
|	DirectDeclarator '(' IdentifierListOpt ')'
	{
		lhs.DirectDeclarator.parent = lhs
	}

// [0](6.7.5)
Pointer:
	'*' TypeQualifierListOpt
|	'*' TypeQualifierListOpt Pointer

PointerOpt:
	/* empty */ {}
|	Pointer

// [0](6.7.5)
//yy:field	attr		int	// tsInline, tsTypedefName, ...
TypeQualifierList:
	TypeQualifier
	{
		lhs.attr = lhs.TypeQualifier.attr
	}
|	TypeQualifierList TypeQualifier
	{
		a := lhs.TypeQualifierList
		b := lhs.TypeQualifier
		if a.attr&b.attr != 0 {
			lx.report.Err(b.Pos(), "invalid type qualifier")
			break
		}

		lhs.attr = a.attr|b.attr
	}

TypeQualifierListOpt:
	/* empty */ {}
|	TypeQualifierList

// [0](6.7.5)
//yy:field	params	[]Parameter
ParameterTypeList:
	ParameterList
	{
		lhs.post()
	}
|	ParameterList ',' "..."
	{
		lhs.post()
	}

ParameterTypeListOpt:
	/* empty */ {}
|	ParameterTypeList

// [0](6.7.5)
ParameterList:
	ParameterDeclaration
|	ParameterList ',' ParameterDeclaration

// [0](6.7.5)
//yy:field	declarator	*Declarator
/*TODO
A declaration of a parameter as ‘‘function returning type’’ shall be adjusted
to ‘‘pointer to function returning type’’, as in 6.3.2.1.
*/
ParameterDeclaration:
	DeclarationSpecifiers Declarator
	{
		lhs.Declarator.setFull(lx)
		lhs.declarator = lhs.Declarator
	}
|	DeclarationSpecifiers AbstractDeclaratorOpt
	{
		if o := lhs.AbstractDeclaratorOpt; o != nil {
			lhs.declarator = o.AbstractDeclarator.declarator
			lhs.declarator.setFull(lx)
			break
		}

		d := &Declarator{
			specifier: lx.scope.specifier,
			DirectDeclarator: &DirectDeclarator{
				Case: 0, // IDENTIFIER
			},
		}
		d.DirectDeclarator.declarator = d
		lhs.declarator = d
		d.setFull(lx)
	}

// [0](6.7.5)
IdentifierList:
	IDENTIFIER
|	IdentifierList ',' IDENTIFIER

//yy:field	params	[]Parameter
IdentifierListOpt:
	/* empty */ {}
|	IdentifierList

IdentifierOpt:
	/* empty */ {}
|	IDENTIFIER

// [0](6.7.6)
//yy:field	Type		Type
//yy:field	declarator	*Declarator
//yy:field	scope		*Bindings
TypeName:
	{
		lx.pushScope(ScopeBlock)
	}
	SpecifierQualifierList AbstractDeclaratorOpt
	{
		if o := lhs.AbstractDeclaratorOpt; o != nil {
			lhs.declarator = o.AbstractDeclarator.declarator
		} else {
			d := &Declarator{
				specifier: lhs.SpecifierQualifierList,
				DirectDeclarator: &DirectDeclarator{
					Case: 0, // IDENTIFIER
					idScope: lx.scope,
				},
			}
			d.DirectDeclarator.declarator = d
			lhs.declarator = d
		}
		lhs.Type = lhs.declarator.setFull(lx)
		lhs.scope = lx.scope
		lx.popScope(xc.Token{})
	}

// [0](6.7.6)
//yy:field	declarator	*Declarator
AbstractDeclarator:
	Pointer
	{
		d := &Declarator{
			specifier: lx.scope.specifier,
			PointerOpt: &PointerOpt {
				Pointer: lhs.Pointer,
			},
			DirectDeclarator: &DirectDeclarator{
				Case: 0, // IDENTIFIER
				idScope: lx.scope,
			},
		}
		d.DirectDeclarator.declarator = d
		lhs.declarator = d
	}
|	PointerOpt DirectAbstractDeclarator
	{
		d := &Declarator{
			specifier: lx.scope.specifier,
			PointerOpt: lhs.PointerOpt,
			DirectDeclarator: lhs.DirectAbstractDeclarator.directDeclarator,
		}
		d.DirectDeclarator.declarator = d
		lhs.declarator = d
	}

AbstractDeclaratorOpt:
	/* empty */ {}
|	AbstractDeclarator

// [0](6.7.6)
//yy:field	directDeclarator	*DirectDeclarator
//yy:field	paramsScope		*Bindings
DirectAbstractDeclarator:
	'(' AbstractDeclarator ')'
	{
		lhs.AbstractDeclarator.declarator.specifier = nil
		lhs.directDeclarator = &DirectDeclarator{
			Case: 1, // '(' Declarator ')'
			Declarator: lhs.AbstractDeclarator.declarator,
		}
		lhs.AbstractDeclarator.declarator.DirectDeclarator.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclaratorOpt '[' ExpressionOpt ']'
	{
		nElements := -1
		if o := lhs.ExpressionOpt; o != nil {
			var err error
			if nElements, err = elements(o.Expression.eval(lx)); err != nil {
				lx.report.Err(o.Expression.Pos(), "%s", err)
			}
		}
		var dd *DirectDeclarator
		switch o := lhs.DirectAbstractDeclaratorOpt; {
		case o == nil:
			dd = &DirectDeclarator{
				Case: 0, // IDENTIFIER
			}
		default:
			dd = o.DirectAbstractDeclarator.directDeclarator
		}
		lhs.directDeclarator = &DirectDeclarator{
			Case: 2, // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			DirectDeclarator: dd,
			ExpressionOpt: lhs.ExpressionOpt,
			elements: nElements,
		}
		dd.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclaratorOpt '[' TypeQualifierList ExpressionOpt ']'
	{
		if o := lhs.ExpressionOpt; o != nil {
			o.Expression.eval(lx)
		}
		var dd *DirectDeclarator
		switch o := lhs.DirectAbstractDeclaratorOpt; {
		case o == nil:
			dd = &DirectDeclarator{
				Case: 0, // IDENTIFIER
			}
		default:
			dd = o.DirectAbstractDeclarator.directDeclarator
		}
		lhs.directDeclarator = &DirectDeclarator{
			Case: 2, // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			DirectDeclarator: dd,
			TypeQualifierListOpt: &TypeQualifierListOpt{ lhs.TypeQualifierList },
			ExpressionOpt: lhs.ExpressionOpt,
		}
		dd.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclaratorOpt '[' "static" TypeQualifierListOpt Expression ']'
	{
		lhs.Expression.eval(lx)
		var dd *DirectDeclarator
		switch o := lhs.DirectAbstractDeclaratorOpt; {
		case o == nil:
			dd = &DirectDeclarator{
				Case: 0, // IDENTIFIER
			}
		default:
			dd = o.DirectAbstractDeclarator.directDeclarator
		}
		lhs.directDeclarator = &DirectDeclarator{
			Case: 2, // DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'
			DirectDeclarator: dd,
			TypeQualifierListOpt: lhs.TypeQualifierListOpt,
			Expression: lhs.Expression,
		}
		dd.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclaratorOpt '[' TypeQualifierList "static" Expression ']'
	{
		lhs.Expression.eval(lx)
		var dd *DirectDeclarator
		switch o := lhs.DirectAbstractDeclaratorOpt; {
		case o == nil:
			dd = &DirectDeclarator{
				Case: 0, // IDENTIFIER
			}
		default:
			dd = o.DirectAbstractDeclarator.directDeclarator
		}
		lhs.directDeclarator = &DirectDeclarator{
			Case: 4, // DirectDeclarator '[' TypeQualifierList "static" Expression ']'
			DirectDeclarator: dd,
			TypeQualifierList: lhs.TypeQualifierList,
			Expression: lhs.Expression,
		}
		dd.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclaratorOpt '[' '*' ']'
	{
		var dd *DirectDeclarator
		switch o := lhs.DirectAbstractDeclaratorOpt; {
		case o == nil:
			dd = &DirectDeclarator{
				Case: 0, // IDENTIFIER
			}
		default:
			dd = o.DirectAbstractDeclarator.directDeclarator
		}
		lhs.directDeclarator = &DirectDeclarator{
			Case: 5, // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
			DirectDeclarator: dd,
		}
		dd.parent = lhs.directDeclarator
	}
|	'('
	{
		lx.pushScope(ScopeParams)
	}
	ParameterTypeListOpt ')'
	{
		lhs.paramsScope, _ = lx.popScope(lhs.Token2)
		switch o := lhs.ParameterTypeListOpt; {
		case o != nil:
			lhs.directDeclarator = &DirectDeclarator{
				Case: 6, // DirectDeclarator '(' ParameterTypeList ')'
				DirectDeclarator: &DirectDeclarator{
					Case: 0, // IDENTIFIER
				},
				ParameterTypeList: o.ParameterTypeList,
			}
		default:
			lhs.directDeclarator = &DirectDeclarator{
				Case: 7, // DirectDeclarator '(' IdentifierListOpt ')'
				DirectDeclarator: &DirectDeclarator{
					Case: 0, // IDENTIFIER
				},
			}
		}
		lhs.directDeclarator.DirectDeclarator.parent = lhs.directDeclarator
	}
|	DirectAbstractDeclarator '('
	{
		lx.pushScope(ScopeParams)
	}
	ParameterTypeListOpt ')'
	{
		lhs.paramsScope, _ = lx.popScope(lhs.Token2)
		switch o := lhs.ParameterTypeListOpt; {
		case o != nil:
			lhs.directDeclarator = &DirectDeclarator{
				Case: 6, // DirectDeclarator '(' ParameterTypeList ')'
				DirectDeclarator: lhs.DirectAbstractDeclarator.directDeclarator,
				ParameterTypeList: o.ParameterTypeList,
			}
		default:
			lhs.directDeclarator = &DirectDeclarator{
				Case: 7, // DirectDeclarator '(' IdentifierListOpt ')'
				DirectDeclarator: lhs.DirectAbstractDeclarator.directDeclarator,
			}
		}
		lhs.directDeclarator.DirectDeclarator.parent = lhs.directDeclarator
	}

DirectAbstractDeclaratorOpt:
	/* empty */ {}
|	DirectAbstractDeclarator

// [0](6.7.8)
Initializer:
	Expression
	{
		lhs.Expression.eval(lx)
	}
|	'{' InitializerList CommaOpt '}'
|	IDENTIFIER ':' Initializer
	{
		if !lx.tweaks.enableLegacyDesignators {
			lx.report.Err(lhs.Pos(), "legacy designators not enabled")
		}
	}

// [0](6.7.8)
InitializerList:
	DesignationOpt Initializer
|	InitializerList ',' DesignationOpt Initializer
|	/* empty */ {}

// [0](6.7.8)
Designation:
	DesignatorList '='

DesignationOpt:
	/* empty */ {}
|	Designation

// [0](6.7.8)
DesignatorList:
	Designator
|	DesignatorList Designator

// [0](6.7.8)
Designator:
	'[' ConstantExpression ']'
|	'.' IDENTIFIER

// [0](6.8)
Statement:
	LabeledStatement
|	CompoundStatement
|	ExpressionStatement
|	SelectionStatement
|	IterationStatement
|	JumpStatement
|	AssemblerStatement

// [0](6.8.1)
LabeledStatement:
	IDENTIFIER ':' Statement
|	"case" ConstantExpression ':' Statement
|	"default" ':' Statement

// [0](6.8.2)
//yy:field	scope	*Bindings	// Scope of the CompoundStatement.
CompoundStatement:
	'{'
	{
		m := lx.scope.mergeScope
		lx.pushScope(ScopeBlock)
		if m != nil {
			lx.scope.merge(m)
		}
		lx.scope.mergeScope = nil
	}
	BlockItemListOpt '}'
	{
		lhs.scope = lx.scope
		lx.popScope(lhs.Token2)
	}

// [0](6.8.2)
BlockItemList:
	BlockItem
|	BlockItemList BlockItem

BlockItemListOpt:
	/* empty */ {}
|	BlockItemList

// [0](6.8.2)
BlockItem:
	Declaration
|	Statement

// [0](6.8.3)
ExpressionStatement:
	ExpressionListOpt ';'

// [0](6.8.4)
SelectionStatement:
	"if" '(' ExpressionList ')' Statement %prec NOELSE
	{
		lhs.ExpressionList.eval(lx)
	}
|	"if" '(' ExpressionList ')' Statement "else" Statement
	{
		lhs.ExpressionList.eval(lx)
	}
|	"switch" '(' ExpressionList ')' Statement
	{
		lhs.ExpressionList.eval(lx)
	}

// [0](6.8.5)
IterationStatement:
	"while" '(' ExpressionList ')' Statement
	{
		lhs.ExpressionList.eval(lx)
	}
|	"do" Statement "while" '(' ExpressionList ')' ';'
	{
		lhs.ExpressionList.eval(lx)
	}
|	"for" '(' ExpressionListOpt ';' ExpressionListOpt ';' ExpressionListOpt ')' Statement
|	"for" '(' Declaration ExpressionListOpt ';' ExpressionListOpt ')' Statement

// [0](6.8.6)
JumpStatement:
	"goto" IDENTIFIER ';'
|	"continue" ';'
|	"break" ';'
|	"return" ExpressionListOpt ';'
|	"goto" Expression ';'
	{
		_, t := lhs.Expression.eval(lx)
		if t == nil {
			break
		}

		for t != nil && t.Kind() == Ptr {
			t = t.Element()
		}

		if t == nil || t.Kind() != Void {
			lx.report.Err(lhs.Pos(), "invalid computed goto argument type, have '%s'", t)
		}

		if !lx.tweaks.enableComputedGotos {
			lx.report.Err(lhs.Pos(), "computed gotos not enabled")
		}
	}

// [0](6.9)
//yy:field	Comments	map[token.Pos]int	// Position -> comment ID. Enable using the KeepComments option.
//yy:field	Declarations	*Bindings
//yy:field	Macros		map[int]*Macro		// Ident ID -> preprocessor macro defined by ident.
//yy:field	Model		*Model			// Model used to parse the TranslationUnit.
//yy:list
TranslationUnit:
	ExternalDeclaration
|	TranslationUnit ExternalDeclaration

// [0](6.9)
ExternalDeclaration:
	FunctionDefinition
|	Declaration
|	BasicAssemblerStatement ';'
|	';'
	{
		if !lx.tweaks.enableEmptyDeclarations {
			lx.report.Err(lhs.Pos(), "C++11 empty declarations are illegal in C99.")
		}
	}

// [0](6.9.1)
FunctionDefinition:
	DeclarationSpecifiers Declarator DeclarationListOpt
	{
		if ds := $1.(*DeclarationSpecifiers); ds.typeSpecifier == 0 {
			ds.typeSpecifier = tsEncode(tsInt)
			$2.(*Declarator).Type = lx.model.IntType
			if !lx.tweaks.enableOmitFuncRetType {
				lx.report.Err($2.Pos(), "missing function return type")
			}
		}
		var fd *FunctionDefinition
		fd.post(lx, $2.(*Declarator), $3.(*DeclarationListOpt))
	}
	FunctionBody
|	{
		lx.scope.specifier = &DeclarationSpecifiers{typeSpecifier: tsEncode(tsInt)}
	}
	Declarator DeclarationListOpt
	{
		if !lx.tweaks.enableOmitFuncRetType {
			lx.report.Err($2.Pos(), "missing function return type")
		}
		var fd *FunctionDefinition
		fd.post(lx, $2.(*Declarator), $3.(*DeclarationListOpt))
	}
	FunctionBody

//yy:field	scope	*Bindings	// Scope of the FunctionBody.
FunctionBody:
	{
		// Handle __func__, [0], 6.4.2.2.
		id, _ := lx.fnDeclarator.Identifier()
		lx.injectFunc = []xc.Token{
			{lex.Char{Rune: STATIC}, idStatic},
			{lex.Char{Rune: CONST}, idConst},
			{lex.Char{Rune: CHAR}, idChar},
			{lex.Char{Rune: IDENTIFIER}, idMagicFunc},
			{lex.Char{Rune: '['}, 0},
			{lex.Char{Rune: ']'}, 0},
			{lex.Char{Rune: '='}, 0},
			{lex.Char{Rune: STRINGLITERAL}, xc.Dict.SID(fmt.Sprintf("%q", xc.Dict.S(id)))},
			{lex.Char{Rune: ';'}, 0},
		}
	}
	CompoundStatement
	{
		lhs.scope = lhs.CompoundStatement.scope
	}
|	
	{
		m := lx.scope.mergeScope
		lx.pushScope(ScopeBlock)
		if m != nil {
			lx.scope.merge(m)
		}
		lx.scope.mergeScope = nil
	}
	AssemblerStatement ';'
	{
		lhs.scope = lx.scope
		lx.popScope(lx.tokPrev)
	}

// [0](6.9.1)
DeclarationList:
	Declaration
|	DeclarationList Declaration

//yy:field	paramsScope	*Bindings
DeclarationListOpt:
	/* empty */ {}
|	{
		lx.pushScope(ScopeParams)
	}
	DeclarationList
	{
		lhs.paramsScope, _ = lx.popScopePos(lhs.Pos())
	}

//yy:list
AssemblerInstructions:
	STRINGLITERAL
|	AssemblerInstructions STRINGLITERAL

BasicAssemblerStatement:
	"asm" VolatileOpt '(' AssemblerInstructions ')'

VolatileOpt:
	/* empty */ {}
|	"volatile"

AssemblerOperand:
	AssemblerSymbolicNameOpt STRINGLITERAL '(' Expression ')'

//yy:list
AssemblerOperands:
	AssemblerOperand
|	AssemblerOperands ',' AssemblerOperand

AssemblerSymbolicNameOpt:
	/* empty */ {}
|	'[' IDENTIFIER ']'

//yy:list
Clobbers:
	STRINGLITERAL
|	Clobbers ',' STRINGLITERAL

AssemblerStatement:
	BasicAssemblerStatement
|	"asm" VolatileOpt '(' AssemblerInstructions ':' AssemblerOperands ')'
|	"asm" VolatileOpt '(' AssemblerInstructions ':' AssemblerOperands ':' AssemblerOperands ')'
|	"asm" VolatileOpt '(' AssemblerInstructions ':' AssemblerOperands ':' AssemblerOperands ':' Clobbers ')'
|	"asm" VolatileOpt "goto" '(' AssemblerInstructions ':' ':' AssemblerOperands ':' Clobbers ':' IdentifierList ')'
|	"asm" VolatileOpt '(' AssemblerInstructions ':' ')' // https://gitlab.com/cznic/cc/issues/59
|	"asm" VolatileOpt '(' AssemblerInstructions ':' ':' AssemblerOperands ')'

StaticAssertDeclaration:
	"_Static_assert" '(' ConstantExpression ',' STRINGLITERAL ')' ';'
	{
		ce := lhs.ConstantExpression
		if ce.Type == nil || ce.Type.Kind() == Undefined || ce.Value == nil || !IsIntType(ce.Type) {
			lx.report.Err(ce.Pos(), "invalid static assert expression (have '%v')", ce.Type)
			break
		}

		if !isNonZero(ce.Value) {
			lx.report.ErrTok(lhs.Token, "%s", lhs.Token4.S())
		}
	}

// ========================================================= PREPROCESSING_FILE

// [0](6.10)
//yy:field	path	string
PreprocessingFile:
	GroupList // No more GroupListOpt due to final '\n' injection.
	{
		lhs.path = lx.file.Name()
	}

// [0](6.10)
GroupList:
	GroupPart
//yy:example "\U00100000int\nf() {}"
|	GroupList GroupPart

GroupListOpt:
	/* empty */ {}
//yy:example "\U00100000 \n#ifndef a\nb\n#elif"
|	GroupList

// [0](6.10)
//yy:ignore
GroupPart:
	ControlLine
	{
		$$ = $1.(Node)
	}
|	IfSection
	{
		$$ = $1.(Node)
	}
|	PPNONDIRECTIVE PPTokenList '\n'
	{
		$$ = $1
	}
|	TextLine
	{
		$$ = $1
	}

//(6.10)
IfSection:
	IfGroup ElifGroupListOpt ElseGroupOpt EndifLine

//(6.10)
IfGroup:
	PPIF PPTokenList '\n' GroupListOpt
|	PPIFDEF IDENTIFIER '\n' GroupListOpt
|	PPIFNDEF IDENTIFIER '\n' GroupListOpt

// [0](6.10)
ElifGroupList:
	ElifGroup
|	ElifGroupList ElifGroup

ElifGroupListOpt:
	/* empty */ {}
|	ElifGroupList

// [0](6.10)
ElifGroup:
	PPELIF PPTokenList '\n' GroupListOpt

// [0](6.10)
ElseGroup:
	PPELSE '\n' GroupListOpt

ElseGroupOpt:
	/* empty */ {}
|	ElseGroup

// [0](6.10)
EndifLine:
	PPENDIF /* PPTokenListOpt */ //TODO Option enabling the non std PPTokenListOpt part.

// [0](6.10)
ControlLine:
	PPDEFINE IDENTIFIER ReplacementList
|	PPDEFINE IDENTIFIER_LPAREN "..." ')' ReplacementList
|	PPDEFINE IDENTIFIER_LPAREN IdentifierList ',' "..." ')' ReplacementList
|	PPDEFINE IDENTIFIER_LPAREN IdentifierListOpt ')' ReplacementList
|	PPERROR PPTokenListOpt
|	PPHASH_NL
|	PPINCLUDE PPTokenList '\n'
|	PPLINE PPTokenList '\n'
|	PPPRAGMA PPTokenListOpt
//yy:example	"\U00100000 \n#undef foo"
|	PPUNDEF IDENTIFIER '\n'

	// Non standard stuff.

|	PPDEFINE IDENTIFIER_LPAREN IdentifierList "..." ')' ReplacementList
	{
		if !lx.tweaks.enableDefineOmitCommaBeforeDDD {
			lx.report.ErrTok(lhs.Token4, "missing comma before \"...\"")
		}
	}
|	PPDEFINE '\n'
	{
		if !lx.tweaks.enableEmptyDefine {
			lx.report.ErrTok(lhs.Token2, "expected identifier")
		}
	}
//yy:example	"\U00100000 \n#undef foo(bar)"
|	PPUNDEF IDENTIFIER PPTokenList '\n'
	{
		toks := decodeTokens(lhs.PPTokenList, nil, false)
		if len(toks) == 0 {
			lhs.Case = 9 // PPUNDEF IDENTIFIER '\n' 
			break
		}

		lx.report.ErrTok(toks[0], "extra tokens after #undef argument")
	}
|	PPINCLUDE_NEXT PPTokenList '\n'

// [0](6.10)
//yy:ignore
TextLine:
	PPTokenListOpt

// [0](6.10)
//yy:ignore
ReplacementList:
	PPTokenListOpt

// [0](6.10)
//yy:ignore
PPTokenList:
	PPTokens
	{
		$$ = PPTokenList(dict.ID(lx.encBuf))
		lx.encBuf = lx.encBuf[:0]
		lx.encPos = 0
	}

//yy:ignore
PPTokenListOpt:
	'\n'
	{
		$$ = 0
	}
|	PPTokenList '\n'

//yy:ignore
PPTokens:
	PPOTHER
|	PPTokens PPOTHER
