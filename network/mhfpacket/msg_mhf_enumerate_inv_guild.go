package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateInvGuild represents the MSG_MHF_ENUMERATE_INV_GUILD
type MsgMhfEnumerateInvGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateInvGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_INV_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateInvGuild) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateInvGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
