package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysRecordLog represents the MSG_SYS_RECORD_LOG
type MsgSysRecordLog struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRecordLog) Opcode() network.PacketID {
	return network.MSG_SYS_RECORD_LOG
}

// Parse parses the packet from binary
func (m *MsgSysRecordLog) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysRecordLog) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}