package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireFestaPersonalPrize represents the MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
type MsgMhfAcquireFestaPersonalPrize struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFestaPersonalPrize) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA_PERSONAL_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFestaPersonalPrize) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFestaPersonalPrize) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
