package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoGuild represents the MSG_MHF_INFO_GUILD
type MsgMhfInfoGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoGuild) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfInfoGuild) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}