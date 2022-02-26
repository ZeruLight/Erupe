// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
// [1]: https://www.spinellis.gr/blog/20060626/cpp.algo.pdf

package cc // import "modernc.org/cc/v2"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go/token"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"modernc.org/golex/lex"
	"modernc.org/ir"
	"modernc.org/mathutil"
	"modernc.org/xc"
)

const (
	maxIncludeLevel = 200 // gcc, std is at least 15.
)

var (
	_ tokenReader = (*cppReader)(nil)
	_ tokenReader = (*tokenBuffer)(nil)
	_ tokenReader = (*tokenPipe)(nil)
	_ tokenWriter = (*tokenBuffer)(nil)
	_ tokenWriter = (*tokenPipe)(nil)
)

type cppToken struct {
	xc.Token
	hs map[int]struct{}
}

func (t *cppToken) has(nm int) bool { _, ok := t.hs[nm]; return ok }

func (t *cppToken) cloneAdd(nm int) map[int]struct{} {
	nhs := map[int]struct{}{nm: {}}
	for k, v := range t.hs {
		nhs[k] = v
	}
	return nhs
}

func (t *cppToken) hsAdd(hs map[int]struct{}) {
	if len(hs) == 0 {
		return
	}

	if len(t.hs) == 0 {
		t.hs = map[int]struct{}{}
	}
	for k := range hs {
		t.hs[k] = struct{}{}
	}
}

type tokenWriter interface {
	write(...cppToken)
}

type tokenReader interface {
	read() cppToken
	unget(cppToken)
	ungets(...cppToken)
}

type tokenPipe struct {
	b  []byte
	ch chan cppToken
	s  []cppToken

	emitWhiteSpace bool
}

func newTokenPipe(n int) *tokenPipe { return &tokenPipe{ch: make(chan cppToken, n)} }

func (*tokenPipe) unget(cppToken)     { panic("internal error") }
func (*tokenPipe) ungets(...cppToken) { panic("internal error") }

func (p *tokenPipe) close() {
	if len(p.s) != 0 {
		p.flush()
	}
	close(p.ch)
}

func (p *tokenPipe) flush() {
	p.b = p.b[:0]
	p.b = append(p.b, '"')
	for _, t := range p.s {
		s := dict.S(t.Val)
		p.b = append(p.b, s[1:len(s)-1]...)
	}
	p.b = append(p.b, '"')
	p.s[0].Val = dict.ID(p.b)
	p.ch <- p.s[0]
	p.s = p.s[:0]
}

func (p *tokenPipe) read() cppToken {
	t, ok := <-p.ch
	if !ok {
		t.Rune = ccEOF
	}
	return t
}

func (p *tokenPipe) write(toks ...cppToken) {
	for _, t := range toks {
		switch t.Rune {
		case '\n', ' ':
			if p.emitWhiteSpace {
				p.ch <- t
			}
		case STRINGLITERAL, LONGSTRINGLITERAL:
			p.s = append(p.s, t)
		default:
			if len(p.s) != 0 {
				p.flush()
			}
			p.ch <- t
		}
	}
}

type tokenBuffer struct {
	toks0 []cppToken
	toks  []cppToken
	ungetBuffer

	last rune
}

func (b *tokenBuffer) write(t ...cppToken) {
	b.toks = append(b.toks, t...)
	if b.toks0 == nil || &b.toks0[0] != &b.toks[0] {
		b.toks0 = b.toks
	}
}

func (b *tokenBuffer) read() (t cppToken) {
	if len(b.ungetBuffer) != 0 {
		return b.ungetBuffer.read()
	}

	if len(b.toks) == 0 {
		t.Rune = ccEOF
		return
	}

	t = b.toks[0]
	b.toks = b.toks[1:]
	if len(b.toks) == 0 {
		b.toks = b.toks0[:0]
	}
	if t.Rune == '#' && (b.last == '\n' || b.last == 0) {
		t.Rune = DIRECTIVE
	}
	b.last = t.Rune
	return t
}

type cppReader struct {
	decBuf []byte
	decPos token.Pos
	tu     [][]uint32
	ungetBuffer

	last rune
}

func (c *cppReader) unget(t cppToken) { c.ungetBuffer = append(c.ungetBuffer, t) }

func (c *cppReader) read() (t cppToken) {
	if len(c.ungetBuffer) != 0 {
		return c.ungetBuffer.read()
	}

more:
	if len(c.decBuf) == 0 {
		if len(c.tu) == 0 {
			t.Rune = ccEOF
			return t
		}

		if len(c.tu[0]) == 0 {
			c.tu = c.tu[1:]
			goto more
		}

		c.decBuf = dict.S(int(c.tu[0][0]))
		c.tu[0] = c.tu[0][1:]
		c.decPos = 0
	}

	c.decBuf, c.decPos, t.Token = decodeToken(c.decBuf, c.decPos)
	if t.Rune == '#' && (c.last == '\n' || c.last == 0) {
		t.Rune = DIRECTIVE
	}
	c.last = t.Rune
	return t
}

type conds []cond

func (c conds) on() bool          { return condOn[c.tos()] }
func (c conds) pop() conds        { return c[:len(c)-1] }
func (c conds) push(n cond) conds { return append(c, n) }
func (c conds) tos() cond         { return c[len(c)-1] }

