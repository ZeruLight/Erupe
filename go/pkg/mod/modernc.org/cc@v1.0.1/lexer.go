// Copyright 2016 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc"

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"strings"

	"modernc.org/golex/lex"
	"modernc.org/xc"
)

// Lexer state
const (
	lsZero             = iota
	lsBOL              // Preprocessor: Beginning of line.
	lsDefine           // Preprocessor: Seen ^#define.
	lsSeekRParen       // Preprocessor: Seen ^#define identifier(
	lsTokens           // Preprocessor: Convert anything to PPOTHER until EOL.
	lsUndef            // Preprocessor: Seen ^#undef.
	lsConstExpr0       // Preprocessor: Parsing constant expression.
	lsConstExpr        // Preprocessor: Parsing constant expression.
	lsTranslationUnit0 //
	lsTranslationUnit  //
)

type trigraphsReader struct {
	*lex.Lexer           //
	pos0       token.Pos //
	sc         int       // Start condition.
}

func (t *trigraphsReader) ReadRune() (rune, int, error) { return lex.RuneEOF, 0, io.EOF }

func (t *trigraphsReader) ReadChar() (c lex.Char, size int, err error) {
	r := rune(t.scan())
	pos0 := t.pos0
	pos := t.Lookahead().Pos()
	t.pos0 = pos
	c = lex.NewChar(t.First.Pos(), r)
	return c, int(pos - pos0), nil
}

type byteReader struct {
	io.Reader
	b [1]byte
}

func (b *byteReader) ReadRune() (r rune, size int, err error) {
	if _, err = b.Read(b.b[:]); err != nil {
		return -1, 0, err
	}

	return rune(b.b[0]), 1, nil
}

type lexer struct {
	*lex.Lexer                             //
	ch                 chan []xc.Token     //
	commentPos0        token.Pos           //
	constExprToks      []xc.Token          //
	constantExpression *ConstantExpression //
	cpp                func([]xc.Token)    //
	encBuf             []byte              // PPTokens
	encBuf1            [30]byte            // Rune, position, optional value ID.
	encPos             token.Pos           // For delta pos encoding
	eof                lex.Char            //
	example            interface{}         //
	exampleRule        int                 //
	externs            map[int]*Declarator //
	file               *token.File         //
	finalNLInjected    bool                //
	fnDeclarator       *Declarator         //
	includePaths       []string            //
	injectFunc         []xc.Token          // [0], 6.4.2.2.
	iota               int64               //
	isPreprocessing    bool                //
	last               xc.Token            //
	model              *Model              //
	preprocessingFile  *PreprocessingFile  //
	report             *xc.Report          //
	sc                 int                 // Start condition.
	scope              *Bindings           //
	scs                int                 // Start condition stack.
	state              int                 // Lexer state.
	sysIncludePaths    []string            //
	t                  *trigraphsReader    //
	textLine           []xc.Token          //
	toC                bool                // Whether to translate preprocessor identifiers to reserved C words.
	tokLast            xc.Token            //
	tokPrev            xc.Token            //
	toks               []xc.Token          // Parsing preprocessor constant expression.
	translationUnit    *TranslationUnit    //
	tweaks             *tweaks             //

	fsm struct {
		comment int
		pos     token.Pos
		state   int
	}
}

func newLexer(nm string, sz int, r io.RuneReader, report *xc.Report, tweaks *tweaks, opts ...lex.Option) (*lexer, error) {
	file := fset.AddFile(nm, -1, sz)
	t := &trigraphsReader{}
	lx, err := lex.New(
		file,
		&byteReader{Reader: r.(io.Reader)},
		lex.ErrorFunc(func(pos token.Pos, msg string) {
			report.Err(pos, msg)
		}),
		lex.RuneClass(func(r rune) int { return int(r) }),
	)
	if err != nil {
		return nil, err
	}

	t.Lexer = lx
	t.pos0 = lx.Lookahead().Pos()
	if tweaks.enableTrigraphs {
		t.sc = scTRIGRAPHS
	}
	r = t

	scope := newBindings(nil, ScopeFile)
	lexer := &lexer{
		externs: map[int]*Declarator{},
		file:    file,
		report:  report,
		scope:   scope,
		scs:     -1, // Stack empty
		t:       t,
		tweaks:  tweaks,
	}
	if lexer.Lexer, err = lex.New(
		file,
		r,
		append(opts, lex.RuneClass(rune2class))...,
	); err != nil {
		return nil, err
	}

	return lexer, nil
}

