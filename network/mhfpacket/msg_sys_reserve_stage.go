package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserveStage represents the MSG_SYS_RESERVE_STAGE
type MsgSysReserveStage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserveStage) Opcode() network.PacketID {
	return network.MSG_SYS_RESERVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysReserveStage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserveStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}