// Macro represents a preprocessor Macro.
type Macro struct {
	Args            []int      // Numeric IDs of argument identifiers.
	DefTok          xc.Token   // Macro name definition token.
	ReplacementToks []xc.Token // The tokens that replace the macro. R/O

	IsFnLike   bool // Whether the macro is function like.
	IsVariadic bool // Whether the macro is variadic.
	ident      bool
}

func newMacro(def xc.Token, repl []xc.Token) *Macro {
	return &Macro{DefTok: def, ReplacementToks: append([]xc.Token(nil), repl...)}
}

// Eval attempts to evaluate m, which must be a simple macro, like `#define foo numeric-literal`.
func (m *Macro) Eval(model Model, macros map[int]*Macro) (op Operand, err error) {
	returned := false

	defer func() {
		e := recover()
		if !returned && err == nil {
			err = fmt.Errorf("PANIC: %v\n%s", e, debugStack())
		}
	}()

	if m.IsFnLike {
		return op, fmt.Errorf("cannot evaluate function-like macro")
	}

	ctx, err := newContext(&Tweaks{})
	if err != nil {
		return op, err
	}

	ctx.model = model
	c := newCPP(ctx)
	c.macros = macros
	if op, _ = c.constExpr(cppToks(m.ReplacementToks), false); op.Type == nil {
		return op, fmt.Errorf("cannot evaluate macro")
	}

	returned = true
	return op, nil
}

func (m *Macro) param(ap [][]cppToken, nm int, out *[]cppToken) bool {
	*out = nil
	if nm == idVaArgs {
		if !m.IsVariadic {
			return false
		}

		if i := len(m.Args); i < len(ap) {
			o := *out
			for i, v := range ap[i:] {
				if i != 0 {
					switch lo := len(o); lo {
					case 0:
						var t cppToken
						t.Rune = ','
						t.Val = 0
						o = append(o, t)
					default:
						t := o[len(o)-1]
						t.Rune = ','
						t.Val = 0
						o = append(o, t)
						t.Rune = ' '
						o = append(o, t)
					}
				}
				o = append(o, v...)
			}
			*out = o
		}
		return true
	}

	if len(m.Args) != 0 && nm == m.Args[len(m.Args)-1] && m.IsVariadic && !m.ident {
		if i := len(m.Args) - 1; i < len(ap) {
			o := *out
			for i, v := range ap[i:] {
				if i != 0 {
					switch lo := len(o); lo {
					case 0:
						var t cppToken
						t.Rune = ','
						t.Val = 0
						o = append(o, t)
					default:
						t := o[len(o)-1]
						t.Rune = ','
						t.Val = 0
						o = append(o, t)
						t.Rune = ' '
						o = append(o, t)
					}
				}
				o = append(o, v...)
			}
			*out = o
		}
		return true
	}

	for i, v := range m.Args {
		if v == nm {
			*out = ap[i]
			return true
		}
	}
	return false
}

type nullReader struct{}

func (nullReader) Read([]byte) (int, error) { return 0, io.EOF }

type cpp struct {
	*context
	includeLevel int
	lx           *lexer
	macroStack   map[int][]*Macro
	macros       map[int]*Macro // name ID: macro
	toks         []cppToken
}

func newCPP(ctx *context) *cpp {
	lx, err := newLexer(ctx, "", 0, nullReader{})
	if err != nil {
		panic(err)
	}

	lx.context = ctx
	r := &cpp{
		context:    ctx,
		lx:         lx,
		macroStack: map[int][]*Macro{},
		macros:     map[int]*Macro{},
	}
	return r
}

