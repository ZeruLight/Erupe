// Copyright 2016 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc"

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"modernc.org/golex/lex"
	"modernc.org/mathutil"
	"modernc.org/xc"
)

func printStack() { debug.PrintStack() }

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "\tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string {
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("TODO: %s:%d:\n", path.Base(fn), fl)
}

func use(...interface{}) int { return 42 }

var _ = use(printStack, caller, dbg, TODO, (*ctype).str, yyDefault, yyErrCode, yyMaxDepth)

// ============================================================================

var (
	o1        = flag.String("1", "", "single file argument of TestPPParse1.")
	oDev      = flag.Bool("dev", false, "enable WIP tests")
	oFailFast = flag.Bool("ff", false, "crash on first reported error (in some tests.)")
	oRe       = flag.String("re", "", "regexp filter.")
	oTmp      = flag.Bool("tmp", false, "keep certain temp files.")
	oTrace    = flag.Bool("trc", false, "print testDev path")

	includes = []string{}

	predefinedMacros = `
#define __STDC_HOSTED__ 1
#define __STDC_VERSION__ 199901L
#define __STDC__ 1

#define __MODEL64

void __GO__(char *s, ...);
`
	sysIncludes = []string{}

	testTweaks = &tweaks{
		enableDefineOmitCommaBeforeDDD: true,
		enableDlrInIdentifiers:         true,
		enableEmptyDefine:              true,
		enableUndefExtraTokens:         true,
	}
)

func newTestReport() *xc.Report {
	r := xc.NewReport()
	r.ErrLimit = -1
	if *oFailFast {
		r.PanicOnError = true
	}
	return r
}

func init() {
	isTesting = true
	log.SetFlags(log.Llongfile)
	flag.BoolVar(&debugIncludes, "dbgi", false, "debug include searches")
	flag.BoolVar(&debugMacros, "dbgm", false, "debug macros")
	flag.BoolVar(&debugTypeStrings, "xtypes", false, "add debug info to type strings")
	flag.BoolVar(&isGenerating, "generating", false, "go generate is executing (false).")
	flag.IntVar(&yyDebug, "yydebug", 0, "")
}

func newTestModel() *Model {
	return &Model{ // 64
		Items: map[Kind]ModelItem{
			Ptr:               {8, 8, 8, nil},
			UintPtr:           {8, 8, 8, nil},
			Void:              {0, 1, 1, nil},
			Char:              {1, 1, 1, nil},
			SChar:             {1, 1, 1, nil},
			UChar:             {1, 1, 1, nil},
			Short:             {2, 2, 2, nil},
			UShort:            {2, 2, 2, nil},
			Int:               {4, 4, 4, nil},
			UInt:              {4, 4, 4, nil},
			Long:              {8, 8, 8, nil},
			ULong:             {8, 8, 8, nil},
			LongLong:          {8, 8, 8, nil},
			ULongLong:         {8, 8, 8, nil},
			Float:             {4, 4, 4, nil},
			Double:            {8, 8, 8, nil},
			LongDouble:        {16, 16, 16, nil},
			Bool:              {1, 1, 1, nil},
			FloatComplex:      {8, 8, 8, nil},
			DoubleComplex:     {16, 16, 16, nil},
			LongDoubleComplex: {16, 16, 16, nil},
		},
	}
}

func printError(w io.Writer, pref string, err error) {
	switch x := err.(type) {
	case scanner.ErrorList:
		for i, v := range x {
			fmt.Fprintf(w, "%s%v\n", pref, v)
			if i == 50 {
				fmt.Fprintln(w, "too many errors")
				break
			}
		}
	default:
		fmt.Fprintf(w, "%s%v\n", pref, err)
	}
}

