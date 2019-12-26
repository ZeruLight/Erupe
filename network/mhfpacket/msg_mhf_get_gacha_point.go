package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGachaPoint represents the MSG_MHF_GET_GACHA_POINT
type MsgMhfGetGachaPoint struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGachaPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GACHA_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetGachaPoint) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGachaPoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
