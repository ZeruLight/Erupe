package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateStage represents the MSG_SYS_CREATE_STAGE
type MsgSysCreateStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateStage) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysCreateStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}