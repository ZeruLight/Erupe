package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysEnumerateStage represents the MSG_SYS_ENUMERATE_STAGE
type MsgSysEnumerateStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnumerateStage) Opcode() network.PacketID {
	return network.MSG_SYS_ENUMERATE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysEnumerateStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnumerateStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}