package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve19E represents the MSG_SYS_reserve19E
type MsgSysReserve19E struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve19E) Opcode() network.PacketID {
	return network.MSG_SYS_reserve19E
}

// Parse parses the packet from binary
func (m *MsgSysReserve19E) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve19E) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
