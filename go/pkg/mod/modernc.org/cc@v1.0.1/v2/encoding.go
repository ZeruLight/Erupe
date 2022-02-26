// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

import (
	"encoding/binary"
	"go/token"
	"reflect"
	"strconv"

	"modernc.org/golex/lex"
	"modernc.org/ir"
	"modernc.org/strutil"
	"modernc.org/xc"
)

var (
	dict       = xc.Dict
	printHooks = strutil.PrettyPrintHooks{}
)

func init() {
	for k, v := range xc.PrintHooks {
		printHooks[k] = v
	}
	delete(printHooks, reflect.TypeOf(token.Pos(0)))
	lcRT := reflect.TypeOf(lex.Char{})
	lcH := func(f strutil.Formatter, v interface{}, prefix, suffix string) {
		c := v.(lex.Char)
		r := c.Rune
		s := yySymName(int(r))
		if x := s[0]; x >= '0' && x <= '9' {
			s = strconv.QuoteRune(r)
		}
		f.Format(prefix)
		f.Format("%s", s)
		f.Format(suffix)
	}

	printHooks[lcRT] = lcH
	printHooks[reflect.TypeOf(xc.Token{})] = func(f strutil.Formatter, v interface{}, prefix, suffix string) {
		t := v.(xc.Token)
		if (t == xc.Token{}) {
			return
		}

		lcH(f, t.Char, prefix, "")
		if s := t.S(); len(s) != 0 {
			f.Format(" %q", s)
		}
		f.Format(suffix)
	}
	for _, v := range []interface{}{
		(*ir.Float32Value)(nil),
		(*ir.Float64Value)(nil),
		(*ir.Int32Value)(nil),
		(*ir.Int64Value)(nil),
		(*ir.StringValue)(nil),
		DirectDeclaratorCase(0),
		ExprCase(0),
		Linkage(0),
		StorageDuration(0),
		TypeKind(0),
		ir.Linkage(0),
	} {
		printHooks[reflect.TypeOf(v)] = func(f strutil.Formatter, v interface{}, prefix, suffix string) {
			f.Format(prefix)
			f.Format("%v", v)
			f.Format(suffix)
		}
	}
}

