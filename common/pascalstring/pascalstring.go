package pascalstring

import (
	"erupe-ce/common/byteframe"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func Uint8(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			panic(err)
		}
		x = xt
	}
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func Uint16(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			panic(err)
		}
		x = xt
	}
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func Uint32(bf *byteframe.ByteFrame, x string, t bool) {
	if t {
		e := japanese.ShiftJIS.NewEncoder()
		xt, _, err := transform.String(e, x)
		if err != nil {
			panic(err)
		}
		x = xt
	}
	bf.WriteUint32(uint32(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}
