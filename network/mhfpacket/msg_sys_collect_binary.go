package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCollectBinary represents the MSG_SYS_COLLECT_BINARY
type MsgSysCollectBinary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCollectBinary) Opcode() network.PacketID {
	return network.MSG_SYS_COLLECT_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysCollectBinary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysCollectBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
