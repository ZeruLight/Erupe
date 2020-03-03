package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve56 represents the MSG_SYS_reserve56
type MsgSysReserve56 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve56) Opcode() network.PacketID {
	return network.MSG_SYS_reserve56
}

// Parse parses the packet from binary
func (m *MsgSysReserve56) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve56) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
