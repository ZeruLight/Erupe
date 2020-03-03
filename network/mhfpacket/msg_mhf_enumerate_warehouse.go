package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateWarehouse represents the MSG_MHF_ENUMERATE_WAREHOUSE
type MsgMhfEnumerateWarehouse struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateWarehouse) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_WAREHOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateWarehouse) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateWarehouse) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
