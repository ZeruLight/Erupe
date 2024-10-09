package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserve192 represents the MSG_SYS_reserve192
type MsgSysReserve192 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve192) Opcode() network.PacketID {
	return network.MSG_SYS_reserve192
}

// Parse parses the packet from binary
func (m *MsgSysReserve192) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve192) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
