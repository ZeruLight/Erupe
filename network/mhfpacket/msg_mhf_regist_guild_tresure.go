package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfRegistGuildTresure represents the MSG_MHF_REGIST_GUILD_TRESURE
type MsgMhfRegistGuildTresure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildTresure) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildTresure) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
