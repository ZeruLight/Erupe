package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateHouse represents the MSG_MHF_UPDATE_HOUSE
type MsgMhfUpdateHouse struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateHouse) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateHouse) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateHouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
