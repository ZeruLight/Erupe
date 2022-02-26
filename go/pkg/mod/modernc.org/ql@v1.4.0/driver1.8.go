// +build go1.8

package ql // import "modernc.org/ql"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const prefix = "$"

var (
	_ driver.ExecerContext      = (*driverConn)(nil)
	_ driver.QueryerContext     = (*driverConn)(nil)
	_ driver.ConnBeginTx        = (*driverConn)(nil)
	_ driver.ConnPrepareContext = (*driverConn)(nil)
)

// BeginTx implements driver.ConnBeginTx.
func (c *driverConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	// Check the transaction level. If the transaction level is non-default
	// then return an error here as the BeginTx driver value is not supported.
	if sql.IsolationLevel(opts.Isolation) != sql.LevelDefault {
		return nil, errors.New("ql: driver does not support non-default isolation level")
	}

	// If a read-only transaction is requested return an error as the
	// BeginTx driver value is not supported.
	if opts.ReadOnly {
		return nil, errors.New("ql: driver does not support read-only transactions")
	}

	if c.ctx == nil {
		c.ctx = NewRWCtx()
	}

	if _, _, err := c.db.db.Execute(c.ctx, txBegin); err != nil {
		return nil, err
	}

	c.tnl++
	return c, nil
}

func (c *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	query, vals, err := replaceNamed(query, args)
	if err != nil {
		return nil, err
	}

	return c.Exec(query, vals)
}

func replaceNamed(query string, args []driver.NamedValue) (string, []driver.Value, error) {
	toks, err := tokenize(query)
	if err != nil {
		return "", nil, err
	}

	a := make([]driver.Value, len(args))
	m := map[string]int{}
	for _, v := range args {
		m[v.Name] = v.Ordinal
		a[v.Ordinal-1] = v.Value
	}
	for i, v := range toks {
		if len(v) > 1 && strings.HasPrefix(v, prefix) {
			if v[1] >= '1' && v[1] <= '9' {
				continue
			}

			nm := v[1:]
			k, ok := m[nm]
			if !ok {
				return query, nil, fmt.Errorf("unknown named parameter %s", nm)
			}

			toks[i] = fmt.Sprintf("$%d", k)
		}
	}
	return strings.Join(toks, " "), a, nil
}

func (c *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	query, vals, err := replaceNamed(query, args)
	if err != nil {
		return nil, err
	}

	return c.Query(query, vals)
}

func (c *driverConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	query, err := filterNamedArgs(query)
	if err != nil {
		return nil, err
	}

	return c.Prepare(query)
}

func filterNamedArgs(query string) (string, error) {
	toks, err := tokenize(query)
	if err != nil {
		return "", err
	}

	n := 0
	for _, v := range toks {
		if len(v) > 1 && strings.HasPrefix(v, prefix) && v[1] >= '1' && v[1] <= '9' {
			m, err := strconv.ParseUint(v[1:], 10, 31)
			if err != nil {
				return "", err
			}

			if int(m) > n {
				n = int(m)
			}
		}
	}
	for i, v := range toks {
		if len(v) > 1 && strings.HasPrefix(v, prefix) {
			if v[1] >= '1' && v[1] <= '9' {
				continue
			}

			n++
			toks[i] = fmt.Sprintf("$%d", n)
		}
	}
	return strings.Join(toks, " "), nil
}

func (s *driverStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	a := make([]driver.Value, len(args))
	for k, v := range args {
		a[k] = v.Value
	}
	return s.Exec(a)
}

func (s *driverStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	a := make([]driver.Value, len(args))
	for k, v := range args {
		a[k] = v.Value
	}
	return s.Query(a)
}

func tokenize(s string) (r []string, _ error) {
	lx, err := newLexer(s)
	if err != nil {
		return nil, err
	}

	var lval yySymType
	for lx.Lex(&lval) != 0 {
		s := string(lx.TokenBytes(nil))
		if s != "" {
			switch s[len(s)-1] {
			case '"':
				s = "\"" + s
			case '`':
				s = "`" + s
			}
		}
		r = append(r, s)
	}
	return r, nil
}
