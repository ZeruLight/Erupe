package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve7E represents the MSG_SYS_reserve7E
type MsgSysReserve7E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve7E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve7E
}

// Parse parses the packet from binary
func (m *MsgSysReserve7E) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve7E) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
