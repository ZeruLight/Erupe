// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"bufio"
	"fmt"
	"go/token"
	"io"

	"modernc.org/golex/lex"
	"modernc.org/mathutil"
	"modernc.org/xc"
)

var (
	noTypedefNameAfter = map[rune]struct{}{
		'*':          {},
		'.':          {},
		ARROW:        {},
		BOOL:         {},
		CHAR:         {},
		COMPLEX:      {},
		DOUBLE:       {},
		ENUM:         {},
		FLOAT:        {},
		GOTO:         {},
		IDENTIFIER:   {},
		INT:          {},
		LONG:         {},
		SHORT:        {},
		SIGNED:       {},
		STRUCT:       {},
		TYPEDEF_NAME: {},
		UNION:        {},
		UNSIGNED:     {},
		VOID:         {},
	}
)

const (
	intBits  = mathutil.IntBits
	bitShift = intBits>>6 + 5
	bitMask  = intBits - 1

	scINITIAL = 0 // Start condition (shared value).
)

const (
	// Character class is an 8 bit encoding of an Unicode rune for the
	// golex generated FSM.
	//
	// Every ASCII rune is its own class.  DO NOT change any of the
	// existing values. Adding new classes is OK.
	ccEOF         = iota + 0x80
	_             // ccError
	ccOther       // Any other rune.
	ccUCNDigit    // [0], Annex D, Universal character names for identifiers - digits.
	ccUCNNonDigit // [0], Annex D, Universal character names for identifiers - non digits.
)

type trigraphs struct {
	*lex.Lexer
	pos token.Pos
	r   *bufio.Reader
	sc  int
}

func newTrigraphs(ctx *context, file *token.File, r io.Reader) (*trigraphs, error) {
	sc := scINITIAL
	if ctx.tweaks.EnableTrigraphs {
		sc = scTRIGRAPHS
	}
	t := &trigraphs{
		pos: file.Pos(0),
		r:   bufio.NewReader(r),
		sc:  sc,
	}
	lx, err := lex.New(
		file,
		t,
		lex.ErrorFunc(func(pos token.Pos, msg string) { ctx.errPos(pos, msg) }),
		lex.RuneClass(func(r rune) int { return int(r) }),
	)
	if err != nil {
		return nil, err
	}

	t.Lexer = lx
	return t, nil
}

func (t *trigraphs) ReadRune() (rune, int, error) { panic("internal error 9") }

func (t *trigraphs) ReadChar() (c lex.Char, size int, err error) {
	size = 1
	b, err := t.r.ReadByte()
	if err != nil {
		return lex.NewChar(t.pos, rune(b)), 0, err
	}

	c = lex.NewChar(t.pos, rune(b))
	t.pos++
	return c, 1, nil
}

type ungetBuffer []cppToken

func (u *ungetBuffer) unget(t cppToken) {
	*u = append(*u, t)
}

func (u *ungetBuffer) read() (t cppToken) {
	s := *u
	n := len(s) - 1
	t = s[n]
	*u = s[:n]
	return t
}

func (u *ungetBuffer) ungets(toks ...cppToken) {
	s := *u
	for i := len(toks) - 1; i >= 0; i-- {
		s = append(s, toks[i])
	}
	*u = s
}

type lexer struct {
	*context
	*lex.Lexer
	ast         Node
	attr        [][]xc.Token
	attr2       [][]xc.Token
	commentPos0 token.Pos
	currFn      *Declarator // [0]6.4.2.2
	last        lex.Char
	mode        int      // CONSTANT_EXPRESSION, TRANSLATION_UNIT
	prev        xc.Token // Most recent result returned by Lex
	sc          int
	ssave       *Scope
	t           *trigraphs
	tc          *tokenPipe

	noTypedefName bool // Do not consider next token a TYPEDEF_NAME
	typedef       bool // Prev token returned was TYPEDEF_NAME

	ungetBuffer
}

