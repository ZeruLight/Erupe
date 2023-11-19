package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfDisplayedAchievement represents the MSG_MHF_DISPLAYED_ACHIEVEMENT
type MsgMhfDisplayedAchievement struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfDisplayedAchievement) Opcode() network.PacketID {
	return network.MSG_MHF_DISPLAYED_ACHIEVEMENT
}

// Parse parses the packet from binary
func (m *MsgMhfDisplayedAchievement) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.ReadUint8() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfDisplayedAchievement) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
