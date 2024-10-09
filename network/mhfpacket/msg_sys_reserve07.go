package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserve07 represents the MSG_SYS_reserve07
type MsgSysReserve07 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve07) Opcode() network.PacketID {
	return network.MSG_SYS_reserve07
}

// Parse parses the packet from binary
func (m *MsgSysReserve07) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve07) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
