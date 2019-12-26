package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLoadRegister represents the MSG_SYS_LOAD_REGISTER
type MsgSysLoadRegister struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLoadRegister) Opcode() network.PacketID {
	return network.MSG_SYS_LOAD_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysLoadRegister) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysLoadRegister) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}