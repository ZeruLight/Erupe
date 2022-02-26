%{
// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on [0], 6.5-6.10. Substantial portions of expression AST size
// optimizations are from [1], license of which follows.
//
// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
// [1]: https://github.com/rsc/c2go/blob/fc8cbfad5a47373828c81c7a56cccab8b221d310/cc/cc.y

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
	"modernc.org/xc"
)
%}

%union {
	Token			xc.Token // Must be exported
	node			Node
}

%token
	/*yy:token "'%c'"            */ CHARCONST		"character constant"
	/*yy:token "1.%d"            */ FLOATCONST		"floating-point constant"
	/*yy:token "%c"              */ IDENTIFIER		"identifier"
	/*yy:token "%d"              */ INTCONST		"integer constant"
	/*yy:token "L'%c'"           */ LONGCHARCONST		"long character constant"
	/*yy:token "L\"%c\""         */ LONGSTRINGLITERAL	"long string constant"
	/*yy:token "%d"              */ PPNUMBER		"preprocessing number"
	/*yy:token "\"%c\""          */ STRINGLITERAL		"string literal"

	/*yy:token "\U00100000"      */	CONSTANT_EXPRESSION	1048576	"constant expression prefix"	// 0x100000 = 1048576
	/*yy:token "\U00100001"      */	TRANSLATION_UNIT	1048577	"translation unit prefix"

	ADDASSIGN			"+="
	ANDAND				"&&"
	ANDASSIGN			"&="
	ARROW				"->"
	ALIGNAS				"_Alignas"
	ALIGNOF				"_Alignof"
	ATOMIC				"_Atomic"
	ATOMIC_LPAREN			"("
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
	DIRECTIVE			// # when it's the first non white token on a line.
	DIVASSIGN			"/="
	DO				"do"
	DOUBLE				"double"
	ELSE				"else"
	ENUM				"enum"
	EQ				"=="
	EXTERN				"extern"
	FLOAT				"float"
	FOR				"for"
	GENERIC				"_Generic"
	GEQ				">="
	GOTO				"goto"
	IF				"if"
	IMAGINARY			"_Imaginary"
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
	NON_REPL			// [0]6.10.3.4-2
	NORETURN			"_Noreturn"
	ORASSIGN			"|="
	OROR				"||"
	PPPASTE				"##"
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
	THREAD_LOCAL			"_Thread_local"
	TYPEDEF				"typedef"
	TYPEDEF_NAME			"typedef name"
	TYPEOF				"typeof"
	UNION				"union"
	UNSIGNED			"unsigned"
	VOID				"void"
	VOLATILE			"volatile"
	WHILE				"while"
	XORASSIGN			"^="

