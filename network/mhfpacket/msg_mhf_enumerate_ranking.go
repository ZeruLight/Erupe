package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateRanking represents the MSG_MHF_ENUMERATE_RANKING
type MsgMhfEnumerateRanking struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateRanking) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateRanking) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	bf.ReadUint16() // Zeroed
	bf.ReadUint8()  // Zeroed
	bf.ReadUint8()  // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateRanking) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