func (c *cpp) parse(src ...Source) (tokenReader, error) {
	var (
		encBuf  []byte
		encBuf1 [30]byte // Rune, position, optional value ID.
		tokBuf  []cppToken
		tu      [][]uint32
	)
	for _, v := range src {
		if pf := v.Cached(); pf != nil {
			tu = append(tu, pf)
			continue
		}

		sz, err := v.Size()
		if err != nil {
			return nil, err
		}

		if sz > mathutil.MaxInt {
			return nil, fmt.Errorf("%v: file too big: %v", v.Name(), sz)
		}

		r, err := v.ReadCloser()
		if err != nil {
			return nil, err
		}

		lx, err := newLexer(c.context, v.Name(), int(sz), r)
		if err != nil {
			return nil, err
		}

		if err := func() (err error) {
			returned := false

			defer func() {
				e := recover()
				if !returned && err == nil {
					err = fmt.Errorf("PANIC: %v\n%s", e, debugStack())
					c.err(nopos, "%v", err)
				}
				if e := r.Close(); e != nil && err == nil {
					err = e
				}
			}()

			var pf []uint32
			var t cppToken
			var toks []cppToken
			for {
				ch := lx.cppScan()
				if ch.Rune == ccEOF {
					break
				}

				tokBuf = tokBuf[:0]
				for {
					t.Char = ch
					t.Val = 0
					if ch.Rune == '\n' {
						toks = append(cppTrimSpace(tokBuf), t)
						break
					}

					if _, ok := tokHasVal[ch.Rune]; ok {
						t.Val = dict.ID(lx.TokenBytes(nil))
					}
					tokBuf = append(tokBuf, t)

					if ch = lx.cppScan(); ch.Rune == ccEOF {
						if !c.tweaks.InjectFinalNL {
							c.errPos(lx.last.Pos(), "file is missing final newline")
						}
						ch.Rune = '\n'
					}
				}

				var encPos token.Pos
				encBuf = encBuf[:0]
				for _, t := range toks {
					n := binary.PutUvarint(encBuf1[:], uint64(t.Rune))
					pos := t.Pos()
					n += binary.PutUvarint(encBuf1[n:], uint64(pos-encPos))
					encPos = pos
					if t.Val != 0 {
						n += binary.PutUvarint(encBuf1[n:], uint64(t.Val))
					}
					encBuf = append(encBuf, encBuf1[:n]...)
				}
				id := dict.ID(encBuf)
				if int64(id) > math.MaxUint32 {
					panic("internal error 4")
				}

				pf = append(pf, uint32(id))
			}
			v.Cache(pf)
			tu = append(tu, pf)
			returned = true
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	return &cppReader{tu: tu}, nil
}
func (c *cpp) eval(r tokenReader, w tokenWriter) (err error) {
	c.macros[idFile] = &Macro{ReplacementToks: []xc.Token{{Char: lex.NewChar(0, STRINGLITERAL)}}}
	c.macros[idLineMacro] = &Macro{ReplacementToks: []xc.Token{{Char: lex.NewChar(0, INTCONST)}}}
	if cs := c.expand(r, w, conds(nil).push(condZero), 0, false); len(cs) != 1 || cs.tos() != condZero {
		return fmt.Errorf("unexpected top of condition stack value: %v", cs)
	}

	return nil
}

// [1]pg 1.
//
// expand(TS ) /* recur, substitute, pushback, rescan */
// {
// 	if TS is {} then
//		// ---------------------------------------------------------- A
// 		return {};
//
// 	else if TS is T^HS • TS’ and T is in HS then
//		//----------------------------------------------------------- B
// 		return T^HS • expand(TS’);
//
// 	else if TS is T^HS • TS’ and T is a "()-less macro" then
//		// ---------------------------------------------------------- C
// 		return expand(subst(ts(T), {}, {}, HS \cup {T}, {}) • TS’ );
//
// 	else if TS is T^HS •(•TS’ and T is a "()’d macro" then
//		// ---------------------------------------------------------- D
// 		check TS’ is actuals • )^HS’ • TS’’ and actuals are "correct for T"
// 		return expand(subst(ts(T), fp(T), actuals,(HS \cap HS’) \cup {T }, {}) • TS’’);
//
//	// ------------------------------------------------------------------ E
// 	note TS must be T^HS • TS’
// 	return T^HS • expand(TS’);
// }
func (c *cpp) expand(r tokenReader, w tokenWriter, cs conds, lvl int, expandDefined bool) conds {
	for {
		t := r.read()
		switch t.Rune {
		// First, if TS is the empty set, the result is the
		// empty set.
		case ccEOF:
			// -------------------------------------------------- A
			// 		return {};
			return cs
		case DIRECTIVE:
			cs = c.directive(r, w, cs)
			t.Rune = '\n'
			t.Val = 0
			w.write(t)
		case IDENTIFIER:
			if !cs.on() {
				break
			}

			nm := t.Val
			if nm == idDefined && expandDefined {
			more:
				switch t = r.read(); t.Rune {
				case ccEOF:
					panic("TODO")
				case IDENTIFIER:
					nm := t.Val
					t.Rune = INTCONST
					t.Val = idZero
					if _, ok := c.macros[nm]; ok {
						t.Val = idOne
					}
					w.write(t)
					continue
				case ' ':
					goto more
				case '(': // defined(name)
					var u cppToken
					switch t = r.read(); t.Rune {
					case ccEOF:
						panic("TODO")
					case IDENTIFIER:
						nm := t.Val
						u = t
						u.Rune = INTCONST
						u.Val = idZero
						if _, ok := c.macros[nm]; ok {
							u.Val = idOne
						}
					more2:
						switch t = r.read(); t.Rune {
						case ccEOF:
							panic("TODO")
						case ' ':
							goto more2
						case ')':
							// ok done
							w.write(u)
							continue
						default:
							panic(t.String())
						}
					default:
						panic(t.String())
					}
				default:
					panic(t.String())
				}
			}

			// Otherwise, if the token sequence begins with a token
			// whose hide set contains that token, then the result
			// is the token sequence beginning with that token
			// (including its hide set) followed by the result of
			// expand on the rest of the token sequence.
			if t.has(nm) {
				// ------------------------------------------ B
				// 		return T^HS • expand(TS’);
				w.write(t)
				continue
			}

			m := c.macros[nm]
			if m != nil && !m.IsFnLike {
				// Otherwise, if the token sequence begins with
				// an object-like macro, the result is the
				// expansion of the rest of the token sequence
				// beginning with the sequence returned by
				// subst invoked with the replacement token
				// sequence for the macro, two empty sets, the
				// union of the macro’s hide set and the macro
				// itself, and an empty set.
				switch nm {
				case idFile:
					m.ReplacementToks[0].Val = dict.SID(fmt.Sprintf("%q", c.position(t).Filename))
				case idLineMacro:
					m.ReplacementToks[0].Val = dict.SID(fmt.Sprint(c.position(t).Line))
				}
				// ------------------------------------------ C
				// 		return expand(subst(ts(T), {}, {}, HS \cup {T}, {}) • TS’ );
				toks := c.subst(m, nil, t.cloneAdd(nm), expandDefined)
				for i, v := range toks {
					toks[i].Char = lex.NewChar(t.Pos(), v.Rune)
				}
				r.ungets(toks...)
				continue
			}

			if m != nil && m.IsFnLike {
				// ------------------------------------------ D
				// 		check TS’ is actuals • )^HS’ • TS’’ and actuals are "correct for T"
				// 		return expand(subst(ts(T), fp(T), actuals,(HS \cap HS’) \cup {T }, {}) • TS’’);
				hs := t.hs
			again:
				switch t2 := r.read(); t2.Rune {
				case '\n', ' ':
					goto again
				case '(':
					// ok
				case ccEOF:
					w.write(t)
					continue
				default:
					w.write(t)
					w.write(t2)
					continue
				}

				ap, hs2 := c.actuals(m, r)
				switch {
				case len(hs2) == 0:
					hs2 = map[int]struct{}{nm: {}}
				default:
					nhs := map[int]struct{}{}
					for k := range hs {
						if _, ok := hs2[k]; ok {
							nhs[k] = struct{}{}
						}
					}
					nhs[nm] = struct{}{}
					hs2 = nhs
				}
				toks := c.subst(m, ap, hs2, expandDefined)
				for i, v := range toks {
					toks[i].Char = lex.NewChar(t.Pos(), v.Rune)
				}
				r.ungets(toks...)
				continue
			}

			w.write(t)
		default:
			// -------------------------------------------------- E
			if !cs.on() {
				break
			}

			w.write(t)
		}
	}
}

func (c *cpp) pragmaActuals(nd Node, line []cppToken) (out []cppToken) {
	first := true
	for {
		if len(line) == 0 {
			c.err(nd, "unexpected EOF")
			return nil
		}

		t := line[0]
		line = line[1:]
		switch t.Rune {
		case '(':
			if !first {
				panic(fmt.Errorf("%v", t))
			}

			first = false
		case STRINGLITERAL:
			out = append(out, t)
		case ')':
			return out
		default:
			panic(fmt.Errorf("%v: %v (%v)", c.position(t), t, yySymName(int(t.Rune))))
		}
	}
}

func (c *cpp) actuals(m *Macro, r tokenReader) (out [][]cppToken, hs map[int]struct{}) {
	var lvl, n int
	for {
		t := r.read()
		if t.Rune < 0 {
			c.err(t, "unexpected EOF")
			return nil, nil
		}

		switch t.Rune {
		case ',':
			if lvl == 0 {
				n++
				continue
			}
		case ')':
			if lvl == 0 {
				for i, v := range out {
					out[i] = cppTrimSpace(v)
				}
				for len(out) < len(m.Args) {
					out = append(out, nil)
				}
				return out, t.hs
			}

			lvl--
		case '(':
			lvl++
		}

		for len(out) <= n {
			out = append(out, []cppToken{})
		}
		if t.Rune == '\n' {
			t.Rune = ' '
		}
		out[n] = append(out[n], t)
	}
}

func (c *cpp) expands(toks []cppToken, expandDefined bool) (out []cppToken) {
	var r, w tokenBuffer
	r.toks = toks
	c.expand(&r, &w, conds(nil).push(condZero), 1, expandDefined)
	return w.toks
}

// [1]pg 2.
//
// subst(IS, FP, AP, HS, OS) /* substitute args, handle stringize and paste */
// {
// 	if IS is {} then
//		// ---------------------------------------------------------- A
// 		return hsadd(HS, OS);
//
// 	else if IS is # • T • IS’ and T is FP[i] then
//		// ---------------------------------------------------------- B
// 		return subst(IS’, FP, AP, HS, OS • stringize(select(i, AP)));
//
// 	else if IS is ## • T • IS’ and T is FP[i] then
//	{
//		// ---------------------------------------------------------- C
// 		if select(i, AP) is {} then /* only if actuals can be empty */
//			// -------------------------------------------------- D
// 			return subst(IS’, FP, AP, HS, OS);
// 		else
//			// -------------------------------------------------- E
// 			return subst(IS’, FP, AP, HS, glue(OS, select(i, AP)));
// 	}
//
// 	else if IS is ## • T^HS’ • IS’ then
//		// ---------------------------------------------------------- F
// 		return subst(IS’, FP, AP, HS, glue(OS, T^HS’));
//
// 	else if IS is T • ##^HS’ • IS’ and T is FP[i] then
//	{
//		// ---------------------------------------------------------- G
// 		if select(i, AP) is {} then /* only if actuals can be empty */
//		{
//			// -------------------------------------------------- H
// 			if IS’ is T’ • IS’’ and T’ is FP[j] then
//				// ------------------------------------------ I
// 				return subst(IS’’, FP, AP, HS, OS • select(j, AP));
// 			else
//				// ------------------------------------------ J
// 				return subst(IS’, FP, AP, HS, OS);
// 		}
//		else
//			// -------------------------------------------------- K
// 			return subst(##^HS’ • IS’, FP, AP, HS, OS • select(i, AP));
//
//	}
//
// 	else if IS is T • IS’ and T is FP[i] then
//		// ---------------------------------------------------------- L
// 		return subst(IS’, FP, AP, HS, OS • expand(select(i, AP)));
//
//	// ------------------------------------------------------------------ M
// 	note IS must be T^HS’ • IS’
// 	return subst(IS’, FP, AP, HS, OS • T^HS’);
// }
//
// A quick overview of subst is that it walks through the input sequence, IS,
// building up an output sequence, OS, by handling each token from left to
// right. (The order that this operation takes is left to the implementation
// also, walking from left to right is more natural since the rest of the
// algorithm is constrained to this ordering.) Stringizing is easy, pasting
// requires trickier handling because the operation has a bunch of
// combinations. After the entire input sequence is finished, the updated hide
// set is applied to the output sequence, and that is the result of subst.
func (c *cpp) subst(m *Macro, ap [][]cppToken, hs map[int]struct{}, expandDefined bool) (out []cppToken) {
	// dbg("%s %v %v", m.def.S(), m.variadic, ap)
	repl := cppToks(m.ReplacementToks)
	var arg []cppToken
	for {
		if len(repl) == 0 {
			// -------------------------------------------------- A
			// 		return hsadd(HS, OS);
			out := cppTrimSpace(out)
			for i := range out {
				out[i].hsAdd(hs)
			}
			return out
		}

		if repl[0].Rune == '#' && len(repl) > 1 && repl[1].Rune == IDENTIFIER && m.param(ap, repl[1].Val, &arg) {
			// -------------------------------------------------- B
			// 		return subst(IS’, FP, AP, HS, OS • stringize(select(i, AP)));
			out = append(out, c.stringize(arg))
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == '#' && len(repl) > 2 && repl[1].Rune == ' ' && repl[2].Rune == IDENTIFIER && m.param(ap, repl[2].Val, &arg) {
			// -------------------------------------------------- B
			// 		return subst(IS’, FP, AP, HS, OS • stringize(select(i, AP)));
			out = append(out, c.stringize(arg))
			repl = repl[3:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 1 && repl[1].Rune == IDENTIFIER && m.param(ap, repl[1].Val, &arg) {
			// -------------------------------------------------- C
			if len(arg) == 0 {
				// ------------------------------------------ D
				// 			return subst(IS’, FP, AP, HS, OS);
				repl = repl[2:]
				continue
			}

			// -------------------------------------------------- E
			// 			return subst(IS’, FP, AP, HS, glue(OS, select(i, AP)));
			_, out = c.glue(out, arg)
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 2 && repl[1].Rune == ' ' && repl[2].Rune == IDENTIFIER && m.param(ap, repl[2].Val, &arg) {
			// -------------------------------------------------- C
			if len(arg) == 0 {
				// ------------------------------------------ D
				// 			return subst(IS’, FP, AP, HS, OS);
				repl = repl[3:]
				continue
			}

			// -------------------------------------------------- E
			// 			return subst(IS’, FP, AP, HS, glue(OS, select(i, AP)));
			_, out = c.glue(out, arg)
			repl = repl[3:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 1 && repl[1].Rune != ' ' {
			// -------------------------------------------------- F
			// 		return subst(IS’, FP, AP, HS, glue(OS, T^HS’));
			_, out = c.glue(out, repl[1:2])
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 2 && repl[1].Rune == ' ' {
			// -------------------------------------------------- F
			// 		return subst(IS’, FP, AP, HS, glue(OS, T^HS’));
			_, out = c.glue(out, repl[2:3])
			repl = repl[3:]
			continue
		}

		if len(repl) > 1 && repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) && repl[1].Rune == PPPASTE {
			// -------------------------------------------------- G
			if len(arg) == 0 {
				// ------------------------------------------ H
				panic(c.position(repl[0]))
			}

			// -------------------------------------------------- K
			// 			return subst(##^HS’ • IS’, FP, AP, HS, OS • select(i, AP));
			out = append(out, arg...)
			repl = repl[1:]
			continue
		}

		if len(repl) > 2 && repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) && repl[1].Rune == ' ' && repl[2].Rune == PPPASTE {
			// -------------------------------------------------- G
			if len(arg) == 0 {
				// ------------------------------------------ H
				if len(repl) > 3 && repl[3].Rune == IDENTIFIER && m.param(ap, repl[3].Val, &arg) {
					// ---------------------------------- I
					panic(c.position(repl[0]))
				}

				// ------------------------------------------ J
				// 				return subst(IS’, FP, AP, HS, OS);
				repl = repl[3:]
				continue
			}

			// -------------------------------------------------- K
			// 			return subst(##^HS’ • IS’, FP, AP, HS, OS • select(i, AP));
			out = append(out, arg...)
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) {
			// -------------------------------------------------- L
			// 		return subst(IS’, FP, AP, HS, OS • expand(select(i, AP)));
			out = append(out, c.expands(arg, expandDefined)...)
			repl = repl[1:]
			continue
		}

		// ---------------------------------------------------------- M
		// 	note IS must be T^HS’ • IS’
		// 	return subst(IS’, FP, AP, HS, OS • T^HS’);
		out = append(out, repl[0])
		repl = repl[1:]
	}
}

// paste last of left side with first of right side
//
// [1] pg. 3
func (c *cpp) glue(ls, rs []cppToken) (n int, out []cppToken) {
	for len(ls) != 0 && ls[len(ls)-1].Rune == ' ' {
		ls = ls[:len(ls)-1]
	}

	for len(rs) != 0 && rs[0].Rune == ' ' {
		rs = rs[1:]
		n++
	}
	if len(rs) == 0 {
		panic("TODO")
	}

	if len(ls) == 0 {
		return n, rs
	}

	l := ls[len(ls)-1]
	ls = ls[:len(ls)-1]
	r := rs[0]
	rs = rs[1:]
	n++

	switch l.Rune {
	case '#':
		switch r.Rune {
		case '#':
			l.Rune = PPPASTE
		default:
			panic(PrettyString([]cppToken{l, r}))
		}
	default:
		switch l.Rune {
		case STRINGLITERAL:
			s := TokSrc(l.Token)
			if len(s) > 2 && s[0] == '"' && s[len(s)-1] == '"' {
				s = s[1 : len(s)-1]
			}
			l.Val = dict.SID(s + TokSrc(r.Token))
		default:
			l.Val = dict.SID(TokSrc(l.Token) + TokSrc(r.Token))
		}
	}
	return n, append(append(ls, l), rs...)
}

// Given a token sequence, stringize returns a single string literal token
// containing the concatenated spellings of the tokens.
//
// [1] pg. 3
func (c *cpp) stringize(s []cppToken) cppToken {
	var a []string
	for _, v := range s {
		switch v.Rune {
		case CHARCONST, LONGCHARCONST, LONGSTRINGLITERAL, STRINGLITERAL:
			s := fmt.Sprintf("%q", TokSrc(v.Token))
			a = append(a, s[1:len(s)-1])
		default:
			a = append(a, TokSrc(v.Token))
		}
	}
	if v := dict.SID(fmt.Sprintf(`"%s"`, strings.Join(a, ""))); v != 0 {
		var t cppToken
		if len(s) != 0 {
			t = s[0]
		}
		t.Rune = STRINGLITERAL
		t.Val = v
		return t
	}

	return cppToken{}
}

func (c *cpp) directive(r tokenReader, w tokenWriter, cs conds) (y conds) {
	line := c.line(r)
	if len(line) == 0 {
		return cs
	}

	if cs.on() {
		if f := c.tweaks.TrackExpand; f != nil && c.tweaks.DefinesOnly {
			if s := cppToksDump(line, ""); strings.HasPrefix(s, "define") {
				f(fmt.Sprintf("#%s", cppToksDump(line, "")))
			}
		}
	}

outer:
	switch t := line[0]; t.Rune {
	case ccEOF:
		// nop
	case IDENTIFIER:
		switch t.Val {
		case idDefine:
			if !cs.on() {
				break
			}

			if len(line) == 1 {
				c.err(t, "empty define not allowed")
				break
			}

			c.define(line[1:])
		case idElif:
			switch cs.tos() {
			case condIfOff:
				if _, ok := c.constExpr(line[1:], true); ok {
					return cs.pop().push(condIfOn)
				}
			case condIfOn:
				return cs.pop().push(condIfSkip)
			case condIfSkip:
				// nop
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cs.tos()))
			}
		case idElse:
			switch cs.tos() {
			case condIfOff:
				return cs.pop().push(condIfOn)
			case condIfOn:
				return cs.pop().push(condIfOff)
			case condIfSkip:
				// nop
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cs.tos()))
			}
		case idError:
			if !cs.on() {
				break
			}

			c.err(t, "%s", cppToksDump(line, ""))
		case idIf:
			if !cs.on() {
				return cs.push(condIfSkip)
			}

			switch _, ok := c.constExpr(line[1:], true); {
			case ok:
				return cs.push(condIfOn)
			default:
				return cs.push(condIfOff)
			}
		case idIfdef:
			if !cs.on() {
				return cs.push(condIfSkip)
			}

			line = cppTrimAllSpace(line[1:])
			if len(line) == 0 {
				c.err(t, "empty #ifdef not allowed")
				break
			}

			if len(line) > 1 {
				c.err(t, "extra tokens after #ifdef not allowed")
				break
			}

			if line[0].Rune != IDENTIFIER {
				c.err(line[0], "expected identifier")
				break
			}

			if _, ok := c.macros[line[0].Val]; ok {
				return cs.push(condIfOn)
			}

			return cs.push(condIfOff)
		case idIfndef:
			if !cs.on() {
				return cs.push(condIfSkip)
			}

			line = cppTrimAllSpace(line[1:])
			if len(line) == 0 {
				c.err(t, "empty #ifndef not allowed")
				break
			}

			if len(line) > 1 {
				c.err(t, "extra tokens after #ifndef not allowed")
				break
			}

			if line[0].Rune != IDENTIFIER {
				c.err(line[0], "expected identifier")
				break
			}

			if _, ok := c.macros[line[0].Val]; ok {
				return cs.push(condIfOff)
			}

			return cs.push(condIfOn)
		case
			idIncludeNext,
			idInclude:

			if !cs.on() {
				break
			}

			line = cppTrimAllSpace(line[1:])
			if len(line) == 0 {
				c.err(t, "empty include not allowed")
				break
			}

			expanded := false
		again:
			switch line[0].Rune {
			case '<':
				if c.tweaks.cppExpandTest {
					w.write(line...)
					return cs
				}

				var nm string
				for _, v := range line[1:] {
					if v.Rune == '>' {
						c.include(t, nm, c.sysIncludePaths, w)
						return cs
					}

					nm += TokSrc(v.Token)
				}
				c.err(t, "invalid include file name specification")
			case STRINGLITERAL:
				if c.tweaks.cppExpandTest {
					w.write(line...)
					return cs
				}

				b := dict.S(line[0].Val)      // `"foo.h"`
				nm := string(b[1 : len(b)-1]) // `foo.h`
				c.include(t, nm, c.includePaths, w)
				return cs
			default:
				if expanded {
					panic(PrettyString(line))
				}

				line = c.expands(cppTrimAllSpace(line), false)
				expanded = true
				if c.tweaks.cppExpandTest {
					w.write(line...)
					return cs
				}

				goto again
			}
		case idEndif:
			switch cs.tos() {
			case condIfOn, condIfOff, condIfSkip:
				return cs.pop()
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cs.tos()))
			}
		case idLine:
			if !cs.on() {
				break
			}

			f := fset.File(line[0].Pos())
			off := f.Offset(line[0].Pos())
			pos := c.position(line[0])
			line = c.expands(cppTrimAllSpace(line[1:]), false)
			switch len(line) {
			case 1: // #line linenum
				n, err := strconv.ParseUint(string(line[0].S()), 10, 31)
				if err != nil {
					break
				}

				f.AddLineInfo(off, pos.Filename, int(n-1))
				//TODO
			case 2: // #line linenum filename
				//TODO
			default:
				// ignore
			}

			// ignored
		case idPragma:
			if !cs.on() {
				break
			}

			for {
				line = line[1:]
				if len(line) == 0 {
					panic(fmt.Errorf("%v", c.position(t)))
				}

				switch t = line[0]; {
				case t.Rune == ' ':
					// nop
				case t.Val == idPushMacro:
					actuals := c.pragmaActuals(t, line[1:])
					if len(actuals) != 1 {
						panic(fmt.Errorf("%v", c.position(t)))
					}

					t := actuals[0]
					switch t.Rune {
					case STRINGLITERAL:
						nm := int(c.strConst(t.Token).Value.(*ir.StringValue).StringID)
						m := c.macros[nm]
						if m != nil {
							c.macroStack[nm] = append(c.macroStack[nm], m)
						}
						break outer
					default:
						panic(fmt.Errorf("%v: %v", c.position(t), yySymName(int(actuals[0].Rune))))
					}
				case t.Val == idPopMacro:
					actuals := c.pragmaActuals(t, line[1:])
					if len(actuals) != 1 {
						panic(fmt.Errorf("%v", c.position(t)))
					}

					t := actuals[0]
					switch t.Rune {
					case STRINGLITERAL:
						nm := int(c.strConst(t.Token).Value.(*ir.StringValue).StringID)
						s := c.macroStack[nm]
						if n := len(s); n != 0 {
							m := s[n-1]
							s = s[:n-1]
							c.macroStack[nm] = s
							c.macros[nm] = m
						}
						break outer
					default:
						panic(fmt.Errorf("%v: %v", c.position(t), yySymName(int(actuals[0].Rune))))
					}
				default:
					if c.tweaks.IgnoreUnknownPragmas {
						break outer
					}

					panic(fmt.Errorf("%v: %#x, %v", c.position(t), t.Rune, t))
				}
			}
		case idUndef:
			if !cs.on() {
				break
			}

			line = cppTrimSpace(line[1:])
			if len(line) == 0 {
				panic("TODO")
			}

			if len(line) > 1 {
				panic("TODO")
			}

			if line[0].Rune != IDENTIFIER {
				panic("TODO")
			}

			delete(c.macros, line[0].Val)
		case idWarning:
			if !cs.on() {
				break
			}

			panic(fmt.Errorf("%v", c.position(t)))
		default:
			panic(fmt.Errorf("%v %v", c.position(t), PrettyString(t)))
		}
	default:
		panic(PrettyString(t))
	}
	return cs
}

func (c *cpp) include(n Node, nm string, paths []string, w tokenWriter) {
	if c.includeLevel == maxIncludeLevel {
		c.err(n, "too many include levels")
	}

	c.includeLevel++

	defer func() { c.includeLevel-- }()

	dir := filepath.Dir(c.position(n).Filename)
	if d, err := filepath.Abs(dir); err == nil {
		dir = d
	}
	var path string
	if n.(cppToken).Val == idIncludeNext {
		nmDir, _ := filepath.Split(nm)
		for i, v := range paths {
			if w, err := filepath.Abs(v); err == nil {
				v = w
			}
			v = filepath.Join(v, nmDir)
			if v == dir {
				paths = paths[i+1:]
				break
			}
		}
	}
	for _, v := range paths {
		if v == "@" {
			v = dir
		}

		var p string
		switch {
		case strings.HasPrefix(nm, "./"):
			p = nm
		default:
			p = filepath.Join(v, nm)
		}
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}

		path = p
		break
	}

	if path == "" {
		wd, _ := os.Getwd()
		c.err(n, "include file not found: %s\nworking dir: %s\nsearch paths:\n\t%s", nm, wd, strings.Join(paths, "\n\t"))
		return
	}

	s, err := NewFileSource2(path, true)
	if err != nil {
		c.err(n, "%s", err.Error())
		return
	}

	if n, _ := s.Size(); n == 0 {
		return
	}

	if f := c.tweaks.TrackIncludes; f != nil {
		f(path)
	}
	r, err := c.parse(s)
	if err != nil {
		c.err(n, "%s", err.Error())
	}

	c.expand(r, w, conds(nil).push(condZero), 0, false)
}

