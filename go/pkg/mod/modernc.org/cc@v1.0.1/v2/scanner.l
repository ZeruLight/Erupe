%{
// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on [0], 6.4.
//
// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
%}

%yyc c
%yyn c = l.Next()
%yym l.Mark()
%yyt l.sc

%x COMMENT

%{
package cc

import (
        "fmt"

        "modernc.org/golex/lex"
)

const (
        _ = iota
        scCOMMENT       // [`/*`, `*/`]
)

func (l *lexer) scan() (r int) {
        c := l.Enter()
%}

binary-constant                 {binary-prefix}{binary-digit}+
binary-digit                    [01]
binary-exponent-part            [pP]{sign}?{digit-sequence}
binary-prefix                   0[bB]
c-char                          [^'\n\x80\\]|{escape-sequence}
c-char-sequence                 {c-char}+
character-constant              '{c-char-sequence}'
comment-close                   ([^*\x80]|\*+[^*/\x80])*\*+\/
decimal-constant                {nonzero-digit}{digit}*
digit                           [0-9]
digit-sequence                  {digit}+
eof                             \x80
escape-sequence                 {simple-sequence}|{octal-escape-sequence}|{hexadecimal-escape-sequence}|{universal-character-name}
exponent-part                   [eE]{sign}?{digit-sequence}
fractional-constant             {digit-sequence}?\.{digit-sequence}|{digit-sequence}\.
h-char                          [^>\n\x80]
h-char-sequence                 {h-char}+
hex-quad                        {hexadecimal-digit}{hexadecimal-digit}{hexadecimal-digit}{hexadecimal-digit}
hexadecimal-constant            {hexadecimal-prefix}{hexadecimal-digit}+
hexadecimal-digit               [0-9a-fA-F]
hexadecimal-digit-sequence      {hexadecimal-digit}+
hexadecimal-escape-sequence     \\x{hexadecimal-digit}+
hexadecimal-prefix              0[xX]
identifier                      {identifier-nondigit}({identifier-nondigit}|{digit}|{ucn-digit}|"$")*
identifier-nondigit             {nondigit}|{universal-character-name}|{ucn-nondigit}
nondigit                        [_a-zA-Z]
nonzero-digit                   [1-9]
octal-constant                  0{octal-digit}*
octal-digit                     [0-7]
octal-escape-sequence           \\{octal-digit}{octal-digit}?{octal-digit}?
pp-number                       ({digit}|\.{digit})({digit}|{identifier-nondigit}|[eEpP]{sign}|\.)*
q-char                          [^\n\x22\x80]
q-char-sequence                 {q-char}+
s-char                          [^\x22\n\x80\\]|{escape-sequence}
s-char-sequence                 {s-char}+
sign                            [-+]
simple-sequence                 \\['\x22?\\abfnrtv]
string-literal                  \x22{s-char-sequence}?\x22
ucn-digit                       \x83
ucn-nondigit                    \x84
universal-character-name        \\u{hex-quad}|\\U{hex-quad}{hex-quad}
unsigned-suffix                 [uU]

%%
                                c = l.Rule0()

[ \t\f\v]+			return ' '

"//"[^\x80\n]*			l.comment(false)
				return ' '

"/*"				l.commentPos0 = l.First.Pos()
                                l.sc = scCOMMENT

<COMMENT>{comment-close}        l.sc = scINITIAL
				l.First = lex.NewChar(l.commentPos0, l.First.Rune)
				l.comment(true)
				return ' '

<COMMENT>{eof}                  l.errPos(l.commentPos0, "unterminated comment")
                                l.sc = scINITIAL
                                return rune2class(lex.RuneEOF)

<*>{eof}                        return rune2class(lex.RuneEOF)

"!="                            return NEQ
"%:"                            return '#'
"%="                            return MODASSIGN
"%>"                            return '}'
"&&"                            return ANDAND
"&="                            return ANDASSIGN
"*="                            return MULASSIGN
"++"                            return INC
"+="                            return ADDASSIGN
"--"                            return DEC
"-="                            return SUBASSIGN
"->"                            return ARROW
"..."                           return DDD
"/="                            return DIVASSIGN
":>"                            return ']'
"<%"                            return '{'
"<:"                            return '['
"<<"                            return LSH
"<<="                           return LSHASSIGN
"<="                            return LEQ
"=="                            return EQ
">="                            return GEQ
">>"                            return RSH
">>="                           return RSHASSIGN
"^="                            return XORASSIGN
"|="                            return ORASSIGN
"||"                            return OROR

"##"                            |
"#%:"                           |
"%:#"                           |
"%:%:"                          return PPPASTE

L{character-constant}           return LONGCHARCONST
L{string-literal}               return LONGSTRINGLITERAL
{character-constant}            return CHARCONST
{identifier}                    return IDENTIFIER
{pp-number}                     return PPNUMBER
{string-literal}                return STRINGLITERAL

%%
        if c, ok := l.Abort(); ok {
                return c
        }
        
        goto yyAction
}
