package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildWeeklyBonusMaster represents the MSG_MHF_GET_GUILD_WEEKLY_BONUS_MASTER
type MsgMhfGetGuildWeeklyBonusMaster struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildWeeklyBonusMaster) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_WEEKLY_BONUS_MASTER
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildWeeklyBonusMaster) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildWeeklyBonusMaster) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
