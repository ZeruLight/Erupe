package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysSetStageBinary represents the MSG_SYS_SET_STAGE_BINARY
type MsgSysSetStageBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_SET_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysSetStageBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetStageBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}