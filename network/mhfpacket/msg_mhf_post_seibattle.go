package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfPostSeibattle represents the MSG_MHF_POST_SEIBATTLE
type MsgMhfPostSeibattle struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint8
	Unk2      uint32
	Unk3      uint8
	Unk4      uint16
	Unk5      uint16
	Unk6      uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostSeibattle) Opcode() network.PacketID {
	return network.MSG_MHF_POST_SEIBATTLE
}

// Parse parses the packet from binary
func (m *MsgMhfPostSeibattle) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint8()
	m.Unk4 = bf.ReadUint16()
	m.Unk5 = bf.ReadUint16()
	m.Unk6 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostSeibattle) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
