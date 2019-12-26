package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildWeeklyBonusActiveCount represents the MSG_MHF_GET_GUILD_WEEKLY_BONUS_ACTIVE_COUNT
type MsgMhfGetGuildWeeklyBonusActiveCount struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildWeeklyBonusActiveCount) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_ACTIVE_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildWeeklyBonusActiveCount) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildWeeklyBonusActiveCount) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}