func newLexer(ctx *context, nm string, sz int, r io.Reader) (*lexer, error) {
	file := fset.AddFile(nm, -1, sz)
	t, err := newTrigraphs(ctx, file, r)
	if err != nil {
		return nil, err
	}

	l := &lexer{
		context: ctx,
		t:       t,
	}

	lx, err := lex.New(
		file,
		l,
		lex.ErrorFunc(func(pos token.Pos, msg string) { l.errPos(pos, msg) }),
		lex.RuneClass(rune2class),
	)
	if err != nil {
		return nil, err
	}

	l.Lexer = lx
	return l, nil
}

func (l *lexer) Error(msg string)             { l.err(l.First, "%v", msg) }
func (l *lexer) ReadRune() (rune, int, error) { panic("internal error 10") }
func (l *lexer) comment(general bool)         { /*TODO*/ }
func (l *lexer) parseExpr() bool              { return l.parse(CONSTANT_EXPRESSION) }

func (l *lexer) Lex(lval *yySymType) (r int) {
more:
	//TODO use follow set to recover from errors.
	l.lex(lval)
	lval.Token.Rune = l.toC(lval.Token.Rune, lval.Token.Val)
	typedef := l.typedef
	l.typedef = false
	noTypedefName := l.noTypedefName
	l.noTypedefName = false
	switch lval.Token.Rune {
	case '(':
		if l.prev.Rune == ATOMIC && l.prev.Pos()+token.Pos(len("_Atomic")) == lval.Token.Pos() {
			lval.Token.Rune = ATOMIC_LPAREN
		}
	case NON_REPL:
		lval.Token.Rune = IDENTIFIER
		fallthrough
	case IDENTIFIER:
		if lval.Token.Val == idAttribute {
			if len(l.attr) != 0 {
				panic(fmt.Errorf("%v:", l.position(lval.Token)))
			}

			l.attr = nil
			l.parseAttr(lval)
			goto more
		}

		if noTypedefName || typedef || !followSetHasTypedefName[lval.yys] {
			break
		}

		if _, ok := noTypedefNameAfter[l.prev.Rune]; ok {
			break
		}

		if l.scope.isTypedef(lval.Token.Val) {
			// https://en.wikipedia.org/wiki/The_lexer_hack
			lval.Token.Rune = TYPEDEF_NAME
			l.typedef = true
		}
	case PPNUMBER:
		lval.Token.Rune = INTCONST
		val := dict.S(lval.Token.Val)
		if !(len(val) > 1 && val[0] == '0' && (val[1] == 'x' || val[1] == 'X')) {
			for _, v := range val {
				switch v {
				case '.', '+', '-', 'e', 'E', 'p', 'P':
					lval.Token.Rune = FLOATCONST
				}
			}
		}
	case ccEOF:
		lval.Token.Rune = lex.RuneEOF
		lval.Token.Val = 0
	}

	if l.prev.Rune == FOR {
		s := l.scope.forStmtEndScope
		if s == nil {
			s = l.scope
		}
		l.newScope().forStmtEndScope = s
	}
	l.prev = lval.Token
	return int(l.prev.Rune)
}

func (l *lexer) attrs() (r [][]xc.Token) {
	l.attr, r = nil, l.attr
	return r
}

func (l *lexer) parseAttr(lval *yySymType) {
	l.lex(lval)
	if lval.Token.Rune != '(' {
		panic("TODO")
	}

	l.lex(lval)
	if lval.Token.Rune != '(' {
		panic("TODO")
	}

	l.parseAttrList(lval)
	l.lex(lval)
	if lval.Token.Rune != ')' {
		panic("TODO")
	}

	l.lex(lval)
	if lval.Token.Rune != ')' {
		panic("TODO")
	}
}

func (l *lexer) parseAttrList(lval *yySymType) {
	for {
		l.lex(lval)
		switch t := lval.Token; t.Rune {
		case IDENTIFIER:
			l.attr = append(l.attr, []xc.Token{t})
		case ')':
			l.unget(cppToken{Token: t})
			return
		case '(':
			l.parseAttrParams(lval)
		case ',':
			// ok
		default:
			panic(fmt.Errorf("%v: %v", l.position(lval.Token), PrettyString(lval.Token)))
		}
	}
}

