package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserve7E represents the MSG_SYS_reserve7E
type MsgSysReserve7E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve7E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve7E
}

// Parse parses the packet from binary
func (m *MsgSysReserve7E) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve7E) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
