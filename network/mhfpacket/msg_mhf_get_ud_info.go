package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdInfo represents the MSG_MHF_GET_UD_INFO
type MsgMhfGetUdInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}