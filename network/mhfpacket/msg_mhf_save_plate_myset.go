package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavePlateMyset represents the MSG_MHF_SAVE_PLATE_MYSET
type MsgMhfSavePlateMyset struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavePlateMyset) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_PLATE_MYSET
}

// Parse parses the packet from binary
func (m *MsgMhfSavePlateMyset) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavePlateMyset) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}