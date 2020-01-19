package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCastBinary represents the MSG_SYS_CAST_BINARY
type MsgSysCastBinary struct {
	Unk0           uint16
	Unk1           uint16
	Type0          uint8
	Type1          uint8
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCastBinary) Opcode() network.PacketID {
	return network.MSG_SYS_CAST_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysCastBinary) Parse(bf *byteframe.ByteFrame) error {
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Type0 = bf.ReadUint8()
	m.Type1 = bf.ReadUint8()
	dataSize := bf.ReadUint16()
	m.RawDataPayload = bf.ReadBytes(uint(dataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCastBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