func (c *cpp) constExpr(toks []cppToken, expandDefined bool) (op Operand, y bool) {
	toks = cppTrimAllSpace(c.expands(cppTrimAllSpace(toks), expandDefined))
	for i, v := range toks {
		if v.Rune == IDENTIFIER {
			toks[i].Rune = INTCONST
			toks[i].Val = idZero
		}
	}
	c.lx.ungetBuffer = c.lx.ungetBuffer[:0]
	c.lx.ungets(toks...)
	if !c.lx.parseExpr() {
		return Operand{}, false
	}

	e := c.lx.ast.(*ConstExpr)
	v := e.eval(c.context)
	if v.Type != Int {
		return v, false
	}

	switch x := v.Value.(type) {
	case *ir.Int64Value:
		return v, x.Value != 0
	default:
		return v, false
	}
}

func (c *cpp) define(line []cppToken) {
	switch line[0].Rune {
	case ' ':
		c.defineMacro(xcToks(line[1:]))
	default:
		panic(PrettyString(line))
	}
}

func (c *cpp) defineMacro(line []xc.Token) {
	if len(line) == 0 {
		panic("internal error")
	}

	if line[0].Rune == ' ' {
		line = line[1:]
	}

	switch t := line[0]; t.Rune {
	case IDENTIFIER:
		nm := t.Val
		if protectedMacro[nm] {
			panic("TODO")
		}
		line := line[1:]
		var repl []xc.Token
		if len(line) != 0 {
			switch line[0].Rune {
			case '\n', ccEOF:
				// nop
			case ' ':
				repl = line[1:]
			case '(':
				c.defineFnMacro(t, line[1:])
				return
			default:
				panic(fmt.Errorf(PrettyString(line[0])))
			}
		}

		if ex := c.macros[nm]; ex != nil {
			if c.identicalReplacementLists(repl, ex.ReplacementToks) {
				return
			}

			c.err(t, "%q replacement lists differ: %q, %q", dict.S(nm), toksDump(ex.ReplacementToks, ""), toksDump(repl, ""))
			return
		}

		if traceMacroDefs {
			fmt.Fprintf(os.Stderr, "#define %s %s\n", dict.S(nm), toksDump(repl, ""))
		}
		c.macros[nm] = newMacro(t, repl)
	default:
		panic(PrettyString(t))
	}
}

