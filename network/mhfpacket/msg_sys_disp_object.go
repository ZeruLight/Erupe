package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysDispObject represents the MSG_SYS_DISP_OBJECT
type MsgSysDispObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysDispObject) Opcode() network.PacketID {
	return network.MSG_SYS_DISP_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysDispObject) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysDispObject) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
