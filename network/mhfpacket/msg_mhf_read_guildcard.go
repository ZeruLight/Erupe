package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadGuildcard represents the MSG_MHF_READ_GUILDCARD
type MsgMhfReadGuildcard struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadGuildcard) Opcode() network.PacketID {
	return network.MSG_MHF_READ_GUILDCARD
}

// Parse parses the packet from binary
func (m *MsgMhfReadGuildcard) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadGuildcard) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}