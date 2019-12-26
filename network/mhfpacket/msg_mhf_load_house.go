package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadHouse represents the MSG_MHF_LOAD_HOUSE
type MsgMhfLoadHouse struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHouse) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHouse) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}