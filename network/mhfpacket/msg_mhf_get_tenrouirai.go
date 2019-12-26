package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetTenrouirai represents the MSG_MHF_GET_TENROUIRAI
type MsgMhfGetTenrouirai struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTenrouirai) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TENROUIRAI
}

// Parse parses the packet from binary
func (m *MsgMhfGetTenrouirai) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTenrouirai) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}