var (
	nopos xc.Token

	// Null pointer, [0]6.3.2.3-3.
	Null = &ir.AddressValue{}

	idAsm                    = dict.SID("asm")
	idAttribute              = dict.SID("__attribute__")
	idBuiltinAlloca          = dict.SID("__builtin_alloca")
	idBuiltinClasifyType     = dict.SID("__builtin_classify_type")
	idBuiltinTypesCompatible = dict.SID("__builtin_types_compatible__") // Implements __builtin_types_compatible_p
	idBuiltinVaList          = dict.SID("__builtin_va_list")
	idChar                   = dict.SID("char")
	idConst                  = dict.SID("const")
	idDefine                 = dict.SID("define")
	idDefined                = dict.SID("defined")
	idElif                   = dict.SID("elif")
	idElse                   = dict.SID("else")
	idEndif                  = dict.SID("endif")
	idError                  = dict.SID("error")
	idFile                   = dict.SID("__FILE__")
	idFuncName               = dict.SID("__func__")
	idIf                     = dict.SID("if")
	idIfdef                  = dict.SID("ifdef")
	idIfndef                 = dict.SID("ifndef")
	idInclude                = dict.SID("include")
	idIncludeNext            = dict.SID("include_next")
	idLine                   = dict.SID("line")
	idLineMacro              = dict.SID("__LINE__")
	idMain                   = dict.SID("main")
	idOne                    = dict.SID("1")
	idPopMacro               = dict.SID("pop_macro")
	idPragma                 = dict.SID("pragma")
	idPtrdiffT               = dict.SID("ptrdiff_t")
	idPushMacro              = dict.SID("push_macro")
	idSizeT                  = dict.SID("size_t")
	idStatic                 = dict.SID("static")
	idUndef                  = dict.SID("undef")
	idVaArgs                 = dict.SID("__VA_ARGS__")
	idVaList                 = dict.SID("va_list")
	idWarning                = dict.SID("warning")
	idWcharT                 = dict.SID("wchar_t")
	idZero                   = dict.SID("0")

	protectedMacro = map[int]bool{
		idFile:      true,
		idLineMacro: true,
	}

	keywords = map[int]rune{
		dict.SID("_Alignas"):       ALIGNAS,
		dict.SID("_Alignof"):       ALIGNOF,
		dict.SID("_Atomic"):        ATOMIC,
		dict.SID("("):              ATOMIC_LPAREN,
		dict.SID("_Bool"):          BOOL,
		dict.SID("_Complex"):       COMPLEX,
		dict.SID("_Generic"):       GENERIC,
		dict.SID("_Imaginary"):     IMAGINARY,
		dict.SID("_Noreturn"):      NORETURN,
		dict.SID("_Static_assert"): STATIC_ASSERT,
		dict.SID("_Thread_local"):  THREAD_LOCAL,
		dict.SID("auto"):           AUTO,
		dict.SID("break"):          BREAK,
		dict.SID("case"):           CASE,
		dict.SID("char"):           CHAR,
		dict.SID("const"):          CONST,
		dict.SID("continue"):       CONTINUE,
		dict.SID("default"):        DEFAULT,
		dict.SID("do"):             DO,
		dict.SID("double"):         DOUBLE,
		dict.SID("else"):           ELSE,
		dict.SID("enum"):           ENUM,
		dict.SID("extern"):         EXTERN,
		dict.SID("float"):          FLOAT,
		dict.SID("for"):            FOR,
		dict.SID("goto"):           GOTO,
		dict.SID("if"):             IF,
		dict.SID("inline"):         INLINE,
		dict.SID("int"):            INT,
		dict.SID("long"):           LONG,
		dict.SID("register"):       REGISTER,
		dict.SID("restrict"):       RESTRICT,
		dict.SID("return"):         RETURN,
		dict.SID("short"):          SHORT,
		dict.SID("signed"):         SIGNED,
		dict.SID("sizeof"):         SIZEOF,
		dict.SID("static"):         STATIC,
		dict.SID("struct"):         STRUCT,
		dict.SID("switch"):         SWITCH,
		dict.SID("typedef"):        TYPEDEF,
		dict.SID("typeof"):         TYPEOF,
		dict.SID("union"):          UNION,
		dict.SID("unsigned"):       UNSIGNED,
		dict.SID("void"):           VOID,
		dict.SID("volatile"):       VOLATILE,
		dict.SID("while"):          WHILE,
	}

	tokConstVals = map[rune]int{
		ADDASSIGN:     dict.SID("+="),
		ALIGNAS:       dict.SID("_Alignas"),
		ALIGNOF:       dict.SID("_Alignof"),
		ANDAND:        dict.SID("&&"),
		ANDASSIGN:     dict.SID("&="),
		ARROW:         dict.SID("->"),
		ATOMIC:        dict.SID("_Atomic"),
		ATOMIC_LPAREN: dict.SID("("),
		AUTO:          dict.SID("auto"),
		BOOL:          dict.SID("_Bool"),
		BREAK:         dict.SID("break"),
		CASE:          dict.SID("case"),
		CHAR:          dict.SID("char"),
		COMPLEX:       dict.SID("_Complex"),
		CONST:         dict.SID("const"),
		CONTINUE:      dict.SID("continue"),
		DDD:           dict.SID("..."),
		DEC:           dict.SID("--"),
		DEFAULT:       dict.SID("default"),
		DIVASSIGN:     dict.SID("/="),
		DO:            dict.SID("do"),
		DOUBLE:        dict.SID("double"),
		ELSE:          dict.SID("else"),
		ENUM:          dict.SID("enum"),
		EQ:            dict.SID("=="),
		EXTERN:        dict.SID("extern"),
		FLOAT:         dict.SID("float"),
		FOR:           dict.SID("for"),
		GENERIC:       dict.SID("_Generic"),
		GEQ:           dict.SID(">="),
		GOTO:          dict.SID("goto"),
		IF:            dict.SID("if"),
		IMAGINARY:     dict.SID("_Imaginary"),
		INC:           dict.SID("++"),
		INLINE:        dict.SID("inline"),
		INT:           dict.SID("int"),
		LEQ:           dict.SID("<="),
		LONG:          dict.SID("long"),
		LSH:           dict.SID("<<"),
		LSHASSIGN:     dict.SID("<<="),
		MODASSIGN:     dict.SID("%="),
		MULASSIGN:     dict.SID("*="),
		NEQ:           dict.SID("!="),
		NORETURN:      dict.SID("_Noreturn"),
		ORASSIGN:      dict.SID("|="),
		OROR:          dict.SID("||"),
		PPPASTE:       dict.SID("##"),
		REGISTER:      dict.SID("register"),
		RESTRICT:      dict.SID("restrict"),
		RETURN:        dict.SID("return"),
		RSH:           dict.SID(">>"),
		RSHASSIGN:     dict.SID(">>="),
		SHORT:         dict.SID("short"),
		SIGNED:        dict.SID("signed"),
		SIZEOF:        dict.SID("sizeof"),
		STATIC:        dict.SID("static"),
		STATIC_ASSERT: dict.SID("_Static_assert"),
		STRUCT:        dict.SID("struct"),
		SUBASSIGN:     dict.SID("-="),
		SWITCH:        dict.SID("switch"),
		THREAD_LOCAL:  dict.SID("_Thread_local"),
		TYPEDEF:       dict.SID("typedef"),
		TYPEOF:        dict.SID("typeof"),
		UNION:         dict.SID("union"),
		UNSIGNED:      dict.SID("unsigned"),
		VOID:          dict.SID("void"),
		VOLATILE:      dict.SID("volatile"),
		WHILE:         dict.SID("while"),
		XORASSIGN:     dict.SID("^="),
	}

	tokHasVal = map[rune]struct{}{
		CHARCONST:         {},
		FLOATCONST:        {},
		IDENTIFIER:        {},
		INTCONST:          {},
		LONGCHARCONST:     {},
		LONGSTRINGLITERAL: {},
		NON_REPL:          {},
		PPNUMBER:          {},
		STRINGLITERAL:     {},
		TYPEDEF_NAME:      {},
	}

	followSetHasTypedefName = [len(yyParseTab)]bool{}

	classifyType = map[TypeKind]int{
		0:                 noTypeClass,
		Void:              voidTypeClass,
		Ptr:               pointerTypeClass,
		Char:              charTypeClass,
		SChar:             charTypeClass,
		UChar:             charTypeClass,
		Short:             integerTypeClass,
		UShort:            integerTypeClass,
		Int:               integerTypeClass,
		UInt:              integerTypeClass,
		Long:              integerTypeClass,
		ULong:             integerTypeClass,
		LongLong:          integerTypeClass,
		ULongLong:         integerTypeClass,
		Float:             realTypeClass,
		Double:            realTypeClass,
		LongDouble:        realTypeClass,
		Bool:              booleanTypeClass,
		FloatComplex:      complexTypeClass,
		DoubleComplex:     complexTypeClass,
		LongDoubleComplex: complexTypeClass,
		Struct:            recordTypeClass,
		Union:             unionTypeClass,
		Enum:              enumeralTypeClass,
		TypedefName:       noTypeClass,
		Function:          functionTypeClass,
		Array:             arrayTypeClass,
	}
)

