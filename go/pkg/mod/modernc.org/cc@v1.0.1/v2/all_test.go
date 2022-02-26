// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"bytes"
	"flag"
	"fmt"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"modernc.org/golex/lex"
	"modernc.org/xc"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func caller3(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(3)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("# TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func use(...interface{}) {}

func init() {
	use(caller, caller3, dbg, TODO, toksDump) //TODOOK
	flag.IntVar(&yyDebug, "yydebug", 0, "")
	flag.BoolVar(&traceMacroDefs, "macros", false, "")
}

// ============================================================================

var (
	oRE = flag.String("re", "", "")

	shellc      = filepath.FromSlash("testdata/_sqlite/sqlite-amalgamation-3210000/shell.c")
	sqlite3c    = filepath.FromSlash("testdata/_sqlite/sqlite-amalgamation-3210000/sqlite3.c")
	searchPaths []string
)

func init() {
	var err error
	searchPaths, err = Paths(true)
	if err != nil {
		panic(err)
	}
}

func testUCNTable(t *testing.T, tab []rune, fOk, fOther func(rune) bool, fcategory func(rune) bool, tag string) {
	m := map[rune]struct{}{}
	for i := 0; i < len(tab); i += 2 {
		l, h := tab[i], tab[i+1]
		if h == 0 {
			h = l
		}
		for r := l; r <= h; r++ {
			m[r] = struct{}{}
		}
	}
	for r := rune(0); r < 0xffff; r++ {
		_, ok := m[r]
		if g, e := fOk(r), ok; g != e {
			t.Errorf("%#04x %v %v", r, g, e)
		}

		if ok {
			if g, e := fOther(r), false; g != e {
				t.Errorf("%#04x %v %v", r, g, e)
			}
		}
	}
}

func TestUCNDigitsTable(t *testing.T) {
	tab := []rune{
		0x0660, 0x0669, 0x06F0, 0x06F9, 0x0966, 0x096F, 0x09E6, 0x09EF, 0x0A66, 0x0A6F,
		0x0AE6, 0x0AEF, 0x0B66, 0x0B6F, 0x0BE7, 0x0BEF, 0x0C66, 0x0C6F, 0x0CE6, 0x0CEF,
		0x0D66, 0x0D6F, 0x0E50, 0x0E59, 0x0ED0, 0x0ED9, 0x0F20, 0x0F33,
	}
	testUCNTable(t, tab, isUCNDigit, isUCNNonDigit, unicode.IsDigit, "unicode.IsDigit")
}

func TestUCNNonDigitsTable(t *testing.T) {
	tab := []rune{
		0x00AA, 0x0000, 0x00B5, 0x0000, 0x00B7, 0x0000, 0x00BA, 0x0000, 0x00C0, 0x00D6,
		0x00D8, 0x00F6, 0x00F8, 0x01F5, 0x01FA, 0x0217, 0x0250, 0x02A8, 0x02B0, 0x02B8,
		0x02BB, 0x0000, 0x02BD, 0x02C1, 0x02D0, 0x02D1, 0x02E0, 0x02E4, 0x037A, 0x0000,
		0x0386, 0x0000, 0x0388, 0x038A, 0x038C, 0x0000, 0x038E, 0x03A1, 0x03A3, 0x03CE,
		0x03D0, 0x03D6, 0x03DA, 0x0000, 0x03DC, 0x0000, 0x03DE, 0x0000, 0x03E0, 0x0000,
		0x03E2, 0x03F3, 0x0401, 0x040C, 0x040E, 0x044F, 0x0451, 0x045C, 0x045E, 0x0481,
		0x0490, 0x04C4, 0x04C7, 0x04C8, 0x04CB, 0x04CC, 0x04D0, 0x04EB, 0x04EE, 0x04F5,
		0x04F8, 0x04F9, 0x0531, 0x0556, 0x0559, 0x0000, 0x0561, 0x0587, 0x05B0, 0x05B9,
		0x05F0, 0x05F2, 0x0621, 0x063A, 0x0640, 0x0652, 0x0670, 0x06B7, 0x06BA, 0x06BE,
		0x06C0, 0x06CE, 0x06D0, 0x06DC, 0x06E5, 0x06E8, 0x06EA, 0x06ED, 0x0901, 0x0903,
		0x0905, 0x0939, 0x093D, 0x0000, 0x093E, 0x094D, 0x0950, 0x0952, 0x0958, 0x0963,
		0x0981, 0x0983, 0x0985, 0x098C, 0x098F, 0x0990, 0x0993, 0x09A8, 0x09AA, 0x09B0,
		0x09B2, 0x0000, 0x09B6, 0x09B9, 0x09BE, 0x09C4, 0x09C7, 0x09C8, 0x09CB, 0x09CD,
		0x09DC, 0x09DD, 0x09DF, 0x09E3, 0x09F0, 0x09F1, 0x0A02, 0x0000, 0x0A05, 0x0A0A,
		0x0A0F, 0x0A10, 0x0A13, 0x0A28, 0x0A2A, 0x0A30, 0x0A32, 0x0A33, 0x0A35, 0x0A36,
		0x0A38, 0x0A39, 0x0A3E, 0x0A42, 0x0A47, 0x0A48, 0x0A4B, 0x0A4D, 0x0A59, 0x0A5C,
		0x0A5E, 0x0000, 0x0A74, 0x0000, 0x0A81, 0x0A83, 0x0A85, 0x0A8B, 0x0A8D, 0x0000,
		0x0A8F, 0x0A91, 0x0A93, 0x0AA8, 0x0AAA, 0x0AB0, 0x0AB2, 0x0AB3, 0x0AB5, 0x0AB9,
		0x0ABD, 0x0AC5, 0x0AC7, 0x0AC9, 0x0ACB, 0x0ACD, 0x0AD0, 0x0000, 0x0AE0, 0x0000,
		0x0B01, 0x0B03, 0x0B05, 0x0B0C, 0x0B0F, 0x0B10, 0x0B13, 0x0B28, 0x0B2A, 0x0B30,
		0x0B32, 0x0B33, 0x0B36, 0x0B39, 0x0B3D, 0x0000, 0x0B3E, 0x0B43, 0x0B47, 0x0B48,
		0x0B4B, 0x0B4D, 0x0B5C, 0x0B5D, 0x0B5F, 0x0B61, 0x0B82, 0x0B83, 0x0B85, 0x0B8A,
		0x0B8E, 0x0B90, 0x0B92, 0x0B95, 0x0B99, 0x0B9A, 0x0B9C, 0x0000, 0x0B9E, 0x0B9F,
		0x0BA3, 0x0BA4, 0x0BA8, 0x0BAA, 0x0BAE, 0x0BB5, 0x0BB7, 0x0BB9, 0x0BBE, 0x0BC2,
		0x0BC6, 0x0BC8, 0x0BCA, 0x0BCD, 0x0C01, 0x0C03, 0x0C05, 0x0C0C, 0x0C0E, 0x0C10,
		0x0C12, 0x0C28, 0x0C2A, 0x0C33, 0x0C35, 0x0C39, 0x0C3E, 0x0C44, 0x0C46, 0x0C48,
		0x0C4A, 0x0C4D, 0x0C60, 0x0C61, 0x0C82, 0x0C83, 0x0C85, 0x0C8C, 0x0C8E, 0x0C90,
		0x0C92, 0x0CA8, 0x0CAA, 0x0CB3, 0x0CB5, 0x0CB9, 0x0CBE, 0x0CC4, 0x0CC6, 0x0CC8,
		0x0CCA, 0x0CCD, 0x0CDE, 0x0000, 0x0CE0, 0x0CE1, 0x0D02, 0x0D03, 0x0D05, 0x0D0C,
		0x0D0E, 0x0D10, 0x0D12, 0x0D28, 0x0D2A, 0x0D39, 0x0D3E, 0x0D43, 0x0D46, 0x0D48,
		0x0D4A, 0x0D4D, 0x0D60, 0x0D61, 0x0E01, 0x0E3A,

		// In [0], Annex D, Thai [0x0E40, 0x0E5B] overlaps with digits
		// [0x0E50, 0x0E59]. Exclude them.
		0x0E40, 0x0E4F,
		0x0E5A, 0x0E5B,

		0x0E81, 0x0E82,
		0x0E84, 0x0000, 0x0E87, 0x0E88, 0x0E8A, 0x0000, 0x0E8D, 0x0000, 0x0E94, 0x0E97,
		0x0E99, 0x0E9F, 0x0EA1, 0x0EA3, 0x0EA5, 0x0000, 0x0EA7, 0x0000, 0x0EAA, 0x0EAB,
		0x0EAD, 0x0EAE, 0x0EB0, 0x0EB9, 0x0EBB, 0x0EBD, 0x0EC0, 0x0EC4, 0x0EC6, 0x0000,
		0x0EC8, 0x0ECD, 0x0EDC, 0x0EDD, 0x0F00, 0x0000, 0x0F18, 0x0F19, 0x0F35, 0x0000,
		0x0F37, 0x0000, 0x0F39, 0x0000, 0x0F3E, 0x0F47, 0x0F49, 0x0F69, 0x0F71, 0x0F84,
		0x0F86, 0x0F8B, 0x0F90, 0x0F95, 0x0F97, 0x0000, 0x0F99, 0x0FAD, 0x0FB1, 0x0FB7,
		0x0FB9, 0x0000, 0x10A0, 0x10C5, 0x10D0, 0x10F6, 0x1E00, 0x1E9B, 0x1EA0, 0x1EF9,
		0x1F00, 0x1F15, 0x1F18, 0x1F1D, 0x1F20, 0x1F45, 0x1F48, 0x1F4D, 0x1F50, 0x1F57,
		0x1F59, 0x0000, 0x1F5B, 0x0000, 0x1F5D, 0x0000, 0x1F5F, 0x1F7D, 0x1F80, 0x1FB4,
		0x1FB6, 0x1FBC, 0x1FBE, 0x0000, 0x1FC2, 0x1FC4, 0x1FC6, 0x1FCC, 0x1FD0, 0x1FD3,
		0x1FD6, 0x1FDB, 0x1FE0, 0x1FEC, 0x1FF2, 0x1FF4, 0x1FF6, 0x1FFC, 0x203F, 0x2040,
		0x207F, 0x0000, 0x2102, 0x0000, 0x2107, 0x0000, 0x210A, 0x2113, 0x2115, 0x0000,
		0x2118, 0x211D, 0x2124, 0x0000, 0x2126, 0x0000, 0x2128, 0x0000, 0x212A, 0x2131,
		0x2133, 0x2138, 0x2160, 0x2182, 0x3005, 0x3007, 0x3021, 0x3029, 0x3041, 0x3093,
		0x309B, 0x309C, 0x30A1, 0x30F6, 0x30FB, 0x30FC, 0x3105, 0x312C, 0x4E00, 0x9FA5,
		0xAC00, 0xD7A3,
	}
	testUCNTable(t, tab, isUCNNonDigit, isUCNDigit, unicode.IsLetter, "unicode.IsLetter")
}

func charStr(c rune) string { return yySymName(int(c)) }

func charsStr(chars []lex.Char, delta token.Pos) (a []string) {
	for _, v := range chars {
		a = append(a, fmt.Sprintf("{%s %d}", charStr(v.Rune), v.Pos()-delta))
	}
	return a
}

type x []struct {
	c   rune
	pos token.Pos
}

type lexerTests []struct {
	src   string
	chars x
}

func testLexer(t *testing.T, newLexer func(i int, src string) (*lexer, error), tab lexerTests) {
nextTest:
	for ti, test := range tab {
		lx, err := newLexer(ti, test.src)
		if err != nil {
			t.Fatal(err)
		}

		delta := token.Pos(lx.File.Base() - 1)
		var chars []lex.Char
		var c lex.Char
		var lval yySymType
		for i := 0; c.Rune >= 0 && i < len(test.src)+2; i++ {
			lx.Lex(&lval)
			c = lval.Token.Char
			chars = append(chars, c)
		}
		if c.Rune >= 0 {
			t.Errorf("%d: scanner stall %v", ti, charsStr(chars, delta))
			continue
		}

		if g, e := lx.error(), error(nil); g != e {
			t.Errorf("%d: lx.err %v %v %v", ti, g, e, charsStr(chars, delta))
			continue
		}

		if g, e := len(chars), len(test.chars); g != e {
			t.Errorf("%d: len(chars) %v %v %v", ti, g, e, charsStr(chars, delta))
			continue
		}

		for i, c := range chars {
			c = chars[i]
			e := test.chars[i]
			g := c.Rune
			if e := e.c; g != e {
				t.Errorf("%d: c[%d] %v %v %v", ti, i, charStr(g), charStr(e), charsStr(chars, delta))
				continue nextTest
			}

			if g, e := c.Pos()-delta, e.pos; g != e {
				t.Errorf("%d: pos[%d] %v %v %v", ti, i, g, e, charsStr(chars, delta))
				continue nextTest
			}
		}
	}
}

func TestLexer(t *testing.T) {
	ctx, err := newContext(&Tweaks{})
	if err != nil {
		t.Fatal(err)
	}

	testLexer(
		t,
		func(i int, src string) (*lexer, error) {
			return newLexer(ctx, fmt.Sprintf("TestLexer.%d", i), len(src), strings.NewReader(src))
		},
		lexerTests{
			{"", x{{-1, 1}}},
			{"%0", x{{'%', 1}, {INTCONST, 2}, {-1, 3}}},
			{"%:%:", x{{PPPASTE, 1}, {-1, 5}}},
			{"%>", x{{'}', 1}, {-1, 3}}},
			{"0", x{{INTCONST, 1}, {-1, 2}}},
			{"01", x{{INTCONST, 1}, {-1, 3}}},
			{"0??/1\n", x{{INTCONST, 1}, {'?', 2}, {'?', 3}, {'/', 4}, {INTCONST, 5}, {'\n', 6}, {-1, 7}}},
			{"0??/1\n2", x{{INTCONST, 1}, {'?', 2}, {'?', 3}, {'/', 4}, {INTCONST, 5}, {'\n', 6}, {INTCONST, 7}, {-1, 8}}},
			{"0??/\n", x{{INTCONST, 1}, {'?', 2}, {'?', 3}, {'/', 4}, {'\n', 5}, {-1, 6}}},
			{"0??/\n2", x{{INTCONST, 1}, {'?', 2}, {'?', 3}, {'/', 4}, {'\n', 5}, {INTCONST, 6}, {-1, 7}}},
			{"0\\1\n", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 3}, {'\n', 4}, {-1, 5}}},
			{"0\\1\n2", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 3}, {'\n', 4}, {INTCONST, 5}, {-1, 6}}},
			{"0\\\n", x{{INTCONST, 1}, {-1, 4}}},
			{"0\\\n2", x{{INTCONST, 1}, {-1, 5}}},
			{"0\x00", x{{INTCONST, 1}, {0, 2}, {-1, 3}}},
			{"0\x001", x{{INTCONST, 1}, {0, 2}, {INTCONST, 3}, {-1, 4}}},
			{":>", x{{']', 1}, {-1, 3}}},
			{"<%", x{{'{', 1}, {-1, 3}}},
			{"<:", x{{'[', 1}, {-1, 3}}},
			{"??!", x{{'?', 1}, {'?', 2}, {'!', 3}, {-1, 4}}},
			{"??!0", x{{'?', 1}, {'?', 2}, {'!', 3}, {INTCONST, 4}, {-1, 5}}},
			{"??!01", x{{'?', 1}, {'?', 2}, {'!', 3}, {INTCONST, 4}, {-1, 6}}},
			{"??!=", x{{'?', 1}, {'?', 2}, {NEQ, 3}, {-1, 5}}},
			{"??'", x{{'?', 1}, {'?', 2}, {'\'', 3}, {-1, 4}}},
			{"??(", x{{'?', 1}, {'?', 2}, {'(', 3}, {-1, 4}}},
			{"??)", x{{'?', 1}, {'?', 2}, {')', 3}, {-1, 4}}},
			{"??-", x{{'?', 1}, {'?', 2}, {'-', 3}, {-1, 4}}},
			{"??/", x{{'?', 1}, {'?', 2}, {'/', 3}, {-1, 4}}},
			{"??/1\n", x{{'?', 1}, {'?', 2}, {'/', 3}, {INTCONST, 4}, {'\n', 5}, {-1, 6}}},
			{"??/1\n2", x{{'?', 1}, {'?', 2}, {'/', 3}, {INTCONST, 4}, {'\n', 5}, {INTCONST, 6}, {-1, 7}}},
			{"??/\n", x{{'?', 1}, {'?', 2}, {'/', 3}, {'\n', 4}, {-1, 5}}},
			{"??/\n2", x{{'?', 1}, {'?', 2}, {'/', 3}, {'\n', 4}, {INTCONST, 5}, {-1, 6}}},
			{"??<", x{{'?', 1}, {'?', 2}, {'<', 3}, {-1, 4}}},
			{"??=??=", x{{'?', 1}, {'?', 2}, {'=', 3}, {'?', 4}, {'?', 5}, {'=', 6}, {-1, 7}}},
			{"??>", x{{'?', 1}, {'?', 2}, {'>', 3}, {-1, 4}}},
			{"???!", x{{'?', 1}, {'?', 2}, {'?', 3}, {'!', 4}, {-1, 5}}},
			{"???!0", x{{'?', 1}, {'?', 2}, {'?', 3}, {'!', 4}, {INTCONST, 5}, {-1, 6}}},
			{"???/\n2", x{{'?', 1}, {'?', 2}, {'?', 3}, {'/', 4}, {'\n', 5}, {INTCONST, 6}, {-1, 7}}},
			{"????!0", x{{'?', 1}, {'?', 2}, {'?', 3}, {'?', 4}, {'!', 5}, {INTCONST, 6}, {-1, 7}}},
			{"???x0", x{{'?', 1}, {'?', 2}, {'?', 3}, {IDENTIFIER, 4}, {-1, 6}}},
			{"???x??!0", x{{'?', 1}, {'?', 2}, {'?', 3}, {IDENTIFIER, 4}, {'?', 5}, {'?', 6}, {'!', 7}, {INTCONST, 8}, {-1, 9}}},
			{"??x0", x{{'?', 1}, {'?', 2}, {IDENTIFIER, 3}, {-1, 5}}},
			{"??x??!0", x{{'?', 1}, {'?', 2}, {IDENTIFIER, 3}, {'?', 4}, {'?', 5}, {'!', 6}, {INTCONST, 7}, {-1, 8}}},
			{"?x0", x{{'?', 1}, {IDENTIFIER, 2}, {-1, 4}}},
			{"?x??!0", x{{'?', 1}, {IDENTIFIER, 2}, {'?', 3}, {'?', 4}, {'!', 5}, {INTCONST, 6}, {-1, 7}}},
			{"@", x{{'@', 1}, {-1, 2}}},
			{"@%", x{{'@', 1}, {'%', 2}, {-1, 3}}},
			{"@%0", x{{'@', 1}, {'%', 2}, {INTCONST, 3}, {-1, 4}}},
			{"@%:", x{{'@', 1}, {'#', 2}, {-1, 4}}},
			{"@%:0", x{{'@', 1}, {'#', 2}, {INTCONST, 4}, {-1, 5}}},
			{"@%:01", x{{'@', 1}, {'#', 2}, {INTCONST, 4}, {-1, 6}}},
			{"@??=", x{{'@', 1}, {'?', 2}, {'?', 3}, {'=', 4}, {-1, 5}}},
			{"\"(a\\\nz", x{{'"', 1}, {'(', 2}, {IDENTIFIER, 3}, {-1, 7}}},
			{"\\1\n", x{{'\\', 1}, {INTCONST, 2}, {'\n', 3}, {-1, 4}}},
			{"\\1\n2", x{{'\\', 1}, {INTCONST, 2}, {'\n', 3}, {INTCONST, 4}, {-1, 5}}},
			{"\\\n", x{{-1, 3}}},
			{"\\\n2", x{{INTCONST, 3}, {-1, 4}}},
			{"\\\r\n", x{{-1, 4}}},
			{"\\\r\n2", x{{INTCONST, 4}, {-1, 5}}},
			{"\r", x{{-1, 2}}},
			{"\r0", x{{INTCONST, 2}, {-1, 3}}},
			{"\r01", x{{INTCONST, 2}, {-1, 4}}},
			{"\x00", x{{0, 1}, {-1, 2}}},
			{"\x000", x{{0, 1}, {INTCONST, 2}, {-1, 3}}},
		},
	)
}