func newSimpleLexer(cpp func([]xc.Token), report *xc.Report, tweaks *tweaks) *lexer {
	return &lexer{
		cpp:     cpp,
		externs: map[int]*Declarator{},
		report:  report,
		scope:   newBindings(nil, ScopeFile),
		tweaks:  tweaks,
	}
}

func (l *lexer) push(sc int) {
	if l.scs >= 0 { // Stack overflow.
		if l.sc != scDIRECTIVE || sc != scCOMMENT {
			panic("internal error")
		}

		// /*-style comment in a line starting with #
		l.pop()
	}

	l.scs = l.sc
	l.sc = sc
}

func (l *lexer) pop() {
	if l.scs < 0 { // Stack underflow
		panic("internal error")
	}
	l.sc = l.scs
	l.scs = -1 // Stack empty.
}

func (l *lexer) pushScope(kind Scope) (old *Bindings) {
	old = l.scope
	l.scope = newBindings(old, kind)
	l.scope.maxAlign = 1
	return old
}

func (l *lexer) popScope(tok xc.Token) (old, new *Bindings) {
	return l.popScopePos(tok.Pos())
}

func (l *lexer) popScopePos(pos token.Pos) (old, new *Bindings) {
	old = l.scope
	new = l.scope.Parent
	if new == nil {
		l.report.Err(pos, "cannot pop scope")
		return nil, old
	}

	l.scope = new
	return old, new
}

const (
	fsmZero = iota
	fsmHasComment
)

var genCommentLeader = []byte("/*")

func (l *lexer) comment(general bool) {
	if l.tweaks.comments != nil {
		b := l.TokenBytes(nil)
		pos := l.First.Pos()
		if general {
			pos = l.commentPos0
			b = append(genCommentLeader, b...)
		}
		if l.Lookahead().Rune == '\n' {
			b = append(b, '\n')
		}

		switch fsm := &l.fsm; fsm.state {
		case fsmHasComment:
			if pos == fsm.pos+token.Pos(len(dict.S(l.fsm.comment))) {
				fsm.comment = dict.ID(append(dict.S(fsm.comment), b...))
				break
			}

			fallthrough
		case fsmZero:
			fsm.state = fsmHasComment
			fsm.comment = dict.ID(b)
			fsm.pos = pos
		}
	}
}

func (l *lexer) scanChar() (c lex.Char) {
again:
	r := rune(l.scan())
	switch r {
	case ' ':
		if l.state != lsTokens || l.tokLast.Rune == ' ' {
			goto again
		}
	case '\n':
		if l.state == lsTokens {
			l.encodeToken(xc.Token{Char: lex.NewChar(l.First.Pos(), ' '), Val: idSpace})
		}
		l.state = lsBOL
		l.sc = scINITIAL
		l.scs = -1 // Stack empty
	case PREPROCESSING_FILE:
		l.state = lsBOL
		l.isPreprocessing = true
	case CONSTANT_EXPRESSION, TRANSLATION_UNIT: //TODO- CONSTANT_EXPRESSION, then must add some manual yy:examples.
		l.toC = true
	}

	fp := l.First.Pos()
	if l.fsm.state == fsmHasComment {
		switch {
		case r == '\n' && fp == l.fsm.pos+token.Pos(len(dict.S(l.fsm.comment)))-1:
			// keep going
		case r != '\n' && fp == l.fsm.pos+token.Pos(len(dict.S(l.fsm.comment))):
			l.tweaks.comments[fp] = dict.ID(bytes.TrimSpace(dict.S(l.fsm.comment)))
			l.fsm.state = fsmZero
		default:
			l.fsm.state = fsmZero
		}
	}

	return lex.NewChar(l.First.Pos(), r)
}

