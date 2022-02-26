/*******************************************************************************
The MIT License (MIT)

Copyright (c) 2013-2019 Hajime Nakagami

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
	"database/sql/driver"

	"context"
)

type firebirdsqlStmt struct {
	wp         *wireProtocol
	stmtHandle int32
	tx         *firebirdsqlTx
	xsqlda     []xSQLVAR
	blr        []byte
	stmtType   int32
}

func (stmt *firebirdsqlStmt) Close() (err error) {
	err = stmt.wp.opFreeStatement(stmt.stmtHandle, 2) // DSQL_drop
	if err != nil {
		return err
	}

	if stmt.wp.acceptType == ptype_lazy_send {
		stmt.wp.lazyResponseCount++
	} else {
		_, _, _, err = stmt.wp.opResponse()
	}

	if stmt.tx.isAutocommit {
		stmt.tx.Commit()
	}
	return
}

func (stmt *firebirdsqlStmt) NumInput() int {
	return -1
}

func (stmt *firebirdsqlStmt) exec(ctx context.Context, args []driver.Value) (result driver.Result, err error) {
	err = stmt.wp.opExecute(stmt.stmtHandle, stmt.tx.transHandle, args)
	if err != nil {
		return
	}

	_, _, _, err = stmt.wp.opResponse()
	if err != nil {
		return
	}

	err = stmt.wp.opInfoSql(stmt.stmtHandle, []byte{isc_info_sql_records})
	if err != nil {
		return
	}

	_, _, buf, err := stmt.wp.opResponse()
	if err != nil {
		return
	}

	var rowcount int64
	if len(buf) >= 32 {
		if stmt.stmtType == isc_info_sql_stmt_select {
			rowcount = int64(bytes_to_int32(buf[20:24]))
		} else {
			rowcount = int64(bytes_to_int32(buf[27:31]) + bytes_to_int32(buf[6:10]) + bytes_to_int32(buf[13:17]))
		}
	} else {
		rowcount = 0
	}

	result = &firebirdsqlResult{
		affectedRows: rowcount,
	}
	return
}

func (stmt *firebirdsqlStmt) Exec(args []driver.Value) (result driver.Result, err error) {
	return stmt.exec(context.Background(), args)
}

func (stmt *firebirdsqlStmt) query(ctx context.Context, args []driver.Value) (driver.Rows, error) {
	var rows driver.Rows
	var err error
	var result []driver.Value

	if stmt.stmtType == isc_info_sql_stmt_exec_procedure {
		err = stmt.wp.opExecute2(stmt.stmtHandle, stmt.tx.transHandle, args, stmt.blr)
		if err != nil {
			return nil, err
		}

		result, err = stmt.wp.opSqlResponse(stmt.xsqlda)
		if err != nil {
			return nil, err
		}

		rows = newFirebirdsqlRows(stmt, result)

		_, _, _, err = stmt.wp.opResponse()
		if err != nil {
			return nil, err
		}
	} else {
		err := stmt.wp.opExecute(stmt.stmtHandle, stmt.tx.transHandle, args)
		if err != nil {
			return nil, err
		}

		_, _, _, err = stmt.wp.opResponse()
		if err != nil {
			return nil, err
		}

		rows = newFirebirdsqlRows(stmt, nil)
	}
	return rows, err
}

func (stmt *firebirdsqlStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	return stmt.query(context.Background(), args)
}

func newFirebirdsqlStmt(fc *firebirdsqlConn, query string) (stmt *firebirdsqlStmt, err error) {
	stmt = new(firebirdsqlStmt)
	stmt.wp = fc.wp
	stmt.tx = fc.tx

	err = fc.wp.opAllocateStatement()
	if err != nil {
		return nil, err
	}

	if fc.wp.acceptType == ptype_lazy_send {
		fc.wp.lazyResponseCount++
		stmt.stmtHandle = -1
	} else {
		stmt.stmtHandle, _, _, err = fc.wp.opResponse()
		if err != nil {
			return
		}
	}

	err = fc.wp.opPrepareStatement(stmt.stmtHandle, stmt.tx.transHandle, query)
	if err != nil {
		return nil, err
	}

	if fc.wp.acceptType == ptype_lazy_send && fc.wp.lazyResponseCount > 0 {
		fc.wp.lazyResponseCount--
		stmt.stmtHandle, _, _, _ = fc.wp.opResponse()
	}

	_, _, buf, err := fc.wp.opResponse()
	if err != nil {
		return
	}

	stmt.stmtType, stmt.xsqlda, err = fc.wp.parse_xsqlda(buf, stmt.stmtHandle)
	if err != nil {
		return nil, err
	}

	stmt.blr = calcBlr(stmt.xsqlda)

	return
}
