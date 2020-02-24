package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateFestaPersonalPrize represents the MSG_MHF_ENUMERATE_FESTA_PERSONAL_PRIZE
type MsgMhfEnumerateFestaPersonalPrize struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateFestaPersonalPrize) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_FESTA_PERSONAL_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateFestaPersonalPrize) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateFestaPersonalPrize) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
