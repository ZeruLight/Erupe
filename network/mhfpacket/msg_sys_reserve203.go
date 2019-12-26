package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve203 represents the MSG_SYS_reserve203
type MsgSysReserve203 struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve203) Opcode() network.PacketID {
	return network.MSG_SYS_reserve203
}

// Parse parses the packet from binary
func (m *MsgSysReserve203) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve203) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}