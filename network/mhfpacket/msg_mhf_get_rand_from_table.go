package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetRandFromTable represents the MSG_MHF_GET_RAND_FROM_TABLE
type MsgMhfGetRandFromTable struct {
	AckHandle uint32
	Results   uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetRandFromTable) Opcode() network.PacketID {
	return network.MSG_MHF_GET_RAND_FROM_TABLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetRandFromTable) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Results = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetRandFromTable) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
