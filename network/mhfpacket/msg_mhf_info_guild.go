package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoGuild represents the MSG_MHF_INFO_GUILD
type MsgMhfInfoGuild struct {
	AckHandle uint32
	Unk0      uint32 // Probably a guild ID, but unverified.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoGuild) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfInfoGuild) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