%type	<node>
	AbstractDeclarator		"abstract declarator"
	AbstractDeclaratorOpt		"optional abstract declarator"
	ArgumentExprList		"argument expression list"
	ArgumentExprListOpt		"optional argument expression list"
	BlockItem			"block item"
	BlockItemList			"block item list"
	BlockItemListOpt		"optional block item list"
	CommaOpt			"optional comma"
	CompoundStmt			"compound statement"
	ConstExpr			"constant expression"
	Declaration			"declaration"
	DeclarationList			"declaration list"
	DeclarationListOpt		"optional declaration list"
	DeclarationSpecifiers		"declaration specifiers"
	DeclarationSpecifiersOpt	"optional declaration specifiers"
	Declarator			"declarator"
	DeclaratorOpt			"optional declarator"
	Designation			"designation"
	Designator			"designator"
	DesignatorList			"designator list"
	DirectAbstractDeclarator	"direct abstract declarator"
	DirectAbstractDeclaratorOpt	"optional direct abstract declarator"
	DirectDeclarator		"direct declarator"
	EnumSpecifier			"enum specifier"
	EnumerationConstant		"enumearation constant"
	Enumerator			"enumerator"
	EnumeratorList			"enumerator list"
	Expr				"expression"
	ExprList			"expression list"
	ExprListOpt			"optional expression list"
	ExprOpt				"optional expression"
	ExprStmt			"expression statement"
	ExternalDeclaration		"external declaration"
	ExternalDeclarationList		"external declaration list"
	FunctionBody			"function body"
	FunctionDefinition		"function definition"
	FunctionSpecifier		"function specifier"
	IdentifierList			"identifier list"
	IdentifierListOpt		"optional identifier list"
	IdentifierOpt			"optional identifier"
	InitDeclarator			"init declarator"
	InitDeclaratorList		"init declarator list"
	InitDeclaratorListOpt		"optional init declarator list"
	Initializer			"initializer"
	InitializerList			"initializer list"
	IterationStmt			"iteration statement"
	JumpStmt			"jump statement"
	LabeledStmt			"labeled statement"
	ParameterDeclaration		"parameter declaration"
	ParameterList			"parameter list"
	ParameterTypeList		"parameter type list"
	ParameterTypeListOpt		"optional parameter type list"
	Pointer				"pointer"
	PointerOpt			"optional pointer"
	SelectionStmt			"selection statement"
	SpecifierQualifierList		"specifier qualifier list"
	SpecifierQualifierListOpt	"optional specifier qualifier list"
	Stmt				"statement"
	StorageClassSpecifier		"storage class specifier"
	StructDeclaration		"struct declaration"
	StructDeclarationList		"struct declaration list"
	StructDeclarator		"struct declarator"
	StructDeclaratorList		"struct declarator list"
	StructOrUnion			"struct-or-union"
	StructOrUnionSpecifier		"struct-or-union specifier"
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
%left		'!' '~' ALIGNOF SIZEOF UNARY
%right		'(' '.' '[' ARROW DEC INC

%%

                        /*yy:ignore */
                        Start:
                        	CONSTANT_EXPRESSION ConstExpr
				{
					lx.ast = $2
				}
                        |	TRANSLATION_UNIT ExternalDeclarationList
				{
					lx.ast = &TranslationUnit{
						ExternalDeclarationList: $2.(*ExternalDeclarationList).reverse(), 
						FileScope: lx.scope,
						FileSet: fset,
						Model: lx.model,
					}
				}

                        // [0]6.4.4.3
			//yy:field	Operand	Operand
                        EnumerationConstant:
                        	IDENTIFIER

/*yy:case TypeName   */ AlignmentSpecifier:
				"_Alignas" '(' TypeName ')'
/*yy:case ConstExpr  */	|	"_Alignas" '(' ConstExpr ')'

                        // [0]6.5.2
                        ArgumentExprList:
                        	Expr
                        |	ArgumentExprList ',' Expr

                        ArgumentExprListOpt:
                        	/* empty */ {}
                        |	ArgumentExprList

                        // [0]6.5.16
			//yy:field	CallArgs	[]Operand	// Promoted arguments of Call.
			//yy:field	Declarator	*Declarator	// Case Ident.
			//yy:field	Operand		Operand
			//yy:field	Scope		*Scope		// Case Ident, CompLit.
			//yy:field	enum		*EnumType
			//yy:field	AssignedTo	bool		// Expression appears at the left side of assignment.
			//yy:field	UseGotos	bool
/*yy:case PreInc     */ Expr:
                        	"++" Expr
/*yy:case PreDec     */ |	"--" Expr
/*yy:case AlignofType*/ |	"_Alignof" '(' TypeName ')' %prec ALIGNOF
/*yy:case AlignofExpr*/ |	"_Alignof" Expr
/*yy:case SizeofType */ |	"sizeof" '(' TypeName ')' %prec SIZEOF
/*yy:case SizeofExpr */ |	"sizeof" Expr
/*yy:case Not        */ |	'!' Expr
/*yy:case Addrof     */ |	'&' Expr %prec UNARY
/*yy:case Statement  */ |	'(' CompoundStmt ')'
/*yy:case PExprList  */ |	'(' ExprList ')'
/*yy:case CompLit    */ |	'(' TypeName ')' '{' InitializerList CommaOpt '}'
				{
					lhs.Scope = lx.scope
				}
