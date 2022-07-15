package pascalstring

import (
  "erupe-ce/common/byteframe"
)

func Uint8(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func Uint16(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func Uint32(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint32(uint32(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}
