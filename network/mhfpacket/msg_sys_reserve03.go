package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserve03 represents the MSG_SYS_reserve03
type MsgSysReserve03 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve03) Opcode() network.PacketID {
	return network.MSG_SYS_reserve03
}

// Parse parses the packet from binary
func (m *MsgSysReserve03) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve03) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
