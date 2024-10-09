package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysAddObject represents the MSG_SYS_ADD_OBJECT
type MsgSysAddObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAddObject) Opcode() network.PacketID {
	return network.MSG_SYS_ADD_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysAddObject) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAddObject) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