/*yy:case Cast       */ |	'(' TypeName ')' Expr %prec CAST
/*yy:case Deref      */ |	'*' Expr %prec UNARY
/*yy:case UnaryPlus  */ |	'+' Expr %prec UNARY
/*yy:case UnaryMinus */ |	'-' Expr %prec UNARY
/*yy:case Cpl        */ |	'~' Expr
/*yy:case Char       */ |	CHARCONST
/*yy:case Ne         */ |	Expr "!=" Expr
/*yy:case ModAssign  */ |	Expr "%=" Expr
/*yy:case LAnd       */ |	Expr "&&" Expr
/*yy:case AndAssign  */ |	Expr "&=" Expr
/*yy:case MulAssign  */ |	Expr "*=" Expr
/*yy:case PostInc    */ |	Expr "++"
/*yy:case AddAssign  */ |	Expr "+=" Expr
/*yy:case PostDec    */ |	Expr "--"
/*yy:case SubAssign  */ |	Expr "-=" Expr
/*yy:case PSelect    */ |	Expr "->" IDENTIFIER
/*yy:case DivAssign  */ |	Expr "/=" Expr
/*yy:case Lsh        */ |	Expr "<<" Expr
/*yy:case LshAssign  */ |	Expr "<<=" Expr
/*yy:case Le         */ |	Expr "<=" Expr
/*yy:case Eq         */ |	Expr "==" Expr
/*yy:case Ge         */ |	Expr ">=" Expr
/*yy:case Rsh        */ |	Expr ">>" Expr
/*yy:case RshAssign  */ |	Expr ">>=" Expr
/*yy:case XorAssign  */ |	Expr "^=" Expr
/*yy:case OrAssign   */ |	Expr "|=" Expr
/*yy:case LOr        */ |	Expr "||" Expr
/*yy:case Mod        */ |	Expr '%' Expr
/*yy:case And        */ |	Expr '&' Expr
/*yy:case Call       */ |	Expr '(' ArgumentExprListOpt ')'
/*yy:case Mul        */ |	Expr '*' Expr
/*yy:case Add        */ |	Expr '+' Expr
/*yy:case Sub        */ |	Expr '-' Expr
/*yy:case Select     */ |	Expr '.' IDENTIFIER
/*yy:case Div        */ |	Expr '/' Expr
/*yy:case Lt         */ |	Expr '<' Expr
/*yy:case Assign     */ |	Expr '=' Expr
/*yy:case Gt         */ |	Expr '>' Expr
/*yy:case Cond       */ |	Expr '?' ExprList ':' Expr
/*yy:case Index      */ |	Expr '[' ExprList ']'
/*yy:case Xor        */ |	Expr '^' Expr
/*yy:case Or         */ |	Expr '|' Expr
/*yy:case Float      */ |	FLOATCONST
/*yy:case Ident      */ |	IDENTIFIER %prec NOSEMI
				{
					lhs.Scope = lx.scope
				}
/*yy:case Int        */ |	INTCONST
/*yy:case LChar      */ |	LONGCHARCONST
/*yy:case LString    */ |	LONGSTRINGLITERAL
/*yy:case String     */ |	STRINGLITERAL

                        ExprOpt:
                        	/* empty */ {}
                        |	Expr

                        // [0]6.5.17
                        //yy:list
			//yy:field	Operand	Operand
                        ExprList:
                        	Expr
                        |	ExprList ',' Expr

                        ExprListOpt:
                        	/* empty */ {}
                        |	ExprList

                        // [0]6.6
			//yy:field	Operand	Operand
                        ConstExpr:
                        	Expr

                        // [0]6.7
			//yy:field	Attributes	[][]xc.Token
			//yy:field	Scope		*Scope
			Declaration:
                        	DeclarationSpecifiers InitDeclaratorListOpt
				{
					lx.attr2 = lx.attr
				}
				';'
				{
					lhs.Scope = lx.scope
					if len(lx.attr2) != 0 {
						lhs.Attributes = lx.attrs()
					}
					lx.scope.typedef = false
				}

                        // [0]6.7