func TestLexerTrigraphs(t *testing.T) {
	ctx, err := newContext(&Tweaks{EnableTrigraphs: true})
	if err != nil {
		t.Fatal(err)
	}

	testLexer(
		t,
		func(i int, src string) (*lexer, error) {
			return newLexer(ctx, fmt.Sprintf("TestLexer.%d", i), len(src), strings.NewReader(src))
		},
		lexerTests{
			{"", x{{-1, 1}}},
			{"%0", x{{'%', 1}, {INTCONST, 2}, {-1, 3}}},
			{"%:%:", x{{PPPASTE, 1}, {-1, 5}}},
			{"%>", x{{'}', 1}, {-1, 3}}},
			{"0", x{{INTCONST, 1}, {-1, 2}}},
			{"01", x{{INTCONST, 1}, {-1, 3}}},
			{"0??/1\n", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 5}, {'\n', 6}, {-1, 7}}},
			{"0??/1\n2", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 5}, {'\n', 6}, {INTCONST, 7}, {-1, 8}}},
			{"0??/\n", x{{INTCONST, 1}, {-1, 6}}},
			{"0??/\n2", x{{INTCONST, 1}, {-1, 7}}},
			{"0\\1\n", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 3}, {'\n', 4}, {-1, 5}}},
			{"0\\1\n2", x{{INTCONST, 1}, {'\\', 2}, {INTCONST, 3}, {'\n', 4}, {INTCONST, 5}, {-1, 6}}},
			{"0\\\n", x{{INTCONST, 1}, {-1, 4}}},
			{"0\\\n2", x{{INTCONST, 1}, {-1, 5}}},
			{"0\x00", x{{INTCONST, 1}, {0, 2}, {-1, 3}}},
			{"0\x001", x{{INTCONST, 1}, {0, 2}, {INTCONST, 3}, {-1, 4}}},
			{":>", x{{']', 1}, {-1, 3}}},
			{"<%", x{{'{', 1}, {-1, 3}}},
			{"<:", x{{'[', 1}, {-1, 3}}},
			{"??!", x{{'|', 1}, {-1, 4}}},
			{"??!0", x{{'|', 1}, {INTCONST, 4}, {-1, 5}}},
			{"??!01", x{{'|', 1}, {INTCONST, 4}, {-1, 6}}},
			{"??!=", x{{ORASSIGN, 1}, {-1, 5}}},
			{"??'", x{{'^', 1}, {-1, 4}}},
			{"??(", x{{'[', 1}, {-1, 4}}},
			{"??)", x{{']', 1}, {-1, 4}}},
			{"??-", x{{'~', 1}, {-1, 4}}},
			{"??/", x{{'\\', 1}, {-1, 4}}},
			{"??/1\n", x{{'\\', 1}, {INTCONST, 4}, {'\n', 5}, {-1, 6}}},
			{"??/1\n2", x{{'\\', 1}, {INTCONST, 4}, {'\n', 5}, {INTCONST, 6}, {-1, 7}}},
			{"??/\n", x{{-1, 5}}},
			{"??/\n2", x{{INTCONST, 5}, {-1, 6}}},
			{"??<", x{{'{', 1}, {-1, 4}}},
			{"??=??=", x{{PPPASTE, 1}, {-1, 7}}},
			{"??>", x{{'}', 1}, {-1, 4}}},
			{"???!", x{{'?', 1}, {'|', 2}, {-1, 5}}},
			{"???!0", x{{'?', 1}, {'|', 2}, {INTCONST, 5}, {-1, 6}}},
			{"???/\n2", x{{'?', 1}, {INTCONST, 6}, {-1, 7}}},
			{"????!0", x{{'?', 1}, {'?', 2}, {'|', 3}, {INTCONST, 6}, {-1, 7}}},
			{"???x0", x{{'?', 1}, {'?', 2}, {'?', 3}, {IDENTIFIER, 4}, {-1, 6}}},
			{"???x??!0", x{{'?', 1}, {'?', 2}, {'?', 3}, {IDENTIFIER, 4}, {'|', 5}, {INTCONST, 8}, {-1, 9}}},
			{"??x0", x{{'?', 1}, {'?', 2}, {IDENTIFIER, 3}, {-1, 5}}},
			{"??x??!0", x{{'?', 1}, {'?', 2}, {IDENTIFIER, 3}, {'|', 4}, {INTCONST, 7}, {-1, 8}}},
			{"?x0", x{{'?', 1}, {IDENTIFIER, 2}, {-1, 4}}},
			{"?x??!0", x{{'?', 1}, {IDENTIFIER, 2}, {'|', 3}, {INTCONST, 6}, {-1, 7}}},
			{"@", x{{'@', 1}, {-1, 2}}},
			{"@%", x{{'@', 1}, {'%', 2}, {-1, 3}}},
			{"@%0", x{{'@', 1}, {'%', 2}, {INTCONST, 3}, {-1, 4}}},
			{"@%:", x{{'@', 1}, {'#', 2}, {-1, 4}}},
			{"@%:0", x{{'@', 1}, {'#', 2}, {INTCONST, 4}, {-1, 5}}},
			{"@%:01", x{{'@', 1}, {'#', 2}, {INTCONST, 4}, {-1, 6}}},
			{"@??=", x{{'@', 1}, {'#', 2}, {-1, 5}}},
			{"\"(a\\\nz", x{{'"', 1}, {'(', 2}, {IDENTIFIER, 3}, {-1, 7}}},
			{"\\1\n", x{{'\\', 1}, {INTCONST, 2}, {'\n', 3}, {-1, 4}}},
			{"\\1\n2", x{{'\\', 1}, {INTCONST, 2}, {'\n', 3}, {INTCONST, 4}, {-1, 5}}},
			{"\\\n", x{{-1, 3}}},
			{"\\\n2", x{{INTCONST, 3}, {-1, 4}}},
			{"\\\r\n", x{{-1, 4}}},
			{"\\\r\n2", x{{INTCONST, 4}, {-1, 5}}},
			{"\r", x{{-1, 2}}},
			{"\r0", x{{INTCONST, 2}, {-1, 3}}},
			{"\r01", x{{INTCONST, 2}, {-1, 4}}},
			{"\x00", x{{0, 1}, {-1, 2}}},
			{"\x000", x{{0, 1}, {INTCONST, 2}, {-1, 3}}},
		},
	)
}

