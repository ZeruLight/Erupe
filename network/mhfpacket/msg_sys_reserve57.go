package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve57 represents the MSG_SYS_reserve57
type MsgSysReserve57 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve57) Opcode() network.PacketID {
	return network.MSG_SYS_reserve57
}

// Parse parses the packet from binary
func (m *MsgSysReserve57) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve57) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
