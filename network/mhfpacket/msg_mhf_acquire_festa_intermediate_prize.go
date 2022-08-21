package mhfpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAcquireFestaIntermediatePrize represents the MSG_MHF_ACQUIRE_FESTA_INTERMEDIATE_PRIZE
type MsgMhfAcquireFestaIntermediatePrize struct {
	AckHandle uint32
	PrizeID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFestaIntermediatePrize) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA_INTERMEDIATE_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFestaIntermediatePrize) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.PrizeID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFestaIntermediatePrize) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint32(m.PrizeID)
	return nil
}
