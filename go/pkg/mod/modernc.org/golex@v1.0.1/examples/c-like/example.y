%{
package main

import (
  "bytes"
  "fmt"
  "os"
  "io"
)

type node struct {
  name string
  children []node
}

func (n node) String() string {
  buf := new(bytes.Buffer)
  n.print(buf, " ")
  return buf.String()
}

func (n node) print(out io.Writer, indent string) {
  fmt.Fprintf(out, "\n%v%v", indent, n.name)
  for _, nn := range n.children { nn.print(out, indent + "  ") }
}

func Node(name string) node { return node{name: name} }
func (n node) append(nn...node) node { n.children = append(n.children, nn...); return n }

%}

%union{
    node node
    token string
}

%token FUNC INT IDENT OP COMMENT

%type <token> FUNC INT IDENT OP COMMENT
%type <node> Input Func Args Statements Expr Call ExprList Statement

%%

Input: /* empty */ { }
     | Input Func { fmt.Println($2) }

Func: FUNC IDENT '(' Args ')' '{' Statements '}' { $$ = Node("func").append(Node("name").append(Node($2))).append($4, $7) }

Args: /* empty */     { $$ = Node("args") }
    | Args ',' IDENT  { $$ = $1.append(Node($3)) }

Statements: /* empty */     { $$ = Node("statements") }
          | Statements Statement { $$ = $1.append($2) }

Statement: IDENT '=' Expr { $$ = Node("assign").append(Node($1), $3) }
         | COMMENT { $$ = Node($1) }

Expr: INT   { $$ = Node($1) }
    | Call
    | Expr OP INT { $$ = Node($2).append($1, Node($3)) }
    | Expr OP Call { $$ = Node($2).append($1, $3) }

Call: IDENT '(' ')'          { $$ = Node("call").append(Node("name").append(Node($1))) }
    | IDENT '(' ExprList ')' { $$ = Node("call").append(Node("name").append(Node($1))).append($3) }

ExprList: Expr               { $$ = Node("expressions").append($1) }
        | ExprList ',' Expr  { $$ = $1.append($3) }

%%

const src = `

func A() {  // Just an example
  a = Привет(42, pi()) / 2
}

`

func main() {
  yyDebug        = 0
  yyErrorVerbose = true
  l := newLexer(bytes.NewBufferString(src), os.Stdout, "file.name")
  yyParse(l)
}
