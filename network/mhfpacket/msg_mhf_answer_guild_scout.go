package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAnswerGuildScout represents the MSG_MHF_ANSWER_GUILD_SCOUT
type MsgMhfAnswerGuildScout struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAnswerGuildScout) Opcode() network.PacketID {
	return network.MSG_MHF_ANSWER_GUILD_SCOUT
}

// Parse parses the packet from binary
func (m *MsgMhfAnswerGuildScout) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAnswerGuildScout) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
