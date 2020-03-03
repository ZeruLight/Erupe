package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfArrangeGuildMember represents the MSG_MHF_ARRANGE_GUILD_MEMBER
type MsgMhfArrangeGuildMember struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfArrangeGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ARRANGE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfArrangeGuildMember) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfArrangeGuildMember) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
