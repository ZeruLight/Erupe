package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfCaravanMyScore represents the MSG_MHF_CARAVAN_MY_SCORE
type MsgMhfCaravanMyScore struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Unk2      int32
	Unk3      int32
	Unk4      uint32
	Unk5      int32
	Unk6      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCaravanMyScore) Opcode() network.PacketID {
	return network.MSG_MHF_CARAVAN_MY_SCORE
}

// Parse parses the packet from binary
func (m *MsgMhfCaravanMyScore) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadInt32()
	m.Unk3 = bf.ReadInt32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadInt32()
	m.Unk6 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCaravanMyScore) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
