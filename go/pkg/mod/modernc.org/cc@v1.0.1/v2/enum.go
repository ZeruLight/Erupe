// Copyright 2017 The CC Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cc // import "modernc.org/cc/v2"

import (
	"modernc.org/ir"
)

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

type cond int

const (
	condZero cond = iota

	condIfOff
	condIfOn
	condIfSkip

	maxCond
)

var (
	condOn = [maxCond]bool{
		condIfOn: true,
		condZero: true,
	}
)

// Linkage describes linkage of identifiers, [0]6.2.2.
type Linkage int

// Values of Linkage
const (
	LinkageNone     Linkage = iota
	LinkageInternal Linkage = Linkage(ir.InternalLinkage)
	LinkageExternal Linkage = Linkage(ir.ExternalLinkage)
)

// StorageDuration describes lifetime of an object, [0]6.2.4.
type StorageDuration int

// Values of StorageDuration
const (
	StorageDurationAutomatic StorageDuration = iota
	StorageDurationStatic
)
