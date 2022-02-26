// Copyright 2014 The ql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ql // import "modernc.org/ql"

import (
	"fmt"

	"errors"
)

var (
	errBeginTransNoCtx          = errors.New("BEGIN TRANSACTION: Must use R/W context, have nil")
	errCommitNotInTransaction   = errors.New("COMMIT: Not in transaction")
	errDivByZero                = errors.New("division by zero")
	errIncompatibleDBFormat     = errors.New("incompatible DB format")
	errNoDataForHandle          = errors.New("read: no data for handle")
	errRollbackNotInTransaction = errors.New("ROLLBACK: Not in transaction")
)

type errDuplicateUniqueIndex []interface{}

func (err errDuplicateUniqueIndex) Error() string {
	return fmt.Sprintf("cannot insert into unique index: duplicate value(s): %v", []interface{}(err))
}

// IsDuplicateUniqueIndexError reports whether err is produced by attempting to
// violate unique index constraints.
func IsDuplicateUniqueIndexError(err error) bool {
	_, ok := err.(errDuplicateUniqueIndex)
	return ok
}