func (c *cpp) identicalReplacementLists(a, b []xc.Token) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		w := b[i]
		if v.Rune != w.Rune || v.Val != w.Val {
			return false
		}
	}

	return true
}

func (c *cpp) defineFnMacro(nmTok xc.Token, line []xc.Token) {
	ident := true
	var params []int
	variadic := false
	for i, v := range line {
		switch v.Rune {
		case IDENTIFIER:
			if !ident {
				panic("TODO")
			}

			params = append(params, v.Val)
			ident = false
		case ')':
			m := newMacro(nmTok, trimSpace(line[i+1:]))
			m.IsFnLike = true
			m.ident = ident
			m.IsVariadic = variadic
			m.Args = params
			if ex := c.macros[nmTok.Val]; ex != nil {
				if c.identicalParamLists(params, ex.Args) && c.identicalReplacementLists(m.ReplacementToks, ex.ReplacementToks) && m.IsVariadic == ex.IsVariadic {
					return
				}

				c.err(nmTok, "parameter and/or replacement lists differ")
				return
			}

			if traceMacroDefs {
				var a [][]byte
				for _, v := range m.Args {
					a = append(a, dict.S(v))
				}
				fmt.Fprintf(os.Stderr, "#define %s(%s) %s\n", dict.S(nmTok.Val), bytes.Join(a, []byte(", ")), toksDump(m.ReplacementToks, ""))
			}
			c.macros[nmTok.Val] = m
			return
		case ',':
			if ident {
				panic("TODO")
			}

			ident = true
		case ' ':
			// nop
		case DDD:
			variadic = true
		default:
			panic(PrettyString(v))
		}
	}
}

func (c *cpp) identicalParamLists(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func (c *cpp) line(r tokenReader) []cppToken {
	c.toks = c.toks[:0]
	for {
		switch t := r.read(); t.Rune {
		case '\n', ccEOF:
			if len(c.toks) == 0 || c.toks[0].Rune != ' ' {
				return c.toks
			}

			for i, v := range c.toks {
				if v.Rune != ' ' {
					n := copy(c.toks, c.toks[i:])
					c.toks = c.toks[:n]
					return c.toks
				}
			}

			c.toks = c.toks[:0]
			return c.toks
		default:
			c.toks = append(c.toks, t)
		}
	}
}
