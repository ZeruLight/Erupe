package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetRejectGuildScout represents the MSG_MHF_GET_REJECT_GUILD_SCOUT
type MsgMhfGetRejectGuildScout struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRejectGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_GET_REJECT_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfGetRejectGuildScout) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRejectGuildScout) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}