package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateWarehouse represents the MSG_MHF_UPDATE_WAREHOUSE
type MsgMhfUpdateWarehouse struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateWarehouse) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_WAREHOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateWarehouse) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateWarehouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
