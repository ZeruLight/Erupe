package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysGetFile represents the MSG_SYS_GET_FILE
type MsgSysGetFile struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetFile) Opcode() network.PacketID {
	return network.MSG_SYS_GET_FILE
}

// Parse parses the packet from binary
func (m *MsgSysGetFile) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetFile) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}