func init() {
	for i, v := range yyFollow {
		for _, v := range v {
			if v == TYPEDEF_NAME {
				followSetHasTypedefName[i] = true
			}
		}
	}
}

func isUCNDigit(r rune) bool {
	return int(r) < len(ucnDigits)<<bitShift && ucnDigits[uint(r)>>bitShift]&(1<<uint(r&bitMask)) != 0
}

func isUCNNonDigit(r rune) bool {
	return int(r) < len(ucnNonDigits)<<bitShift && ucnNonDigits[uint(r)>>bitShift]&(1<<uint(r&bitMask)) != 0
}

func rune2class(r rune) (c int) {
	switch {
	case r == lex.RuneEOF:
		return ccEOF
	case r < 128:
		return int(r)
	case isUCNDigit(r):
		return ccUCNDigit
	case isUCNNonDigit(r):
		return ccUCNNonDigit
	default:
		return ccOther
	}
}

func decodeToken(b []byte, pos token.Pos) ([]byte, token.Pos, xc.Token) {
	r, n := binary.Uvarint(b)
	b = b[n:]
	d, n := binary.Uvarint(b)
	b = b[n:]
	np := pos + token.Pos(d)
	c := lex.NewChar(np, rune(r))
	var v uint64
	if _, ok := tokHasVal[c.Rune]; ok {
		v, n = binary.Uvarint(b)
		b = b[n:]
	}
	return b, np, xc.Token{Char: c, Val: int(v)}
}