/*yy:case Func       */ DeclarationSpecifiers:
                        	FunctionSpecifier DeclarationSpecifiersOpt
/*yy:case Storage    */ |	StorageClassSpecifier DeclarationSpecifiersOpt
/*yy:case Qualifier  */ |	TypeQualifier DeclarationSpecifiersOpt
/*yy:case Specifier  */ |	TypeSpecifier DeclarationSpecifiersOpt
/*yy:case Alignment  */ |	AlignmentSpecifier DeclarationSpecifiersOpt

                        DeclarationSpecifiersOpt:
                        	/* empty */ {}
                        |	DeclarationSpecifiers

                        // [0]6.7
                        InitDeclaratorList:
                        	InitDeclarator
                        |	InitDeclaratorList ',' InitDeclarator

                        InitDeclaratorListOpt:
                        	/* empty */ {}
                        |	InitDeclaratorList

                        // [0]6.7
/*yy:case Base       */ InitDeclarator:
                        	Declarator
/*yy:case Init       */ |	Declarator '=' Initializer

                        // [0]6.7.1
/*yy:case Auto       */ StorageClassSpecifier:
                        	"auto"
/*yy:case Extern     */ |	"extern"
/*yy:case Register   */ |	"register"
/*yy:case Static     */ |	"static"
/*yy:case Typedef    */ |	"typedef"
				{
					lx.scope.typedef = true
				}

                        // [0]6.7.2
			//yy:field	scope	*Scope
			//yy:field	typ	Type	// typeof
/*yy:case Bool       */ TypeSpecifier:
                        	"_Bool"
/*yy:case Complex    */ |	"_Complex"
/*yy:case Imaginary  */ |	"_Imaginary"
/*yy:case Char       */ |	"char"
/*yy:case Double     */ |	"double"
/*yy:case Float      */ |	"float"
/*yy:case Int        */ |	"int"
/*yy:case Long       */ |	"long"
/*yy:case Short      */ |	"short"
/*yy:case Signed     */ |	"signed"
/*yy:case Unsigned   */ |	"unsigned"
/*yy:case Void       */ |	"void"
/*yy:case Enum       */ |	EnumSpecifier
/*yy:case Struct     */ |	StructOrUnionSpecifier
/*yy:example "\U00100001 typedef int foo; foo bar;" */
/*yy:case Name       */ |	TYPEDEF_NAME
				{
					lhs.scope = lx.scope
				}
/*yy:case TypeofExpr */	|	"typeof" '(' Expr ')'
/*yy:case Typeof     */	|	"typeof" '(' TypeName ')'
/*yy:case Atomic     */	|	"_Atomic" ATOMIC_LPAREN TypeName ')'

                        // [0]6.7.2.1
			//yy:field	scope	*Scope	// Declare the struct tag in scope.parent.
			//yy:field	typ	Type
/*yy:case Tag        */ StructOrUnionSpecifier:
                        	StructOrUnion IDENTIFIER
				{
					lhs.scope = lx.scope
				}
/*yy:case Empty      */ |	StructOrUnion IdentifierOpt '{'
				{
					lx.noTypedefName = true // https://gitlab.com/cznic/sqlite2go/issues/9
				}
				'}'
				{
					if !lx.tweaks.EnableEmptyStructs {
						lx.err($1, "empty structs/unions not allowed")
					}
				}
/*yy:case Define     */ |	StructOrUnion IdentifierOpt '{'
				{
					lx.newStructScope()
				}
				StructDeclarationList
				{
					lx.noTypedefName = true // https://gitlab.com/cznic/sqlite2go/issues/9
				}
				'}'
				{
					lhs.scope, _ = lx.popScope()
				}

                        // [0]6.7.2.1
