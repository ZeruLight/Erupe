package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysBackStage represents the MSG_SYS_BACK_STAGE
type MsgSysBackStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysBackStage) Opcode() network.PacketID {
	return network.MSG_SYS_BACK_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysBackStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysBackStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}