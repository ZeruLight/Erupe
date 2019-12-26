package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysSetStagePass represents the MSG_SYS_SET_STAGE_PASS
type MsgSysSetStagePass struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetStagePass) Opcode() network.PacketID {
	return network.MSG_SYS_SET_STAGE_PASS
}

// Parse parses the packet from binary
func (m *MsgSysSetStagePass) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetStagePass) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}