/*yy:case Struct     */ StructOrUnion:
                        	"struct"
/*yy:case Union      */ |	"union"

                        // [0]6.7.2.1
                        StructDeclarationList:
                        	StructDeclaration
                        |	StructDeclarationList StructDeclaration

                        // [0]6.7.2.1
/*yy:case Base       */ StructDeclaration:
				SpecifierQualifierList StructDeclaratorList ';'
/*yy:case Anon       */	|	SpecifierQualifierList ';'
				{
					if !lx.tweaks.EnableAnonymousStructFields {
						lx.err($1, "anonymous structs/unions members not allowed")
					}
				}

                        // [0]6.7.2.1
/*yy:case Qualifier  */ SpecifierQualifierList:
                        	TypeQualifier SpecifierQualifierListOpt
/*yy:case Specifier  */ |	TypeSpecifier SpecifierQualifierListOpt

                        SpecifierQualifierListOpt:
                        	/* empty */ {}
                        |	SpecifierQualifierList

                        // [0]6.7.2.1
                        StructDeclaratorList:
                        	StructDeclarator
                        |	StructDeclaratorList ',' StructDeclarator

                        // [0]6.7.2.1
			//yy:field	Bits	int
/*yy:case Base       */ StructDeclarator:
                        	Declarator
/*yy:case Bits       */ |	DeclaratorOpt ':' ConstExpr

                        CommaOpt:
                        	/* empty */ {}
                        |	','

                        // [0]6.7.2.2
			//yy:field	Tag	int
			//yy:field	scope	*Scope	// Where to declare enumeration constants.
			//yy:field	typ	Type
/*yy:case Tag        */ EnumSpecifier:
                        	"enum" IDENTIFIER
				{
					lhs.scope = lx.scope
				}
/*yy:case Define     */ |	"enum" IdentifierOpt '{' EnumeratorList  CommaOpt '}'
				{
					lhs.scope = lx.scope
				}

                        // [0]6.7.2.2
                        EnumeratorList:
                        	Enumerator
                        |	EnumeratorList ',' Enumerator

                        // [0]6.7.2.2
/*yy:case Base       */ Enumerator:
                        	EnumerationConstant
/*yy:case Init       */ |	EnumerationConstant '=' ConstExpr

                        // [0]6.7.3
/*yy:case Const      */ TypeQualifier:
                        	"const"
/*yy:case Restrict   */ |	"restrict"
/*yy:case Volatile   */ |	"volatile"
/*yy:case Atomic     */ |	"_Atomic"

                        // [0]6.7.4
/*yy:case Inline     */	FunctionSpecifier:
				"inline"
/*yy:case Noreturn   */	|	"_Noreturn"

                        // [0]6.7.5
			//yy:field	AssignedTo		int			// Declarator appears at the left side of assignment.
			//yy:field	Attributes		[][]xc.Token
			//yy:field	Bits			int			// StructDeclarator: bit width when a bit field.
			//yy:field	DeclarationSpecifier	*DeclarationSpecifier	// Nil for embedded declarators.
			//yy:field	Definition		*Declarator		// Declaration -> definition.
			//yy:field	Field			int			// Declaration order# if struct field declarator.
			//yy:field	FunctionDefinition	*FunctionDefinition	// When the declarator defines a function.
			//yy:field	Initializer		*Initializer		// Only when part of an InitDeclarator.
			//yy:field	Linkage			Linkage			// Linkage of the declared name, [0]6.2.2.
			//yy:field	Parameters		[]*Declarator		// Of the function declarator.
			//yy:field	Referenced		int
			//yy:field	Scope			*Scope			// Declaration scope.
			//yy:field	ScopeNum		int			// Sequential scope number within function body.
			//yy:field	StorageDuration		StorageDuration		// Storage duration of the declared name, [0]6.2.4.
			//yy:field	Type			Type			// Declared type.
			//yy:field	TypeQualifiers		[]*TypeQualifier	// From the PointerOpt production, if any.
			//yy:field	unnamed			int
			//yy:field	vars			[]*Declarator		// Function declarator only.
			//yy:field	AddressTaken		bool
			//yy:field	Alloca			bool			// Function declarator: Body calls __builtin_alloca
			//yy:field	Embedded		bool			// [0]6.7.5-3: Not a full declarator.
			//yy:field	IsField			bool
			//yy:field	IsFunctionParameter	bool
			//yy:field	IsBuiltin		bool
                        Declarator:
                        	PointerOpt DirectDeclarator
				{
					lhs.Attributes = lx.attrs()
					lhs.Scope = lx.scope
					lx.scope.insertTypedef(lx.context, lhs.Name(), lx.scope.typedef)
				}

