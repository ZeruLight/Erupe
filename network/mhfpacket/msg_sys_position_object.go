package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysPositionObject represents the MSG_SYS_POSITION_OBJECT
type MsgSysPositionObject struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysPositionObject) Opcode() network.PacketID {
	return network.MSG_SYS_POSITION_OBJECT
}

// Parse parses the packet from binary
func (m *MsgSysPositionObject) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysPositionObject) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}