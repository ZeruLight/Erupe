package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadHouse represents the MSG_MHF_LOAD_HOUSE
type MsgMhfLoadHouse struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint32
	Unk2      uint8
	Unk3      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHouse) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHouse) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint8()
	m.Unk3 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