/*yy:case IdentList  */	Parameters:
				IdentifierListOpt
/*yy:case ParamTypes */ |	ParameterTypeList

                        DeclaratorOpt:
                        	/* empty */ {}
                        |	Declarator

                        // [0]6.7.5
			//yy:field	paramScope	*Scope
/*yy:case Paren      */ DirectDeclarator:
                        	'(' Declarator ')'
				{
					lhs.Declarator.Embedded = true
				}
/*yy:case Parameters  */ |	DirectDeclarator
				{
					lx.newScope()
					lx.fixDeclarator($1)
				}
				'(' Parameters
				{
					lx.postFixDeclarator(lx.context)
				}
				')'
				{
					lhs.paramScope, _ = lx.popScope()
				}
/*yy:case ArraySize  */ |	DirectDeclarator '[' "static" TypeQualifierListOpt Expr ']'
/*yy:case ArraySize2 */ |	DirectDeclarator '[' TypeQualifierList "static" Expr ']'
/*yy:case ArrayVar   */ |	DirectDeclarator '[' TypeQualifierListOpt '*' ']'
/*yy:case Array      */ |	DirectDeclarator '[' TypeQualifierListOpt ExprOpt ']'
/*yy:case Ident      */ |	IDENTIFIER

                        // [0]6.7.5
/*yy:case Base       */ Pointer:
                        	'*' TypeQualifierListOpt
/*yy:case Ptr        */ |	'*' TypeQualifierListOpt Pointer

                        PointerOpt:
                        	/* empty */ {}
                        |	Pointer

                        // [0]6.7.5
                        TypeQualifierList:
                        	TypeQualifier
                        |	TypeQualifierList TypeQualifier

                        TypeQualifierListOpt:
                        	/* empty */ {}
                        |	TypeQualifierList

                        // [0]6.7.5
/*yy:case Base       */ ParameterTypeList:
                        	ParameterList
/*yy:case Dots       */ |	ParameterList ',' "..."

                        ParameterTypeListOpt:
                        	/* empty */ {}
                        |	ParameterTypeList

                        // [0]6.7.5
                        ParameterList:
                        	ParameterDeclaration
                        |	ParameterList ',' ParameterDeclaration

                        // [0]6.7.5
/*yy:case Abstract   */ ParameterDeclaration:
                        	DeclarationSpecifiers AbstractDeclaratorOpt
				{
					lx.scope.typedef = false
				}
/*yy:case Declarator */ |	DeclarationSpecifiers Declarator
				{
					lx.scope.typedef = false
				}

                        // [0]6.7.5
                        IdentifierList:
                        	IDENTIFIER
                        |	IdentifierList ',' IDENTIFIER

                        IdentifierListOpt:
                        	/* empty */ {}
                        |	IdentifierList

                        IdentifierOpt:
                        	/* empty */ {}
                        |	IDENTIFIER

                        // [0]6.7.6
			//yy:field	Type			Type
                        TypeName:
                        	SpecifierQualifierList AbstractDeclaratorOpt

                        // [0]6.7.6
			//yy:field	DeclarationSpecifier	*DeclarationSpecifier
			//yy:field	Type			Type
			//yy:field	TypeQualifiers		[]*TypeQualifier	// From the PointerOpt production, if any.
