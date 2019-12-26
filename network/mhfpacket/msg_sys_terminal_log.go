package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysTerminalLog represents the MSG_SYS_TERMINAL_LOG
type MsgSysTerminalLog struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTerminalLog) Opcode() network.PacketID {
	return network.MSG_SYS_TERMINAL_LOG
}

// Parse parses the packet from binary
func (m *MsgSysTerminalLog) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysTerminalLog) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}