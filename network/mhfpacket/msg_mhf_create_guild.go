package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCreateGuild represents the MSG_MHF_CREATE_GUILD
type MsgMhfCreateGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCreateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_CREATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfCreateGuild) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCreateGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