/*yy:case Pointer    */ AbstractDeclarator:
                        	Pointer
/*yy:case Abstract   */ |	PointerOpt DirectAbstractDeclarator

                        AbstractDeclaratorOpt:
                        	/* empty */ {}
                        |	AbstractDeclarator

                        // [0]6.7.6
/*yy:case Abstract   */ DirectAbstractDeclarator:
                        	'(' AbstractDeclarator ')'
/*yy:case ParamList  */ |	'(' ParameterTypeListOpt ')'
/*yy:case DFn        */ |	DirectAbstractDeclarator '(' ParameterTypeListOpt ')'
/*yy:case DArrSize   */ |	DirectAbstractDeclaratorOpt '[' "static" TypeQualifierListOpt Expr ']'
/*yy:case DArrVL     */ |	DirectAbstractDeclaratorOpt '[' '*' ']'
/*yy:case DArr       */ |	DirectAbstractDeclaratorOpt '[' ExprOpt ']'
/*yy:case DArrSize2  */ |	DirectAbstractDeclaratorOpt '[' TypeQualifierList "static" Expr ']'
/*yy:case DArr2      */ |	DirectAbstractDeclaratorOpt '[' TypeQualifierList ExprOpt ']'

                        DirectAbstractDeclaratorOpt:
                        	/* empty */ {}
                        |	DirectAbstractDeclarator

                        // [0]6.7.8
/*yy:case CompLit    */ Initializer:
                        	'{' InitializerList CommaOpt '}'
/*yy:case Expr       */ |	Expr

                        // [0]6.7.8
			//yy:field	Operand	Operand	//TODO-
			//yy:field	Len	int
                        InitializerList:
                        	/* empty */ {}
                        |	Initializer
                        |	Designation Initializer
                        |	InitializerList ',' Initializer
                        |	InitializerList ',' Designation Initializer

                        // [0]6.7.8
			//yy:field	List	[]int64
			//yy:field	Type	Type
                        Designation:
                        	DesignatorList '='

                        // [0]6.7.8
                        DesignatorList:
                        	Designator
                        |	DesignatorList Designator

                        // [0]6.7.8
/*yy:case Field      */ Designator:
                        	'.' IDENTIFIER
/*yy:case Index      */ |	'[' ConstExpr ']'

                        // [0]6.8
			//yy:field	UseGotos	bool
/*yy:case Block      */ Stmt:
				CompoundStmt
/*yy:case Expr       */ |	ExprStmt
/*yy:case Iter       */ |	IterationStmt
/*yy:case Jump       */ |	JumpStmt
/*yy:case Labeled    */ |	LabeledStmt
/*yy:case Select     */ |	SelectionStmt

                        // [0]6.8.1
			//yy:field	UseGotos	bool
/*yy:case SwitchCase */ LabeledStmt:
                        	"case" ConstExpr ':' Stmt
/*yy:case Default    */ |	"default" ':' Stmt
/*yy:case Label      */ |	IDENTIFIER ':' Stmt
				{
					lx.scope.insertLabel(lx.context, lhs)
				}
/*yy:case Label2      */ |	TYPEDEF_NAME ':' Stmt
				{
					lx.scope.insertLabel(lx.context, lhs)
				}

			statementEnd:
				{
					if s := lx.scope; s.forStmtEndScope != nil {
						switch yychar {
						case '}':
							var lval yySymType
							lx.lex(&lval)
							lval.Token.Rune = lx.toC(lval.Token.Rune, lval.Token.Val)
							lx.unget(cppToken{Token: lval.Token})
							switch lval.Token.Rune {
							case ELSE:
								// nop
							default:
								lx.scope = s.forStmtEndScope
							}
						case ELSE:
							// nop
						default:
							lx.scope = s.forStmtEndScope
						}
					}
				}

                        // [0]6.8.2
			//yy:field	scope	*Scope
			//yy:field	UseGotos	bool
                        CompoundStmt:
				'{'
				{
					lx.newScope()
					lx.insertParamNames()
				}
				BlockItemListOpt
				{
					lx.ssave, _ = lx.popScope()
				}
				statementEnd
				'}'
				{
					lhs.scope = lx.ssave
				}

                        // [0]6.8.2
                        BlockItemList:
                        	BlockItem
                        |	BlockItemList BlockItem

                        BlockItemListOpt:
                        	/* empty */ {}
                        |	BlockItemList

                        // [0]6.8.2
