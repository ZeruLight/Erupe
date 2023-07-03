package decryption

/*
	This code is HEAVILY based from
	https://github.com/Chakratos/ReFrontier/blob/master/ReFrontier/Unpack.cs
*/

import (
	"erupe-ce/common/byteframe"
	"io"
)

var m_shiftIndex int = 0
var m_flag byte = byte(0)

func UnpackSimple(data []byte) []byte {
	m_shiftIndex = 0
	m_flag = byte(0)

	bf := byteframe.NewByteFrameFromBytes(data)
	bf.SetLE()
	header := bf.ReadUint32()

	println("Decrypting")

	if header == 0x1A524B4A {
		bf.Seek(0x2, io.SeekCurrent)
		jpkType := bf.ReadUint16()
		println("JPK Type: ", jpkType)

		switch jpkType {
		case 3:
			startOffset := bf.ReadInt32()
			outSize := bf.ReadInt32()
			outBuffer := make([]byte, outSize)
			bf.Seek(int64(startOffset), io.SeekStart)
			ProcessDecode(bf, outBuffer)

			return outBuffer
		}
	}

	println("Skipping")

	return data
}

func ProcessDecode(data *byteframe.ByteFrame, outBuffer []byte) {
	outIndex := 0

	for int(data.Index()) < len(data.Data()) && outIndex < len(outBuffer)-1 {
		if JPKBitshift(data) == 0 {
			outBuffer[outIndex] = ReadByte(data)
			outIndex++
			continue
		} else {
			if JPKBitshift(data) == 0 {
				len := (JPKBitshift(data) << 1) | JPKBitshift(data)
				off := ReadByte(data)
				JPKCopy(outBuffer, int(off), int(len)+3, &outIndex)
				continue
			} else {
				hi := ReadByte(data)
				lo := ReadByte(data)
				var len int = int((hi & 0xE0)) >> 5
				var off int = ((int(hi) & 0x1F) << 8) | int(lo)
				if len != 0 {
					JPKCopy(outBuffer, off, len+2, &outIndex)
					continue
				} else {
					if JPKBitshift(data) == 0 {
						len := (JPKBitshift(data) << 3) | (JPKBitshift(data) << 2) | (JPKBitshift(data) << 1) | JPKBitshift(data)
						JPKCopy(outBuffer, off, int(len)+2+8, &outIndex)
						continue
					} else {
						temp := ReadByte(data)
						if temp == 0xFF {
							for i := 0; i < off+0x1B; i++ {
								outBuffer[outIndex] = ReadByte(data)
								outIndex++
								continue
							}
						} else {
							JPKCopy(outBuffer, off, int(temp)+0x1a, &outIndex)
						}
					}
				}
			}
		}
	}
}

func JPKBitshift(data *byteframe.ByteFrame) byte {
	m_shiftIndex--

	if m_shiftIndex < 0 {
		m_shiftIndex = 7
		m_flag = ReadByte(data)
	}

	return (byte)((m_flag >> m_shiftIndex) & 1)
}

func JPKCopy(outBuffer []byte, offset int, length int, index *int) {
	for i := 0; i < length; i++ {
		outBuffer[*index] = outBuffer[*index-offset-1]
		*index++
	}
}

func ReadByte(bf *byteframe.ByteFrame) byte {
	value := bf.ReadUint8()
	if value < 0 {
		println("Not implemented")
	}
	return byte(value)
}
