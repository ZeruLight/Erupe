package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysWaitStageBinary represents the MSG_SYS_WAIT_STAGE_BINARY
type MsgSysWaitStageBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysWaitStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_WAIT_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysWaitStageBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysWaitStageBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}