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

type firebirdsqlTx struct {
	fc             *firebirdsqlConn
	isolationLevel int
	isAutocommit   bool
	transHandle    int32
	needBegin      bool
}

func (tx *firebirdsqlTx) begin() (err error) {
	var tpb []byte
	switch tx.isolationLevel {
	case ISOLATION_LEVEL_READ_COMMITED_LEGACY:
		tpb = []byte{
			byte(isc_tpb_version3),
			byte(isc_tpb_write),
			byte(isc_tpb_wait),
			byte(isc_tpb_read_committed),
			byte(isc_tpb_no_rec_version),
		}
	case ISOLATION_LEVEL_READ_COMMITED:
		tpb = []byte{
			byte(isc_tpb_version3),
			byte(isc_tpb_write),
			byte(isc_tpb_wait),
			byte(isc_tpb_read_committed),
			byte(isc_tpb_rec_version),
		}
	case ISOLATION_LEVEL_REPEATABLE_READ:
		tpb = []byte{
			byte(isc_tpb_version3),
			byte(isc_tpb_write),
			byte(isc_tpb_wait),
			byte(isc_tpb_concurrency),
		}
	case ISOLATION_LEVEL_SERIALIZABLE:
		tpb = []byte{
			byte(isc_tpb_version3),
			byte(isc_tpb_write),
			byte(isc_tpb_wait),
			byte(isc_tpb_consistency),
		}
	case ISOLATION_LEVEL_READ_COMMITED_RO:
		tpb = []byte{
			byte(isc_tpb_version3),
			byte(isc_tpb_read),
			byte(isc_tpb_wait),
			byte(isc_tpb_read_committed),
			byte(isc_tpb_rec_version),
		}
	}
	err = tx.fc.wp.opTransaction(tpb)
	if err != nil {
		return
	}
	tx.transHandle, _, _, err = tx.fc.wp.opResponse()
	tx.needBegin = false
	tx.fc.transactionSet[tx] = struct{}{}
	return
}

func (tx *firebirdsqlTx) commitRetainging() (err error) {
	err = tx.fc.wp.opCommitRetaining(tx.transHandle)
	if err != nil {
		return
	}
	_, _, _, err = tx.fc.wp.opResponse()
	tx.isAutocommit = tx.fc.isAutocommit
	return
}

func (tx *firebirdsqlTx) Commit() (err error) {
	err = tx.fc.wp.opCommit(tx.transHandle)
	if err != nil {
		return err
	}
	_, _, _, err = tx.fc.wp.opResponse()
	tx.isAutocommit = tx.fc.isAutocommit
	tx.needBegin = true
	return
}

func (tx *firebirdsqlTx) Rollback() (err error) {
	err = tx.fc.wp.opRollback(tx.transHandle)
	if err != nil {
		return nil
	}
	_, _, _, err = tx.fc.wp.opResponse()
	tx.isAutocommit = tx.fc.isAutocommit
	tx.needBegin = true
	return
}

func newFirebirdsqlTx(fc *firebirdsqlConn, isolationLevel int, isAutocommit bool, withBegin bool) (tx *firebirdsqlTx, err error) {
	tx = new(firebirdsqlTx)
	tx.fc = fc
	tx.isolationLevel = isolationLevel
	tx.isAutocommit = isAutocommit
	tx.needBegin = false

	if withBegin {
		err = tx.begin()

		if err != nil {
			return nil, err
		}

	} else {
		tx.needBegin = true
	}

	return
}
