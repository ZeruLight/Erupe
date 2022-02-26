%{
// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on [0], 6.4.

// Implements translation phases 1 and 2 of [0], 5.1.1.2.
//
// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
%}

%yyc c
%yyn c = t.Next()
%yym t.Mark()
%yyt t.sc

%s TRIGRAPHS

%{
package cc

import (
        "fmt"
)

const (
        _ = iota
        scTRIGRAPHS
)

func (t *trigraphs) scan() (r int) {
        c := t.Enter()
%}

%%
        c = t.Rule0()

<TRIGRAPHS>"??!"        return '|'
<TRIGRAPHS>"??'"        return '^'
<TRIGRAPHS>"??("        return '['
<TRIGRAPHS>"??)"        return ']'
<TRIGRAPHS>"??-"        return '~'
<TRIGRAPHS>"??/"        return '\\'
<TRIGRAPHS>"??<"        return '{'
<TRIGRAPHS>"??="        return '#'
<TRIGRAPHS>"??>"        return '}'

\\\r?\n|\r |
<TRIGRAPHS>"??/"\r?\n

%%
        if c, ok := t.Abort(); ok {
                return c
        }
        
        goto yyAction
}
