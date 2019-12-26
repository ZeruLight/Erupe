package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetKijuInfo represents the MSG_MHF_GET_KIJU_INFO
type MsgMhfGetKijuInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetKijuInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_KIJU_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetKijuInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetKijuInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}