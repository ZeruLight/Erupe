package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve208 represents the MSG_SYS_reserve208
type MsgSysReserve208 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve208) Opcode() network.PacketID {
	return network.MSG_SYS_reserve208
}

// Parse parses the packet from binary
func (m *MsgSysReserve208) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve208) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
