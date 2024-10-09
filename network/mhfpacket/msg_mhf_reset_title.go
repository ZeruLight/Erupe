package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfResetTitle represents the MSG_MHF_RESET_TITLE
type MsgMhfResetTitle struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfResetTitle) Opcode() network.PacketID {
	return network.MSG_MHF_RESET_TITLE
}

// Parse parses the packet from binary
func (m *MsgMhfResetTitle) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfResetTitle) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