func errString(err error) string {
	var b bytes.Buffer
	printError(&b, "", err)
	return b.String()
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

func charStr(c rune) string {
	return yySymName(int(c))
}

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

func testLexer(t *testing.T, newLexer func(i int, src string, report *xc.Report) (*lexer, error), tab lexerTests) {
nextTest:
	for ti, test := range tab {
		//dbg("==== %v", ti)
		report := xc.NewReport()
		lx, err := newLexer(ti, test.src, report)
		if err != nil {
			t.Fatal(err)
		}

		delta := token.Pos(lx.file.Base() - 1)
		var chars []lex.Char
		var c lex.Char
		for i := 0; c.Rune != ccEOF && i < len(test.src)+2; i++ {
			c = lx.scanChar()
			chars = append(chars, c)
		}
		if c.Rune != ccEOF {
			t.Errorf("%d: scanner stall %v", ti, charsStr(chars, delta))
			continue
		}

		if g, e := report.Errors(true), error(nil); g != e {
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
			if c.Rune == ccEOF {
				g = -1
			}
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
	testLexer(
		t,
		func(i int, src string, report *xc.Report) (*lexer, error) {
			return newLexer(fmt.Sprintf("TestLexer.%d", i), len(src), strings.NewReader(src), report, testTweaks)
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

func TestLexer2(t *testing.T) {
	testLexer(
		t,
		func(i int, src string, report *xc.Report) (*lexer, error) {
			tweaks := *testTweaks
			tweaks.enableTrigraphs = true
			return newLexer(fmt.Sprintf("TestLexer.%d", i), len(src), strings.NewReader(src), report, &tweaks)
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

func testPreprocessor(t *testing.T, fname string) string {
	var buf bytes.Buffer
	_, err := Parse(
		"",
		[]string{fname},
		newTestModel(),
		preprocessOnly(),
		Cpp(func(toks []xc.Token) {
			//dbg("____ cpp toks\n%s", PrettyString(toks))
			for _, v := range toks {
				buf.WriteString(TokSrc(v))
			}
			buf.WriteByte('\n')
		}),
		EnableDefineOmitCommaBeforeDDD(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}
	return strings.TrimSpace(buf.String())
}

func TestStdExample6_10_3_3_4(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.3-4.h"), `char p[] = "x ## y";`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

func TestStdExample6_10_3_5_3(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.5-3.h"),
		`f(2 * (y+1)) + f(2 * (f(2 * (z[0])))) % f(2 * (0)) + t(1);
f(2 * (2+(3,4)-0,1)) | f(2 * (~ 5)) &
f(2 * (0,1))^m(0,1);
int i[] = { 1, 23, 4, 5,  };
char c[2][6] = { "hello", "" };`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

func TestStdExample6_10_3_5_4(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.5-4.h"),
		`printf("x1= %d, x2= %s", x1, x2);
fputs(
"strncmp(\"abc\\0d\", \"abc\", '\\4') == 0: @\n", s);
vers2.h included from testdata/example-6.10.3.5-4.h
"hello";
"hello, world"`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

func TestStdExample6_10_3_5_5(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.5-5.h"),
		`int j[] = { 123, 45, 67, 89,
10, 11, 12,  };`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

func TestStdExample6_10_3_5_6(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.5-6.h"),
		`ok`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

func TestStdExample6_10_3_5_7(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/example-6.10.3.5-7.h"),
		`fprintf(stderr, "Flag");
fprintf(stderr, "X = %d\n", x);
puts("The first, second, and third items.");
((x>y)?puts("x>y"): printf("x is %d but y is %d", x, y));`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

type cppCmpError struct {
	error
}

func testDev1(ppPredefine, cppPredefine, parsePredefine string, cppOpts []string, wd, src string, ppOpts, parseOpts []Opt) error {
	fp := filepath.Join(wd, src)
	if re := *oRe; re != "" {
		ok, err := regexp.MatchString(re, fp)
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}
	}

	logf, err := os.Create("log-" + filepath.Base(src))
	if err != nil {
		return err
	}

	defer logf.Close()

	logw := bufio.NewWriter(logf)

	defer logw.Flush()

	if *oTrace {
		fmt.Println(fp)
		fmt.Println(logf.Name())
	}

	var got, exp []xc.Token
	var lpos token.Position

	var tw tweaks
	_, err = Parse(
		ppPredefine,
		[]string{src},
		newTestModel(),
		append(
			ppOpts,
			getTweaks(&tw),
			preprocessOnly(),
			Cpp(func(toks []xc.Token) {
				if len(toks) != 0 {
					p := toks[0].Position()
					if p.Filename != lpos.Filename {
						fmt.Fprintf(logw, "# %d %q\n", p.Line, p.Filename)
					}
					lpos = p
				}
				for _, v := range toks {
					logw.WriteString(TokSrc(toC(v, &tw)))
					if v.Rune != ' ' {
						got = append(got, v)
					}
				}
				logw.WriteByte('\n')
			}),
			disableWarnings(),
			disablePredefinedLineMacro(),
		)...,
	)
	if err != nil {
		return err
	}

	out, err := exec.Command("cpp", append(cppOpts, src)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %v", src, err)
	}

	f, err := ioutil.TempFile("", "cc-test-")
	if err != nil {
		return err
	}

	if *oTrace {
		fmt.Println(f.Name())
	}
	defer func() {
		if !*oTmp {
			os.Remove(f.Name())
		}
		f.Close()
	}()

	if _, err := f.Write(out); err != nil {
		return err
	}

	if _, err := Parse(
		cppPredefine,
		[]string{f.Name()},
		newTestModel(),
		preprocessOnly(),
		Cpp(func(toks []xc.Token) {
			for _, tok := range toks {
				if tok.Rune != ' ' {
					exp = append(exp, tok)
				}
			}
		}),
		disableWarnings(),
	); err != nil {
		return err
	}

	for i, g := range got {
		if i >= len(exp) {
			break
		}

		g = toC(g, &tw)
		e := toC(exp[i], &tw)
		if g.Rune != e.Rune || g.Val != e.Val {

			if g.Rune == STRINGLITERAL && e.Rune == STRINGLITERAL && bytes.Contains(g.S(), []byte(fakeTime)) {
				continue
			}

			if g.Rune == IDENTIFIER && e.Rune == INTCONST && g.Val == idLine {
				n, err := strconv.ParseUint(string(e.S()), 10, mathutil.IntBits-1)
				if err != nil {
					return err
				}

				d := g.Position().Line - int(n)
				if d < 0 {
					d = -d
				}
				if d <= 3 {
					continue
				}
			}

			return cppCmpError{fmt.Errorf("%d\ngot %s\nexp %s", i, PrettyString(g), PrettyString(e))}
		}
	}

	if g, e := len(got), len(exp); g != e {
		return cppCmpError{fmt.Errorf("%v: got %d tokens, expected %d tokens (∆ %d)", src, g, e, g-e)}
	}

	logf2, err := os.Create("log2-" + filepath.Base(src))
	if err != nil {
		return err
	}

	defer logf2.Close()

	logw2 := bufio.NewWriter(logf2)

	defer logw2.Flush()

	if *oTrace {
		fmt.Println(logf2.Name())
	}

	_, err = Parse(
		parsePredefine,
		[]string{src},
		newTestModel(),
		append(
			parseOpts,
			disableWarnings(),
			Cpp(func(toks []xc.Token) {
				if len(toks) != 0 {
					p := toks[0].Position()
					if p.Filename != lpos.Filename {
						fmt.Fprintf(logw2, "# %d %q\n", p.Line, p.Filename)
					}
					lpos = p
				}
				for _, v := range toks {
					logw2.WriteString(TokSrc(toC(v, &tw)))
				}
				logw2.WriteByte('\n')
			}),
		)...,
	)
	return err
}

func testDev(t *testing.T, ppPredefine, cppPredefine, parsePredefine string, cppOpts, src []string, wd string, ppOpts, parseOpts []Opt) {
	if !dirExists(t, wd) {
		t.Logf("skipping: %v", wd)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(cwd)

	for _, src := range src {
		fi, err := os.Stat(src)
		if err != nil {
			t.Error(err)
			continue
		}

		if !fi.Mode().IsRegular() {
			t.Errorf("not a regular file: %s", filepath.Join(wd, src))
			continue
		}

		if err := testDev1(ppPredefine, cppPredefine, parsePredefine, cppOpts, wd, src, ppOpts, parseOpts); err != nil {
			t.Error(errString(err))
		}
	}
}

func dirExists(t *testing.T, dir string) bool {
	dir = filepath.FromSlash(dir)
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		t.Fatal(err)
	}

	if !fi.IsDir() {
		t.Fatal(dir, "is not a directory")
	}

	return true
}

func TestPreprocessor(t *testing.T) {
	if err := testDev1("", "", "", nil, "", "testdata/arith-1.h", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestDevSDL(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		EnableIncludeNext(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	testDev(
		t,
		predefined,
		predefined,
		predefined+`
#define __inline inline
`,
		nil,
		[]string{
			"SDL.h",
		},
		"testdata/dev/SDL-1.2.15/include/",
		ppOpts,
		parseOpts,
	)
}

func TestDevSqlite(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		EnableIncludeNext(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	testDev(
		t,
		predefined,
		predefined,
		predefined+`
#define __const const
#define __inline inline
#define __restrict restrict
`,
		nil,
		[]string{
			"shell.c",
			"sqlite3.c",
			"sqlite3.h",
			"sqlite3ext.h",
		},
		"testdata/dev/sqlite3",
		ppOpts,
		parseOpts,
	)
}

func TestDevVim(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
			"proto",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		EnableDefineOmitCommaBeforeDDD(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
			"proto",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined + `
#define _FORTIFY_SOURCE 1
#define HAVE_CONFIG_H
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict restrict
#define __typeof typeof
`,
		[]string{
			"-I.",
			"-Iproto",
			"-DHAVE_CONFIG_H",
			"-U_FORTIFY_SOURCE",
			"-D_FORTIFY_SOURCE=1",
		},
		[]string{
			"auto/pathdef.c",
			"blowfish.c",
			"buffer.c",
			"channel.c",
			"charset.c",
			"crypt.c",
			"crypt_zip.c",
			"diff.c",
			"digraph.c",
			"edit.c",
			"eval.c",
			"ex_cmds.c",
			"ex_cmds2.c",
			"ex_docmd.c",
			"ex_eval.c",
			"ex_getln.c",
			"fileio.c",
			"fold.c",
			"getchar.c",
			"hardcopy.c",
			"hashtab.c",
			"if_cscope.c",
			"if_xcmdsrv.c",
			"json.c",
			"main.c",
			"mark.c",
			"mbyte.c",
			"memfile.c",
			"memline.c",
			"menu.c",
			"message.c",
			"misc1.c",
			"misc2.c",
			"move.c",
			"netbeans.c",
			"normal.c",
			"ops.c",
			"option.c",
			"os_unix.c",
			"popupmnu.c",
			"quickfix.c",
			"regexp.c",
			"screen.c",
			"search.c",
			"sha256.c",
			"spell.c",
			"syntax.c",
			"tag.c",
			"term.c",
			"ui.c",
			"undo.c",
			"version.c",
			"window.c",
		},
		"testdata/dev/vim/vim/src",
		ppOpts,
		parseOpts,
	)
}

func TestDevBash(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
			"include",
			"lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
			"include",
			"lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined + `
#define PROGRAM "bash"
#define CONF_HOSTTYPE "x86_64"
#define CONF_OSTYPE "linux-gnu"
#define CONF_MACHTYPE "x86_64-unknown-linux-gnu"
#defien CONF_VENDOR "unknown"
#define LOCALEDIR "/usr/local/share/locale"
#define PACKAGE "bash"
#define SHELL
#define HAVE_CONFIG_H
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
#define __typeof typeof
		`,
		[]string{
			`-DPROGRAM="bash"`,
			`-DCONF_HOSTTYPE="x86_64"`,
			`-DCONF_OSTYPE="linux-gnu"`,
			`-DCONF_MACHTYPE="x86_64-unknown-linux-gnu"`,
			`-DCONF_VENDOR="unknown"`,
			`-DLOCALEDIR="/usr/local/share/locale"`,
			`-DPACKAGE="bash"`,
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-Iinclude",
			"-Ilib",
		},
		[]string{
			"alias.c",
			"array.c",
			"arrayfunc.c",
			"assoc.c",
			"bashhist.c",
			"bashline.c",
			"bracecomp.c",
			"braces.c",
			"copy_cmd.c",
			"dispose_cmd.c",
			"error.c",
			"eval.c",
			"expr.c",
			"findcmd.c",
			"flags.c",
			"general.c",
			"hashcmd.c",
			"hashlib.c",
			"input.c",
			"jobs.c",
			"list.c",
			"locale.c",
			"mailcheck.c",
			"make_cmd.c",
			"mksyntax.c",
			"pathexp.c",
			"pcomplete.c",
			"pcomplib.c",
			"print_cmd.c",
			"redir.c",
			"shell.c",
			"sig.c",
			"stringlib.c",
			"subst.c",
			"support/bashversion.c",
			"support/mksignames.c",
			"support/signames.c",
			"syntax.c",
			"test.c",
			"trap.c",
			"unwind_prot.c",
			"variables.c",
			"version.c",
			"version.c",
			"xmalloc.c",
			"y.tab.c",
			//"execute_cmd.c", // Composite type K&R fn def style vs prototype decl lefts an undefined param.
		},
		"testdata/dev/bash-4.3/",
		ppOpts,
		parseOpts,
	)

	ppOpts = []Opt{
		IncludePaths([]string{
			".",
			"..",
			"../include",
			"../lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts = []Opt{
		IncludePaths([]string{
			".",
			"..",
			"../include",
			"../lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p = predefined + `
#define HAVE_CONFIG_H
#define SHELL
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __restrict __restrict__
#define __inline inline
`,
		[]string{
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I..",
			"-I../include",
			"-I../lib",
		},
		[]string{
			"builtins.c",
			"common.c",
			"evalfile.c",
			"evalstring.c",
			"mkbuiltins.c",
			"psize.c",
		},
		"testdata/dev/bash-4.3/builtins",
		ppOpts,
		parseOpts,
	)

	ppOpts = []Opt{
		IncludePaths([]string{
			".",
			"../..",
			"../../include",
			"../../lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts = []Opt{
		IncludePaths([]string{
			".",
			"../..",
			"../../include",
			"../../lib",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p = predefined + `
#define HAVE_CONFIG_H
#define SHELL
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I../..",
			"-I../../include",
			"-I../../lib",
		},
		[]string{
			"glob.c",
			"gmisc.c",
			"smatch.c",
			"strmatch.c",
			"xmbsrtowcs.c",
		},
		"testdata/dev/bash-4.3/lib/glob",
		ppOpts,
		parseOpts,
	)

	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I../..",
			"-I../../include",
			"-I../../lib",
		},
		[]string{
			"casemod.c",
			"clktck.c",
			"clock.c",
			"eaccess.c",
			"fmtullong.c",
			"fmtulong.c",
			"fmtumax.c",
			"fnxform.c",
			"fpurge.c",
			"getenv.c",
			"input_avail.c",
			"itos.c",
			"mailstat.c",
			"makepath.c",
			"mbscasecmp.c",
			"mbschr.c",
			"mbscmp.c",
			"netconn.c",
			"netopen.c",
			"oslib.c",
			"pathcanon.c",
			"pathphys.c",
			"setlinebuf.c",
			"shmatch.c",
			"shmbchar.c",
			"shquote.c",
			"shtty.c",
			"snprintf.c",
			"spell.c",
			"stringlist.c",
			"stringvec.c",
			"strnlen.c",
			"strtrans.c",
			"timeval.c",
			"tmpfile.c",
			"uconvert.c",
			"ufuncs.c",
			"unicode.c",
			"wcsdup.c",
			"wcsnwidth.c",
			"winsize.c",
			"zcatfd.c",
			"zgetline.c",
			"zmapfd.c",
			"zread.c",
			"zwrite.c",
		},
		"testdata/dev/bash-4.3/lib/sh",
		ppOpts,
		parseOpts,
	)

	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I../..",
			"-I../../include",
			"-I../../lib",
		},
		[]string{
			"bind.c",
			"callback.c",
			"colors.c",
			"compat.c",
			"complete.c",
			"display.c",
			"funmap.c",
			"histexpand.c",
			"histfile.c",
			"history.c",
			"histsearch.c",
			"input.c",
			"isearch.c",
			"keymaps.c",
			"kill.c",
			"macro.c",
			"mbutil.c",
			"misc.c",
			"nls.c",
			"parens.c",
			"parse-colors.c",
			"readline.c",
			"rltty.c",
			"savestring.c",
			"search.c",
			"shell.c",
			"signals.c",
			"terminal.c",
			"text.c",
			"tilde.c",
			"undo.c",
			"util.c",
			"vi_mode.c",
			"xfree.c",
			"xmalloc.c",
		},
		"testdata/dev/bash-4.3/lib/readline",
		ppOpts,
		parseOpts,
	)

	p = predefined + `
#define HAVE_CONFIG_H
#define SHELL
#define RCHECK
#define botch programming_error
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DSHELL",
			"-DHAVE_CONFIG_H",
			"-DRCHECK",
			"-Dbotch=programming_error",
			"-I.",
			"-I../..",
			"-I../../include",
			"-I../../lib",
		},
		[]string{
			"malloc.c",
			"trace.c",
			"stats.c",
			"table.c",
			"watch.c",
		},
		"testdata/dev/bash-4.3/lib/malloc",
		ppOpts,
		parseOpts,
	)
}

func TestDevMake(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined + `
#define LOCALEDIR "/usr/local/share/locale"
#define LIBDIR "/usr/local/lib"
#define INCLUDEDIR "/usr/local/include"
#define HAVE_CONFIG_H
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
#define __typeof typeof
`,
		[]string{
			"-DLOCALEDIR=\"/usr/local/share/locale\"",
			"-DLIBDIR=\"/usr/local/lib\"",
			"-DINCLUDEDIR=\"/usr/local/include\"",
			"-DHAVE_CONFIG_H",
			"-I.",
		},
		[]string{
			"ar.c",
			"arscan.c",
			"commands.c",
			"default.c",
			"dir.c",
			"expand.c",
			"file.c",
			"function.c",
			"getopt.c",
			"getopt1.c",
			"guile.c",
			"hash.c",
			"implicit.c",
			"job.c",
			"load.c",
			"loadapi.c",
			"main.c",
			"misc.c",
			"output.c",
			"read.c",
			"remake.c",
			"remote-stub.c",
			"rule.c",
			"signame.c",
			"strcache.c",
			"variable.c",
			"version.c",
			"vpath.c",
		},
		"testdata/dev/make-4.1/",
		ppOpts,
		parseOpts,
	)
}

func TestDevBc(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
			"..",
			"./../h",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
			"..",
			"./../h",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined + `
#define HAVE_CONFIG_H
`
	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I..",
			"-I./../h",
		},
		[]string{
			"getopt.c",
			"getopt1.c",
			"number.c",
			"vfprintf.c",
		},
		"testdata/dev/bc-1.06/lib/",
		ppOpts,
		parseOpts,
	)

	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I..",
			"-I./../h",
		},
		[]string{
			"main.c",
			"bc.c",
			"scan.c",
			"execute.c",
			"load.c",
			"storage.c",
			"util.c",
			"global.c",
		},
		"testdata/dev/bc-1.06/bc",
		ppOpts,
		parseOpts,
	)

	testDev(
		t,
		p,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I..",
			"-I./../h",
		},
		[]string{
			"dc.c",
			"misc.c",
			"eval.c",
			"stack.c",
			"array.c",
			"numeric.c",
			"string.c",
		},
		"testdata/dev/bc-1.06/dc",
		ppOpts,
		parseOpts,
	)
}

func TestDevEmacs(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
			"../lib",
			"../src",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
			"../lib",
			"../src",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined + `
#define HAVE_CONFIG_H
#define _GCC_MAX_ALIGN_T
`
	testDev(
		t,
		p+`
#define _Noreturn __attribute__ ((__noreturn__))
`,
		p,
		p+`
#define __const const
#define __getopt_argv_const const
#define __inline inline
#define __restrict __restrict__
`,
		[]string{
			"-std=gnu99",
			"-DHAVE_CONFIG_H",
			"-I.",
			"-I../lib",
			"-I../src",
		},
		[]string{
			"acl-errno-valid.c",
			"allocator.c",
			"binary-io.c",
			"c-ctype.c",
			"c-strcasecmp.c",
			"c-strncasecmp.c",
			"careadlinkat.c",
			"close-stream.c",
			"count-one-bits.c",
			"count-trailing-zeros.c",
			"dtoastr.c",
			"dtotimespec.c",
			"fcntl.c",
			"file-has-acl.c",
			"filemode.c",
			"getopt.c",
			"getopt1.c",
			"gettime.c",
			"md5.c",
			"openat-die.c",
			"pipe2.c",
			"pthread_sigmask.c",
			"qcopy-acl.c",
			"qset-acl.c",
			"save-cwd.c",
			"sha1.c",
			"sha256.c",
			"sha512.c",
			"sig2str.c",
			"stat-time.c",
			"strftime.c",
			"timespec-add.c",
			"timespec-sub.c",
			"timespec.c",
			"u64.c",
			"unistd.c",
			"utimens.c",
		},
		"testdata/dev/emacs-24.5/lib",
		ppOpts,
		parseOpts,
	)

	p = predefined + `
 #define CTAGS
 #define EMACS_NAME "GNU Emacs"
 #define HAVE_CONFIG_H
 #define HAVE_SHARED_GAME_DIR "/usr/local/var/games/emacs"
 #define VERSION "24.5"
 `
	testDev(
		t,
		p+`
#define _GCC_MAX_ALIGN_T
#define _Noreturn __attribute__ ((__noreturn__))
`,
		p,
		p+`
#define __const const
#define __inline inline
#define __restrict __restrict__
#define __typeof typeof
`,
		[]string{
			"-std=gnu99",
			"-I.",
			"-I../lib",
			"-I../src",
			"-DEMACS_NAME=\"GNU Emacs\"",
			"-DCTAGS",
			"-DHAVE_SHARED_GAME_DIR=\"/usr/local/var/games/emacs\"",
			"-DVERSION=\"24.5\"",
		},
		[]string{
			"./../src/regex.c",
			"./ebrowse.c",
			"./emacsclient.c",
			"./etags.c",
			"./hexl.c",
			"./make-docfile.c",
			"./movemail.c",
			"./pop.c",
			"./profile.c",
			"./test-distrib.c",
			"./update-game-score.c",
		},
		"testdata/dev/emacs-24.5/lib-src/",
		ppOpts,
		parseOpts,
	)

	ppOpts = []Opt{
		IncludePaths([]string{
			".",
			"../lib",
			"/usr/include/gtk-3.0",
			"/usr/include/pango-1.0",
			"/usr/include/gio-unix-2.0/",
			"/usr/include/atk-1.0",
			"/usr/include/cairo",
			"/usr/include/gdk-pixbuf-2.0",
			"/usr/include/freetype2",
			"/usr/include/glib-2.0",
			"/usr/lib/x86_64-linux-gnu/glib-2.0/include",
			"/usr/include/pixman-1",
			"/usr/include/libpng12",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		EnableDefineOmitCommaBeforeDDD(),
		EnableIncludeNext(),
		devTest(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts = []Opt{
		IncludePaths([]string{
			".",
			"../lib",
			"/usr/include/gtk-3.0",
			"/usr/include/pango-1.0",
			"/usr/include/gio-unix-2.0/",
			"/usr/include/atk-1.0",
			"/usr/include/cairo",
			"/usr/include/gdk-pixbuf-2.0",
			"/usr/include/freetype2",
			"/usr/include/glib-2.0",
			"/usr/lib/x86_64-linux-gnu/glib-2.0/include",
			"/usr/include/pixman-1",
			"/usr/include/libpng12",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p = predefined + `
#define _GCC_MAX_ALIGN_T
#define emacs
`

	testDev(
		t,
		p+`
#define _Noreturn __attribute__ ((__noreturn__))
`,
		p,
		p+`
#define _Alignas(x)
#define __const const
#define __inline inline
#define __restrict __restrict__
#define __typeof typeof
`,
		[]string{
			"-std=gnu99",
			"-Demacs",
			"-I.",
			"-I../lib",
			"-I/usr/include/gtk-3.0",
			"-I/usr/include/pango-1.0",
			"-I/usr/include/gio-unix-2.0/",
			"-I/usr/include/atk-1.0",
			"-I/usr/include/cairo",
			"-I/usr/include/gdk-pixbuf-2.0",
			"-I/usr/include/freetype2",
			"-I/usr/include/glib-2.0",
			"-I/usr/lib/x86_64-linux-gnu/glib-2.0/include",
			"-I/usr/include/pixman-1",
			"-I/usr/include/libpng12",
		},
		[]string{
			"alloc.c",
			"atimer.c",
			"bidi.c",
			"buffer.c",
			"callint.c",
			"callproc.c",
			"casefiddle.c",
			"casetab.c",
			"category.c",
			"ccl.c",
			"character.c",
			"charset.c",
			"chartab.c",
			"cm.c",
			"cmds.c",
			"coding.c",
			"composite.c",
			"data.c",
			"decompress.c",
			"dired.c",
			"dispnew.c",
			"doc.c",
			"doprnt.c",
			"editfns.c",
			"emacs.c",
			"eval.c",
			"fileio.c",
			"filelock.c",
			"floatfns.c",
			"fns.c",
			"font.c",
			"fontset.c",
			"frame.c",
			"fringe.c",
			"ftfont.c",
			"gfilenotify.c",
			"gnutls.c",
			"gtkutil.c",
			"indent.c",
			"insdel.c",
			"intervals.c",
			"keyboard.c",
			"keymap.c",
			"lastfile.c",
			"lread.c",
			"macros.c",
			"marker.c",
			"menu.c",
			"minibuf.c",
			"print.c",
			"process.c",
			"profiler.c",
			"region-cache.c",
			"scroll.c",
			"search.c",
			"sound.c",
			"syntax.c",
			"sysdep.c",
			"term.c",
			"terminal.c",
			"terminfo.c",
			"textprop.c",
			"undo.c",
			"vm-limit.c",
			"window.c",
			"xdisp.c",
			"xfaces.c",
			"xfns.c",
			"xfont.c",
			"xgselect.c",
			"xmenu.c",
			"xml.c",
			"xrdb.c",
			"xselect.c",
			"xsmfns.c",
			"xterm.c",
			/// "bytecode.c",      // [lo ... hi] = expr
			/// "emacsgtkfixed.c", // /usr/include/gtk-3.0/gtk/gtkversion.h:98:9: cannot redefine macro: argument names differ
			/// "ftxfont.c",       // ftxfont.c:145:7: undefined: ftfont_driver
			/// "unexelf.c",       // /usr/include/x86_64-linux-gnu/bits/link.h:97:3: unexpected identifier __int128_t, expected one of ['}', _Bool, _Complex, _Static_assert, char, const, double, enum, float, int, long, restrict, short, signed, struct, typedefname, typeof, union, unsigned, void, volatile]
			/// "xftfont.c",       // lisp.h:2041:13: unexpected typedefname, expected optional type qualifier list or pointer or one of ['(', ')', '*', ',', '[', const, identifier, restrict, volatile]
			/// "xsettings.c",     // xsettings.c:431:36: unexpected '{', expected expression list or type name or one of ['!', '&', '(', '*', '+', '-', '~', ++, --, _Alignof, _Bool, _Complex, char, character constant, const, double, enum, float, floating-point constant, identifier, int, integer constant, long, long character constant, long string constant, restrict, short, signed, sizeof, string literal, struct, typedefname, typeof, union, unsigned, void, volatile]
			///"image.c",         // /usr/include/gif_lib.h:269:44: unexpected typedefname, expected optional type qualifier list or pointer or one of ['(', ')', '*', ',', '[', const, identifier, restrict, volatile]
		},
		"testdata/dev/emacs-24.5/src/",
		ppOpts,
		parseOpts,
	)
}
func TestDevM4(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths([]string{
			".",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		EnableIncludeNext(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths([]string{
			".",
		}),
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	p := predefined

	testDev(
		t,
		p,
		p,
		p+`
#define __const
#define __inline inline
#define __restrict __restrict__
#define __typeof typeof
`,
		[]string{
			"-I.",
		},
		[]string{
			"asnprintf.c",
			"asprintf.c",
			"basename-lgpl.c",
			"basename.c",
			"binary-io.c",
			"c-ctype.c",
			"c-strcasecmp.c",
			"c-strncasecmp.c",
			"clean-temp.c",
			"cloexec.c",
			"close-stream.c",
			"closein.c",
			"closeout.c",
			"dirname-lgpl.c",
			"dirname.c",
			"dup-safer-flag.c",
			"dup-safer.c",
			"exitfail.c",
			"fatal-signal.c",
			"fclose.c",
			"fcntl.c",
			"fd-hook.c",
			"fd-safer-flag.c",
			"fd-safer.c",
			"fflush.c",
			"filenamecat-lgpl.c",
			"filenamecat.c",
			"fopen-safer.c",
			"fpurge.c",
			"freadahead.c",
			"freading.c",
			"fseek.c",
			"fseeko.c",
			"gl_avltree_oset.c",
			"gl_linkedhash_list.c",
			"gl_list.c",
			"gl_oset.c",
			"gl_xlist.c",
			"gl_xoset.c",
			"glthread/lock.c",
			"glthread/threadlib.c",
			"glthread/tls.c",
			"localcharset.c",
			"malloca.c",
			"math.c",
			"memchr2.c",
			"mkstemp-safer.c",
			"pipe-safer.c",
			"pipe2-safer.c",
			"pipe2.c",
			"printf-args.c",
			"printf-frexp.c",
			"printf-parse.c",
			"progname.c",
			"quotearg.c",
			"sig-handler.c",
			"stripslash.c",
			"tempname.c",
			"tmpdir.c",
			"unistd.c",
			"vasprintf.c",
			"verror.c",
			"version-etc-fsf.c",
			"version-etc.c",
			"wait-process.c",
			"wctype-h.c",
			"xalloc-die.c",
			"xasprintf.c",
			"xmalloc.c",
			"xmalloca.c",
			"xprintf.c",
			"xsize.c",
			"xstrndup.c",
			"xvasprintf.c",
			/// "c-stack.c", c-stack.c:119:3: unexpected '{', expected expression list or type name or one of ['!', '&', '(', '*', '+', '-', '~', ++, --, _Alignof, _Bool, _Complex, char, character constant, const, double, enum, float, floating-point constant, identifier, int, integer constant, long, long character constant, long string constant, restrict, short, signed, sizeof, string literal, struct, typedefname, typeof, union, unsigned, void, volatile]
			/// "execute.c", spawn.h:457:9: cannot redefine macro using a replacement list of different length
			/// "isnanl.c", float+.h:145:32: array size must be positive: -1
			/// "printf-frexpl.c", printf-frexp.c:72:3: unexpected '{', expected expression list or type name or one of ['!', '&', '(', '*', '+', '-', '~', ++, --, _Alignof, _Bool, _Complex, char, character constant, const, double, enum, float, floating-point constant, identifier, int, integer constant, long, long character constant, long string constant, restrict, short, signed, sizeof, string literal, struct, typedefname, typeof, union, unsigned, void, volatile]
			/// "spawn-pipe.c", spawn.h:457:9: cannot redefine macro using a replacement list of different length
			/// "vasnprintf.c", vasnprintf.c:3624:25: unexpected '{', expected expression list or type name or one of ['!', '&', '(', '*', '+', '-', '~', ++, --, _Alignof, _Bool, _Complex, char, character constant, const, double, enum, float, floating-point constant, identifier, int, integer constant, long, long character constant, long string constant, restrict, short, signed, sizeof, string literal, struct, typedefname, typeof, union, unsigned, void, volatile]
		},
		"testdata/dev/m4-1.4.17/lib/",
		ppOpts,
		parseOpts,
	)
}

func TestDevStbVorbis(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		EnableIncludeNext(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	testDev(
		t,
		predefined,
		predefined,
		predefined+`
#define __inline inline
#define __const const
#define __restrict restrict
`,
		nil,
		[]string{
			"stb_vorbis.c",
		},
		"testdata/dev/stb",
		ppOpts,
		parseOpts,
	)
}

func TestDevGMP(t *testing.T) {
	predefined, includePaths, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Logf("skipping: %v", err)
		return
	}

	ppOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		EnableIncludeNext(),
	}
	if *oFailFast {
		ppOpts = append(ppOpts, CrashOnError())
	}
	parseOpts := []Opt{
		IncludePaths(includePaths),
		SysIncludePaths(sysIncludePaths),
		devTest(),
		gccEmu(),
	}
	if *oFailFast {
		parseOpts = append(parseOpts, CrashOnError())
	}

	testDev(
		t,
		predefined,
		predefined,
		predefined,
		nil,
		[]string{
			"gmp.h",
		},
		"testdata/dev/gmp-6.1.0/",
		ppOpts,
		parseOpts,
	)
}

func TestPPParse1(t *testing.T) {
	path := *o1
	if path == "" {
		return
	}

	testReport := newTestReport()
	testReport.ClearErrors()
	_, err := ppParse(path, testReport, testTweaks)
	if err != nil {
		t.Fatal(err)
	}

	if err := testReport.Errors(true); err != nil {
		t.Fatal(errString(err))
	}
}

func TestFinalInjection(t *testing.T) {
	const src = "int f() {}"

	if strings.HasSuffix(src, "\n") {
		t.Fatal("internal error")
	}

	ast, err := ppParseString("test.c", src, xc.NewReport(), &tweaks{})
	if err != nil {
		t.Fatal(errString(err))
	}

	t.Log(PrettyString(ast))
}

func TestRedecl(t *testing.T) {
	testParse(t, []string{"testdata/redecl.c"}, "")
}

func TestParse1(t *testing.T) {
	path := *o1
	if path == "" {
		return
	}

	testParse(t, []string{path}, "")
}

func testParse(t *testing.T, paths []string, ignoreError string, opts ...Opt) *TranslationUnit {
	last := paths[len(paths)-1]
	ln := filepath.Base(last)
	f, err := os.Create("log-" + ln)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	b := bufio.NewWriter(f)
	defer b.Flush()

	b.WriteString("// +build ignore\n\n")
	var a []string
	crash := nopOpt()
	if *oFailFast {
		crash = CrashOnError()
	}
	opts = append(
		opts,
		IncludePaths(includes),
		SysIncludePaths(sysIncludes),
		Cpp(func(toks []xc.Token) {
			a = a[:0]
			for _, v := range toks {
				a = append(a, TokSrc(v))
			}
			fmt.Fprintf(b, "%s\n", strings.Join(a, " "))
		}),
		crash,
		ErrLimit(-1),
	)
	ast, err := Parse(
		predefinedMacros,
		paths,
		newTestModel(),
		opts...,
	)
	if err != nil {
		if s := strings.TrimSpace(errString(err)); s != ignoreError {
			t.Fatal(errString(err))
		}
	}

	t.Log(paths)
	return ast
}

func ddsStr(dds []*DirectDeclarator) string {
	buf := bytes.NewBufferString("|")
	for i, dd := range dds {
		if i == 0 {
			fmt.Fprintf(buf, "(@%p)", &dds[0])
		}
		switch dd.Case {
		case 0: // IDENTIFIER
			fmt.Fprintf(buf, "IDENTIFIER(%s: %s)", dd.Token.Position(), dd.Token.S())
		case 1: // '(' Declarator ')'
			buf.WriteString("(")
			buf.WriteString(strings.Repeat("*", dd.Declarator.stars()))
			fmt.Fprintf(buf, "Declarator.%v)", dds[i-1].Case)
		case 2: // DirectDeclarator '[' TypeQualifierListOpt ExpressionOpt ']'
			fmt.Fprintf(buf, "DirectDeclarator[TypeQualifierListOpt ExpressionOpt.%v]", dd.elements)
		case 3: // DirectDeclarator '[' "static" TypeQualifierListOpt Expression ']'
			fmt.Fprintf(buf, "DirectDeclarator[static TypeQualifierListOpt Expression.%v]", dd.elements)
		case 4: // DirectDeclarator '[' TypeQualifierList "static" Expression ']'
			fmt.Fprintf(buf, "DirectDeclarator[TypeQualifierList static Expression.%v]", dd.elements)
		case 5: // DirectDeclarator '[' TypeQualifierListOpt '*' ']'
			fmt.Fprintf(buf, "DirectDeclarator[TypeQualifierListOpt*.%v]", dd.elements)
		case 6: // DirectDeclarator '(' ParameterTypeList ')'
			buf.WriteString("DirectDeclarator(ParameterTypeList)")
		case 7: // DirectDeclarator '(' IdentifierListOpt ')'
			buf.WriteString("DirectDeclarator(IdentifierListOpt)")
		}
		buf.WriteString("|")
	}
	return buf.String()
}

func (n *ctype) str() string {
	return fmt.Sprintf("R%v S%v %v", n.resultStars, n.stars, ddsStr(n.dds))
}

func (n *ctype) str0() string {
	return fmt.Sprintf("R%v S%v %v", n.resultStars, n.stars, ddsStr(n.dds0))
}

func TestIssue3(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue3.h"}, newTestModel()); err != nil {
		t.Fatal(errString(err))
	}
}

func TestIssue8(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue8.h"}, newTestModel()); err != nil {
		t.Fatal(errString(err))
	}
}

func TestIssue4(t *testing.T) {
	_, err := Parse("", []string{"testdata/issue4.c"}, newTestModel())
	if err == nil {
		t.Fatal("unexpected sucess")
	}

	l, ok := err.(scanner.ErrorList)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}

	if g, e := l.Len(), 2; g != e {
		t.Fatal(g, e)
	}

	if g, e := l[0].Error(), "testdata/issue4.c:5:13: redeclaration of foo as different kind of symbol, previous declaration at testdata/issue4.c:4:5"; g != e {
		t.Fatal(g, e)
	}

	if g, e := l[1].Error(), "testdata/issue4.c:9:15: redeclaration of foo2 as different kind of symbol, previous declaration at testdata/issue4.c:8:7"; g != e {
		t.Fatal(g, e)
	}
}

func unpackType(typ Type) Type {
	for {
		switch typ.Kind() {
		case Ptr, Array:
			typ = typ.Element()
		default:
			return typ
		}
	}
}

func TestIssue9(t *testing.T) {
	const exp = `original:  typedef short[64] Array
unpacked:  typedef short Short
original: JBLOCK short(*)[64] Ptr
unpacked: JBLOCK short Short
original: JBLOCKROW short(**)[64] Ptr
unpacked: JBLOCKROW short Short
original: JBLOCKARRAY short(***)[64] Ptr
unpacked: JBLOCKARRAY short Short
original:  short[64] Array
unpacked:  short Short
original:  short[64] Array
unpacked:  short Short
original: JBLOCK short[64] Array
unpacked: JBLOCK short Short
original: JBLOCK short[64] Array
unpacked: JBLOCK short Short
original:  short(*)[64] Ptr
unpacked:  short Short
original:  short(*)[64] Ptr
unpacked:  short Short
original:  short(*)[64] Ptr
unpacked:  short Short
original: JBLOCKROW short(*)[64] Ptr
unpacked: JBLOCKROW short Short
original: JBLOCKROW short(*)[64] Ptr
unpacked: JBLOCKROW short Short
original:  short(**)[64] Ptr
unpacked:  short Short
original:  short(**)[64] Ptr
unpacked:  short Short
original:  short(**)[64] Ptr
unpacked:  short Short
original: JBLOCKARRAY short(**)[64] Ptr
unpacked: JBLOCKARRAY short Short
original: JBLOCKARRAY short(**)[64] Ptr
unpacked: JBLOCKARRAY short Short
original:  short(***)[64] Ptr
unpacked:  short Short
original:  short(***)[64] Ptr
unpacked:  short Short
original:  short(***)[64] Ptr
unpacked:  short Short
original: JBLOCKIMAGE short(***)[64] Ptr
unpacked: JBLOCKIMAGE short Short
original: JBLOCKIMAGE short(***)[64] Ptr
unpacked: JBLOCKIMAGE short Short
`

	tu, err := Parse("", []string{"testdata/issue9.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	var buf bytes.Buffer
	for tu != nil {
		declr := tu.ExternalDeclaration.Declaration.InitDeclaratorListOpt.InitDeclaratorList.InitDeclarator.Declarator
		name := string(xc.Dict.S(declr.RawSpecifier().TypedefName()))
		fmt.Fprintln(&buf, "original:", name, declr.Type, declr.Type.Kind())
		unpacked := unpackType(declr.Type)
		fmt.Fprintln(&buf, "unpacked:", name, unpacked, unpacked.Kind())
		tu = tu.TranslationUnit
	}
	if g, e := buf.String(), exp; g != e {
		t.Fatalf("got:\n%s\nexp:\n%s", g, e)
	}
}

func TestEnumConstToks(t *testing.T) {
	tu, err := Parse("", []string{"testdata/enum.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	sc := tu.Declarations
	foo := sc.Lookup(NSIdentifiers, xc.Dict.SID("foo"))
	if foo.Node == nil {
		t.Fatal("internal error")
	}

	switch x := foo.Node.(type) {
	case *DirectDeclarator:
		typ := x.TopDeclarator().Type
		if g, e := typ.Kind(), Enum; g != e {
			t.Fatal(g, e)
		}

		l := typ.EnumeratorList()
		if g, e := PrettyString(
			l),
			`[]cc.EnumConstant{ // len 2
· 0: cc.EnumConstant{
· · DefTok: testdata/enum.c:4:2: IDENTIFIER "c",
· · Value: 18,
· · Tokens: []xc.Token{ // len 3
· · · 0: testdata/enum.c:4:6: INTCONST "42",
· · · 1: testdata/enum.c:4:9: '-',
· · · 2: testdata/enum.c:4:11: INTCONST "24",
· · },
· },
· 1: cc.EnumConstant{
· · DefTok: testdata/enum.c:5:2: IDENTIFIER "d",
· · Value: 592,
· · Tokens: []xc.Token{ // len 3
· · · 0: testdata/enum.c:5:6: INTCONST "314",
· · · 1: testdata/enum.c:5:10: '+',
· · · 2: testdata/enum.c:5:12: INTCONST "278",
· · },
· },
}`; g != e {
			t.Fatalf("got\n%s\nexp\n%s", g, e)
		}
		var a []string
		for _, e := range l {
			var b []string
			for _, t := range e.Tokens {
				b = append(b, TokSrc(t))
			}
			a = append(a, strings.Join(b, " "))
		}
		if g, e := strings.Join(a, "\n"), "42 - 24\n314 + 278"; g != e {
			t.Fatalf("got\n%s\nexp\n%s", g, e)
		}
	default:
		t.Fatalf("%T", x)
	}
}

func TestPaste(t *testing.T) {
	testParse(t, []string{"testdata/paste.c"}, "")
}

func TestPaste2(t *testing.T) {
	testParse(t, []string{"testdata/paste2.c"}, "")
}

func TestFunc(t *testing.T) {
	testParse(t, []string{"testdata/func.c"}, "")
}

func TestEmptyMacroArg(t *testing.T) {
	testParse(t, []string{"testdata/empty.c"}, "")
}

func TestFuncFuncParams(t *testing.T) {
	testParse(t, []string{"testdata/funcfunc.c"}, "")
}

func TestAnonStructField(t *testing.T) {
	testParse(
		t,
		[]string{"testdata/anon.c"},
		"testdata/anon.c:4:7: only unnamed structs and unions are allowed",
		EnableAnonymousStructFields(),
	)
}

func tokStr(toks []xc.Token) string {
	var b []byte
	for _, v := range toks {
		switch v.Rune {
		case sentinel:
			b = append(b, []byte("@:")...)
		case IDENTIFIER_NONREPL:
			b = append(b, []byte("-:")...)
		}
		b = append(b, xc.Dict.S(tokVal(v))...)
	}
	return string(b)
}

func tokStr2(toks [][]xc.Token) string {
	var a []string
	for _, v := range toks {
		a = append(a, tokStr(v))
	}
	return strings.Join(a, ", ")
}

func TestIssue50(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue50.h"}, newTestModel()); err == nil {
		t.Fatal("unexpected success")
	}
}

// https://gitlab.com/cznic/cc/issues/57
func TestIssue57(t *testing.T) {
	tu, err := Parse("", []string{"testdata/issue57.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	bool_func := tu.Declarations.Identifiers[dict.SID("bool_func")].Node.(*DirectDeclarator).TopDeclarator()
	typ := bool_func.Type
	if g, e := typ.String(), "int(*)()"; g != e {
		t.Fatalf("%q %q", g, e)
	}
	typ = typ.Element() // deref function pointer
	if g, e := typ.Result().String(), "int"; g != e {
		t.Fatalf("%q %q", g, e)
	}
	// bool_t -> ok!
	if g, e := typ.Result().RawDeclarator().RawSpecifier().TypedefName(), dict.SID("bool_t"); g != e {
		t.Fatal(g, e)
	}

	func1 := tu.Declarations.Identifiers[dict.SID("func1")].Node.(*DirectDeclarator).TopDeclarator()
	typ = func1.Type
	if g, e := typ.String(), "int(*)()"; g != e {
		t.Fatalf("%q %q", g, e)
	}
	typ = typ.Element() // deref function pointer
	if g, e := typ.String(), "int()"; g != e {
		t.Fatalf("%q %q", g, e)
	}
	if g, e := typ.Result().String(), "int"; g != e {
		t.Fatalf("%q %q", g, e)
	}
	// try to get bool_t the way we got it above
	if g, e := typ.Result().RawDeclarator().RawSpecifier().TypedefName(), dict.SID("bool_t"); g != e {
		t.Fatal(string(xc.Dict.S(g)), string(xc.Dict.S(e))) // bool_func, how to get bool_t?
	}
}

// https://gitlab.com/cznic/cc/issues/62
func TestIssue62(t *testing.T) {
	tu, err := Parse("", []string{"testdata/issue62.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	for ; tu != nil; tu = tu.TranslationUnit {
		d := tu.ExternalDeclaration.Declaration.Declarator()
		var e Linkage
		tag := string(xc.Dict.S(d.Type.Tag()))
		t.Logf("%s: %s", position(d.Pos()), tag)
		switch {
		case strings.HasPrefix(tag, "global"):
			e = External
		case strings.HasPrefix(tag, "local"):
			e = Internal
		}
		if g := d.Linkage; g != e {
			t.Fatalf("%v %v", g, e)
		}
	}
}

// https://gitlab.com/cznic/cc/issues/64
func TestIssue64(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue64.c"}, newTestModel()); err == nil {
		t.Fatal("expected error")
	} else {
		t.Log(errString(err))
	}
}

// https://gitlab.com/cznic/cc/issues/65
func TestIssue65(t *testing.T) {
	tu, err := Parse("", []string{"testdata/issue65.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	foo, ok := tu.Declarations.Identifiers[xc.Dict.SID("foo")]
	if !ok {
		t.Fatal("undefined: foo")
	}

	ft := foo.Node.(*DirectDeclarator).TopDeclarator().Type
	m, _ := ft.Members()
	tab := map[string]int{
		"i": -1,
		"j": 0,
		"k": 1,
		"l": 3,
		"m": -1,
	}
	for _, v := range m {
		ofs, ok := tab[string(xc.Dict.S(v.Name))]
		if !ok {
			t.Fatal(PrettyString(v))
		}

		if ofs < 0 {
			if v.Bits != 0 {
				t.Fatal(PrettyString(v))
			}
			continue
		}

		if v.Bits == 0 {
			t.Fatal(PrettyString(v))
		}

		if g, e := v.BitOffsetOf, ofs; g != e {
			t.Log(PrettyString(v))
			t.Fatal(g, e)
		}
	}
}

// https://gitlab.com/cznic/cc/issues/66
func TestIssue66(t *testing.T) {
	tu, err := Parse("", []string{"testdata/issue66.c"}, newTestModel())
	if err != nil {
		t.Fatal(errString(err))
	}

	e := tu.ExternalDeclaration.Declaration.InitDeclaratorListOpt.InitDeclaratorList.InitDeclarator.Initializer.Expression
	if e.Value == nil {
		t.Fatal("expected constant expression")
	}

	switch g := e.Value.(type) {
	case uintptr:
		if e := uintptr(13); g != e {
			t.Fatal(g, e)
		}
	default:
		t.Fatalf("%T(%#v)", g, g)
	}
}

// https://gitlab.com/cznic/cc/issues/67
func TestIssue67(t *testing.T) {
	tu, err := Parse("", []string{"testdata/issue67.c"}, newTestModel(), KeepComments())
	if err != nil {
		t.Fatal(errString(err))
	}

	var a []string
	for k, v := range tu.Comments {
		a = append(a, fmt.Sprintf("%s: %q", xc.FileSet.Position(k), xc.Dict.S(v)))
	}
	sort.Strings(a)
	if g, e := strings.Join(a, "\n"), `testdata/issue67.c:14:1: "/* abc11 */\n/* def12\n */"
testdata/issue67.c:19:1: "/* abc16\n */\n/* def18 */"
testdata/issue67.c:23:1: "/* abc21 */\n// def22"
testdata/issue67.c:27:1: "// def25\n/* abc26 */"
testdata/issue67.c:2:1: "// bar1"
testdata/issue67.c:32:1: "// def31"
testdata/issue67.c:5:1: "/*\nbaz3\n*/"
testdata/issue67.c:9:1: "// abc7\n// def8"`; g != e {
		t.Fatalf("got\n%s\nexp\n%s", g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/68
func TestIssue68(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue68.h"}, newTestModel()); err == nil {
		t.Fatal("expected error")
	}

	if _, err := Parse("", []string{"testdata/issue68.h"}, newTestModel(), EnableEmptyDeclarations()); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/69
func TestIssue69(t *testing.T) {
	if _, err := Parse("", []string{"testdata/issue69.h"}, newTestModel()); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/72
func TestIssue72(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue72.h"}, newTestModel(),
		EnableWideEnumValues(),
	); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/74
func TestIssue74EnableWideBitFieldTypes(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue74.h"}, newTestModel(),
		EnableWideBitFieldTypes(),
	); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/77
func TestIssue77(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue77.c"}, newTestModel(),
	); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/78
func TestIssue78(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue78.c"}, newTestModel(),
	); err == nil {
		t.Fatal("expected error")
	}

	tu, err := Parse(
		"", []string{"testdata/issue78.c"}, newTestModel(), EnableOmitFuncRetType(),
	)
	if err != nil {
		t.Fatal(err)
	}

	b := tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("f"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	typ := b.Node.(*DirectDeclarator).TopDeclarator().Type
	if typ == nil {
		t.Fatal("missing type")
	}

	if typ = typ.Result(); typ == nil {
		t.Fatal("missing result type")
	}

	if g, e := typ.String(), "int"; g != e {
		t.Fatalf("%q %q", g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/80
func TestIssue80(t *testing.T) {
	tu, err := Parse(
		"", []string{"testdata/issue80.c"}, newTestModel(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}

	b := tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("s"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	typ := b.Node.(*DirectDeclarator).TopDeclarator().Type
	if typ == nil {
		t.Fatal("missing type")
	}

	if g, e := typ.Kind(), Array; g != e {
		t.Errorf("Kind: %v %v", g, e)
	}

	if g, e := typ.Elements(), 7; g != e {
		t.Errorf("Elements: %v %v", g, e)
	}

	if g, e := typ.SizeOf(), 7; g != e {
		t.Fatalf("Sizeof: %v %v", g, e)
	}

	b = tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("t"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	typ = b.Node.(*DirectDeclarator).TopDeclarator().Type
	if typ == nil {
		t.Fatal("missing type")
	}

	if g, e := typ.Kind(), Ptr; g != e {
		t.Errorf("Kind: %v %v", g, e)
	}

	if g, e := typ.Elements(), -1; g != e {
		t.Errorf("Elements: %v %v", g, e)
	}

	if g, e := typ.SizeOf(), 8; g != e {
		t.Fatalf("Sizeof: %v %v", g, e)
	}

	b = tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("u"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	typ = b.Node.(*DirectDeclarator).TopDeclarator().Type
	if typ == nil {
		t.Fatal("missing type")
	}

	if g, e := typ.Kind(), Array; g != e {
		t.Errorf("Kind: %v %v", g, e)
	}

	if g, e := typ.Elements(), 11; g != e {
		t.Errorf("Elements: %v %v", g, e)
	}

	if g, e := typ.SizeOf(), 11; g != e {
		t.Fatalf("Sizeof: %v %v", g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/81
func TestIssue81(t *testing.T) {
	tu, err := Parse(
		"", []string{"testdata/issue81.c"}, newTestModel(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_ = tu
	for l := tu; l != nil; l = l.TranslationUnit {
		d := l.ExternalDeclaration.Declaration
		for l := d.InitDeclaratorListOpt.InitDeclaratorList; l != nil; l = l.InitDeclaratorList {
			x := l.InitDeclarator.Initializer.Expression
			s := xc.Dict.S(int(x.Value.(StringLitID)))
			if g, e := len(s), 3; g != e {
				t.Fatalf("%v |% x| \n%v %v", position(x.Pos()), s, g, e)
			}

			if g, e := s, []byte{0, 255, 0}; !bytes.Equal(g, e) {
				t.Fatalf("%v |% x| |% x|", position(x.Pos()), g, e)
			}
		}
	}
}

// https://gitlab.com/cznic/cc/issues/82
func TestIssue82(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/issue82.c"),
		`d(2)
d(2, 3)`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/84
func TestIssue84(t *testing.T) {
	if g, e := testPreprocessor(t, "testdata/issue84.c"),
		`c(1, 2, 3);
c(1, 2);
c(1, );`; g != e {
		t.Fatalf("\ngot\n%s\nexp\n%s", g, e)
	}
}

var vectorAttr = regexp.MustCompile(`__attribute__ *\(\((__)?vector_size(__)? *\(`)

func testDir(t *testing.T, dir string) {

	var re *regexp.Regexp
	if s := *oRe; s != "" {
		re = regexp.MustCompile(s)
	}

	dir = filepath.FromSlash(dir)
	t.Log(dir)
	m, err := filepath.Glob(filepath.Join(dir, "*.c"))
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(m)
	predefined, _, sysIncludePaths, err := HostConfig()
	if err != nil {
		t.Fatal(err)
	}

	blacklist := []string{
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/20011217-2.c", // (((((union { double __d; int __i[2]; }) {__d: __x}).__i[1]
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/20020320-1.c", // static T *p = x;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr37056.c",    // ? ({void *__s = (u.buf + off); __s;}) : ...
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr42196-1.c",  // __complex__ int c;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr42196-2.c",  // __complex__ int ci;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr42196-3.c",  // __complex__ int ci;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr54559.c",    // return x + y * (T) (__extension__ 1.0iF);
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr54713-2.c",  // #include: typedef int V __attribute__((vector_size (N * sizeof (int))));
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr54713-3.c",  // #include: typedef int V __attribute__((vector_size (N * sizeof (int))));
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/pr67143.c",    // __sync_add_and_fetch(&a, 536870912);

		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/20041124-1.c",    // struct s { _Complex unsigned short x; };
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/20041201-1.c",    // typedef struct { _Complex char a; _Complex char b; } Scc2;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/20071211-1.c",    // __asm__ volatile ("" : : : "memory");
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pr19449.c",       // int z = __builtin_choose_expr (!__builtin_constant_p (y), 3, 4);
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pr38151.c",       // _Complex int b;
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pr39228.c",       // if (testl (1.18973149535723176502e+4932L) < 1)
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pr56982.c",       // __asm__ volatile ("" : : : "memory");
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pr71626-2.c",     // #include: typedef __INTPTR_TYPE__ V __attribute__((__vector_size__(sizeof (__INTPTR_TYPE__))));
		"testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/pushpop_macro.c", // #pragma push_macro("_")
	}

	const attempt2prototypes = `
void exit();
void abort();
`

	var pass, gccFail int
	defer func() {
		t.Logf("pass %v, gccFail %v (sum %v), total test cases %v", pass, gccFail, pass+gccFail, len(m))
	}()
outer:
	for i, v := range m {
		if re != nil && !re.MatchString(v) {
			continue
		}

		for _, w := range blacklist {
			if strings.HasSuffix(filepath.ToSlash(v), w) {
				continue outer
			}
		}

		b, err := ioutil.ReadFile(v)
		if err != nil {
			t.Fatal(err)
		}

		if vectorAttr.Match(b) {
			continue
		}

		attempt := 1
	retry:
		func() {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("PANIC\n%s\n%v", debug.Stack(), e)
				}
			}()

			s := predefined
			if attempt == 2 {
				s += attempt2prototypes
			}
			err = testDev1(
				s,
				s,
				s,
				[]string{},
				"",
				v,
				[]Opt{
					ErrLimit(-1),
					SysIncludePaths(sysIncludePaths),
					EnableIncludeNext(),
					EnableDefineOmitCommaBeforeDDD(),
				},
				[]Opt{
					ErrLimit(-1),
					SysIncludePaths(sysIncludePaths),
					EnableIncludeNext(),
					EnableWideBitFieldTypes(),
					EnableEmptyDeclarations(),
					gccEmu(),
				},
			)
		}()

		if err != nil {
			//dbg("%T(%v)", err, err)
			switch err.(type) {
			case cppCmpError:
				// fail w/o retry.
			default:
				if attempt == 1 { // retry with {abort,exit} prototype.
					attempt++
					goto retry
				}

				s := errString(err)
				if !strings.Contains(s, "PANIC") && !strings.Contains(s, "TODO") && !strings.Contains(s, "undefined: __builtin_") {
					if out, err := exec.Command("gcc", "-o", os.DevNull, "-c", "-std=c99", "--pedantic", "-fmax-errors=10", v).CombinedOutput(); len(out) != 0 || err != nil {
						// Auto blacklist if gcc fails to compile as well.
						if n := 4000; len(out) > n {
							out = out[:n]
						}
						t.Logf("%s\n==== gcc reports\n%s\n%v", s, out, err)
						gccFail++
						continue
					}

				}
			}

			t.Errorf("%v\n%v/%v, %v ok(+%v=%v)\nFAIL\n%s (%T)", v, i+1, len(m), pass, gccFail, pass+gccFail, errString(err), err)
			return
		}

		pass++
		if re != nil {
			t.Logf("%v: %v ok", v, pass)
		}
	}
}

func TestTCCTests(t *testing.T) {
	if !*oDev {
		t.Log("enable with -dev")
		return
	}

	testDir(t, "testdata/tcc-0.9.26/tests/tests2/")
}

func TestGCCTests(t *testing.T) {
	if !*oDev {
		t.Log("enable with -dev")
		return
	}

	testDir(t, "testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compat/")
	testDir(t, "testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/compile/")
	testDir(t, "testdata/gcc-6.3.0/gcc/testsuite/gcc.c-torture/execute/")
}

// https://gitlab.com/cznic/cc/issues/85
func TestIssue85(t *testing.T) {
	tu, err := Parse(
		"", []string{"testdata/issue85.c"}, newTestModel(), EnableOmitFuncRetType(),
	)
	if err != nil {
		t.Fatal(err)
	}

	b := tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("i"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	d := b.Node.(*DirectDeclarator).TopDeclarator()
	if g, e := d.Linkage, External; g != e {
		t.Fatal(g, e)
	}

	if g, e := d.Type.Specifier().IsExtern(), false; g != e {
		t.Fatal(g, e)
	}

	b = tu.Declarations.Lookup(NSIdentifiers, xc.Dict.SID("j"))
	if b.Node == nil {
		t.Fatal("lookup fail")
	}

	d = b.Node.(*DirectDeclarator).TopDeclarator()
	if g, e := d.Linkage, External; g != e {
		t.Fatal(g, e)
	}

	if g, e := d.Type.Specifier().IsExtern(), true; g != e {
		t.Fatal(g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/86
func TestIssue86(t *testing.T) {
	_, err := Parse(
		"", []string{"testdata/issue86.c"}, newTestModel(), EnableOmitFuncRetType(),
	)
	if err == nil {
		t.Fatal("missed error")
	}

	if g, e := err.Error(), "testdata/issue86.c:2:12: 'j' initialized and declared 'extern'"; g != e {
		t.Fatalf("%q %q", g, e)
	}

	t.Log(err)
}

func TestArray(t *testing.T) {
	ast, err := Parse(
		"", []string{"testdata/array.c"}, newTestModel(), EnableOmitFuncRetType(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}

	expr := ast.TranslationUnit.ExternalDeclaration.FunctionDefinition.FunctionBody.
		CompoundStatement.BlockItemListOpt.BlockItemList.BlockItemList.BlockItem.
		Statement.ExpressionStatement.ExpressionListOpt.ExpressionList.Expression

	if g, e := expr.Type.Kind(), Ptr; g != e {
		t.Fatal(g, e)
	}

	dd := expr.IdentResolutionScope().Lookup(NSIdentifiers, dict.SID("a")).Node.(*DirectDeclarator)
	if g, e := dd.TopDeclarator().Type.Kind(), Array; g != e {
		t.Fatal(g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/87
func TestIssue87(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue87.c"}, newTestModel(),
	); err == nil {
		t.Fatal("missed error")
	}

	if _, err := Parse(
		"", []string{"testdata/issue87.c"}, newTestModel(), AllowCompatibleTypedefRedefinitions(),
	); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/88
func TestIssue88(t *testing.T) {
	ast, err := Parse(
		"", []string{"testdata/issue88.c"}, newTestModel(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}

	exp := ast.TranslationUnit.ExternalDeclaration.FunctionDefinition.FunctionBody.
		CompoundStatement.BlockItemListOpt.BlockItemList.BlockItemList.BlockItem.
		Statement.ExpressionStatement.ExpressionListOpt.ExpressionList.Expression

	if g := exp.BinOpType; g != nil {
		t.Fatalf("unexpected non-nil BinOpType %s", g)
	}
}

// https://gitlab.com/cznic/cc/issues/89
func TestIssue89(t *testing.T) {
	ast, err := Parse(
		"", []string{"testdata/issue89.c"}, newTestModel(), EnableImplicitFuncDef(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}

	exp := ast.TranslationUnit.ExternalDeclaration.FunctionDefinition.FunctionBody.
		CompoundStatement.BlockItemListOpt.BlockItemList.BlockItemList.BlockItemList.BlockItem.
		Statement.ExpressionStatement.ExpressionListOpt.ExpressionList.Expression.
		ArgumentExpressionListOpt.ArgumentExpressionList.Expression

	if g := exp.Type; g == nil {
		t.Errorf("'a.f': missing expression type")
	}
	if g := exp.Expression.Type; g == nil {
		t.Errorf("'a': missing expression type")
	}
}

// https://gitlab.com/cznic/cc/issues/90
func TestIssue90(t *testing.T) {
	ast, err := Parse(
		"", []string{"testdata/issue90.c"}, newTestModel(), EnableImplicitFuncDef(),
	)
	if err != nil {
		t.Fatal(errString(err))
	}

	expr := ast.TranslationUnit.ExternalDeclaration.FunctionDefinition.FunctionBody.
		CompoundStatement.BlockItemListOpt.BlockItemList.BlockItemList.BlockItem.
		Statement.ExpressionStatement.ExpressionListOpt.ExpressionList.Expression

	if g, e := expr.Type.Kind(), UInt; g != e {
		t.Errorf("expr: %v %v", g, e)
	}
	if g, e := expr.Expression.Type.Kind(), UInt; g != e {
		t.Errorf("expr.Expression: %v %v", g, e)
	}
	if g, e := expr.Expression2.Type.Kind(), UInt; g != e {
		t.Errorf("expr.Expression2: %v %v", g, e)
	}
}

// https://gitlab.com/cznic/cc/issues/92
func TestIssue92(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue92.c"}, newTestModel(),
	); err == nil {
		t.Fatal("missed error")
	}

	if _, err := Parse(
		"", []string{"testdata/issue92.c"}, newTestModel(), AllowCompatibleTypedefRedefinitions(),
	); err != nil {
		t.Fatal(err)
	}
}

// https://gitlab.com/cznic/cc/issues/93
func TestIssue93(t *testing.T) {
	if _, err := Parse(
		"", []string{"testdata/issue93.c"}, newTestModel(),
	); err != nil {
		t.Fatal(err)
	}
}

func ExampleCharTypes() {
	ast, err := Parse(
		"", []string{"testdata/chartypes.c"}, newTestModel(),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(ast.Declarations.Identifiers[dict.SID("c")].Node.(*DirectDeclarator).declarator.Type)
	fmt.Println(ast.Declarations.Identifiers[dict.SID("d")].Node.(*DirectDeclarator).declarator.Type)
	fmt.Println(ast.Declarations.Identifiers[dict.SID("e")].Node.(*DirectDeclarator).declarator.Type)
	// Output:
	// char
	// signed char
	// unsigned char
}
