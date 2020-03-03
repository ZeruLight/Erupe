package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateGuild represents the MSG_MHF_ENUMERATE_GUILD
type MsgMhfEnumerateGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuild) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
