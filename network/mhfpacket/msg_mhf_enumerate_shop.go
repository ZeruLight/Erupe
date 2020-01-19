package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateShop represents the MSG_MHF_ENUMERATE_SHOP
type MsgMhfEnumerateShop struct {
	AckHandle uint32
	Unk0      uint8 // Shop ID maybe? I seen 0 -> 10.
	Unk1      uint32
	Unk2      uint16
	Unk3      uint8
	Unk4      uint8
	Unk5      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateShop) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_SHOP
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateShop) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint8()
	m.Unk4 = bf.ReadUint8()
	m.Unk5 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateShop) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
