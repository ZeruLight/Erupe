package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfOperateGuildMember represents the MSG_MHF_OPERATE_GUILD_MEMBER
type MsgMhfOperateGuildMember struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuildMember) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuildMember) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuildMember) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
