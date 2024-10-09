package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireFestaPersonalPrize represents the MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
type MsgMhfAcquireFestaPersonalPrize struct {
	AckHandle uint32
	PrizeID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFestaPersonalPrize) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFestaPersonalPrize) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.PrizeID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFestaPersonalPrize) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
