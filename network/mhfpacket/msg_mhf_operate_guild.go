package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfOperateGuild represents the MSG_MHF_OPERATE_GUILD
type MsgMhfOperateGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuild) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuild) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuild) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
