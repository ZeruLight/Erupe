package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddUdTacticsPoint represents the MSG_MHF_ADD_UD_TACTICS_POINT
type MsgMhfAddUdTacticsPoint struct {
	AckHandle   uint32
	QuestFileID uint16
	Points      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdTacticsPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_TACTICS_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdTacticsPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.QuestFileID = bf.ReadUint16()
	m.Points = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdTacticsPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
