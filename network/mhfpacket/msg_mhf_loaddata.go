package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoaddata represents the MSG_MHF_LOADDATA
type MsgMhfLoaddata struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoaddata) Opcode() network.PacketID {
	return network.MSG_MHF_LOADDATA
}

// Parse parses the packet from binary
func (m *MsgMhfLoaddata) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoaddata) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}