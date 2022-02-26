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
	"bytes"
	"container/list"
	"database/sql/driver"
	"io"
	"reflect"
	"strings"
)

type firebirdsqlRows struct {
	stmt            *firebirdsqlStmt
	currentChunkRow *list.Element
	moreData        bool
	result          []driver.Value
}

func newFirebirdsqlRows(stmt *firebirdsqlStmt, result []driver.Value) *firebirdsqlRows {
	rows := new(firebirdsqlRows)
	rows.stmt = stmt
	rows.result = result
	if stmt.stmtType == isc_info_sql_stmt_select {
		rows.moreData = true
	}
	return rows
}

func (rows *firebirdsqlRows) Columns() []string {
	columns := make([]string, len(rows.stmt.xsqlda))
	for i, x := range rows.stmt.xsqlda {
		columns[i] = x.aliasname
		if rows.stmt.tx.fc.columnNameToLower {
			columns[i] = strings.ToLower(columns[i])
		}
	}
	return columns
}

func (rows *firebirdsqlRows) Close() (er error) {
	rows.stmt.Close()

	return
}

func (rows *firebirdsqlRows) Next(dest []driver.Value) (err error) {
	if rows.stmt.stmtType == isc_info_sql_stmt_exec_procedure {
		if rows.result != nil {
			for i, v := range rows.result {
				dest[i] = v
			}
			rows.result = nil
		} else {
			err = io.EOF
		}

		return
	}

	if rows.currentChunkRow != nil {
		rows.currentChunkRow = rows.currentChunkRow.Next()
	}

	if rows.currentChunkRow == nil && rows.moreData == true {
		// Get one chunk
		var chunk *list.List
		err = rows.stmt.wp.opFetch(rows.stmt.stmtHandle, rows.stmt.blr)
		if err != nil {
			return err
		}
		chunk, rows.moreData, err = rows.stmt.wp.opFetchResponse(rows.stmt.stmtHandle, rows.stmt.tx.transHandle, rows.stmt.xsqlda)

		if err == nil {
			rows.currentChunkRow = chunk.Front()
		} else {
			return
		}
	}

	if rows.currentChunkRow == nil {
		err = io.EOF
		return
	}
	row, _ := rows.currentChunkRow.Value.([]driver.Value)
	for i, v := range row {
		if rows.stmt.xsqlda[i].sqltype == SQL_TYPE_BLOB && v != nil {
			blobId := v.([]byte)
			var blob []byte
			blob, err = rows.stmt.wp.getBlobSegments(blobId, rows.stmt.tx.transHandle)
			if rows.stmt.xsqlda[i].sqlsubtype == 1 {
				dest[i] = bytes.NewBuffer(blob).String()
			} else {
				dest[i] = blob
			}

		} else {
			dest[i] = v
		}
	}

	return
}

func (rows *firebirdsqlRows) ColumnTypeDatabaseTypeName(index int) string {
	return rows.stmt.xsqlda[index].typename()
}

func (rows *firebirdsqlRows) ColumnTypeLength(index int) (length int64, ok bool) {
	return int64(rows.stmt.xsqlda[index].displayLength()), true
}

func (rows *firebirdsqlRows) ColumnTypeNullable(index int) (nullable bool, ok bool) {
	return rows.stmt.xsqlda[index].null_ok, true
}

func (rows *firebirdsqlRows) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return int64(rows.stmt.xsqlda[index].displayLength()), int64(rows.stmt.xsqlda[index].sqlscale), rows.stmt.xsqlda[index].hasPrecisionScale()
}

func (rows *firebirdsqlRows) ColumnTypeScanType(index int) reflect.Type {
	return rows.stmt.xsqlda[index].scantype()
}
