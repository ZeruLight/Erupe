package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateGuildMember represents the MSG_MHF_ENUMERATE_GUILD_MEMBER
type MsgMhfEnumerateGuildMember struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuildMember) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuildMember) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}