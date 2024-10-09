package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAddGuildWeeklyBonusExceptionalUser represents the MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER
type MsgMhfAddGuildWeeklyBonusExceptionalUser struct {
	AckHandle uint32
	NumUsers  uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_GUILD_WEEKLY_BONUS_EXCEPTIONAL_USER
}

// Parse parses the packet from binary
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.NumUsers = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddGuildWeeklyBonusExceptionalUser) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