func (l *lexer) parseAttrParams(lval *yySymType) {
	for {
		l.lex(lval)
		switch t := lval.Token; t.Rune {
		case IDENTIFIER, STRINGLITERAL:
			n := len(l.attr)
			l.attr[n-1] = append(l.attr[n-1], t)
		case ')':
			return
		default:
			panic(fmt.Errorf("%v: %v", l.position(lval.Token), PrettyString(lval.Token)))
		}
	}
}

func (l *lexer) ReadChar() (c lex.Char, size int, err error) {
	if c = l.t.Lookahead(); c.Rune == lex.RuneEOF {
		return c, 0, io.EOF
	}

	ch := l.t.scan()
	return lex.NewChar(l.t.First.Pos(), rune(ch)), 1, nil
}

func (l *lexer) Reduced(rule, state int, lval *yySymType) (stop bool) {
	if rule != l.exampleRule {
		return false
	}

	switch x := lval.node.(type) {
	case interface {
		fragment() interface{}
	}:
		l.exampleAST = x.fragment()
	default:
		l.exampleAST = x
	}
	return true
}

func (l *lexer) cppScan() lex.Char {
again:
	r := l.scan()
	if r == ' ' && l.last.Rune == ' ' {
		goto again
	}

	l.last = lex.NewChar(l.First.Pos(), rune(r))
	return l.last
}

func (l *lexer) lex(lval *yySymType) {
	if len(l.ungetBuffer) != 0 {
		lval.Token = l.ungetBuffer.read().Token
		return
	}

	if l.tc != nil {
		lval.Token = l.tc.read().Token
		l.First = lval.Token.Char
		return
	}

	ch := l.scanChar()
	lval.Token = xc.Token{Char: ch}
	if _, ok := tokHasVal[ch.Rune]; ok {
		lval.Token = xc.Token{Char: ch, Val: dict.ID(l.TokenBytes(nil))}
	}
}

// static const char __func__[] = "function-name"; // [0], 6.4.2.2.
func (l *lexer) declareFuncName() {
	pos := l.First.Pos() // '{'
	l.ungets(
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, STATIC), Val: idStatic}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, CONST), Val: idConst}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, CHAR), Val: idChar}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, IDENTIFIER), Val: idFuncName}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, '[')}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, ']')}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, '=')}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, STRINGLITERAL), Val: dict.SID(`"` + string(dict.S(l.currFn.Name())) + `"`)}},
		cppToken{Token: xc.Token{Char: lex.NewChar(pos, ';')}},
	)
}

func (l *lexer) insertParamNames() {
	if l.currFn == nil {
		return
	}

	defer func() { l.currFn = nil }()

	fp := l.currFn.fpScope(l.context)
	if fp == nil {
		return
	}

	for k, v := range fp.typedefs {
		l.scope.insertTypedef(l.context, k, v)
	}
}

func (l *lexer) parse(mode int) bool {
	var tok xc.Token
	tok.Rune = rune(mode)
	l.ungetBuffer = append(l.ungetBuffer, cppToken{Token: tok})
	l.mode = mode
	l.last.Rune = '\n'
	return yyParse(l) == 0
}

func (l *lexer) scanChar() (c lex.Char) {
again:
	r := l.scan()
	if r == ' ' {
		goto again
	}

	l.last = lex.NewChar(l.First.Pos(), rune(r))
	switch r {
	case CONSTANT_EXPRESSION, TRANSLATION_UNIT:
		l.mode = r
	}
	return l.last
}

func (l *lexer) fixDeclarator(n Node) {
	if dd := n.(*DirectDeclarator); dd.Case == DirectDeclaratorParen {
		nm := dd.Declarator.Name()
		//dbg("removing %q from %p", dict.S(nm), l.scope.Parent)
		delete(l.scope.Parent.typedefs, nm)
		l.scope.fixDecl = nm
	}
}

func (l *lexer) postFixDeclarator(ctx *context) {
	if nm := l.scope.fixDecl; nm != 0 {
		//dbg("reinserting %q into %p", dict.S(nm), l.scope.Parent)
		l.scope.Parent.insertTypedef(ctx, nm, false)
	}
}