func exampleAST(rule int, src string) interface{} {
	ctx, err := newContext(&Tweaks{
		EnableAnonymousStructFields: true,
		EnableEmptyStructs:          true,
		EnableOmitFuncDeclSpec:      true,
	})
	if err != nil {
		return fmt.Sprintf("TODO: %v", err) //TODOOK
	}

	ctx.exampleRule = rule
	src = strings.TrimSpace(src)
	r, n := utf8.DecodeRuneInString(src)
	src = src[n:]
	l, err := newLexer(ctx, fmt.Sprintf("example%v.c", rule), len(src), strings.NewReader(src))
	if err != nil {
		return fmt.Sprintf("TODO: %v", err) //TODOOK
	}

	l.unget(cppToken{Token: xc.Token{Char: lex.Char{Rune: r}}})
	yyParse(l)
	if err := ctx.error(); err != nil {
		return fmt.Sprintf("TODO: %v", err) //TODOOK
	}

	if ctx.exampleAST == nil {
		return "TODO: nil" //TODOOK
	}

	return ctx.exampleAST
}

func testCPPParseSource(ctx *context, src Source) (*cpp, tokenReader, error) {
	if ctx == nil {
		var err error
		if ctx, err = newContext(&Tweaks{}); err != nil {
			return nil, nil, err
		}
	}

	c := newCPP(ctx)
	r, err := c.parse(src)
	if err != nil {
		return nil, nil, err
	}

	return c, r, nil
}