func (l *lexer) scanToken() (tok xc.Token) {
	switch l.state {
	case lsConstExpr0:
		tok = xc.Token{Char: lex.NewChar(0, CONSTANT_EXPRESSION)}
		l.state = lsConstExpr
	case lsConstExpr:
		if len(l.toks) == 0 {
			tok = xc.Token{Char: lex.NewChar(l.tokLast.Pos(), lex.RuneEOF)}
			break
		}

		tok = l.toks[0]
		l.toks = l.toks[1:]
	case lsTranslationUnit0:
		tok = xc.Token{Char: lex.NewChar(0, TRANSLATION_UNIT)}
		l.state = lsTranslationUnit
		l.toC = true
	case lsTranslationUnit:
	again:
		for len(l.textLine) == 0 {
			var ok bool
			if l.textLine, ok = <-l.ch; !ok {
				return xc.Token{Char: lex.NewChar(l.tokLast.Pos(), lex.RuneEOF)}
			}

			if l.cpp != nil {
				l.cpp(l.textLine)
			}
		}
		tok = l.textLine[0]
		l.textLine = l.textLine[1:]
		if tok.Rune == ' ' {
			goto again
		}

		tok = l.scope.lexerHack(tok, l.tokLast)
	default:
		c := l.scanChar()
		if c.Rune == ccEOF {
			c = lex.NewChar(c.Pos(), lex.RuneEOF)
			if l.isPreprocessing && l.last.Rune != '\n' && !l.finalNLInjected {
				l.finalNLInjected = true
				l.eof = c
				c.Rune = '\n'
				l.state = lsBOL
				return xc.Token{Char: c}
			}

			return xc.Token{Char: c}
		}

		val := 0
		if tokHasVal[c.Rune] {
			b := l.TokenBytes(nil)
			val = dict.ID(b)
			//TODO handle ID UCNs
			//TODO- chars := l.Token()
			//TODO- switch c.Rune {
			//TODO- case IDENTIFIER, IDENTIFIER_LPAREN:
			//TODO- 	b := l.TokenBytes(func(buf *bytes.Buffer) {
			//TODO- 		for i := 0; i < len(chars); {
			//TODO- 			switch c := chars[i]; {
			//TODO- 			case c.Rune == '$' && !l.tweaks.enableDlrInIdentifiers:
			//TODO- 				l.report.Err(c.Pos(), "identifier character set extension '$' not enabled")
			//TODO- 				i++
			//TODO- 			case c.Rune == '\\':
			//TODO- 				r, n := decodeUCN(chars[i:])
			//TODO- 				buf.WriteRune(r)
			//TODO- 				i += n
			//TODO- 			case c.Rune < 0x80: // ASCII
			//TODO- 				buf.WriteByte(byte(c.Rune))
			//TODO- 				i++
			//TODO- 			default:
			//TODO- 				panic("internal error")
			//TODO- 			}
			//TODO- 		}
			//TODO- 	})
			//TODO- 	val = dict.ID(b)
			//TODO- default:
			//TODO- 	panic("internal error: " + yySymName(int(c.Rune)))
			//TODO- }
		}
		tok = xc.Token{Char: c, Val: val}
		if !l.isPreprocessing {
			tok = l.scope.lexerHack(tok, l.tokLast)
		}
	}
	if l.toC {
		tok = toC(tok, l.tweaks)
	}
	l.tokPrev = l.tokLast
	l.tokLast = tok
	return tok
}

