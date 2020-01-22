package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCheckWeeklyStamp represents the MSG_MHF_CHECK_WEEKLY_STAMP
type MsgMhfCheckWeeklyStamp struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      bool
	Unk2      uint16 // Hardcoded 0 in the binary
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckWeeklyStamp) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_WEEKLY_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfCheckWeeklyStamp) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadBool()
	m.Unk2 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckWeeklyStamp) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
