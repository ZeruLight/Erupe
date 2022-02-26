package byteframe

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
)

func TestGetData(t *testing.T) {
	want := []byte{0x01, 0x02, 0x03, 0x04}
	bf := NewByteFrameFromBytes(want)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestGetDataFromCurrent(t *testing.T) {
	have := []byte{0x01, 0x02, 0x03, 0x04}
	want := []byte{0x03, 0x04}
	bf := NewByteFrameFromBytes(have)

	bf.index += 2

	got := bf.DataFromCurrent()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint8(t *testing.T) {
	want := []byte{0xFF}
	bf := NewByteFrame()

	bf.WriteUint8(0xFF)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteBool(t *testing.T) {
	want := []byte{1}
	bf := NewByteFrame()

	bf.WriteBool(true)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint16LE(t *testing.T) {
	want := []byte{0xBB, 0xAA}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteUint16(0xAABB)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint16BE(t *testing.T) {
	want := []byte{0xAA, 0xBB}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteUint16(0xAABB)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint32LE(t *testing.T) {
	want := []byte{0xDD, 0xCC, 0xBB, 0xAA}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteUint32(0xAABBCCDD)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint32BE(t *testing.T) {
	want := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteUint32(0xAABBCCDD)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint64LE(t *testing.T) {
	want := []byte{0x33, 0x22, 0x11, 0x00, 0xDD, 0xCC, 0xBB, 0xAA}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteUint64(0xAABBCCDD00112233)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteUint64BE(t *testing.T) {
	want := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0x00, 0x11, 0x22, 0x33}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteUint64(0xAABBCCDD00112233)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt8(t *testing.T) {
	want := []byte{0xFF}
	bf := NewByteFrame()

	bf.WriteInt8(-1)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt16LE(t *testing.T) {
	want := []byte{0xFE, 0xFF}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteInt16(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt16BE(t *testing.T) {
	want := []byte{0xFF, 0xFE}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteInt16(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt32LE(t *testing.T) {
	want := []byte{0xFE, 0xFF, 0xFF, 0xFF}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteInt32(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt32BE(t *testing.T) {
	want := []byte{0xFF, 0xFF, 0xFF, 0xFE}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteInt32(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt64LE(t *testing.T) {
	want := []byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteInt64(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteInt64BE(t *testing.T) {
	want := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteInt64(-2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteFloat32LE(t *testing.T) {
	want := []byte{0x01, 0x00, 0x70, 0x42}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteFloat32(60.0000038)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteFloat32BE(t *testing.T) {
	want := []byte{0x42, 0x70, 0x00, 0x01}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteFloat32(60.0000038)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteFloat64LE(t *testing.T) {
	want := []byte{0x18, 0x70, 0xE0, 0x1F, 0x00, 0x00, 0x4E, 0x40}
	bf := NewByteFrame()
	bf.SetLE()

	bf.WriteFloat64(60.0000038)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteFloat64BE(t *testing.T) {
	want := []byte{0x40, 0x4E, 0x00, 0x00, 0x1F, 0xE0, 0x70, 0x18}
	bf := NewByteFrame()
	bf.SetBE()

	bf.WriteFloat64(60.0000038)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteBytes(t *testing.T) {
	want := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}
	bf := NewByteFrame()

	bf.WriteBytes([]byte{0x00})
	bf.WriteBytes([]byte{0x11})
	bf.WriteBytes([]byte{0x22, 0x33, 0x44})
	bf.WriteBytes([]byte{0x55, 0x66})
	bf.WriteBytes([]byte{0x77})

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestWriteNullTerminatedBytes(t *testing.T) {
	want := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21, 0x00, 0x00}
	have0 := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
	have1 := []byte{0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21}
	have2 := []byte{}

	bf := NewByteFrame()
	bf.WriteNullTerminatedBytes(have0)
	bf.WriteNullTerminatedBytes(have1)
	bf.WriteNullTerminatedBytes(have2)

	got := bf.Data()
	if !bytes.Equal(got, want) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got), hex.Dump(want))
	}
}

func TestReadUint8(t *testing.T) {
	have := []byte{0xFF}
	want := uint8(0xFF)
	bf := NewByteFrameFromBytes(have)

	got := bf.ReadUint8()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadBool(t *testing.T) {
	have := []byte{0x1}
	want := true
	bf := NewByteFrameFromBytes(have)

	got := bf.ReadBool()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint16LE(t *testing.T) {
	have := []byte{0xBB, 0xAA}
	want := uint16(0xAABB)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadUint16()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint16BE(t *testing.T) {
	have := []byte{0xAA, 0xBB}
	want := uint16(0xAABB)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadUint16()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint32LE(t *testing.T) {
	have := []byte{0xDD, 0xCC, 0xBB, 0xAA}
	want := uint32(0xAABBCCDD)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadUint32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint32BE(t *testing.T) {
	have := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	want := uint32(0xAABBCCDD)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadUint32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint64LE(t *testing.T) {
	have := []byte{0x33, 0x22, 0x11, 0x00, 0xDD, 0xCC, 0xBB, 0xAA}
	want := uint64(0xAABBCCDD00112233)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadUint64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadUint64BE(t *testing.T) {
	have := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0x00, 0x11, 0x22, 0x33}
	want := uint64(0xAABBCCDD00112233)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadUint64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt8(t *testing.T) {
	have := []byte{0xFF}
	want := int8(-1)
	bf := NewByteFrameFromBytes(have)

	got := bf.ReadInt8()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt16LE(t *testing.T) {
	have := []byte{0xFE, 0xFF}
	want := int16(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadInt16()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt16BE(t *testing.T) {
	have := []byte{0xFF, 0xFE}
	want := int16(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadInt16()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt32LE(t *testing.T) {
	have := []byte{0xFE, 0xFF, 0xFF, 0xFF}
	want := int32(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadInt32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt32BE(t *testing.T) {
	have := []byte{0xFF, 0xFF, 0xFF, 0xFE}
	want := int32(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadInt32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt64LE(t *testing.T) {
	have := []byte{0xFE, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	want := int64(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadInt64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadInt64BE(t *testing.T) {
	have := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}
	want := int64(-2)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadInt64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadFloat32LE(t *testing.T) {
	have := []byte{0x01, 0x00, 0x70, 0x42}
	want := float32(60.0000038)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadFloat32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadFloat32BE(t *testing.T) {
	have := []byte{0x42, 0x70, 0x00, 0x01}
	want := float32(60.0000038)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadFloat32()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadFloat64LE(t *testing.T) {
	have := []byte{0x18, 0x70, 0xE0, 0x1F, 0x00, 0x00, 0x4E, 0x40}
	want := float64(60.0000038)
	bf := NewByteFrameFromBytes(have)
	bf.SetLE()

	got := bf.ReadFloat64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadFloat64BE(t *testing.T) {
	have := []byte{0x40, 0x4E, 0x00, 0x00, 0x1F, 0xE0, 0x70, 0x18}
	want := float64(60.0000038)
	bf := NewByteFrameFromBytes(have)
	bf.SetBE()

	got := bf.ReadFloat64()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadBytes(t *testing.T) {
	have := []byte{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, 0xCC, 0xCC, 0xCC, 0xCC}
	want0 := []byte{0xAA, 0xAA, 0xAA, 0xAA}
	want1 := []byte{0xBB, 0xBB, 0xBB, 0xBB, 0xCC, 0xCC, 0xCC, 0xCC}
	bf := NewByteFrameFromBytes(have)

	got0 := bf.ReadBytes(4)
	got1 := bf.ReadBytes(8)

	if !bytes.Equal(got0, want0) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got0), hex.Dump(want0))
	}

	if !bytes.Equal(got1, want1) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got1), hex.Dump(want1))
	}
}

func TestReadWriteSeek(t *testing.T) {
	bf := NewByteFrameFromBytes([]byte{0xAA, 0xAA, 0xAA, 0xAA, 0xBB, 0xBB, 0xBB, 0xBB, 0xCC, 0xCC, 0xCC, 0xCC})

	a := bf.ReadUint32()
	b := bf.ReadUint32()
	c := bf.ReadUint32()
	if a != 0xAAAAAAAA || b != 0xBBBBBBBB || c != 0xCCCCCCCC {
		t.Error("error on initial read")
	}

	bf.Seek(4, io.SeekStart)
	b = bf.ReadUint32()
	c = bf.ReadUint32()
	if b != 0xBBBBBBBB || c != 0xCCCCCCCC {
		t.Error("error after seek start")
	}

	bf.Seek(0, io.SeekStart)
	bf.WriteUint32(0xDDDDDDDD)

	bf.Seek(-4, io.SeekCurrent)
	a = bf.ReadUint32()
	b = bf.ReadUint32()
	c = bf.ReadUint32()
	if a != 0xDDDDDDDD || b != 0xBBBBBBBB || c != 0xCCCCCCCC {
		t.Error("error after seek start->write->seek current")
	}

	bf.Seek(-4, io.SeekEnd)
	c = bf.ReadUint32()
	if c != 0xCCCCCCCC {
		t.Error("error after seek end")
	}
}

func TestReadNullTerminatedBytes(t *testing.T) {
	have := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21, 0x00, 0x00}
	want0 := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}
	want1 := []byte{0x57, 0x6F, 0x72, 0x6C, 0x64, 0x21}
	want2 := []byte{}
	bf := NewByteFrameFromBytes(have)

	got0 := bf.ReadNullTerminatedBytes()
	got1 := bf.ReadNullTerminatedBytes()
	got2 := bf.ReadNullTerminatedBytes()

	if !bytes.Equal(got0, want0) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got0), hex.Dump(want0))
	}
	if !bytes.Equal(got1, want1) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got1), hex.Dump(want1))
	}
	if !bytes.Equal(got2, want2) {
		t.Errorf("got\n\t%s\nwant\n\t%s", hex.Dump(got2), hex.Dump(want2))
	}

}
