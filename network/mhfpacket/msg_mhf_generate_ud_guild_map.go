package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGenerateUdGuildMap represents the MSG_MHF_GENERATE_UD_GUILD_MAP
type MsgMhfGenerateUdGuildMap struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGenerateUdGuildMap) Opcode() network.PacketID {
	return network.MSG_MHF_GENERATE_UD_GUILD_MAP
}

// Parse parses the packet from binary
func (m *MsgMhfGenerateUdGuildMap) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGenerateUdGuildMap) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
