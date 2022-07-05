package clientctx

import "erupe-ce/common/stringsupport"

// ClientContext holds contextual data required for packet encoding/decoding.
type ClientContext struct {
	StrConv *stringsupport.StringConverter
}
