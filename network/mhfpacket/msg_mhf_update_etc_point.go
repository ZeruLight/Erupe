package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateEtcPoint represents the MSG_MHF_UPDATE_ETC_POINT
type MsgMhfUpdateEtcPoint struct {
	AckHandle uint32
	PointType uint8
	Delta     int16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateEtcPoint) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_ETC_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateEtcPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.PointType = bf.ReadUint8()
	m.Delta = bf.ReadInt16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateEtcPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
