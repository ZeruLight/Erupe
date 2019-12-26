package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavePlateBox represents the MSG_MHF_SAVE_PLATE_BOX
type MsgMhfSavePlateBox struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavePlateBox) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_PLATE_BOX
}

// Parse parses the packet from binary
func (m *MsgMhfSavePlateBox) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavePlateBox) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}