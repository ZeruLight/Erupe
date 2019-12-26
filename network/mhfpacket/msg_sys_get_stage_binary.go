package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysGetStageBinary represents the MSG_SYS_GET_STAGE_BINARY
type MsgSysGetStageBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_GET_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysGetStageBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetStageBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}