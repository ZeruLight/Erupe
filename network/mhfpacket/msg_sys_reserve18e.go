package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserve18E represents the MSG_SYS_reserve18E
type MsgSysReserve18E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve18E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve18E
}

// Parse parses the packet from binary
func (m *MsgSysReserve18E) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve18E) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
