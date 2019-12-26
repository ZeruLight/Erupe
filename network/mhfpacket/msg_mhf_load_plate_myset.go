package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadPlateMyset represents the MSG_MHF_LOAD_PLATE_MYSET
type MsgMhfLoadPlateMyset struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPlateMyset) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PLATE_MYSET
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPlateMyset) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPlateMyset) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}