func testCPPParseFile(ctx *context, nm string) (*cpp, tokenReader, error) {
	return testCPPParseSource(ctx, MustFileSource(nm))
}

func testCPPParseString(ctx *context, name, src string) (*cpp, tokenReader, error) {
	return testCPPParseSource(ctx, NewStringSource(name, src))
}

func TestCPPParse0(t *testing.T) {
	ctx, err := newContext(&Tweaks{})
	if err != nil {
		t.Fatal(err)
	}

	for i, v := range []string{
		"",
		"\n",
		"foo\n",
		`#if 1
#endif
`,
		`#if 1
# /* foo */
#endif
`,
	} {
		if _, _, err := testCPPParseString(ctx, "test", v); err != nil {
			t.Error(i, err)
		}
	}
}

func TestCPPExpand(t *testing.T) {
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	if err := filepath.Walk(filepath.FromSlash("testdata/cpp-expand/"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (!strings.HasSuffix(path, ".c") && !strings.HasSuffix(path, ".h")) {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		t.Log(path)
		ctx, err := newContext(&Tweaks{
			cppExpandTest: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		ctx.model = model
		b, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		c, r, err := testCPPParseFile(ctx, path)
		if err != nil {
			t.Fatal(ErrString(err))
		}

		var tb tokenBuffer
		if err := c.eval(r, &tb); err != nil {
			t.Fatal(ErrString(err))
		}

		switch {
		case strings.Contains(filepath.ToSlash(path), "/mustfail/"):
			err := c.error()
			if err != nil {
				t.Logf(ErrString(err))
				return nil
			}

			t.Fatalf("unexpected success: %s", path)
		default:
			if err := c.error(); err != nil {
				t.Fatal(ErrString(err))
			}
		}

		var a []string
		for {
			t := tb.read()
			if t.Rune == ccEOF {
				break
			}

			a = append(a, TokSrc(t.Token))
		}
		s := strings.Join(a, "")
		exp, err := ioutil.ReadFile(path + ".expect")
		if err != nil {
			t.Fatal(err)
		}

		if g, e := s, string(exp); g != e {
			t.Errorf("\n---- src %s\n%s---- got\n%s---- exp %s\n%s", path, b, g, path+".expect", e)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func (b *tokenBuffer) WriteTo(fset *token.FileSet, w io.Writer) {
	var lpos token.Position
	for {
		t := b.read()
		if t.Rune == ccEOF {
			return
		}

		pos := fset.Position(t.Pos())
		if pos.Filename != lpos.Filename {
			fmt.Fprintf(w, "# %d %v\n", pos.Line, pos.Filename)
		}
		lpos = pos
		w.Write([]byte(TokSrc(t.Token)))
	}
}

func (b *tokenBuffer) Bytes(fset *token.FileSet) []byte {
	var buf bytes.Buffer
	b.WriteTo(fset, &buf)
	return buf.Bytes()
}

func TestPreprocessSQLite(t *testing.T) {
	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := newContext(&Tweaks{})
	if err != nil {
		t.Fatal(err)
	}

	ctx.model = model
	cpp := newCPP(ctx)
	cpp.includePaths = []string{"@"}
	cpp.sysIncludePaths = searchPaths
	r, err := cpp.parse(MustBuiltin(), MustFileSource(sqlite3c))
	if err != nil {
		t.Fatalf("%v: %v", sqlite3c, err)
	}

	var w tokenBuffer
	if err := cpp.eval(r, &w); err != nil {
		t.Fatalf("%v: %v", sqlite3c, ErrString(err))
	}

	if err := cpp.error(); err != nil {
		t.Fatalf("%v: %v", sqlite3c, ErrString(err))
	}

	if n := len(cpp.lx.ungetBuffer); n != 0 {
		t.Fatal(n)
	}
}

func TestParseSQLite(t *testing.T) {
	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := newContext(&Tweaks{
		EnableAnonymousStructFields: true,
		EnableEmptyStructs:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx.model = model
	ctx.includePaths = []string{"@"}
	ctx.sysIncludePaths = searchPaths
	if _, err := ctx.parse([]Source{MustBuiltin(), MustFileSource(sqlite3c)}); err != nil {
		t.Fatalf("%v", ErrString(err))
	}
}

func TestFunc(t *testing.T) {
	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := newContext(&Tweaks{InjectFinalNL: true})
	if err != nil {
		t.Fatal(err)
	}

	ctx.model = model
	tu, err := ctx.parse(
		[]Source{NewStringSource("testfunc.c", `int (*foo(char bar))(double baz){}`)},
	)
	if err != nil {
		t.Fatalf("%v", ErrString(err))
	}

	if err := tu.ExternalDeclarationList.check(ctx); err != nil {
		t.Fatal(err)
	}

	if err := ctx.error(); err != nil {
		t.Fatal(err)
	}

	fileScope := ctx.scope
	n := fileScope.LookupIdent(dict.SID("foo"))
	d, ok := n.(*Declarator)
	if !ok {
		t.Fatalf("%T", n)
	}

	fd := d
	if g, e := d.Type.String(), "function (char) returning pointer to function (double) returning int"; g != e {
		t.Fatalf("got %q\nexp %q", g, e)
	}

	if g, e := fmt.Sprint(d.Type.(*FunctionType).Params), "[char]"; g != e {
		t.Fatalf("got %q\nexp %q", g, e)
	}

	fnScope := tu.ExternalDeclarationList.ExternalDeclaration.FunctionDefinition.FunctionBody.CompoundStmt.scope
	n = fnScope.LookupIdent(dict.SID("bar"))
	if d, ok = n.(*Declarator); !ok {
		t.Fatalf("%T", n)
	}

	if g, e := fmt.Sprint(d.Type), "char"; g != e {
		t.Fatalf("got %q\nexp %q", g, e)
	}

	names := fd.ParameterNames()
	if g, e := len(names), 1; g != e {
		t.Fatal(g, e)
	}

	if g, e := names[0], dict.SID("bar"); g != e {
		t.Fatal(g, e)
	}

	params := fd.Parameters
	if g, e := len(params), 1; g != e {
		t.Fatal(g, e)
	}

	if g, e := params[0].Name(), dict.SID("bar"); g != e {
		t.Fatal(g, e)
	}
}

func TestTypecheckSQLite(t *testing.T) {
	if _, err := Translate(
		&Tweaks{
			EnableAnonymousStructFields: true,
			EnableEmptyStructs:          true,
		},
		[]string{"@"},
		searchPaths,
		MustBuiltin(),
		MustFileSource(sqlite3c),
	); err != nil {
		t.Fatal(err)
	}
}

func TestTypecheckSQLiteShell(t *testing.T) {
	if _, err := Translate(
		&Tweaks{
			EnableAnonymousStructFields: true,
			EnableEmptyStructs:          true,
		},
		[]string{"@"},
		searchPaths,
		MustBuiltin(),
		MustCrt0(),
		MustFileSource(shellc),
	); err != nil {
		t.Fatal(err)
	}
}

func TestTypecheckTCCTests(t *testing.T) {
	blacklist := map[string]struct{}{
		"34_array_assignment.c": {}, // gcc: main.c:16:6: error: incompatible types when assigning to type ‘int[4]’ from type ‘int *’
		"46_grep.c":             {}, // gcc: 46_grep.c:489:12: error: ‘documentation’ undeclared (first use in this function)
	}
	m, err := filepath.Glob("testdata/tcc-0.9.26/tests/tests2/*.c")
	if err != nil {
		t.Fatal(err)
	}

	for _, pth := range m {
		if _, ok := blacklist[filepath.Base(pth)]; ok {
			continue
		}

		if _, err := Translate(
			&Tweaks{
				EnableBinaryLiterals:        true,
				EnableEmptyStructs:          true,
				EnableImplicitDeclarations:  true,
				EnableReturnExprInVoidFunc:  true,
				EnableAnonymousStructFields: true,
			},
			[]string{"@"},
			searchPaths,
			MustBuiltin(),
			MustCrt0(),
			MustFileSource(pth),
		); err != nil {
			t.Fatal(ErrString(err))
		}
	}
}

func TestParseJhjourdan(t *testing.T) {
	var blacklist = map[string]struct{}{
		"bitfield_declaration_ambiguity.fail.c": {}, // fails only during typecheck
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	model, err := NewModel()
	if err != nil {
		t.Fatal(err)
	}

	var ok, n int
	if err := filepath.Walk(filepath.FromSlash("testdata/jhjourdan/"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".c") {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		ctx, err := newContext(&Tweaks{})
		if err != nil {
			t.Fatal(err)
		}

		ctx.model = model
		ctx.includePaths = []string{"@"}
		ctx.sysIncludePaths = searchPaths
		n++
		shouldFail := strings.HasSuffix(path, ".fail.c")
		//dbg("", path)
		switch _, err := ctx.parse([]Source{MustBuiltin(), MustFileSource(path)}); {
		case err != nil:
			if !shouldFail {
				t.Errorf("%v", ErrString(err))
				return nil
			}
		default:
			if shouldFail {
				t.Errorf("%v: unexpected success", path)
				return nil
			}
		}

		ok++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("jhjourdan parse\tok %v n %v\n", ok, n)
}

func TestTypecheckJhjourdan(t *testing.T) {
	var blacklist = map[string]struct{}{
		"bitfield_declaration_ambiguity.c":   {}, //TODO
		"dangling_else_lookahead.if.c":       {}, //TODO
		"designator.c":                       {}, //TODO
		"expressions.c":                      {}, //TODO
		"function_parameter_scope_extends.c": {}, //TODO
		"if_scopes.c":                        {}, //TODO
		"loop_scopes.c":                      {}, //TODO
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	var ok, n int
	if err := filepath.Walk(filepath.FromSlash("testdata/jhjourdan/"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".c") {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		if re != nil && !re.MatchString(path) {
			return nil
		}

		n++
		shouldFail := strings.HasSuffix(path, ".fail.c")
		_, err = Translate(
			&Tweaks{},
			[]string{"@"},
			searchPaths,
			MustBuiltin(),
			MustFileSource(path),
		)
		switch {
		case err != nil:
			if !shouldFail {
				dbg("%q, err: %v, shouldFail: %v", path, err, shouldFail)
				t.Errorf("%v", ErrString(err))
				return nil
			}
		default:
			if shouldFail {
				dbg("%q, err: %v, shouldFail: %v", path, err, shouldFail)
				t.Errorf("%v: unexpected success", path)
				return nil
			}
		}

		ok++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Logf("jhjourdan typecheck\tok %v n %v\n", ok, n)
}
