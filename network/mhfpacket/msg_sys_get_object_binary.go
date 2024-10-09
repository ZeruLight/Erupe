package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysGetObjectBinary represents the MSG_SYS_GET_OBJECT_BINARY
type MsgSysGetObjectBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetObjectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_OBJECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetObjectBinary) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetObjectBinary) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
