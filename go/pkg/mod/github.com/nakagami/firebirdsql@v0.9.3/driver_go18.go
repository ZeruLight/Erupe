// +build go1.8

/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2016-2019 Hajime Nakagami

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*******************************************************************************/

package firebirdsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sort"
)

func (stmt *firebirdsqlStmt) ExecContext(ctx context.Context, namedargs []driver.NamedValue) (result driver.Result, err error) {
	sort.SliceStable(namedargs, func(i, j int) bool {
		return namedargs[i].Ordinal < namedargs[j].Ordinal
	})
	args := make([]driver.Value, len(namedargs))
	for i, nv := range namedargs {
		args[i] = nv.Value
	}

	return stmt.exec(ctx, args)
}

func (stmt *firebirdsqlStmt) QueryContext(ctx context.Context, namedargs []driver.NamedValue) (rows driver.Rows, err error) {
	args := make([]driver.Value, len(namedargs))
	for i, nv := range namedargs {
		args[i] = nv.Value
	}

	return stmt.query(ctx, args)
}

func (fc *firebirdsqlConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if opts.ReadOnly {
		return fc.begin(ISOLATION_LEVEL_READ_COMMITED_RO)
	}

	switch (sql.IsolationLevel)(opts.Isolation) {
	case sql.LevelDefault:
		return fc.begin(ISOLATION_LEVEL_READ_COMMITED)
	case sql.LevelReadCommitted:
		return fc.begin(ISOLATION_LEVEL_READ_COMMITED)
	case sql.LevelRepeatableRead:
		return fc.begin(ISOLATION_LEVEL_REPEATABLE_READ)
	case sql.LevelSerializable:
		return fc.begin(ISOLATION_LEVEL_SERIALIZABLE)
	default:
	}
	return nil, errors.New("This isolation level is not supported.")
}

func (fc *firebirdsqlConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return fc.prepare(ctx, query)
}

func (fc *firebirdsqlConn) ExecContext(ctx context.Context, query string, namedargs []driver.NamedValue) (result driver.Result, err error) {
	args := make([]driver.Value, len(namedargs))
	for i, nv := range namedargs {
		args[i] = nv.Value
	}
	return fc.exec(ctx, query, args)
}

// This file implements the optional pinger interface for the database/sql package
func (fc *firebirdsqlConn) Ping(ctx context.Context) (err error) {
	if fc == nil {
		return errors.New("Connection was closed")
	}

	rows, err := fc.query(ctx, "SELECT 1 from rdb$database", nil)
	if err != nil {
		return driver.ErrBadConn
	}
	rows.Close()

	return nil
}

func (fc *firebirdsqlConn) QueryContext(ctx context.Context, query string, namedargs []driver.NamedValue) (rows driver.Rows, err error) {
	args := make([]driver.Value, len(namedargs))
	for i, nv := range namedargs {
		args[i] = nv.Value
	}
	return fc.query(ctx, query, args)
}

// ================== Implementation of the Connector interface ====================

type firebirdConnector struct {
	dsn *firebirdDsn
}

func (d *firebirdConnector) OpenConnector(dsns string) (driver.Connector, error) {
	dsn, err := parseDSN(dsns)
	if err != nil {
		return nil, err
	}
	return &firebirdConnector{dsn: dsn}, nil
}

func (fc *firebirdConnector) Driver() driver.Driver {
	return &firebirdsqlDriver{}
}

func (fc *firebirdConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return newFirebirdsqlConn(fc.dsn)
}