// Lex implements yyLexer
func (l *lexer) Lex(lval *yySymType) int {
	var tok xc.Token
	if x := l.injectFunc; l.exampleRule == 0 && len(x) != 0 {
		tok = x[0]
		l.injectFunc = x[1:]
	} else {
		tok = l.scanToken()
	}
	//dbg("Lex %s", PrettyString(tok))
	if l.constExprToks != nil {
		l.constExprToks = append(l.constExprToks, tok)
	}
	l.last = tok
	if tok.Rune == lex.RuneEOF {
		lval.Token = tok
		return 0
	}

	switch l.state {
	case lsBOL:
		switch tok.Rune {
		case PREPROCESSING_FILE, '\n':
			// nop
		case '#':
			l.push(scDIRECTIVE)
			tok = l.scanToken()
			switch tok.Rune {
			case '\n':
				tok.Char = lex.NewChar(tok.Pos(), PPHASH_NL)
			case PPDEFINE:
				l.push(scDEFINE)
				l.state = lsDefine
			case PPELIF, PPENDIF, PPERROR, PPIF, PPLINE, PPPRAGMA:
				l.sc = scINITIAL
				l.state = lsTokens
			case PPELSE, PPIFDEF, PPIFNDEF:
				l.state = lsZero
			case PPUNDEF:
				l.state = lsUndef
			case PPINCLUDE:
				l.sc = scHEADER
				l.state = lsTokens
			case PPINCLUDE_NEXT:
				if l.tweaks.enableIncludeNext {
					l.sc = scHEADER
					l.state = lsTokens
					break
				}

				l.state = lsTokens
				tok.Char = lex.NewChar(tok.Pos(), PPNONDIRECTIVE)
				tok.Val = xc.Dict.SID("include_next")
			default:
				l.state = lsTokens
				tok.Char = lex.NewChar(tok.Pos(), PPNONDIRECTIVE)
				l.pop()
			}
		default:
			l.encodeToken(tok)
			tok.Char = lex.NewChar(tok.Pos(), PPOTHER)
			l.state = lsTokens
		}
	case lsDefine:
		l.pop()
		switch tok.Rune {
		case IDENTIFIER:
			l.state = lsTokens
		case IDENTIFIER_LPAREN:
			l.state = lsSeekRParen
		default:
			l.state = lsZero
		}
	case lsSeekRParen:
		if tok.Rune == ')' {
			l.state = lsTokens
		}
	case lsTokens:
		l.encodeToken(tok)
		tok.Char = lex.NewChar(tok.Pos(), PPOTHER)
	case lsUndef:
		l.state = lsTokens
	}

	lval.Token = tok
	return int(tok.Char.Rune)
}

// Error Implements yyLexer.
func (l *lexer) Error(msg string) {
	msg = strings.Replace(msg, "$end", "EOF", -1)
	t := l.last
	parts := strings.Split(msg, ", expected ")
	if len(parts) == 2 && strings.HasPrefix(parts[0], "unexpected ") && tokHasVal[t.Rune] {
		msg = fmt.Sprintf("%s %s, expected %s", parts[0], t.S(), parts[1])
	}
	l.report.ErrTok(t, "%s", msg)
}

// Reduced implements yyLexerEx
func (l *lexer) Reduced(rule, state int, lval *yySymType) (stop bool) {
	if n := l.exampleRule; n >= 0 && rule != n {
		return false
	}

	switch x := lval.node.(type) {
	case interface {
		fragment() interface{}
	}:
		l.example = x.fragment()
	default:
		l.example = x
	}
	return true
}

func (l *lexer) parsePPConstExpr0(list PPTokenList, p *pp) (interface{}, Type) {
	l.toks = l.toks[:0]
	p.expand(&tokenBuf{decodeTokens(list, nil, true)}, true, func(toks []xc.Token) {
		l.toks = append(l.toks, toks...)
	})
	w := 0
	for _, tok := range l.toks {
		switch tok.Rune {
		case ' ':
			// nop
		case IDENTIFIER:
			if p.macros.m[tok.Val] != nil {
				l.report.ErrTok(tok, "expected constant expression")
				return nil, nil
			}

			tok.Rune = INTCONST
			tok.Val = id0
			fallthrough
		default:
			l.toks[w] = tok
			w++
		}
	}
	l.toks = l.toks[:w]
	l.state = lsConstExpr0
	if yyParse(l) == 0 {
		e := l.constantExpression
		return e.Value, e.Type
	}

	return nil, nil
}

func (l *lexer) parsePPConstExpr(list PPTokenList, p *pp) bool {
	if v, _ := l.parsePPConstExpr0(list, p); v != nil {
		return isNonZero(v)
	}

	return false
}
