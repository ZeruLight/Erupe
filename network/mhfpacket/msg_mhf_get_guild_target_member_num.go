package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildTargetMemberNum represents the MSG_MHF_GET_GUILD_TARGET_MEMBER_NUM
type MsgMhfGetGuildTargetMemberNum struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildTargetMemberNum) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_TARGET_MEMBER_NUM
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildTargetMemberNum) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildTargetMemberNum) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