/*yy:case Decl       */ BlockItem:
                        	Declaration
/*yy:case Stmt       */ |	Stmt

                        // [0]6.8.3
			//yy:field	UseGotos	bool
                        ExprStmt:
                        	ExprListOpt statementEnd ';'

                        // [0]6.8.4
			//yy:field	Cases		[]*LabeledStmt
			//yy:field	SwitchOp	Operand	// Promoted switch operand
			//yy:field	UseGotos	bool
/*yy:case IfElse     */ SelectionStmt:
                        	"if" '(' ExprList ')' Stmt "else" Stmt
/*yy:case If         */ |	"if" '(' ExprList ')' Stmt %prec NOELSE
/*yy:case Switch     */ |	"switch" '(' ExprList ')' Stmt

                        // [0]6.8.5
			//yy:field	UseGotos	bool
/*yy:case Do         */ IterationStmt:
                        	"do" Stmt "while" '(' ExprList ')' statementEnd ';'
/*yy:case ForDecl    */ |	"for" '(' Declaration ExprListOpt ';' ExprListOpt ')' Stmt
/*yy:case For        */ |	"for" '(' ExprListOpt ';' ExprListOpt ';' ExprListOpt ')' Stmt
/*yy:case While      */ |	"while" '(' ExprList ')' Stmt

                        // [0]6.8.6
			//yy:field	ReturnOperand	Operand
			//yy:field	scope		*Scope
			//yy:field	UseGotos	bool
/*yy:case Break      */ JumpStmt:
                        	"break" statementEnd ';'
/*yy:case Continue   */ |	"continue" statementEnd ';'
/*yy:case Goto       */ |	"goto" IDENTIFIER statementEnd ';'
				{
					lhs.scope = lx.scope
				}
/*yy:case Return     */ |	"return" ExprListOpt statementEnd ';'

                        // [0]6.9
                        //yy:list
                        ExternalDeclarationList:
                        	ExternalDeclaration
                        |	ExternalDeclarationList ExternalDeclaration

                        // [0]6.9
/*yy:case Decl       */ ExternalDeclaration:
				Declaration
/*yy:case Func       */ |	FunctionDefinition

                        // [0]6.9.1
/*yy:case Spec       */	FunctionDefinition:
                                DeclarationSpecifiers Declarator
				{
					lx.scope.typedef = false
					lx.currFn = $2.(*Declarator)
				}
				DeclarationListOpt FunctionBody
				{
					lhs.Declarator.FunctionDefinition = lhs
					if lx.scope.Parent != nil {
						panic("internal error")
					}
				}
/*yy:case Int        */ |	Declarator
				{
					if !lx.tweaks.EnableOmitFuncDeclSpec {
						lx.err($1, "omitting function declaration specifiers not allowed")
					}
					lx.scope.typedef = false
					lx.currFn = $1.(*Declarator)
				}
				DeclarationListOpt FunctionBody
				{
					lhs.Declarator.FunctionDefinition = lhs
					if lx.scope.Parent != nil {
						panic("internal error")
					}
				}

			FunctionBody:
				{
					lx.declareFuncName() // [0], 6.4.2.2.
				}
				CompoundStmt

                        // [0]6.9.1
                        DeclarationList:
                        	Declaration
                        |	DeclarationList Declaration

                        DeclarationListOpt:
                        	/* empty */ {}
                        |	DeclarationList

                        VolatileOpt:
                        	/* empty */ {}
                        |	"volatile"
