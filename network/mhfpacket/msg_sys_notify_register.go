package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysNotifyRegister represents the MSG_SYS_NOTIFY_REGISTER
type MsgSysNotifyRegister struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysNotifyRegister) Opcode() network.PacketID {
	return network.MSG_SYS_NOTIFY_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysNotifyRegister) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysNotifyRegister) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