// TokSrc returns t in its source form.
func TokSrc(t xc.Token) string {
	if x, ok := tokConstVals[t.Rune]; ok {
		return string(dict.S(x))
	}

	if _, ok := tokHasVal[t.Rune]; ok {
		return string(t.S())
	}

	return string(t.Rune)
}

// escape-sequence		{simple-sequence}|{octal-escape-sequence}|{hexadecimal-escape-sequence}|{universal-character-name}
// simple-sequence		\\['\x22?\\abfnrtv]
// octal-escape-sequence	\\{octal-digit}{octal-digit}?{octal-digit}?
// hexadecimal-escape-sequence	\\x{hexadecimal-digit}+
func decodeEscapeSequence(runes []rune) (rune, int) {
	if runes[0] != '\\' {
		panic("internal error")
	}

	r := runes[1]
	switch r {
	case '\'', '"', '?', '\\':
		return r, 2
	case 'a':
		return 7, 2
	case 'b':
		return 8, 2
	case 'f':
		return 12, 2
	case 'n':
		return 10, 2
	case 'r':
		return 13, 2
	case 't':
		return 9, 2
	case 'v':
		return 11, 2
	case 'x':
		v, n := 0, 2
	loop2:
		for _, r := range runes[2:] {
			switch {
			case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
				v = v<<4 | decodeHex(r)
				n++
			default:
				break loop2
			}
		}
		return -rune(v & 0xff), n
	case 'u', 'U':
		return decodeUCN(runes)
	}

	if r < '0' || r > '7' {
		panic("internal error")
	}

	v, n := 0, 1
loop:
	for _, r := range runes[1:] {
		switch {
		case r >= '0' && r <= '7':
			v = v<<3 | (int(r) - '0')
			n++
		default:
			break loop
		}
	}
	return -rune(v), n
}

func decodeHex(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r) - '0'
	default:
		x := int(r) &^ 0x20
		return x - 'A' + 10
	}
}

// universal-character-name	\\u{hex-quad}|\\U{hex-quad}{hex-quad}
func decodeUCN(runes []rune) (rune, int) {
	if runes[0] != '\\' {
		panic("internal error")
	}

	runes = runes[1:]
	switch runes[0] {
	case 'u':
		return rune(decodeHexQuad(runes[1:])), 6
	case 'U':
		return rune(decodeHexQuad(runes[1:])<<16 | decodeHexQuad(runes[5:])), 10
	default:
		panic("internal error")
	}
}

// hex-quad	{hexadecimal-digit}{hexadecimal-digit}{hexadecimal-digit}{hexadecimal-digit}
func decodeHexQuad(runes []rune) int {
	n := 0
	for _, r := range runes[:4] {
		n = n<<4 | decodeHex(r)
	}
	return n
}

// Values from GCC's typeclass.h
const (
	noTypeClass = iota - 1
	voidTypeClass
	integerTypeClass
	charTypeClass
	enumeralTypeClass
	booleanTypeClass
	pointerTypeClass
	referenceTypeClass
	offsetTypeClass
	realTypeClass
	complexTypeClass
	functionTypeClass
	methodTypeClass
	recordTypeClass
	unionTypeClass
	arrayTypeClass
	stringTypeClass
	langTypeClass
)
